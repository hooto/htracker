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

	"github.com/lynkdb/iomix/sko"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (cn *Conn) Commit(rr *sko.ObjectWriter) *sko.ObjectResult {

	if len(cn.opts.Cluster.Masters) > 0 {

		if cn.opts.ClientConnectEnable {
			return cn.objectCommitRemote(rr, 0)
		}

		rs, err := cn.cluster.Commit(nil, rr)
		if err != nil {
			return sko.NewObjectResultServerError(err)
		}
		return rs
	}

	return cn.objectCommitLocal(rr, 0)
}

func (cn *Conn) objectCommitLocal(rr *sko.ObjectWriter, cLog uint64) *sko.ObjectResult {

	if err := rr.CommitValid(); err != nil {
		return sko.NewObjectResultClientError(err)
	}

	cn.mu.Lock()
	defer cn.mu.Unlock()

	meta, err := cn.objectMetaGet(rr)
	if meta == nil && err != nil {
		return sko.NewObjectResultServerError(err)
	}

	if meta == nil {

		if sko.AttrAllow(rr.Mode, sko.ObjectWriterModeDelete) {
			return sko.NewObjectResultOK()
		}

	} else {

		if (cLog > 0 && meta.Version == cLog) ||
			sko.AttrAllow(rr.Mode, sko.ObjectWriterModeCreate) ||
			(rr.Meta.Expired == meta.Expired && meta.DataCheck == rr.Meta.DataCheck) {

			rs := sko.NewObjectResultOK()
			rs.Meta = &sko.ObjectMeta{
				Version: meta.Version,
				IncrId:  meta.IncrId,
				Created: meta.Created,
				Updated: meta.Updated,
			}

			return rs
		}

		if rr.PrevVersion > 0 && rr.PrevVersion != meta.Version {
			return sko.NewObjectResultClientError(errors.New("invalid prev-version"))
		}

		if rr.PrevDataCheck > 0 && rr.PrevDataCheck != meta.DataCheck {
			return sko.NewObjectResultClientError(errors.New("invalid prev-data-check"))
		}

		if meta.IncrId > 0 {
			rr.Meta.IncrId = meta.IncrId
		}

		if meta.Created > 0 {
			rr.Meta.Created = meta.Created
		}
	}

	if rr.Meta.Created < 1 {
		rr.Meta.Created = rr.Meta.Updated
	}

	if rr.IncrNamespace != "" {

		if rr.Meta.IncrId == 0 {
			rr.Meta.IncrId, err = cn.objectIncrSet(rr.IncrNamespace, 1, 0)
			if err != nil {
				return sko.NewObjectResultServerError(err)
			}
		} else {
			cn.objectIncrSet(rr.IncrNamespace, 0, rr.Meta.IncrId)
		}
	}

	if cLog == 0 {
		if meta != nil && meta.Version > 0 {
			cLog = meta.Version
		}

		cLog, err = cn.objectLogVersionSet(1, cLog)
		if err != nil {
			return sko.NewObjectResultServerError(err)
		}
	}
	rr.Meta.Version = cLog

	if sko.AttrAllow(rr.Mode, sko.ObjectWriterModeDelete) {

		rr.Meta.Attrs = sko.ObjectMetaAttrDelete

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

			if !cn.opts.Feature.WriteLogDisable {
				batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)
			}

			err = cn.db.Write(batch, nil)
		}

	} else {

		if bsMeta, bsData, err := rr.PutEncode(); err == nil {

			batch := new(leveldb.Batch)

			if !cn.opts.Feature.WriteMetaDisable {
				batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsMeta)
			}
			batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			if !cn.opts.Feature.WriteLogDisable {
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

			err = cn.db.Write(batch, nil)
		}
	}

	if err != nil {
		return sko.NewObjectResultServerError(err)
	}

	rs := sko.NewObjectResultOK()
	rs.Meta = &sko.ObjectMeta{
		Version: cLog,
		IncrId:  rr.Meta.IncrId,
		Updated: rr.Meta.Updated,
	}

	return rs
}

func (cn *Conn) objectCommitRemote(rr *sko.ObjectWriter, cLog uint64) *sko.ObjectResult {

	err := rr.CommitValid()
	if err != nil {
		return sko.NewObjectResultClientError(err)
	}

	addrs := cn.opts.Cluster.masterAddrs(3)
	if len(addrs) < 1 {
		return sko.NewObjectResultClientError(errors.New("no master found"))
	}

	for _, addr := range addrs {

		conn, err := clientConn(addr, cn.authKey(addr))
		if err != nil {
			continue
		}

		ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
		defer fc()

		rs, err := sko.NewObjectClient(conn).Commit(ctx, rr)
		if err != nil {
			return sko.NewObjectResultServerError(err)
		}

		return rs
	}

	return sko.NewObjectResultServerError(errors.New("no cluster nodes"))
}

func (cn *Conn) Query(rr *sko.ObjectReader) *sko.ObjectResult {

	if cn.opts.ClientConnectEnable {
		return cn.objectQueryRemote(rr)
	}

	rs := sko.NewObjectResultOK()

	if sko.AttrAllow(rr.Mode, sko.ObjectReaderModeKey) {

		for _, k := range rr.Keys {

			bs, err := cn.db.Get(keyEncode(nsKeyData, k), nil)
			if err == nil {

				item, err := sko.ObjectItemDecode(bs)
				if err == nil {
					rs.Items = append(rs.Items, item)
				} else {
					rs.StatusMessage(sko.ResultServerError, err.Error())
				}

			} else {

				if err.Error() != ldbNotFound {
					rs.StatusMessage(sko.ResultServerError, err.Error())
					break
				}

				if len(rr.Keys) == 1 {
					rs.StatusMessage(sko.ResultNotFound, "")
				}
			}
		}

	} else if sko.AttrAllow(rr.Mode, sko.ObjectReaderModeKeyRange) {

		if err := cn.objectQueryKeyRange(rr, rs); err != nil {
			rs.StatusMessage(sko.ResultServerError, err.Error())
		}

	} else if sko.AttrAllow(rr.Mode, sko.ObjectReaderModeLogRange) {

		if err := cn.objectQueryLogRange(rr, rs); err != nil {
			rs.StatusMessage(sko.ResultServerError, err.Error())
		}

	} else {

		rs.StatusMessage(sko.ResultClientError, "invalid mode")
	}

	if rs.Status == 0 {
		rs.Status = sko.ResultOK
	}

	return rs
}

func (cn *Conn) objectQueryRemote(rr *sko.ObjectReader) *sko.ObjectResult {

	addrs := cn.opts.Cluster.masterAddrs(3)
	if len(addrs) < 1 {
		return sko.NewObjectResultClientError(errors.New("no master found"))
	}

	for _, addr := range addrs {

		conn, err := clientConn(addr, cn.authKey(addr))
		if err != nil {
			continue
		}

		ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
		defer fc()

		rs, err := sko.NewObjectClient(conn).Query(ctx, rr)
		if err != nil {
			return sko.NewObjectResultServerError(err)
		}

		return rs
	}

	return sko.NewObjectResultServerError(errors.New("no cluster nodes"))
}

func (cn *Conn) objectQueryKeyRange(rr *sko.ObjectReader, rs *sko.ObjectResult) error {

	var (
		offset    = keyEncode(nsKeyData, bytesClone(rr.KeyOffset))
		cutset    = keyEncode(nsKeyData, bytesClone(rr.KeyCutset))
		limitNum  = rr.LimitNum
		limitSize = rr.LimitSize
	)

	if limitNum > sko.ObjectReaderLimitNumMax {
		limitNum = sko.ObjectReaderLimitNumMax
	} else if limitNum < 1 {
		limitNum = 1
	}

	if limitSize < 1 {
		limitSize = sko.ObjectReaderLimitSizeDef
	} else if limitSize > sko.ObjectReaderLimitSizeMax {
		limitSize = sko.ObjectReaderLimitSizeMax
	}

	var (
		iter   iterator.Iterator
		values = [][]byte{}
	)

	if sko.AttrAllow(rr.Mode, sko.ObjectReaderModeRevRange) {

		// offset = append(offset, 0xff)

		iter = cn.db.NewIterator(&util.Range{
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

		iter = cn.db.NewIterator(&util.Range{
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
		if item, err := sko.ObjectItemDecode(bs); err == nil {
			rs.Items = append(rs.Items, item)
		}
	}

	if limitNum < 1 || limitSize < 1 {
		rs.Next = true
	}

	return nil
}

func (cn *Conn) objectQueryLogRange(rr *sko.ObjectReader, rs *sko.ObjectResult) error {

	var (
		offset    = keyEncode(nsKeyLog, uint64ToBytes(rr.LogOffset))
		cutset    = keyEncode(nsKeyLog, []byte{0xff})
		limitNum  = rr.LimitNum
		limitSize = rr.LimitSize
	)

	if limitNum > sko.ObjectReaderLimitNumMax {
		limitNum = sko.ObjectReaderLimitNumMax
	} else if limitNum < 1 {
		limitNum = 1
	}

	if limitSize < 1 {
		limitSize = sko.ObjectReaderLimitSizeDef
	} else if limitSize > sko.ObjectReaderLimitSizeMax {
		limitSize = sko.ObjectReaderLimitSizeMax
	}

	var (
		tto  = uint64(time.Now().UnixNano()/1e6) - 3000
		iter = cn.db.NewIterator(&util.Range{
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

		meta, err := sko.ObjectMetaDecode(iter.Value())
		if err != nil || meta == nil {
			break
		}

		//
		if sko.AttrAllow(meta.Attrs, sko.ObjectMetaAttrDelete) {
			rs.Items = append(rs.Items, &sko.ObjectItem{
				Meta: meta,
			})
		} else {

			bs, err := cn.db.Get(keyEncode(nsKeyData, meta.Key), nil)
			if err != nil {
				break
			}

			limitSize -= int64(len(bs))
			if limitSize < 1 {
				break
			}

			if item, err := sko.ObjectItemDecode(bs); err == nil {
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

func (cn *Conn) NewReader(key []byte) *sko.ClientReader {
	return sko.NewClientReader(cn, key)
}

func (cn *Conn) NewWriter(key []byte, value interface{}) *sko.ClientWriter {
	return sko.NewClientWriter(cn, key, value)
}

func (cn *Conn) objectMetaGet(rr *sko.ObjectWriter) (*sko.ObjectMeta, error) {

	data, err := cn.db.Get(keyEncode(nsKeyMeta, rr.Meta.Key), nil)
	if err == nil {
		return sko.ObjectMetaDecode(data)
	} else {
		if err.Error() == ldbNotFound {
			err = nil
		}
	}

	return nil, err
}

func (cn *Conn) objectLogVersionSet(incr, set uint64) (uint64, error) {

	cn.logMu.Lock()
	defer cn.logMu.Unlock()

	if incr == 0 && set == 0 {
		return cn.logOffset, nil
	}

	if cn.logCutset <= 100 {

		if bs, err := cn.db.Get(keySysLogCutset, nil); err != nil {
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

			if err := cn.db.Put(keySysLogCutset,
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			cn.logCutset = cutset
		}

		cn.logOffset += incr
	}

	return cn.logOffset, nil
}

func (cn *Conn) objectIncrSet(ns string, incr, set uint64) (uint64, error) {

	cn.incrMu.Lock()
	defer cn.incrMu.Unlock()

	if incr == 0 && set == 0 {
		return cn.incrOffset, nil
	}

	if cn.incrCutset <= 100 {

		if bs, err := cn.db.Get(keySysIncrCutset(ns), nil); err != nil {
			if err.Error() != ldbNotFound {
				return 0, err
			}
		} else {
			if cn.incrCutset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return 0, err
			}
			if cn.incrOffset < cn.incrCutset {
				cn.incrOffset = cn.incrCutset
			}
		}
	}

	if cn.incrOffset < 100 {
		cn.incrOffset = 100
	}

	if set > 0 && set > cn.incrOffset {
		incr += (set - cn.incrOffset)
	}

	if incr > 0 {

		if (cn.incrOffset + incr) >= cn.incrCutset {

			cutset := cn.incrOffset + incr + 100

			if n := cutset % 100; n > 0 {
				cutset += n
			}

			if err := cn.db.Put(keySysIncrCutset(ns),
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			cn.incrCutset = cutset
		}

		cn.incrOffset += incr
	}

	return cn.incrOffset, nil
}
