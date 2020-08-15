// Copyright 2019 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package kvspec

import (
	"fmt"
)

func NewBatchRequest(tableName string) *BatchRequest {
	return &BatchRequest{
		TableName: tableName,
	}
}

func (it *BatchRequest) TableNameSet(name string) *BatchRequest {
	it.TableName = name
	return it
}

func (it *BatchRequest) KeyQuery(keys ...[]byte) *ObjectReader {
	r := NewObjectReader(nil)
	if len(keys) > 0 {
		for _, key := range keys {
			r.KeySet(key)
		}
	}
	it.Items = append(it.Items, &BatchItem{
		Reader: r,
	})
	return r
}

func (it *BatchRequest) KeyRangeQuery(offset, cutset []byte) *ObjectReader {
	r := NewObjectReader(nil).KeyRangeSet(offset, cutset)
	it.Items = append(it.Items, &BatchItem{
		Reader: r,
	})
	return r
}

func (it *BatchRequest) Put(key []byte, value interface{}, opts ...interface{}) *ObjectWriter {
	w := NewObjectWriter(key, value, opts...)
	it.Items = append(it.Items, &BatchItem{
		Writer: w,
	})
	return w
}

func (it *BatchRequest) Create(key []byte, value interface{}, opts ...interface{}) *ObjectWriter {
	w := NewObjectWriter(key, value, opts...).
		ModeCreateSet(true)
	it.Items = append(it.Items, &BatchItem{
		Writer: w,
	})
	return w
}

func (it *BatchRequest) Delete(key []byte) *ObjectWriter {
	w := NewObjectWriter(key, nil).ModeDeleteSet(true)
	it.Items = append(it.Items, &BatchItem{
		Writer: w,
	})
	return w
}

func (it *BatchRequest) NewResult(status uint64, errMessage string) *BatchResult {

	rs := &BatchResult{
		Status:  status,
		Message: errMessage,
	}

	/**
	for i := 0; i < len(it.Items); i++ {
		rs.Items = append(rs.Items, &BatchResult{})
	}
	*/

	return rs
}

func NewBatchResultAccessDenied(args ...interface{}) *BatchResult {
	msg := "Access Denied"
	if len(args) > 0 {
		msg += " : " + fmt.Sprint(args...)
	}
	return &BatchResult{Status: ResultClientError, Message: msg}
}

func (it *BatchResult) OK() bool {
	return it.Status == ResultOK
	/**
	ok := 0
	for _, v := range it.Items {
		if v.Status == ResultOK {
			ok += 1
		}
	}
	return ok == len(it.Items)
	*/
}
