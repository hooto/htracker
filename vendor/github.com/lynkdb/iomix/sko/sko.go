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

//go:generate protoc --go_out=plugins=grpc:. sko.proto
//go:generate protobuf_slice "*.proto"
//--go:generate sed -i "s/json:\"id,omitempty\"/json:\"id\"/g" sko.pb.go

package sko // import "github.com/lynkdb/iomix/sko"

const (
	// 1:version|1:meta-size|meta-bytes|data-bytes
	objectRawBytesVersion1 uint8 = 2
)

const (
	ObjectWriterModePut    uint64 = 1 << 0
	ObjectWriterModeCreate uint64 = 1 << 1
	ObjectWriterModeDelete uint64 = 1 << 2
)

const (
	ObjectReaderModeKey      uint64 = 1 << 0
	ObjectReaderModeKeyRange uint64 = 1 << 1
	ObjectReaderModeLogRange uint64 = 1 << 2
	ObjectReaderModeRevRange uint64 = 1 << 3
)

const (
	ObjectReaderLimitNumMax  int64 = 1000
	ObjectReaderLimitSizeMax int64 = 8 * 1024 * 1024
	ObjectReaderLimitSizeDef int64 = ObjectReaderLimitSizeMax
)

const (
	ObjectMetaAttrDelete uint64 = 1 << 32
)

const (
	objectMetaKeyLenMin int = 4
	objectMetaKeyLenMax int = 128
)

const (
	ObjectClusterNodeMax int = 7
)
