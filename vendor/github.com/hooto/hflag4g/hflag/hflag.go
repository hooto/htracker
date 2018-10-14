// Copyright 2016 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package hflag

import (
	"os"
	"strings"

	"github.com/lessos/lessgo/types"
)

const (
	Version = "0.9.0"
)

var (
	args_kv = map[string]types.Bytex{}
)

func init() {

	if len(os.Args) < 2 {
		return
	}

	for i, k := range os.Args {

		if k[0] != '-' || len(k) < 2 {
			continue
		}

		k = strings.Trim(k, "-")

		if len(os.Args) <= (i+1) || os.Args[i+1][0] == '-' {
			args_kv[k] = types.Bytex([]byte(""))
			continue
		}

		v := os.Args[i+1]

		args_kv[k] = types.Bytex([]byte(v))
	}
}

func ValueOK(key string) (types.Bytex, bool) {

	if v, ok := args_kv[key]; ok {
		return v, ok
	}

	return nil, false
}

func Value(key string) types.Bytex {

	if v, ok := ValueOK(key); ok {
		return v
	}

	return types.Bytex{}
}

func Each(fn func(key, val string)) {
	for k, v := range args_kv {
		fn(k, v.String())
	}
}
