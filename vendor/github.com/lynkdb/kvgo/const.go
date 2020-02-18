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

const (
	Version = "0.2.0"
)

const (
	nsKeySys  uint8 = 16
	nsKeyMeta uint8 = 17
	nsKeyData uint8 = 18
	nsKeyLog  uint8 = 19
	nsKeyTtl  uint8 = 20
)

const (
	ldbNotFound            = "leveldb: not found"
	objAcceptTTL           = uint64(3000)
	workerLocalExpireSleep = 200e6
	workerLocalExpireLimit = 200
	workerClusterSleep     = 1e9
)

var (
	keySysInstanceId = append([]byte{nsKeySys}, []byte("inst:id")...)
	keySysLogCutset  = append([]byte{nsKeySys}, []byte("log:cutset")...)
)

func keySysLogAsync(hostport string) []byte {
	return append([]byte{nsKeySys}, []byte("log:async:"+hostport)...)
}

func keySysIncrCutset(ns string) []byte {
	return append([]byte{nsKeySys}, []byte("incr:cutset:"+ns)...)
}
