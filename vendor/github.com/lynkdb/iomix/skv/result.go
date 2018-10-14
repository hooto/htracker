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

package skv

import (
	"hash/crc32"

	"github.com/golang/protobuf/proto"
	"github.com/lessos/lessgo/types"
)

const (
	_ uint8 = iota
	ResultOK
	ResultError
	ResultNotFound
	ResultBadArgument
	ResultNoAuth
	ResultServerError
	ResultNetError
	ResultTimeout
	ResultUnknown
)

type Result interface {
	Status() uint8
	OK() bool
	NotFound() bool
	ErrorString() string
	Bytes() []byte
	Bytex() types.Bytex
	String() string
	Int() int
	Int8() int8
	Int16() int16
	Int32() int32
	Int64() int64
	Uint() uint
	Uint8() uint8
	Uint16() uint16
	Uint32() uint32
	Uint64() uint64
	Bool() bool
	Float32() float32
	Float64() float64
	ListLen() int
	List() []Result
	KvLen() int
	KvEach(fn func(entry *ResultEntry) int) int
	KvEntry(i int) (entry *ResultEntry)
	KvList() []*ResultEntry
	KvSize() int
	KvPairs() []Result
	KvKey() []byte
	Decode(obj interface{}) error
	Meta() *KvMeta
}

//
type ResultEntry struct {
	Key   []byte
	Value []byte
}

func (re *ResultEntry) ValueSize() int64 {
	if bs := re.Bytes(); len(bs) > 1 {
		return int64(len(bs) - 1)
	}
	return 0
}

func (re *ResultEntry) Crc32() uint32 {
	if bs := re.Bytes(); len(bs) > 1 {
		return crc32.ChecksumIEEE(bs[1:])
	}
	return 0
}

func (re *ResultEntry) Bytes() []byte {
	if len(re.Value) > 1 && (re.Value[0] == value_ns_prog || re.Value[0] == value_ns_protobuf) {
		meta_len := int(re.Value[1])
		if len(re.Value) > (meta_len + 2) {
			return re.Value[(meta_len + 2):]
		}
	}
	return re.Value
}

func (re *ResultEntry) Bytex() types.Bytex {
	if bs := re.Bytes(); len(bs) > 1 && bs[0] == value_ns_bytes {
		return types.Bytex(bs[1:])
	}
	return types.Bytex{}
}

func (re *ResultEntry) String() string {
	return re.Bytex().String()
}

func (re *ResultEntry) Decode(obj interface{}) error {
	return ValueDecode(re.Bytes(), obj)
}

func (re *ResultEntry) Int64() int64 {
	return KvValueBytes(re.Bytes()).Int64()
}

func (re *ResultEntry) Uint64() uint64 {
	return KvValueBytes(re.Bytes()).Uint64()
}

func (re *ResultEntry) Meta() *KvMeta {
	if len(re.Value) > 1 && re.Value[0] == value_ns_prog {
		meta_len := int(re.Value[1])
		if (meta_len + 2) <= len(re.Value) {
			var meta KvMeta
			if err := proto.Unmarshal(re.Value[2:(2+meta_len)], &meta); err == nil {
				return &meta
			}
		}
	}

	return nil
}
