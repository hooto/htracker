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
	"errors"
	"sync"

	"github.com/lynkdb/iomix/skv"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	t_prog_incr_mu sync.Mutex
)

func (cn *Conn) KvProgNew(key skv.KvProgKey, val skv.KvEntry, opts *skv.KvProgWriteOptions) skv.Result {

	if opts == nil {
		opts = &skv.KvProgWriteOptions{}
	}

	opts.Actions = opts.Actions | skv.KvProgOpCreate

	return cn.KvProgPut(key, val, opts)
}

func (cn *Conn) KvProgPut(key skv.KvProgKey, val skv.KvEntry, opts *skv.KvProgWriteOptions) skv.Result {

	if !key.Valid() || !val.Valid() {
		return newResultBadArgument()
	}

	if opts == nil {
		opts = &skv.KvProgWriteOptions{}
	}

	var (
		p_rs      *Result
		p_meta    *skv.KvMeta
		prev_diff = true
	)

	if opts.OpAllow(skv.KvProgOpCreate) {
		if rs := cn.rawGet(key.Encode(ns_prog_def)); rs.OK() {
			return newResult(0, nil)
		} else {
			p_rs, p_meta = rs, rs.Meta()
		}
	}

	// TODO
	// if i, entry := key.LastEntry(); entry != nil {

	// 	switch entry.Type {

	// 	case skv.KvProgKeyEntryIncr:
	// 		if i < 1 {
	// 			return newResultBadArgument()
	// 		}
	// 		pmeta := key.EncodeIndex(ns_prog_def, i-1)
	// 		if len(pmeta) < 1 {
	// 			return newResultBadArgument()
	// 		}
	// 		kiv, err := cn.progExtIncrby(pmeta, 1)
	// 		if err != nil {
	// 			return newResultBadArgument()
	// 		}
	// 		if err := key.Set(i, kiv); err != nil {
	// 			return newResultBadArgument()
	// 		}
	// 		prev_diff = false

	// 	case skv.KvProgKeyEntryBytes:
	// 	case skv.KvProgKeyEntryUint:

	// 	default:
	// 		return newResultBadArgument()
	// 	}
	// }

	if prev_diff && (opts.PrevSum > 0 || opts.OpAllow(skv.KvProgOpFoldMeta)) {
		if p_rs == nil {
			if rs := cn.rawGet(key.Encode(ns_prog_def)); rs.OK() {
				p_rs, p_meta = rs, rs.Meta()
			}
		}
	}

	if opts.PrevSum > 0 && p_rs != nil {
		if p_rs.Crc32() != opts.PrevSum {
			return newResultBadArgument()
		}
	}

	if opts.OpAllow(skv.KvProgOpMetaSum) {
		val.KvMeta().Sum = 1
	}
	if opts.OpAllow(skv.KvProgOpMetaSize) {
		val.KvMeta().Size = 1
	}

	batch := new(leveldb.Batch)

	if opts.Expired > 0 {
		val.KvMeta().Expired = opts.Expired
		ttl_key := t_ns_cat(ns_prog_ttl,
			append(uint64_to_bytes(opts.Expired), key.Encode(ns_prog_def)...))
		batch.Put(ttl_key, []byte{0x00})
	}

	if opts.OpAllow(skv.KvProgOpFoldMeta) {
		fmeta := cn.rawGet(key.EncodeFoldMeta(ns_prog_def)).Meta()
		if fmeta == nil {
			fmeta = &skv.KvMeta{}
		}
		if p_rs == nil || p_meta == nil || p_meta.Num == 0 {
			fmeta.Num++
			fmeta.Size += uint64(val.ValueSize())
		} else {
			if si := (val.ValueSize() - p_rs.ValueSize()); si > 0 {
				fmeta.Size += uint64(si)
			} else if fmeta.Size > uint64(-si) {
				fmeta.Size -= uint64(-si)
			} else {
				fmeta.Size = 0
			}
		}
		if bs := fmeta.Encode(); len(bs) > 1 {
			cn.rawPut(key.EncodeFoldMeta(ns_prog_def), bs, 0)
		}
		val.KvMeta().Num = 1
	}

	batch.Put(key.Encode(ns_prog_def), val.Encode())
	if err := cn.db.Write(batch, nil); err != nil {
		return newResult(skv.ResultServerError, err)
	}

	return newResult(0, nil)
}

func (cn *Conn) KvProgGet(key skv.KvProgKey) skv.Result {
	if len(key.Encode(ns_prog_def)) == 0 {
		return newResultBadArgument()
	}
	return cn.rawGet(key.Encode(ns_prog_def))
}

func (cn *Conn) KvProgDel(key skv.KvProgKey, opts *skv.KvProgWriteOptions) skv.Result {

	if len(key.Encode(ns_prog_def)) == 0 {
		return newResultBadArgument()
	}

	rs := cn.rawGet(key.Encode(ns_prog_def))
	if rs.OK() {
		if meta := rs.Meta(); meta != nil && meta.Num == 1 {

			if fmeta := cn.rawGet(key.EncodeFoldMeta(ns_prog_def)).Meta(); fmeta != nil {
				if fmeta.Size > uint64(rs.ValueSize()) {
					fmeta.Size -= uint64(rs.ValueSize())
				} else {
					fmeta.Size = 0
				}
				if fmeta.Num <= 1 {
					cn.rawDel(key.EncodeFoldMeta(ns_prog_def))
				} else {
					fmeta.Num--
					if bs := fmeta.Encode(); len(bs) > 1 {
						cn.rawPut(key.EncodeFoldMeta(ns_prog_def), bs, 0)
					}
				}
			}
		}

		rs = cn.rawDel(key.Encode(ns_prog_def))
	}

	return rs
}

func (cn *Conn) KvProgScan(offset, cutset skv.KvProgKey, limit int) skv.Result {

	var (
		plen = offset.FoldLen()
		off  = offset.Encode(ns_prog_def)
		cut  = cutset.Encode(ns_prog_def)
		rs   = newResult(0, nil)
	)

	for i := len(cut); i < 200; i += 4 {
		cut = append(cut, []byte{0xff, 0xff, 0xff, 0xff}...)
	}

	if limit > skv.ScanLimitMax {
		limit = skv.ScanLimitMax
	} else if limit < 1 {
		limit = 1
	}

	iter := cn.db.NewIterator(&util.Range{
		Start: off,
		Limit: cut,
	}, nil)

	for iter.Next() {

		if limit < 1 {
			break
		}

		if len(iter.Key()) <= plen || len(iter.Value()) < 2 {
			continue
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key()[plen:])))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) KvProgRevScan(offset, cutset skv.KvProgKey, limit int) skv.Result {

	var (
		plen = offset.FoldLen()
		off  = offset.Encode(ns_prog_def)
		cut  = cutset.Encode(ns_prog_def)
		rs   = newResult(0, nil)
	)

	for i := len(off); i < 200; i += 4 {
		off = append(off, []byte{0xff, 0xff, 0xff, 0xff}...)
	}

	for i := len(cut); i < 200; i += 4 {
		cut = append(cut, []byte{0x00, 0x00, 0x00, 0x00}...)
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

		if len(iter.Key()) <= plen || len(iter.Value()) < 2 {
			continue
		}

		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Key()[plen:])))
		rs.Items = append(rs.Items, newResultData(bytes_clone(iter.Value())))

		limit--
	}

	iter.Release()

	if iter.Error() != nil {
		return newResult(skv.ResultServerError, iter.Error())
	}

	return rs
}

func (cn *Conn) progExtIncrby(key []byte, incr int64) (uint64, error) {

	if incr < 1 {
		return 0, errors.New("BadArgument::INCR")
	}

	key_enc := append(ns_prog_x_incr, key...)

	t_prog_incr_mu.Lock()
	defer t_prog_incr_mu.Unlock()

	if rs := cn.rawGet(key_enc); !rs.OK() {
		if !rs.NotFound() {
			return 0, errors.New("server error")
		}
	} else {
		incr += rs.Int64()
		if incr < 1 {
			return 0, errors.New("BadArgument::INCR")
		}
	}

	if value_enc, err := skv.ValueEncode(incr, nil); err == nil {
		if rs := cn.rawPut(key_enc, value_enc, 0); rs.OK() {
			return uint64(incr), nil
		}
	}

	return 0, errors.New("server error")
}
