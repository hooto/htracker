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

//go:generate protoc --go_out=plugins=grpc:. skv.proto

const (
	ScanLimitMax = 100000
)

type Connector interface {
	KvConnector
	KvProgConnector
	PvConnector
	Close() error
}

const (
	PathEventCreated uint8 = 1
	PathEventUpdated uint8 = 2
	PathEventDeleted uint8 = 3
)

type PathEventInterface interface {
	Path() string
	Action() uint8
	Version() uint64
}

type PathEventHandler func(ev PathEventInterface)

// Path-Value APIs
type PvConnector interface {
	PvNew(path string, value interface{}, opts *KvProgWriteOptions) Result
	PvDel(path string, opts *KvProgWriteOptions) Result
	PvPut(path string, value interface{}, opts *KvProgWriteOptions) Result
	PvGet(path string) Result
	PvScan(fold, offset, cutset string, limit int) Result
	PvRevScan(fold, offset, cutset string, limit int) Result
}
