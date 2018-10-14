// Copyright 2015 Eryx <evorui аt gmail dοt com>, All rights reserved.
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
	"hash/crc32"

	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"
)

type Result struct {
	status uint8
	Data   []byte
	Cap    int
	Items  []*Result
	key    []byte
}

func newResult(status uint8, err error) *Result {

	r := &Result{
		status: status,
	}

	if err == nil {
		if r.status == 0 {
			r.status = skv.ResultOK
		}
	} else {
		if r.status == 0 {
			r.status = skv.ResultError
		}
		r.Data = []byte(err.Error())
	}

	return r
}

func newResultData(data []byte) *Result {
	return &Result{
		status: skv.ResultOK,
		Data:   data,
	}
}

func newResultError() *Result {
	return newResult(skv.ResultError, nil)
}

func newResultNotFound() *Result {
	return newResult(skv.ResultNotFound, nil)
}

func newResultBadArgument() *Result {
	return newResult(skv.ResultBadArgument, nil)
}

func newResultNetError() *Result {
	return newResult(skv.ResultNetError, nil)
}

func (rs *Result) Status() uint8 {
	return rs.status
}

func (rs *Result) OK() bool {
	return rs.status == skv.ResultOK
}

func (rs *Result) NotFound() bool {
	return rs.status == skv.ResultNotFound
}

func (rs *Result) ErrorString() string {
	return string(rs.Data)
}

func (rs *Result) ValueSize() int64 {
	if bs := rs.Bytes(); len(bs) > 1 {
		return int64(len(bs) - 1)
	}
	return 0
}

func (rs *Result) Bytes() []byte {
	return skv.KvValueBytes(rs.Data).Bytes()
}

func (rs *Result) Bytex() types.Bytex {
	return skv.KvValueBytes(rs.Bytes()).Bytex()
}

func (rs *Result) String() string {
	return skv.KvValueBytes(rs.Bytes()).String()
}

func (rs *Result) Crc32() uint32 {
	if bs := rs.Bytes(); len(bs) > 1 {
		return crc32.ChecksumIEEE(bs[1:])
	}
	return 0
}

func (rs *Result) Int() int {
	return skv.KvValueBytes(rs.Bytes()).Int()
}

func (rs *Result) Int8() int8 {
	return skv.KvValueBytes(rs.Bytes()).Int8()
}

func (rs *Result) Int16() int16 {
	return skv.KvValueBytes(rs.Bytes()).Int16()
}

func (rs *Result) Int32() int32 {
	return skv.KvValueBytes(rs.Bytes()).Int32()
}

func (rs *Result) Int64() int64 {
	return skv.KvValueBytes(rs.Bytes()).Int64()
}

func (rs *Result) Uint() uint {
	return skv.KvValueBytes(rs.Bytes()).Uint()
}

func (rs *Result) Uint8() uint8 {
	return skv.KvValueBytes(rs.Bytes()).Uint8()
}

func (rs *Result) Uint16() uint16 {
	return skv.KvValueBytes(rs.Bytes()).Uint16()
}

func (rs *Result) Uint32() uint32 {
	return skv.KvValueBytes(rs.Bytes()).Uint32()
}

func (rs *Result) Uint64() uint64 {
	return skv.KvValueBytes(rs.Bytes()).Uint64()
}

func (rs *Result) Bool() bool {
	return skv.KvValueBytes(rs.Bytes()).Bool()
}

func (rs *Result) Float32() float32 {
	return float32(rs.Float64())
}

func (rs *Result) Float64() float64 {
	return skv.KvValueBytes(rs.Bytes()).Float64()
}

func (rs *Result) KvLen() int {
	return len(rs.Items) / 2
}

func (rs *Result) KvEach(fn func(entry *skv.ResultEntry) int) int {
	for i := 1; i < len(rs.Items); i += 2 {
		if rn := fn(&skv.ResultEntry{
			Key:   rs.Items[i-1].Data,
			Value: rs.Items[i].Data,
		}); rn != 0 {
			return (i + 1) / 2
		}
	}
	return rs.KvLen()
}

func (rs *Result) KvEntry(i int) *skv.ResultEntry {
	if i < 0 {
		i = 0
	} else {
		i = i * 2
	}
	if i+1 < len(rs.Items) {
		return &skv.ResultEntry{
			Key:   rs.Items[i].Data,
			Value: rs.Items[i+1].Data,
		}
	}
	return nil
}

func (rs *Result) KvList() []*skv.ResultEntry {
	ls := []*skv.ResultEntry{}
	for i := 1; i < len(rs.Items); i += 2 {
		ls = append(ls, &skv.ResultEntry{
			Key:   rs.Items[i-1].Data,
			Value: rs.Items[i].Data,
		})
	}
	return ls
}

func (rs *Result) Decode(obj interface{}) error {
	return skv.ValueDecode(rs.Bytes(), obj)
}

func (rs *Result) ListLen() int {
	return len(rs.Items)
}

func (rs *Result) List() []skv.Result {
	ls := []skv.Result{}
	for _, v := range rs.Items {
		ls = append(ls, v)
	}
	return ls
}

func (rs *Result) Meta() *skv.KvMeta {
	return skv.KvValueBytes(rs.Data).Meta()
}

func (rs *Result) KvSize() int {
	return len(rs.Items) / 2
}

func (rs *Result) KvPairs() []skv.Result {
	ls := []skv.Result{}
	for i := 1; i < len(rs.Items); i += 2 {
		ls = append(ls, &Result{
			key:  rs.Items[i-1].Data,
			Data: rs.Items[i].Data,
		})
	}
	return ls
}

func (rs *Result) KvKey() []byte {
	return rs.key
}
