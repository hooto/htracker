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
	"context"
	"strconv"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/lynkdb/iomix/sko"
)

func (cn *Conn) workerLocal() {

	for {
		time.Sleep(workerLocalExpireSleep)

		if err := cn.workerLocalExpiredRefresh(); err != nil {
			hlog.Printf("warn", "local ttl clean err %s", err.Error())
		}
	}
}

func (cn *Conn) workerClusterReplica() {

	for {
		time.Sleep(workerClusterSleep)

		if err := cn.workerClusterReplicaLogAsync(); err != nil {
			hlog.Printf("warn", "cluster log async err %s", err.Error())
		}
	}
}

func (cn *Conn) workerLocalExpiredRefresh() error {

	iter := cn.db.NewIterator(&util.Range{
		Start: keyEncode(nsKeyTtl, uint64ToBytes(0)),
		Limit: keyEncode(nsKeyTtl, uint64ToBytes(uint64(time.Now().UnixNano()/1e6))),
	}, nil)
	defer iter.Release()

	for {

		var (
			num   = 0
			batch = new(leveldb.Batch)
		)

		for iter.Next() {

			meta, err := sko.ObjectMetaDecode(bytesClone(iter.Value()))
			if err != nil {
				return err
			}

			data, err := cn.db.Get(keyEncode(nsKeyMeta, meta.Key), nil)
			if err == nil {

				cmeta, err := sko.ObjectMetaDecode(data)
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
		}

		if num > 0 {
			cn.db.Write(batch, nil)
		}

		if num < workerLocalExpireLimit {
			break
		}
	}

	return nil
}

func (cn *Conn) workerClusterReplicaLogAsync() error {

	if len(cn.opts.Cluster.Masters) < 1 {
		return nil
	}

	ctx, fc := context.WithTimeout(context.Background(), time.Second*10)
	defer fc()

	for _, hp := range cn.opts.Cluster.Masters {

		if hp.Addr == cn.opts.Server.Bind {
			continue
		}

		var (
			offset = uint64(0)
			num    = int64(0)
		)

		if bs, err := cn.db.Get(keySysLogAsync(hp.Addr), nil); err != nil {
			if err.Error() != ldbNotFound {
				return err
			}
		} else {
			if offset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return err
			}
		}

		conn, err := clientConn(hp.Addr, cn.authKey(hp.Addr))
		if err != nil {
			continue
		}

		rr := sko.NewObjectReader(nil).LogOffsetSet(offset).LimitNumSet(100)

		for {

			rs, err := sko.NewObjectClient(conn).Query(ctx, rr)
			if err != nil || !rs.OK() || len(rs.Items) < 1 {
				break
			}

			for _, item := range rs.Items {

				ow := &sko.ObjectWriter{
					Meta: item.Meta,
					Data: item.Data,
				}

				if sko.AttrAllow(item.Meta.Attrs, sko.ObjectMetaAttrDelete) {
					ow.ModeDeleteSet(true)
				}

				if rs2 := cn.objectCommitLocal(ow, item.Meta.Version); rs2.OK() {
					rr.LogOffset = item.Meta.Version
					num += 1
				}
			}

			if !rs.Next {
				break
			}
		}

		if rr.LogOffset > offset {
			cn.db.Put(keySysLogAsync(hp.Addr),
				[]byte(strconv.FormatUint(rr.LogOffset, 10)), nil)
		}

		if num > 0 {
			hlog.Printf("debug", "kvgo log async num %d, ver %d", num, rr.LogOffset)
		}
	}

	return nil
}
