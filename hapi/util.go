// Copyright 2018 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package hapi

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/encoding/json"
)

func ObjPrint(name string, v interface{}) {
	js, _ := json.Encode(v, "  ")
	fmt.Println(name, string(js))
}

func Uint32ToBytes(v uint32) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return bs
}

func Uint32ToHexString(v uint32) string {
	return hex.EncodeToString(Uint32ToBytes(v))
}

func ObjectId(u32 uint32, n int) string {
	if n < 2 {
		n = 2
	} else if n > 24 {
		n = 24
	} else if (n % 2) == 1 {
		n += 1
	}
	return Uint32ToHexString(u32) + idhash.RandHexString(n)
}
