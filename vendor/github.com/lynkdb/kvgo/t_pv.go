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
	"path/filepath"
	"strings"

	"github.com/lynkdb/iomix/skv"
)

var (
	t_pv_event_handler skv.PathEventHandler
)

func (cn *Conn) PvNew(path string, value interface{}, opts *skv.KvProgWriteOptions) skv.Result {
	return cn.KvProgNew(pv_path_parser(path), skv.NewKvEntry(value), opts)
}

func (cn *Conn) PvDel(path string, opts *skv.KvProgWriteOptions) skv.Result {
	return cn.KvProgDel(pv_path_parser(path), opts)
}

func (cn *Conn) PvPut(path string, value interface{}, opts *skv.KvProgWriteOptions) skv.Result {
	return cn.KvProgPut(pv_path_parser(path), skv.NewKvEntry(value), opts)
}

func (cn *Conn) PvGet(path string) skv.Result {
	return cn.KvProgGet(pv_path_parser(path))
}

func (cn *Conn) PvScan(fold, offset, cutset string, limit int) skv.Result {
	return cn.KvProgScan(pv_path_parser_add(fold, offset), pv_path_parser_add(fold, cutset), limit)
}

func (cn *Conn) PvRevScan(fold, offset, cutset string, limit int) skv.Result {
	return cn.KvProgRevScan(pv_path_parser_add(fold, offset), pv_path_parser_add(fold, cutset), limit)
}

func pv_path_parser(path string) skv.KvProgKey {
	values := strings.Split(pv_path_clean(path), "/")
	k := skv.KvProgKey{}
	for _, value := range values {
		k.Append(value)
	}
	return k
}

func pv_path_parser_add(path, add string) skv.KvProgKey {
	k := pv_path_parser(path)
	k.Append(add)
	return k
}

func pv_path_clean(path string) string {
	return strings.Trim(strings.Trim(filepath.Clean(path), "/"), ".")
}
