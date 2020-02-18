// Copyright 2019 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package sko // import "github.com/lynkdb/iomix/sko"

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"

	"github.com/golang/protobuf/proto"
)

func AttrAllow(base, comp uint64) bool {
	return (comp & base) == comp
}

func AttrAppend(base, comp uint64) uint64 {
	return base | comp
}

func AttrRemove(base, comp uint64) uint64 {
	return (base | comp) - (comp)
}

func NewObjectItem(key []byte) *ObjectItem {
	return &ObjectItem{
		Meta: &ObjectMeta{
			Key: key,
		},
	}
}

func (it *ObjectItem) DataValue() DataValue {
	if it.Data != nil {
		return DataValue(it.Data.Value)
	}
	return DataValue{}
}

func (it *ObjectItem) Decode(obj interface{}) error {
	return it.DataValue().Decode(&obj, nil)
}

func ObjectMetaDecode(bs []byte) (*ObjectMeta, error) {

	if len(bs) > 2 &&
		bs[0] == objectRawBytesVersion1 &&
		bs[1] > 0 && (int(bs[1])+2) <= len(bs) {
		var meta ObjectMeta
		if err := protobufDecode(bs[2:(int(bs[1])+2)], &meta); err == nil {
			return &meta, nil
		} else {
			return nil, err
		}
	}
	return nil, errors.New("Invalid Meta/Bytes")
}

func objectMetaKeyValid(key []byte) bool {

	if len(key) < objectMetaKeyLenMin ||
		len(key) > objectMetaKeyLenMax {
		return false
	}

	for _, v := range key {

		if (v >= 'a' && v <= 'z') ||
			(v >= 'A' && v <= 'Z') ||
			(v >= '0' && v <= '9') ||
			(v == ':') || (v == '-') || (v == '_') || (v == '/') || (v == '.') {
			continue
		}

		return false
	}

	return true
}

func ObjectItemDecode(bs []byte) (*ObjectItem, error) {

	meta, err := ObjectMetaDecode(bs)
	if err != nil {
		return nil, err
	}

	offset := int(bs[1]) + 2
	if offset >= len(bs) {
		return nil, errors.New("Invalid Data/Bytes")
	}

	var data ObjectData
	if err := protobufDecode(bs[offset:], &data); err != nil {
		return nil, err
	}

	return &ObjectItem{
		Meta: meta,
		Data: &data,
	}, nil
}

func bytesCrc32Checksum(bs []byte) uint64 {
	sumCheck := crc32.ChecksumIEEE(bs)
	if sumCheck == 0 {
		sumCheck = 1
	}
	return uint64(sumCheck)
}

func protobufEncode(obj proto.Message) ([]byte, error) {
	return proto.Marshal(obj)
}

func protobufDecode(bs []byte, obj proto.Message) error {
	return proto.Unmarshal(bs, obj)
}

func Uint32ToHexString(v uint32) string {
	return hex.EncodeToString(Uint32ToBytes(v))
}

func Uint64ToHexString(v uint64) string {
	return hex.EncodeToString(Uint64ToBytes(v))
}

func Uint32ToBytes(v uint32) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return bs
}

func Uint64ToBytes(v uint64) []byte {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)
	return bs
}
