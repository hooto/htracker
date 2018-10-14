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

package skv // import "github.com/lynkdb/iomix/skv"

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	prog_ttl_zero int64 = 1500000000000000000
)

const (
	KvProgKeyEntryUnknown uint32 = 0
	KvProgKeyEntryBytes   uint32 = 1
	KvProgKeyEntryUint    uint32 = 4
	// KvProgKeyEntryIncr    uint32 = 16
)

func NewKvProgKey(values ...interface{}) KvProgKey {
	k := KvProgKey{}
	for _, value := range values {
		k.Append(value)
	}
	return k
}

func newKvProgKeyEntry(value interface{}) (*KvProgKeyEntry, error) {

	set := &KvProgKeyEntry{}

	switch value.(type) {

	case []byte:
		set.Type = KvProgKeyEntryBytes
		if bs := value.([]byte); len(bs) > 0 {
			set.Data = bs
		}

	case string:
		set.Type = KvProgKeyEntryBytes
		if bs := []byte(value.(string)); len(bs) > 0 {
			set.Data = bs
		}

	case uint8:
		set.Type, set.Data = KvProgKeyEntryUint, []byte{value.(uint8)}

	case uint16:
		set.Type, set.Data = KvProgKeyEntryUint, make([]byte, 2)
		binary.BigEndian.PutUint16(set.Data, value.(uint16))

	case uint32:
		set.Type, set.Data = KvProgKeyEntryUint, make([]byte, 4)
		binary.BigEndian.PutUint32(set.Data, value.(uint32))

	case uint64:
		set.Type, set.Data = KvProgKeyEntryUint, make([]byte, 8)
		binary.BigEndian.PutUint64(set.Data, value.(uint64))

	default:
		return nil, errors.New("Invalid Data Type")
	}

	return set, nil
}

func (k *KvProgKey) Append(value interface{}) error {
	if len(k.Items) > 20 {
		return errors.New("too many Items")
	}

	set, err := newKvProgKeyEntry(value)
	if err != nil {
		return err
	}
	k.Items = append(k.Items, set)

	// if len(k_enc) > 0 {
	// 	k_enc = []byte{}
	// }
	// if len(k_fold_meta) > 0 {
	// 	k_fold_meta = []byte{}
	// }

	return nil
}

func (k *KvProgKey) AppendTypeValue(t uint32, value interface{}) error {
	if len(k.Items) > 20 {
		return errors.New("too many Items")
	}

	switch t {
	// case KvProgKeyEntryIncr:
	// 	k.Items = append(k.Items, &KvProgKeyEntry{Type: t})

	case KvProgKeyEntryBytes, KvProgKeyEntryUint:
		return k.Append(value)

	default:
		return errors.New("Invalid Data Type")
	}

	return nil
}

func (k *KvProgKey) Set(idx int, value interface{}) error {

	if idx+1 > len(k.Items) {
		return errors.New("Invalid index")
	}

	set, err := newKvProgKeyEntry(value)
	if err != nil {
		return err
	}

	k.Items[idx] = set

	return nil
}

func (k *KvProgKey) LastEntry() (int, *KvProgKeyEntry) {
	if i := (len(k.Items) - 1); i >= 0 {
		return i, k.Items[i]
	}
	return -1, nil
}

func (k *KvProgKey) Value(i int) []byte {
	if i > 0 && i <= len(k.Items) {
		return k.Items[i-1].Data
	}
	return []byte{}
}

func (k *KvProgKey) Size() int {
	return len(k.Items)
}

func (k *KvProgKey) Valid() bool {
	return len(k.Items) > 0
}

func (k *KvProgKey) FoldLen() int {

	n := len(k.Items)
	if n > 0 {
		for i, v := range k.Items {
			if (i + 1) < len(k.Items) {
				n += len(v.Data)
			}
		}
		n += 1
	}

	return n
}

func (k *KvProgKey) Encode(ns uint8) []byte {
	// if len(k_enc) > 0 {
	// 	return k_enc
	// }
	if len(k.Items) == 0 {
		return []byte{}
	}

	k_enc := []byte{ns, uint8(len(k.Items))}
	for i, v := range k.Items {
		if (i + 1) < len(k.Items) {
			if len(v.Data) > 0 {
				k_enc = append(k_enc, uint8(len(v.Data)))
				k_enc = append(k_enc, v.Data...)
			} else {
				k_enc = append(k_enc, uint8(0))
			}
		} else if len(v.Data) > 0 {
			k_enc = append(k_enc, v.Data...)
		}
	}

	return k_enc
}

func (k *KvProgKey) EncodeFoldMeta(ns uint8) []byte {
	if len(k.Items) == 0 {
		return []byte{}
	}
	// if len(k_fold_meta) > 0 {
	// 	return k_fold_meta
	// }

	k_fold_meta := []byte{ns, uint8(len(k.Items))}
	for i := 0; i < (len(k.Items) - 1); i++ {
		k_fold_meta = append(k_fold_meta, uint8(len(k.Items[i].Data)))
		k_fold_meta = append(k_fold_meta, k.Items[i].Data...)
	}

	return k_fold_meta
}

func (k *KvProgKey) EncodeIndex(ns uint8, idx int) []byte {
	if len(k.Items) == 0 {
		return []byte{}
	}
	if idx < 0 || (idx+1) > len(k.Items) {
		return []byte{}
	}

	enc := []byte{ns, uint8(len(k.Items))}
	for i := 0; i <= idx; i++ {
		enc = append(enc, uint8(len(k.Items[i].Data)))
		enc = append(enc, k.Items[i].Data...)
	}

	return enc
}

func KvProgKeyDecode(bs []byte) *KvProgKey {
	if len(bs) > 2 {
		var (
			k   = &KvProgKey{}
			off = 2
		)
		for i := 0; i < int(bs[1])-1; i++ {
			nlen := int(bs[off])
			if (off + nlen + 1) <= len(bs) {
				k.Items = append(k.Items, &KvProgKeyEntry{
					Data: bs[(off + 1):(off + nlen + 1)],
				})
				off += (nlen + 1)
			} else {
				return nil
			}
		}
		if off < len(bs) {
			k.Items = append(k.Items, &KvProgKeyEntry{Data: bs[off:]})
		}
		return k
	}
	return nil
}

//
func NewKvEntry(value interface{}) KvEntry {
	obj := KvEntry{}
	obj.Set(value)
	return obj
}

func (o *KvEntry) Valid() bool {
	if o.Meta == nil && len(o.Value) < 1 {
		return false
	}
	return true
}

func (o *KvEntry) Encode() []byte {
	// if len(o_enc) > 1 {
	// 	return o_enc
	// }
	o_enc := []byte{value_ns_prog, 0}
	if o.Meta != nil {

		if len(o.Value) > 1 {
			if o.Meta.Sum > 0 {
				o.Meta.Sum = crc32.ChecksumIEEE(o.Value[1:])
			}
			if o.Meta.Size > 0 {
				o.Meta.Size = uint64(len(o.Value) - 1)
			}
		}

		if bs, err := proto.Marshal(o.Meta); err == nil {
			if len(bs) > 0 && len(bs) < 200 {
				o_enc[1] = uint8(len(bs))
				o_enc = append(o_enc, bs...)
			}
		}
	}
	if len(o.Value) > 1 {
		o_enc = append(o_enc, o.Value...)
	}
	return o_enc
}

func (o *KvEntry) Crc32() uint32 {
	if len(o.Value) > 1 {
		return crc32.ChecksumIEEE(o.Value[1:])
	}
	return 0
}

func (o *KvEntry) ValueSize() int64 {
	return int64(len(o.Value) - 1)
}

func (o *KvEntry) Set(value interface{}) error {
	var err error
	if o.Value, err = ValueEncode(value, nil); err == nil {
		// if len(o_enc) > 0 {
		// 	o_enc = []byte{}
		// }
	}
	return err
}

func (o *KvEntry) ValueBytes() KvValueBytes {
	return KvValueBytes(o.Value)
}

func (o *KvEntry) KvMeta() *KvMeta {
	if o.Meta == nil {
		o.Meta = &KvMeta{}
	}
	return o.Meta
}

type KvProgKeyValue struct {
	Key KvProgKey
	Val KvEntry
}

const (
	KvProgOpMetaSum  uint64 = 1 << 1
	KvProgOpMetaSize uint64 = 1 << 2
	KvProgOpCreate   uint64 = 1 << 13
	// KvProgOpForce     uint64 = 1 << 14
	KvProgOpFoldMeta uint64 = 1 << 15
	// KvProgOpLogEnable uint64 = 1 << 16
)

func (o *KvProgWriteOptions) OpSet(v uint64) *KvProgWriteOptions {
	o.Actions = (o.Actions | v)
	return o
}

func (o *KvProgWriteOptions) OpAllow(v uint64) bool {
	return (v & o.Actions) == v
}

func (m *KvMeta) Encode() []byte {
	if bs, err := proto.Marshal(m); err == nil {
		return append([]byte{value_ns_prog, uint8(len(bs))}, bs...)
	}
	return []byte{}
}

func (m *KvMeta) Timeout() bool {
	if m.Expired > 0 && m.Expired <= uint64(time.Now().UTC().UnixNano()) {
		return true
	}
	return false
}

//
// KvProgrammable Key/Value
type KvProgConnector interface {
	KvProgNew(k KvProgKey, v KvEntry, opts *KvProgWriteOptions) Result
	KvProgPut(k KvProgKey, v KvEntry, opts *KvProgWriteOptions) Result
	KvProgGet(k KvProgKey) Result
	KvProgDel(k KvProgKey, opts *KvProgWriteOptions) Result
	KvProgScan(offset, cutset KvProgKey, limit int) Result
	KvProgRevScan(offset, cutset KvProgKey, limit int) Result
}
