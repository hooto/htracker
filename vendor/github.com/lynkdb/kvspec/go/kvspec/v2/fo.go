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
	"io"
	"path/filepath"
	"strings"
)

const (
	FileObjectBlockAttrVersion1   uint64 = 1 << 1
	FileObjectBlockAttrBlockSize2 uint64 = 1 << 2
	FileObjectBlockAttrBlockSize4 uint64 = 1 << 4
	FileObjectBlockSize2          int64  = 2 * 1024 * 1024
	FileObjectBlockSize4          int64  = 4 * 1024 * 1024
)

type ClientFileObjectConnector interface {
	Close() error
	FoFilePut(srcPath string, dstPath string) *ObjectResult
	FoFileOpen(path string) (io.ReadSeeker, error)
}

type FileObjectBlock struct {
	Num       uint32 `json:"num,omitempty"`
	Path      string `json:"path,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Attrs     uint64 `json:"attrs,omitempty"`
	Data      []byte `json:"data,omitempty"`
	CommitKey string `json:"commit_key,omitempty"`
}

func FileObjectPathEncode(path string) string {

	s := strings.Trim(filepath.Clean(path), ".")
	if len(s) == 0 {
		return "/"
	}

	if len(s) > 1 && path[len(path)-1] == '/' {
		s += "/"
	}

	if s[0] != '/' {
		return ("/" + s)
	}

	return s
}

func NewFileObjectBlock(path string, size int64,
	num uint32, data []byte) FileObjectBlock {
	return FileObjectBlock{
		Path:  FileObjectPathEncode(path),
		Size:  size,
		Num:   num,
		Attrs: FileObjectBlockAttrVersion1 | FileObjectBlockAttrBlockSize2,
		Data:  data,
	}
}

func (it *FileObjectBlock) AttrAllow(v uint64) bool {
	return AttrAllow(it.Attrs, v)
}

func (it *FileObjectBlock) BlockSize() int64 {
	if it.AttrAllow(FileObjectBlockAttrBlockSize2) {
		return FileObjectBlockSize2
	} else if it.AttrAllow(FileObjectBlockAttrBlockSize4) {
		return FileObjectBlockSize4
	}
	return 0
}

func (it *FileObjectBlock) Valid() bool {

	if len(it.Path) < 1 || it.Size < 1 || len(it.Data) < 1 {
		return false
	}

	blockSize := it.BlockSize()
	if blockSize == 0 {
		return false
	}

	numMax := uint32(it.Size / blockSize)
	if (it.Size % blockSize) == 0 {
		numMax -= 1
	}
	if it.Num > numMax {
		return false
	}
	if it.Num < numMax {
		if int64(len(it.Data)) != blockSize {
			return false
		}
	} else {
		if int64(len(it.Data)) != (it.Size % blockSize) {
			return false
		}
	}
	return true
}
