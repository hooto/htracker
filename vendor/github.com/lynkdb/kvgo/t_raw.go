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
	"sync"

	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	t_raw_incr_mu sync.Mutex
)

func (cn *Conn) rawNew(key, value []byte, ttl int64) *Result {

	if data, err := cn.db.Get(key, nil); err == nil && len(data) > 0 {
		return newResult(0, nil)
	}

	return cn.rawPut(key, value, ttl)
}

func (cn *Conn) rawDel(keys ...[]byte) *Result {

	batch := new(leveldb.Batch)

	for _, key := range keys {
		batch.Delete(key)
	}

	if err := cn.db.Write(batch, nil); err != nil {
		return newResult(skv.ResultServerError, err)
	}

	return newResult(0, nil)
}

func (cn *Conn) rawGet(key []byte) *Result {

	data, err := cn.db.Get(key, nil)

	if err != nil {

		if err.Error() == "leveldb: not found" {
			return newResultNotFound()
		}

		return newResult(skv.ResultServerError, err)
	}

	return &Result{
		status: skv.ResultOK,
		Data:   data,
	}
}

func (cn *Conn) rawPut(key, value []byte, ttl int64) *Result {

	if len(key) < 2 {
		return newResultBadArgument()
	}

	if ttl > 0 {

		if ttl < 1000 {
			return newResultBadArgument()
		}

		if ok := cn.raw_ssttlat_put(key, uint64(types.MetaTimeNow().AddMillisecond(ttl))); !ok {
			return newResultBadArgument()
		}
	}

	if err := cn.db.Put(key, value, nil); err != nil {
		return newResult(skv.ResultServerError, err)
	}

	return newResult(0, nil)
}

func (cn *Conn) RawScan(offset, cutset []byte, limit int) *Result {
	return cn.rawScan(offset, cutset, limit)
}

func (cn *Conn) rawScan(offset, cutset []byte, limit int) *Result {

	if len(cutset) < 1 {
		cutset = offset
	}

	for i := len(cutset); i < 200; i++ {
		cutset = append(cutset, 0xff)
	}

	if limit > skv.ScanLimitMax {
		limit = skv.ScanLimitMax
	} else if limit < 1 {
		limit = 1
	}

	var (
		rs   = newResult(0, nil)
		iter = cn.db.NewIterator(&util.Range{
			Start: offset,
			Limit: cutset,
		}, nil)
	)

	for iter.Next() {

		if limit < 1 {
			break
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key())))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) rawRevScan(offset, cutset []byte, limit int) *Result {

	if len(offset) < 1 {
		offset = cutset
	}

	for i := len(offset); i < 256; i++ {
		offset = append(offset, 0x00)
	}

	for i := len(cutset); i < 256; i++ {
		cutset = append(cutset, 0xff)
	}

	if limit > skv.ScanLimitMax {
		limit = skv.ScanLimitMax
	} else if limit < 1 {
		limit = 1
	}

	var (
		rs   = newResult(0, nil)
		iter = cn.db.NewIterator(&util.Range{Start: offset, Limit: cutset}, nil)
	)

	for ok := iter.Last(); ok; ok = iter.Prev() {

		if limit < 1 {
			break
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key())))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) raw_ssttlat_put(key []byte, ttlat uint64) bool {

	if ttlat < 20060102150405000 {
		return true
	}

	//
	meta := skv.KvMeta{}

	if rs := cn.rawGet(t_ns_cat(ns_meta, key)); rs.OK() {

		if err := rs.Decode(&meta); err != nil {
			return false
		}

	} else if !rs.NotFound() {
		return false
	}

	if ttlat == meta.Expired {
		return true
	}

	batch := new(leveldb.Batch)

	if meta.Expired > 0 {
		batch.Delete(t_ns_cat(ns_ttl, append(uint64_to_bytes(meta.Expired), key...)))
	}

	meta.Expired = ttlat

	if value_enc, err := skv.ValueEncode(&meta, nil); err == nil {

		batch.Put(t_ns_cat(ns_meta, key), value_enc)
		batch.Put(t_ns_cat(ns_ttl, append(uint64_to_bytes(ttlat), key...)), []byte{0x00})

		if err := cn.db.Write(batch, nil); err != nil {
			return false
		}
	}

	return true
}
