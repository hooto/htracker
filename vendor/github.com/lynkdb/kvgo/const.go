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

//go:generate protoc --proto_path=./ --go_out=./ --go_opt=paths=source_relative --go-grpc_out=. kvgo.proto

import (
	"github.com/hooto/hauth/go/hauth/v1"
	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

const (
	Version = "0.3.1"
)

const (
	nsKeySys  uint8 = 16
	nsKeyMeta uint8 = 17
	nsKeyData uint8 = 18
	nsKeyLog  uint8 = 19
	nsKeyTtl  uint8 = 20
)

const (
	grpcMsgByteMax = 12 * int(kv2.MiB)
)

const (
	ldbNotFound                = "leveldb: not found"
	objAcceptTTL               = uint64(3000)
	workerLocalExpireSleep     = 200e6
	workerLocalExpireLimit     = 200
	workerLogRangeWaitTimeMax  = int64(10e3)
	workerLogRangeWaitSleep    = int64(200)
	workerReplicaLogAsyncSleep = 1e9
	workerTableRefreshTime     = int64(600)
)

var (
	keySysInstanceId = append([]byte{nsKeySys}, []byte("inst:id")...)
	keySysLogCutset  = append([]byte{nsKeySys}, []byte("log:cutset")...)
)

func keySysLogAsync(hostAddr, tableName string) []byte {
	if hostAddr == "" {
		return append([]byte{nsKeySys}, []byte("log:async:")...)
	}
	return append([]byte{nsKeySys}, []byte("log:async:"+hostAddr+":"+tableName)...)
}

func keySysIncrCutset(ns string) []byte {
	return append([]byte{nsKeySys}, []byte("incr:cutset:"+ns)...)
}

const (
	sysTableName   = "sys"
	sysTableIncrNS = "sys_table_id"
)

func nsSysTable(name string) []byte {
	return []byte("sys:table:" + name)
}

func nsSysTableStatus(name string) []byte {
	return []byte("sys:table-status:" + name)
}

func nsSysAuthRole(name string) []byte {
	return []byte("sys:role:" + name)
}

func nsSysAccessKey(id string) []byte {
	return []byte("sys:ak:" + id)
}

var (
	authPermSysAll     = "sys/all"
	authPermTableList  = "table/list"
	authPermTableRead  = "table/read"
	authPermTableWrite = "table/write"
	AuthScopeTable     = "kvgo/table"
	defaultScopes      = []string{
		AuthScopeTable,
	}
	defaultRoles = []*hauth.Role{
		{
			Name:  "sa",
			Title: "System Administrator",
			Permissions: []string{
				authPermSysAll,
				authPermTableList,
				authPermTableRead,
				authPermTableWrite,
			},
		},
		{
			Name:  "client",
			Title: "General Client",
			Permissions: []string{
				authPermTableList,
				authPermTableRead,
				authPermTableWrite,
			},
		},
	}
)
