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
	"strconv"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func (cn *Conn) Commit(rr *kv2.ObjectWriter) *kv2.ObjectResult {

	if len(cn.opts.Cluster.MainNodes) > 0 {

		if cn.opts.ClientConnectEnable {
			return cn.objectCommitRemote(rr, 0)
		}

		rs, err := cn.public.Commit(nil, rr)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}
		return rs
	}

	return cn.commitLocal(rr, 0)
}

func (cn *Conn) commitLocal(rr *kv2.ObjectWriter, cLog uint64) *kv2.ObjectResult {

	if err := rr.CommitValid(); err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	cn.mu.Lock()
	defer cn.mu.Unlock()

	meta, err := cn.objectMetaGet(rr)
	if meta == nil && err != nil {
		return kv2.NewObjectResultServerError(err)
	}

	tdb := cn.tabledb(rr.TableName)
	if tdb == nil {
		return kv2.NewObjectResultClientError(errors.New("table not found"))
	}

	if meta == nil {

		if kv2.AttrAllow(rr.Mode, kv2.ObjectWriterModeDelete) {
			return kv2.NewObjectResultOK()
		}

	} else {

		if rr.PrevVersion > 0 && rr.PrevVersion != meta.Version {
			return kv2.NewObjectResultClientError(errors.New("invalid prev_version"))
		}

		if rr.PrevDataCheck > 0 && rr.PrevDataCheck != meta.DataCheck {
			return kv2.NewObjectResultClientError(errors.New("invalid prev_data_check"))
		}

		if rr.PrevAttrs > 0 && !kv2.AttrAllow(meta.Attrs, rr.PrevAttrs) {
			return kv2.NewObjectResultClientError(errors.New("invalid prev_attrs"))
		}

		if rr.PrevIncrId > 0 && rr.PrevIncrId != meta.IncrId {
			return kv2.NewObjectResultClientError(errors.New("invalid prev_incr_id"))
		}

		if (cLog > 0 && meta.Version == cLog) ||
			kv2.AttrAllow(rr.Mode, kv2.ObjectWriterModeCreate) ||
			(rr.Meta.Expired == meta.Expired &&
				(rr.Meta.IncrId == 0 || rr.Meta.IncrId == meta.IncrId) &&
				(rr.PrevIncrId == 0 || rr.PrevIncrId == meta.IncrId) &&
				rr.Meta.DataCheck == meta.DataCheck) {

			rs := kv2.NewObjectResultOK()
			rs.Meta = &kv2.ObjectMeta{
				Version: meta.Version,
				IncrId:  meta.IncrId,
				Created: meta.Created,
				Updated: meta.Updated,
			}
			return rs
		}

		if rr.Meta.IncrId == 0 && meta.IncrId > 0 {
			rr.Meta.IncrId = meta.IncrId
		}

		if meta.Attrs > 0 {
			rr.Meta.Attrs |= meta.Attrs
		}

		if meta.Created > 0 {
			rr.Meta.Created = meta.Created
		}
	}

	if rr.Meta.Updated < 1 {
		rr.Meta.Updated = uint64(time.Now().UnixNano() / 1e6)
	}

	if rr.Meta.Created < 1 {
		rr.Meta.Created = rr.Meta.Updated
	}

	if rr.IncrNamespace != "" {

		if rr.Meta.IncrId == 0 {
			rr.Meta.IncrId, err = cn.objectIncrSet(tdb, rr.IncrNamespace, 1, 0)
			if err != nil {
				return kv2.NewObjectResultServerError(err)
			}
		} else {
			cn.objectIncrSet(tdb, rr.IncrNamespace, 0, rr.Meta.IncrId)
		}
	}

	cLogOn := true
	if cLog == 0 {
		if meta != nil && meta.Version > 0 {
			cLog = meta.Version
		}

		cLog, err = cn.objectLogVersionSet(tdb, 1, cLog)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}
	} else {
		_, err = cn.objectLogVersionSet(tdb, 0, cLog)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}
		cLogOn = false
	}
	rr.Meta.Version = cLog

	if kv2.AttrAllow(rr.Mode, kv2.ObjectWriterModeDelete) {

		rr.Meta.Attrs = kv2.ObjectMetaAttrDelete

		if bsMeta, err := rr.MetaEncode(); err == nil {

			batch := new(leveldb.Batch)

			if meta != nil {
				if !cn.opts.Feature.WriteMetaDisable {
					batch.Delete(keyEncode(nsKeyMeta, rr.Meta.Key))
				}
				batch.Delete(keyEncode(nsKeyData, rr.Meta.Key))
				if !cn.opts.Feature.WriteLogDisable {
					batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
				}
			}

			if cLogOn && !cn.opts.Feature.WriteLogDisable {
				batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)
			}

			err = tdb.db.Write(batch, nil)
		}

	} else {

		if bsMeta, bsData, err := rr.PutEncode(); err == nil {

			batch := new(leveldb.Batch)

			if kv2.AttrAllow(rr.Meta.Attrs, kv2.ObjectMetaAttrDataOff) {
				batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsData)
			} else if kv2.AttrAllow(rr.Meta.Attrs, kv2.ObjectMetaAttrMetaOff) {
				batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			} else {
				if !cn.opts.Feature.WriteMetaDisable {
					batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsMeta)
				}
				batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			}

			if cLogOn && !cn.opts.Feature.WriteLogDisable {
				batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)
			}

			if rr.Meta.Expired > 0 {
				batch.Put(keyExpireEncode(nsKeyTtl, rr.Meta.Expired, rr.Meta.Key), bsMeta)
			}

			if meta != nil {
				if meta.Version < cLog && !cn.opts.Feature.WriteLogDisable {
					batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
				}
				if meta.Expired > 0 && meta.Expired != rr.Meta.Expired {
					batch.Delete(keyExpireEncode(nsKeyTtl, meta.Expired, rr.Meta.Key))
				}
			}

			err = tdb.db.Write(batch, nil)
		}
	}

	if err != nil {
		return kv2.NewObjectResultServerError(err)
	}

	rs := kv2.NewObjectResultOK()
	rs.Meta = &kv2.ObjectMeta{
		Version: cLog,
		IncrId:  rr.Meta.IncrId,
		Updated: rr.Meta.Updated,
	}

	return rs
}

func (cn *Conn) objectCommitRemote(rr *kv2.ObjectWriter, cLog uint64) *kv2.ObjectResult {

	err := rr.CommitValid()
	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	mainNodes := cn.opts.Cluster.randMainNodes(3)
	if len(mainNodes) < 1 {
		return kv2.NewObjectResultClientError(errors.New("no master found"))
	}

	for _, v := range mainNodes {

		conn, err := clientConn(v.Addr, v.AccessKey, v.AuthTLSCert, false)
		if err != nil {
			continue
		}

		ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
		defer fc()

		rs, err := kv2.NewPublicClient(conn).Commit(ctx, rr)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}

		return rs
	}

	return kv2.NewObjectResultServerError(errors.New("no cluster nodes"))
}

func (cn *Conn) Query(rr *kv2.ObjectReader) *kv2.ObjectResult {

	if cn.opts.ClientConnectEnable ||
		(len(cn.opts.Cluster.MainNodes) > 0 && cn.opts.Server.Bind == "") {
		return cn.objectQueryRemote(rr)
	}

	return cn.objectLocalQuery(rr)
}

func (cn *Conn) objectLocalQuery(rr *kv2.ObjectReader) *kv2.ObjectResult {

	rs := kv2.NewObjectResultOK()

	tdb := cn.tabledb(rr.TableName)
	if tdb == nil {
		rs.StatusMessage(kv2.ResultClientError, "table not found")
		return rs
	}

	if kv2.AttrAllow(rr.Mode, kv2.ObjectReaderModeKey) {

		for _, k := range rr.Keys {

			var (
				bs  []byte
				err error
			)

			if kv2.AttrAllow(rr.Attrs, kv2.ObjectMetaAttrDataOff) {
				bs, err = tdb.db.Get(keyEncode(nsKeyMeta, k), nil)
			} else {
				bs, err = tdb.db.Get(keyEncode(nsKeyData, k), nil)
			}

			if err == nil {

				item, err := kv2.ObjectItemDecode(bs)
				if err == nil {
					rs.Items = append(rs.Items, item)
				} else {
					rs.StatusMessage(kv2.ResultServerError, err.Error())
				}

			} else {

				if err.Error() != ldbNotFound {
					rs.StatusMessage(kv2.ResultServerError, err.Error())
					break
				}

				if len(rr.Keys) == 1 {
					rs.StatusMessage(kv2.ResultNotFound, "")
				}
			}
		}

	} else if kv2.AttrAllow(rr.Mode, kv2.ObjectReaderModeKeyRange) {

		if err := cn.objectQueryKeyRange(rr, rs); err != nil {
			rs.StatusMessage(kv2.ResultServerError, err.Error())
		}

	} else if kv2.AttrAllow(rr.Mode, kv2.ObjectReaderModeLogRange) {

		if err := cn.objectQueryLogRange(rr, rs); err != nil {
			rs.StatusMessage(kv2.ResultServerError, err.Error())
		}

	} else {

		rs.StatusMessage(kv2.ResultClientError, "invalid mode")
	}

	if rs.Status == 0 {
		rs.Status = kv2.ResultOK
	}

	return rs
}

func (cn *Conn) objectQueryRemote(rr *kv2.ObjectReader) *kv2.ObjectResult {

	mainNodes := cn.opts.Cluster.randMainNodes(3)
	if len(mainNodes) < 1 {
		return kv2.NewObjectResultClientError(errors.New("no master found"))
	}

	for _, v := range mainNodes {

		conn, err := clientConn(v.Addr, v.AccessKey, v.AuthTLSCert, false)
		if err != nil {
			continue
		}

		ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
		defer fc()

		rs, err := kv2.NewPublicClient(conn).Query(ctx, rr)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}

		return rs
	}

	return kv2.NewObjectResultServerError(errors.New("no cluster nodes"))
}

func (cn *Conn) objectQueryKeyRange(rr *kv2.ObjectReader, rs *kv2.ObjectResult) error {

	tdb := cn.tabledb(rr.TableName)
	if tdb == nil {
		return errors.New("table not found")
	}

	nsKey := nsKeyData
	if kv2.AttrAllow(rr.Attrs, kv2.ObjectMetaAttrDataOff) {
		nsKey = nsKeyMeta
	}

	var (
		offset    = keyEncode(nsKey, bytesClone(rr.KeyOffset))
		cutset    = keyEncode(nsKey, bytesClone(rr.KeyCutset))
		limitNum  = rr.LimitNum
		limitSize = rr.LimitSize
	)

	if limitNum > kv2.ObjectReaderLimitNumMax {
		limitNum = kv2.ObjectReaderLimitNumMax
	} else if limitNum < 1 {
		limitNum = 1
	}

	if limitSize < 1 {
		limitSize = kv2.ObjectReaderLimitSizeDef
	} else if limitSize > kv2.ObjectReaderLimitSizeMax {
		limitSize = kv2.ObjectReaderLimitSizeMax
	}

	var (
		iter   iterator.Iterator
		values = [][]byte{}
	)

	if kv2.AttrAllow(rr.Mode, kv2.ObjectReaderModeRevRange) {

		// offset = append(offset, 0xff)

		iter = tdb.db.NewIterator(&util.Range{
			Start: cutset,
			Limit: offset,
		}, nil)

		for ok := iter.Last(); ok; ok = iter.Prev() {

			if limitNum < 1 {
				break
			}

			if bytes.Compare(iter.Key(), offset) >= 0 {
				continue
			}

			if bytes.Compare(iter.Key(), cutset) < 0 {
				break
			}

			if len(iter.Value()) < 2 {
				continue
			}

			limitSize -= int64(len(iter.Value()))
			if limitSize < 1 {
				break
			}

			limitNum -= 1
			values = append(values, bytesClone(iter.Value()))
		}

	} else {

		cutset = append(cutset, 0xff)

		iter = tdb.db.NewIterator(&util.Range{
			Start: offset,
			Limit: cutset,
		}, nil)

		for iter.Next() {

			if limitNum < 1 {
				break
			}

			if bytes.Compare(iter.Key(), offset) <= 0 {
				continue
			}

			if bytes.Compare(iter.Key(), cutset) >= 0 {
				break
			}

			if len(iter.Value()) < 2 {
				continue
			}

			limitSize -= int64(len(iter.Value()))
			if limitSize < 1 {
				break
			}

			limitNum -= 1
			values = append(values, bytesClone(iter.Value()))
		}
	}

	iter.Release()

	if iter.Error() != nil {
		return iter.Error()
	}

	for _, bs := range values {
		if item, err := kv2.ObjectItemDecode(bs); err == nil {
			rs.Items = append(rs.Items, item)
		}
	}

	if limitNum < 1 || limitSize < 1 {
		rs.Next = true
	}

	return nil
}

func (cn *Conn) objectQueryLogRange(rr *kv2.ObjectReader, rs *kv2.ObjectResult) error {

	tdb := cn.tabledb(rr.TableName)
	if tdb == nil {
		return errors.New("table not found")
	}

	var (
		offset    = keyEncode(nsKeyLog, uint64ToBytes(rr.LogOffset))
		cutset    = keyEncode(nsKeyLog, []byte{0xff})
		limitNum  = rr.LimitNum
		limitSize = rr.LimitSize
	)

	if limitNum > kv2.ObjectReaderLimitNumMax {
		limitNum = kv2.ObjectReaderLimitNumMax
	} else if limitNum < 1 {
		limitNum = 1
	}

	if limitSize < 1 {
		limitSize = kv2.ObjectReaderLimitSizeDef
	} else if limitSize > kv2.ObjectReaderLimitSizeMax {
		limitSize = kv2.ObjectReaderLimitSizeMax
	}

	var (
		tto  = uint64(time.Now().UnixNano()/1e6) - 3000
		iter = tdb.db.NewIterator(&util.Range{
			Start: offset,
			Limit: cutset,
		}, nil)
	)

	for iter.Next() {

		if limitNum < 1 {
			break
		}

		if bytes.Compare(iter.Key(), offset) <= 0 {
			continue
		}

		if bytes.Compare(iter.Key(), cutset) >= 0 {
			break
		}

		if len(iter.Value()) < 2 {
			continue
		}

		meta, err := kv2.ObjectMetaDecode(iter.Value())
		if err != nil || meta == nil {
			if err != nil {
				hlog.Printf("debug", "db-log-range err %s", err.Error())
			}
			break
		}

		//
		if kv2.AttrAllow(meta.Attrs, kv2.ObjectMetaAttrDelete) {
			rs.Items = append(rs.Items, &kv2.ObjectItem{
				Meta: meta,
			})
		} else {

			var nsKey = nsKeyData
			if kv2.AttrAllow(meta.Attrs, kv2.ObjectMetaAttrDataOff) {
				nsKey = nsKeyMeta
			}

			bs, err := tdb.db.Get(keyEncode(nsKey, meta.Key), nil)

			if err != nil && err.Error() == ldbNotFound {
				if nsKey == nsKeyData {
					nsKey = nsKeyMeta
				} else {
					nsKey = nsKeyData
				}
				bs, err = tdb.db.Get(keyEncode(nsKey, meta.Key), nil)
			}

			if err != nil {
				hlog.Printf("debug", "db-log-range err %s", err.Error())
				break
			}

			limitSize -= int64(len(bs))
			if limitSize < 1 {
				break
			}

			if item, err := kv2.ObjectItemDecode(bs); err == nil {
				if item.Meta.Updated >= tto {
					break
				}
				rs.Items = append(rs.Items, item)
			}
		}

		limitNum -= 1
	}

	iter.Release()

	if iter.Error() != nil {
		return iter.Error()
	}

	if limitNum < 1 || limitSize < 1 {
		rs.Next = true
	}

	return nil
}

func (cn *Conn) NewReader(keys ...[]byte) *kv2.ClientReader {
	return kv2.NewClientReader(cn, keys...)
}

func (cn *Conn) NewWriter(key []byte, value interface{}, opts ...interface{}) *kv2.ClientWriter {
	return kv2.NewClientWriter(cn, key, value, opts...)
}

func (cn *Conn) objectMetaGet(rr *kv2.ObjectWriter) (*kv2.ObjectMeta, error) {

	tdb := cn.tabledb(rr.TableName)
	if tdb == nil {
		return nil, errors.New("table not found")
	}

	var nsKey = nsKeyMeta
	if kv2.AttrAllow(rr.Meta.Attrs, kv2.ObjectMetaAttrMetaOff) ||
		cn.opts.Feature.WriteMetaDisable {
		nsKey = nsKeyData
	}
	data, err := tdb.db.Get(keyEncode(nsKey, rr.Meta.Key), nil)

	if err != nil && err.Error() == ldbNotFound {
		if nsKey == nsKeyData {
			nsKey = nsKeyMeta
		} else {
			nsKey = nsKeyData
		}
		data, err = tdb.db.Get(keyEncode(nsKey, rr.Meta.Key), nil)
	}

	if err == nil {
		return kv2.ObjectMetaDecode(data)
	} else {
		if err.Error() == ldbNotFound {
			err = nil
		}
	}

	return nil, err
}

func (cn *Conn) objectLogVersionSet(tdb *dbTable, incr, set uint64) (uint64, error) {

	cn.logMu.Lock()
	defer cn.logMu.Unlock()

	if incr == 0 && set == 0 {
		return cn.logOffset, nil
	}

	if cn.logCutset <= 100 {

		if bs, err := tdb.db.Get(keySysLogCutset, nil); err != nil {
			if err.Error() != ldbNotFound {
				return 0, err
			}
		} else {
			if cn.logCutset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return 0, err
			}
			if cn.logOffset < cn.logCutset {
				cn.logOffset = cn.logCutset
			}
		}
	}

	if cn.logOffset < 100 {
		cn.logOffset = 100
	}

	if set > 0 && set > cn.logOffset {
		incr += (set - cn.logOffset)
	}

	if incr > 0 {

		if (cn.logOffset + incr) >= cn.logCutset {

			cutset := cn.logOffset + incr + 100

			if n := cutset % 100; n > 0 {
				cutset += n
			}

			if err := tdb.db.Put(keySysLogCutset,
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			cn.logCutset = cutset
		}

		cn.logOffset += incr
	}

	return cn.logOffset, nil
}

func (cn *Conn) objectIncrSet(tdb *dbTable, ns string, incr, set uint64) (uint64, error) {

	cn.incrMu.Lock()
	defer cn.incrMu.Unlock()

	incrSet := tdb.incrSets[ns]
	if incrSet == nil {
		incrSet = &dbTableIncrSet{
			offset: 0,
			cutset: 0,
		}
		tdb.incrSets[ns] = incrSet
	}

	if incr == 0 && set == 0 {
		return incrSet.offset, nil
	}

	if incrSet.cutset <= 100 {

		if bs, err := tdb.db.Get(keySysIncrCutset(ns), nil); err != nil {
			if err.Error() != ldbNotFound {
				return 0, err
			}
		} else {
			if incrSet.cutset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return 0, err
			}
			if incrSet.offset < incrSet.cutset {
				incrSet.offset = incrSet.cutset
			}
		}
	}

	if incrSet.offset < 100 {
		incrSet.offset = 100
	}

	if set > 0 && set > incrSet.offset {
		incr += (set - incrSet.offset)
	}

	if incr > 0 {

		if (incrSet.offset + incr) >= incrSet.cutset {

			cutset := incrSet.offset + incr + 100

			if err := tdb.db.Put(keySysIncrCutset(ns),
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			incrSet.cutset = cutset
		}

		incrSet.offset += incr
	}

	return incrSet.offset, nil
}

func (it *Conn) Connector() kv2.ClientConnector {
	return it
}
