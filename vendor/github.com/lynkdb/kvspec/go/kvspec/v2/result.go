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
	"errors"
	"fmt"
)

const (
	_ uint64 = iota
	ResultOK
	ResultError
	ResultNotFound
	ResultClientError
	ResultServerError
	ResultUnAuth
)

func NewObjectResult(status uint64, err error) *ObjectResult {

	rs := &ObjectResult{
		Status: status,
	}

	if err != nil {
		rs.Message = err.Error()
		if status == ResultOK {
			rs.Status = ResultError
		}
	} else if status == 0 {
		rs.Status = ResultOK
	}

	return rs
}

func NewObjectResultOK() *ObjectResult {
	return NewObjectResult(ResultOK, nil)
}

func NewObjectResultNotFound() *ObjectResult {
	return NewObjectResult(ResultNotFound, nil)
}

func NewObjectResultClientError(err error) *ObjectResult {
	return NewObjectResult(ResultClientError, err)
}

func NewObjectResultAccessDenied(args ...interface{}) *ObjectResult {
	msg := "Access Denied"
	if len(args) > 0 {
		msg += " : " + fmt.Sprint(args...)
	}
	return &ObjectResult{
		Status:  ResultClientError,
		Message: msg,
	}
}

func NewObjectResultServerError(err error) *ObjectResult {
	return NewObjectResult(ResultServerError, err)
}

func (it *ObjectResult) StatusMessage(status uint64, msg string) *ObjectResult {
	it.Status, it.Message = status, msg
	return it
}

func (it *ObjectResult) OK() bool {
	return it.Status == ResultOK
}

func (it *ObjectResult) NotFound() bool {
	return it.Status == ResultNotFound
}

func (it *ObjectResult) Error() error {
	return errors.New(it.Message)
}

func (it *ObjectResult) DataValue() DataValue {
	if len(it.Items) > 0 && it.Items[0].Data != nil {
		return DataValue(it.Items[0].Data.Value)
	}
	return DataValue{}
}

func (it *ObjectResult) Decode(obj interface{}, opts ...interface{}) error {
	return it.DataValue().Decode(&obj, opts...)
}
