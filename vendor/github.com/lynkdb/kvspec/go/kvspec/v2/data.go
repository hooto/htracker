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

package kvspec

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/golang/protobuf/proto"
)

const (
	dataValueVersionBytes    uint8 = 0
	dataValueVersionProtobuf uint8 = 1
	dataValueVersionJson     uint8 = 2
)

type DataValue []byte

type DataValueCodec interface {
	Encode(object interface{}) ([]byte, error)
	Decode(value []byte, object interface{}) error
}

type dataValueCodecDefault struct{}

var (
	dataValueCodecStd      = &dataValueCodecDefault{}
	DataValueProtobufCodec = &dataValueProtobufCodec{}
)

func (it *ObjectData) Valid() error {
	if len(it.Value) < 2 || it.Value[0] > dataValueVersionJson {
		// TODO
		// return errors.New("Invalid Value Version")
	}
	if it.Check > 0 {
		if sc := bytesCrc32Checksum(it.Value); sc != it.Check {
			return errors.New("Invalid Value Sum Check")
		}
	} else {
		it.Check = bytesCrc32Checksum(it.Value)
	}
	return nil
}

func (v DataValue) Bytes() []byte {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		return v[1:]
	}
	return nil
}

func (v DataValue) String() string {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		return string(v[1:])
	}
	return ""
}

func (v DataValue) Int() int {
	return int(v.Int64())
}

func (v DataValue) Int8() int8 {
	return int8(v.Int64())
}

func (v DataValue) Int16() int16 {
	return int16(v.Int64())
}

func (v DataValue) Int32() int32 {
	return int32(v.Int64())
}

func (v DataValue) Int64() int64 {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		if i64, err := strconv.ParseInt(string(v[1:]), 10, 64); err == nil {
			return i64
		}
	}
	return 0
}

func (v DataValue) Uint() uint {
	return uint(v.Uint64())
}

func (v DataValue) Uint8() uint8 {
	return uint8(v.Uint64())
}

func (v DataValue) Uint16() uint16 {
	return uint16(v.Uint64())
}

func (v DataValue) Uint32() uint32 {
	return uint32(v.Uint64())
}

func (v DataValue) Uint64() uint64 {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		if u64, err := strconv.ParseUint(string(v[1:]), 10, 64); err == nil {
			return u64
		}
	}

	return 0
}

func (v DataValue) Bool() bool {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		if b, err := strconv.ParseBool(string(v[1:])); err == nil {
			return b
		}
	}
	return false
}

func (v DataValue) Float64() float64 {
	if len(v) > 1 && v[0] == dataValueVersionBytes {
		if f64, err := strconv.ParseFloat(string(v[1:]), 64); err == nil {
			return f64
		}
	}
	return 0
}

func (v DataValue) Decode(object interface{}, opts ...interface{}) error {

	var codec DataValueCodec
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch opt.(type) {
		case DataValueCodec:
			codec = opt.(DataValueCodec)
		}
	}

	if codec == nil {
		codec = dataValueCodecStd
	}

	return codec.Decode(v, object)
}

func (it *dataValueCodecDefault) Decode(value []byte, object interface{}) error {

	if len(value) > 1 && object != nil {

		switch value[0] {

		case dataValueVersionProtobuf:
			if obj, ok := object.(proto.Message); ok {
				if err := StdProto.Decode(value[1:], obj); err != nil {
					return errors.New("Invalid Value/ProtoBuf " + err.Error())
				}
				return nil
			}

		case dataValueVersionJson:
			return json.Unmarshal(value[1:], object)

		default:
			if value[1] == '{' || value[1] == '[' {
				return json.Unmarshal(value[1:], object)
			}
		}
	}

	return errors.New("Invalid Value/Version")
}

func (it *dataValueCodecDefault) Encode(value interface{}) ([]byte, error) {

	var valueEnc []byte

	switch value.(type) {

	case []byte:
		valueEnc = append([]byte{dataValueVersionBytes}, value.([]byte)...)

	case string:
		valueEnc = append([]byte{dataValueVersionBytes}, []byte(value.(string))...)

	//
	case uint:
		valueEnc = valueEncodeUint(uint64(value.(uint)))

	case uint8:
		valueEnc = valueEncodeUint(uint64(value.(uint8)))

	case uint16:
		valueEnc = valueEncodeUint(uint64(value.(uint16)))

	case uint32:
		valueEnc = valueEncodeUint(uint64(value.(uint32)))

	case uint64:
		valueEnc = valueEncodeUint(value.(uint64))

	//
	case int:
		valueEnc = valueEncodeInt(int64(value.(int)))

	case int8:
		valueEnc = valueEncodeInt(int64(value.(int8)))

	case int16:
		valueEnc = valueEncodeInt(int64(value.(int16)))

	case int32:
		valueEnc = valueEncodeInt(int64(value.(int32)))

	case int64:
		valueEnc = valueEncodeInt(value.(int64))

	//
	// case proto.Message:
	// 	if bs, err := proto.Marshal(value.(proto.Message)); err != nil {
	// 		return nil, errors.New("Invalid ProtoBuf " + err.Error())
	// 	} else {
	// 		valueEnc = append([]byte{dataValueVersionProtobuf}, bs...)
	// 	}

	//
	case map[string]interface{}, struct{}, interface{}:
		if bsJson, err := json.Marshal(value); err != nil {
			return nil, errors.New("Invalid JSON")
		} else {
			valueEnc = append([]byte{dataValueVersionJson}, bsJson...)
		}

	default:
		return nil, errors.New("Invalid")
	}

	return valueEnc, nil
}

func valueEncodeUint(num uint64) []byte {
	return append([]byte{dataValueVersionBytes}, []byte(strconv.FormatUint(num, 10))...)
}

func valueEncodeInt(num int64) []byte {
	return append([]byte{dataValueVersionBytes}, []byte(strconv.FormatInt(num, 10))...)
}

type dataValueProtobufCodec struct{}

func (it *dataValueProtobufCodec) Encode(object interface{}) ([]byte, error) {
	obj, ok := object.(proto.Message)
	if !ok {
		return nil, errors.New("Invalid Protobuf Value")
	}
	return proto.Marshal(obj)
}

func (it *dataValueProtobufCodec) Decode(value []byte, object interface{}) error {
	obj, ok := object.(proto.Message)
	if !ok {
		return errors.New("Invalid Protobuf Message define")
	}
	return proto.Unmarshal(value, obj)
}
