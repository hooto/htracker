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
	"time"
)

// Key-Value types
type KvWriteOptions struct {
	Ttl       int64     // time to live in milliseconds
	ExpireAt  time.Time // UTC time
	LogEnable bool
	Encoder   KvValueEncoder
}

// Key-Value APIs
type KvConnector interface {
	KvNew(key []byte, value interface{}, opts *KvWriteOptions) Result
	KvDel(key ...[]byte) Result
	KvPut(key []byte, value interface{}, opts *KvWriteOptions) Result
	KvGet(key []byte) Result
	KvScan(offset, cutset []byte, limit int) Result
	KvRevScan(offset, cutset []byte, limit int) Result
	KvIncr(key []byte, increment int64) Result
}
