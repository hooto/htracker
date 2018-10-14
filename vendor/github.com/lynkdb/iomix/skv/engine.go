// Copyright 2015 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package skv

type KvEngine interface {
	KvNew(key []byte, value interface{}, opts *KvWriteOptions) Result
	KvPut(key []byte, value interface{}, opts *KvWriteOptions) Result
	KvGet(key []byte) Result
	KvDel(key ...[]byte) Result
	KvScan(offset, cutset []byte, limit int) Result
	KvRevScan(offset, cutset []byte, limit int) Result
	KvIncr(key []byte, increment int64) Result
	KvIncrf(key []byte, increment float64) Result
	KvBatch(batch *KvEngineBatch, opts *KvWriteOptions) error
	Close() error
}

type KvEngineBatch struct {
	Cmds []KvEngineBatchCommand
}

const (
	KvEngineBatchDel uint8 = 1
	KvEngineBatchPut uint8 = 2
)

type KvEngineBatchCommand struct {
	Type  uint8
	Key   []byte
	Value []byte
}

func (it *KvEngineBatch) Del(key []byte) {
	it.Cmds = append(it.Cmds, KvEngineBatchCommand{
		Type: KvEngineBatchDel,
		Key:  key,
	})
}

func (it *KvEngineBatch) Put(key []byte, value []byte) {
	it.Cmds = append(it.Cmds, KvEngineBatchCommand{
		Type:  KvEngineBatchPut,
		Key:   key,
		Value: value,
	})
}
