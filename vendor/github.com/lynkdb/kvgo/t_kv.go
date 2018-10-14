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

	"github.com/lynkdb/iomix/skv"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	t_kv_incr_mu           sync.Mutex
	t_kv_write_options_def = &skv.KvWriteOptions{}
)

func (cn *Conn) KvNew(key []byte, value interface{}, opts *skv.KvWriteOptions) skv.Result {
	return newResult(0, nil)
}

func (cn *Conn) KvDel(keys ...[]byte) skv.Result {

	keys_kv := [][]byte{}

	for _, k := range keys {
		keys_kv = append(keys_kv, t_ns_cat(ns_kv, k))
	}

	return cn.rawDel(keys_kv...)
}

func (cn *Conn) KvPut(key []byte, value interface{}, opts *skv.KvWriteOptions) skv.Result {

	value_enc, err := skv.ValueEncode(value, nil)
	if err != nil {
		return newResultBadArgument()
	}

	if opts == nil {
		opts = t_kv_write_options_def
	}

	return cn.rawPut(t_ns_cat(ns_kv, key), value_enc, opts.Ttl)
}

func (cn *Conn) KvGet(key []byte) skv.Result {
	return cn.rawGet(t_ns_cat(ns_kv, key))
}

func (cn *Conn) KvScan(offset, cutset []byte, limit int) skv.Result {

	if len(cutset) < 1 {
		cutset = bytes_clone(offset)
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
			Start: t_ns_cat(ns_kv, offset),
			Limit: t_ns_cat(ns_kv, cutset),
		}, nil)
	)

	for iter.Next() {

		if limit < 1 {
			break
		}

		if len(iter.Value()) < 2 {
			continue
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key()[1:])))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) KvRevScan(offset, cutset []byte, limit int) skv.Result {

	var (
		off = t_ns_cat(ns_kv, offset)
		cut = t_ns_cat(ns_kv, cutset)
		rs  = newResult(0, nil)
	)

	for i := len(off); i < 200; i++ {
		off = append(off, 0xff)
	}

	if limit > skv.ScanLimitMax {
		limit = skv.ScanLimitMax
	} else if limit < 1 {
		limit = 1
	}

	iter := cn.db.NewIterator(&util.Range{
		Start: cut,
		Limit: off,
	}, nil)

	for ok := iter.Last(); ok; ok = iter.Prev() {

		if limit < 1 {
			break
		}

		if len(iter.Key()) < 2 {
			continue
		}

		if len(iter.Value()) < 2 {
			continue
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key()[1:])))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) KvIncr(key []byte, incr int64) skv.Result {

	if incr == 0 {
		return newResult(0, nil)
	}

	t_kv_incr_mu.Lock()
	defer t_kv_incr_mu.Unlock()

	key_kv := t_ns_cat(ns_kv, key)

	if rs := cn.rawGet(key_kv); !rs.OK() {

		if !rs.NotFound() {
			return newResult(skv.ResultServerError, nil)
		}

	} else {
		incr += rs.Int64()
	}

	value_enc, err := skv.ValueEncode(incr, nil)
	if err != nil {
		return newResultBadArgument()
	}

	rs := cn.rawPut(key_kv, value_enc, 0)
	if rs.OK() {
		rs.Data = value_enc
	}

	return rs
}
