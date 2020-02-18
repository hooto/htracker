// Copyright 2019 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package iamauth

//go:generate protoc --go_out=plugins=grpc:. auth.proto
//go:generate protobuf_slice "*.proto"

import (
	"encoding/base64"
	"strings"

	"github.com/lessos/lessgo/crypto/idhash"
)

var (
	authKeyDefault = &AuthKey{
		AccessKey: "00000000",
		SecretKey: idhash.RandBase64String(40),
	}
	base64Std = base64.StdEncoding.WithPadding(base64.NoPadding)
	base64Url = base64.URLEncoding.WithPadding(base64.NoPadding)
)

func base64nopad(s string) string {
	if i := strings.IndexByte(s, '='); i > 0 {
		return s[:i]
	}
	return s
}

func base64pad(s string) string {
	if n := len(s) % 4; n > 0 {
		s += strings.Repeat("=", 4-n)
	}
	return s
}
