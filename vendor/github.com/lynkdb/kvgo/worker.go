// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kvgo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func (cn *Conn) workerLocal() {
	cn.workmu.Lock()
	if cn.workerLocalRunning {
		cn.workmu.Unlock()
		return
	}
	cn.workerLocalRunning = true
	cn.workmu.Unlock()

	go cn.workerLocalReplicaOfRefresh()

	for !cn.close {

		if err := cn.workerLocalExpiredRefresh(); err != nil {
			hlog.Printf("warn", "local ttl clean err %s", err.Error())
		}

		if err := cn.workerLocalTableRefresh(); err != nil {
			hlog.Printf("warn", "local table refresh err %s", err.Error())
		}

		time.Sleep(workerLocalExpireSleep)
	}
}

func (cn *Conn) workerLocalReplicaOfRefresh() {

	hlog.Printf("info", "replica-of servers %d", len(cn.opts.Cluster.ReplicaOfNodes))

	for !cn.close {
		time.Sleep(workerReplicaLogAsyncSleep)

		if err := cn.workerLocalReplicaOfLogAsync(); err != nil {
			hlog.Printf("warn", "replica-of log async err %s", err.Error())
		}
	}
}

func (cn *Conn) workerLocalExpiredRefresh() error {

	for _, t := range cn.tables {
		if err := cn.workerLocalExpiredRefreshTable(t); err != nil {
			hlog.Printf("warn", "cluster ttl refresh error %s", err.Error())
		}
	}

	return nil
}

func (cn *Conn) workerLocalExpiredRefreshTable(dt *dbTable) error {

	iter := dt.db.NewIterator(&util.Range{
		Start: keyEncode(nsKeyTtl, uint64ToBytes(0)),
		Limit: keyEncode(nsKeyTtl, uint64ToBytes(uint64(time.Now().UnixNano()/1e6))),
	}, nil)
	defer iter.Release()

	for !cn.close {

		var (
			num   = 0
			batch = new(leveldb.Batch)
		)

		for iter.Next() {

			meta, err := kv2.ObjectMetaDecode(bytesClone(iter.Value()))
			if err != nil {
				return err
			}

			data, err := dt.db.Get(keyEncode(nsKeyMeta, meta.Key), nil)
			if err == nil {

				cmeta, err := kv2.ObjectMetaDecode(data)
				if err == nil && cmeta.Version == meta.Version {
					batch.Delete(keyEncode(nsKeyMeta, meta.Key))
					batch.Delete(keyEncode(nsKeyData, meta.Key))
					batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
				}

			} else if err.Error() != ldbNotFound {
				return err
			}

			batch.Delete(keyExpireEncode(nsKeyTtl, meta.Expired, meta.Key))
			num += 1

			if num >= workerLocalExpireLimit {
				break
			}

			if cn.close {
				break
			}
		}

		if num > 0 {
			dt.db.Write(batch, nil)
		}

		if num < workerLocalExpireLimit {
			break
		}
	}

	return nil
}

func (cn *Conn) workerLocalTableRefresh() error {

	tn := time.Now().Unix()

	if (cn.workerTableRefreshed + workerTableRefreshTime) > tn {
		return nil
	}

	rgS := []util.Range{
		{
			Start: []byte{},
			Limit: []byte{0xff},
		},
	}
	rgK := &util.Range{
		Start: keyEncode(nsKeyMeta, []byte{}),
		Limit: keyEncode(nsKeyMeta, []byte{0xff}),
	}
	rgIncr := &util.Range{
		Start: keySysIncrCutset(""),
		Limit: append(keySysIncrCutset(""), []byte{0xff}...),
	}
	rgAsync := &util.Range{
		Start: keySysLogAsync("", ""),
		Limit: append(keySysLogAsync("", ""), []byte{0xff}...),
	}

	if cn.opts.Feature.WriteMetaDisable {
		rgK.Start = keyEncode(nsKeyData, []byte{})
		rgK.Limit = keyEncode(nsKeyData, []byte{0xff})
	}

	for _, t := range cn.tables {

		if err := cn.workerLocalLogCleanTable(t); err != nil {
			hlog.Printf("warn", "worker log clean table %s, err %s",
				t.tableName, err.Error())
		}

		// db size
		s, err := t.db.SizeOf(rgS)
		if err != nil {
			hlog.Printf("warn", "get db size error %s", err.Error())
			continue
		}
		if len(s) < 1 {
			continue
		}

		// db keys
		kn := uint64(0)
		iter := t.db.NewIterator(rgK, nil)
		for ; iter.Next(); kn++ {
		}
		iter.Release()

		tableStatus := kv2.TableStatus{
			Name:    t.tableName,
			KeyNum:  kn,
			DbSize:  uint64(s[0]),
			Options: map[string]int64{},
		}

		// log-id
		if bs, err := t.db.Get(keySysLogCutset, nil); err == nil {
			if logid, err := strconv.ParseInt(string(bs), 10, 64); err == nil {
				tableStatus.Options["log_id"] = logid
			}
		}

		// incr
		iterIncr := t.db.NewIterator(rgIncr, nil)
		for iterIncr.Next() {

			if bytes.Compare(iterIncr.Key(), rgIncr.Start) <= 0 {
				continue
			}

			if bytes.Compare(iterIncr.Key(), rgIncr.Limit) > 0 {
				break
			}

			incrid, err := strconv.ParseInt(string(iterIncr.Value()), 10, 64)
			if err != nil {
				continue
			}

			key := bytes.TrimPrefix(iterIncr.Key(), rgIncr.Start)
			if len(key) > 0 {
				tableStatus.Options[fmt.Sprintf("incr_id_%s", string(key))] = incrid
			}
		}
		iterIncr.Release()

		// async
		iterAsync := t.db.NewIterator(rgAsync, nil)
		for iterAsync.Next() {

			if bytes.Compare(iterAsync.Key(), rgAsync.Start) <= 0 {
				continue
			}

			if bytes.Compare(iterAsync.Key(), rgAsync.Limit) > 0 {
				break
			}

			logid, err := strconv.ParseInt(string(iterAsync.Value()), 10, 64)
			if err != nil {
				continue
			}

			key := bytes.TrimPrefix(iterAsync.Key(), rgAsync.Start)
			if len(key) > 0 {
				tableStatus.Options[fmt.Sprintf("async_%s", string(key))] = logid
			}
		}
		iterAsync.Release()

		//
		rr := kv2.NewObjectWriter(nsSysTableStatus(t.tableName), tableStatus).
			TableNameSet(sysTableName)
		rs := cn.commitLocal(rr, 0)
		if !rs.OK() {
			hlog.Printf("warn", "refresh table (%s) status error %s", t.tableName, err.Error())
		}

		if cn.close {
			break
		}
	}

	cn.workerTableRefreshed = tn

	return nil
}

func (cn *Conn) workerLocalReplicaOfLogAsync() error {

	ups := map[string]bool{}

	for _, hp := range cn.opts.Cluster.MainNodes {

		if cn.close {
			break
		}

		if hp.Addr == cn.opts.Server.Bind {
			continue
		}

		if _, ok := ups[hp.Addr]; ok {
			continue
		}

		for _, dt := range cn.tables {

			go func(hp *ClientConfig, dt *dbTable) {
				if err := cn.workerLocalReplicaOfLogAsyncTable(hp, &ConfigReplicaTableMap{
					From: dt.tableName,
					To:   dt.tableName,
				}); err != nil {
					hlog.Printf("warn", "worker replica-of log-async table %s -> %s, err %s",
						dt.tableName, dt.tableName, err.Error())
				}
			}(hp, dt)

			if cn.close {
				break
			}
		}

		ups[hp.Addr] = true
	}

	for _, hp := range cn.opts.Cluster.ReplicaOfNodes {

		if cn.close {
			break
		}

		if hp.Addr == cn.opts.Server.Bind || len(hp.TableMaps) == 0 {
			continue
		}

		if _, ok := ups[hp.Addr]; ok {
			continue
		}

		for _, tm := range hp.TableMaps {

			go func(hp *ClientConfig, tm *ConfigReplicaTableMap) {
				if err := cn.workerLocalReplicaOfLogAsyncTable(hp, tm); err != nil {
					hlog.Printf("warn", "worker replica-of log-async table %s -> %s, err %s",
						tm.From, tm.To, err.Error())
				}
			}(hp.ClientConfig, tm)

			if cn.close {
				break
			}
		}

		ups[hp.Addr] = true
	}

	return nil
}

func (cn *Conn) workerLocalReplicaOfLogAsyncTable(hp *ClientConfig, tm *ConfigReplicaTableMap) error {

	dt := cn.tabledb(tm.To)
	if dt == nil {
		return errors.New("no table found in local server")
	}

	lkey := hp.Addr + "/" + tm.From

	dt.logAsyncMu.Lock()
	if _, ok := dt.logAsyncSets[lkey]; ok {
		dt.logAsyncMu.Unlock()
		return nil
	}
	dt.logAsyncSets[lkey] = true
	dt.logAsyncMu.Unlock()

	defer func() {
		dt.logAsyncMu.Lock()
		delete(dt.logAsyncSets, lkey)
		dt.logAsyncMu.Unlock()
	}()

	var (
		offset = uint64(0)
		num    = 0
		retry  = 0
	)

	if bs, err := dt.db.Get(keySysLogAsync(hp.Addr, tm.From), nil); err != nil {
		if err.Error() != ldbNotFound {
			return err
		}
	} else {
		if offset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
			return err
		}
	}

	conn, err := clientConn(hp.Addr, hp.AccessKey, hp.AuthTLSCert, false)
	if err != nil {
		return err
	}

	rr := kv2.NewObjectReader().
		TableNameSet(tm.From).
		LogOffsetSet(offset).LimitNumSet(100)
	rr.LimitSize = kv2.ObjectReaderLimitSizeMax
	rr.WaitTime = 10000

	// hlog.Printf("info", "pull from %s/%s at %d", hp.Addr, tm.From, offset)

	for !cn.close {

		ctx, fc := context.WithTimeout(context.Background(), time.Second*1000)
		rs, err := kv2.NewPublicClient(conn).Query(ctx, rr)
		fc()

		if cn.close {
			break
		}

		if err != nil {

			hlog.Printf("warn", "kvgo log async from %s/%s, err %s",
				hp.Addr, dt.tableName, err.Error())

			retry += 1
			if retry >= 3 {
				break
			}

			time.Sleep(1e9)
			conn, err = clientConn(hp.Addr, hp.AccessKey, hp.AuthTLSCert, true)
			continue
		}
		retry = 0

		for _, item := range rs.Items {

			ow := &kv2.ObjectWriter{
				Meta: item.Meta,
				Data: item.Data,
			}

			if kv2.AttrAllow(item.Meta.Attrs, kv2.ObjectMetaAttrDelete) {
				ow.ModeDeleteSet(true)
			}

			ow.TableNameSet(tm.To)

			if rs2 := cn.commitLocal(ow, item.Meta.Version); rs2.OK() {
				rr.LogOffset = item.Meta.Version
				num += 1
			} else {
				hlog.Printf("info", "kvgo log async from %s/%s to local/%s, err %s",
					hp.Addr, tm.From, tm.To, rs2.Message)
				rs.Next = false
				break
			}
		}

		if rr.LogOffset > offset {
			dt.db.Put(keySysLogAsync(hp.Addr, tm.From),
				[]byte(strconv.FormatUint(rr.LogOffset, 10)), nil)
		}

		if !rs.Next {
			break
		}
	}

	if num > 0 {
		hlog.Printf("debug", "kvgo log async from %s/%s to local/%s, num %d, offset %d",
			hp.Addr, tm.From, tm.To, num, rr.LogOffset)
	}

	return nil
}

func (cn *Conn) workerLocalLogCleanTable(tdb *dbTable) error {

	var (
		offset = keyEncode(nsKeyLog, uint64ToBytes(0))
		cutset = keyEncode(nsKeyLog, []byte{0xff})
	)

	var (
		iter = tdb.db.NewIterator(&util.Range{
			Start: offset,
			Limit: cutset,
		}, nil)
		sets  = map[string]uint64{}
		ndel  = 0
		batch = new(leveldb.Batch)
	)

	for ok := iter.Last(); ok; ok = iter.Prev() {

		if bytes.Compare(iter.Key(), cutset) >= 0 {
			continue
		}

		if bytes.Compare(iter.Key(), offset) <= 0 {
			break
		}

		if len(iter.Value()) >= 2 {

			logMeta, err := kv2.ObjectMetaDecode(iter.Value())
			if err == nil && logMeta != nil {

				tdb.objectLogVersionSet(0, logMeta.Version, 0)

				if _, ok := sets[string(logMeta.Key)]; !ok {

					if bs, err := tdb.db.Get(keyEncode(nsKeyMeta, logMeta.Key), nil); err == nil {
						meta, err := kv2.ObjectMetaDecode(bs)
						if err == nil && meta.Version > 0 && meta.Version != logMeta.Version {
							batch.Delete(iter.Key())
							ndel += 1
							continue
						}
					}

					sets[string(logMeta.Key)] = logMeta.Version
					continue
				}
			}
		}

		batch.Delete(iter.Key())
		ndel += 1

		if ndel >= 1000 {
			tdb.db.Write(batch, nil)
			batch = new(leveldb.Batch)
			ndel = 0
			hlog.Printf("info", "table %s, log clean %d/%d", tdb.tableName, ndel, len(sets))
		}
	}

	if ndel > 0 {
		tdb.db.Write(batch, nil)
		hlog.Printf("info", "table %s, log clean %d/%d", tdb.tableName, ndel, len(sets))
	}

	iter.Release()

	return nil
}
