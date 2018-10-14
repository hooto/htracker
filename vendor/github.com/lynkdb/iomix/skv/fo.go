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
	"io"
	"path/filepath"
	"strings"
)

const (
	FileObjectEntryAttrVersion1   uint64 = 1 << 1
	FileObjectEntryAttrIsDir      uint64 = 1 << 2
	FileObjectEntryAttrBlockSize4 uint64 = 1 << 4
	FileObjectEntryAttrCommiting  uint64 = 1 << 22
	FileObjectBlockSize4          uint64 = 4 * 1024 * 1024
)

type FileObjectConnector interface {
	FoMpInit(sets FileObjectEntryInit) Result
	FoMpPut(sets FileObjectEntryBlock) Result
	FoMpGet(sets FileObjectEntryBlock) Result
	FoGet(key string) Result
	FoScan(offset, cutset string, limit int) Result
	FoRevScan(offset, cutset string, limit int) Result
	FoFilePut(src_path string, dst_path string) Result
	FoFileOpen(path string) (io.ReadSeeker, error)
}

func (it *FileObjectEntryMeta) AttrAllow(v uint64) bool {
	return ((v & it.Attrs) == v)
}

func NewFileObjectEntryInit(path string, size uint64) FileObjectEntryInit {
	return FileObjectEntryInit{
		Path:  FileObjectPathEncode(path),
		Size:  size,
		Attrs: FileObjectEntryAttrVersion1 | FileObjectEntryAttrBlockSize4,
	}
}

func (it *FileObjectEntryInit) Valid() bool {
	if len(it.Path) < 1 || it.Size < 1 {
		return false
	}
	return true
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

func NewFileObjectEntryBlock(path string, size uint64, block_num uint32, data []byte, commit_key string) FileObjectEntryBlock {
	return FileObjectEntryBlock{
		Path:      FileObjectPathEncode(path),
		Size:      size,
		Attrs:     FileObjectEntryAttrVersion1 | FileObjectEntryAttrBlockSize4,
		Num:       block_num,
		Data:      data,
		CommitKey: commit_key,
	}
}

func (it *FileObjectEntryBlock) Valid() bool {
	if len(it.Path) < 1 || it.Size < 1 || len(it.Data) < 1 {
		return false
	}
	num_max := uint32(it.Size / FileObjectBlockSize4)
	if (it.Size % FileObjectBlockSize4) == 0 {
		num_max -= 1
	}
	if it.Num > num_max {
		return false
	}
	if it.Num < num_max {
		if uint64(len(it.Data)) != FileObjectBlockSize4 {
			return false
		}
	} else {
		if uint64(len(it.Data)) != (it.Size % FileObjectBlockSize4) {
			return false
		}
	}
	return true
}
