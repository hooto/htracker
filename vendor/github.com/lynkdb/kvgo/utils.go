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
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	mrand "math/rand"
)

func keyExpireEncode(ns byte, expired uint64, key []byte) []byte {
	return append(append([]byte{ns}, uint64ToBytes(expired)...), key...)
}

func keyEncode(ns byte, key []byte) []byte {
	return append([]byte{ns}, key...)
}

func bytesClone(src []byte) []byte {

	dst := make([]byte, len(src))
	copy(dst, src)

	return dst
}

func uint64ToBytes(v uint64) []byte {

	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, v)

	return bs
}

func randHexString(length int) string {

	length = length / 2
	if length < 1 {
		length = 1
	}
	if n := length % 2; n > 0 {
		length += n
	}

	bs := make([]byte, length)
	if _, err := rand.Read(bs); err != nil {
		for i := range bs {
			bs[i] = uint8(mrand.Intn(256))
		}
	}

	return hex.EncodeToString(bs)
}
