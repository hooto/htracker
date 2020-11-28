// Copyright 2018 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

import (
	"errors"
	"io"
	"os"

	"github.com/hooto/hlog4g/hlog"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func foFilePathBlock(path string, n uint32) []byte {
	return []byte(path + ":" + kv2.Uint32ToHexString(n))
}

type FileObjectConn struct {
	kv2.ClientTable
}

func NewFileObjectConn(c kv2.ClientTable) (*FileObjectConn, error) {
	return &FileObjectConn{
		ClientTable: c,
	}, nil
}

func (cn *FileObjectConn) FoFilePut(srcPath, dstPath string) *kv2.ObjectResult {

	fp, err := os.Open(srcPath)
	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}
	defer fp.Close()

	st, err := fp.Stat()
	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	if st.Size() < 1 {
		return kv2.NewObjectResultClientError(errors.New("invalid file size"))
	}

	var (
		block0    = kv2.NewFileObjectBlock(dstPath, st.Size(), 0, nil)
		blockSize = block0.BlockSize()
	)
	if blockSize == 0 {
		return kv2.NewObjectResultClientError(errors.New("protocol error"))
	}

	if ors := cn.NewReader(foFilePathBlock(block0.Path, 0)).Query(); ors.OK() {

		var prev kv2.FileObjectBlock

		if err := ors.Decode(&prev); err != nil {
			return kv2.NewObjectResultClientError(err)
		}

		if prev.Size != st.Size() {
			return kv2.NewObjectResultClientError(errors.New("protocol error"))
		}

		block0.Attrs = prev.Attrs
	}

	num := uint32(block0.Size / blockSize)
	if num > 0 && (block0.Size%blockSize) == 0 {
		num -= 1
	}

	for n := uint32(0); n <= num; n++ {

		bsize := int(blockSize)
		if n == num {
			bsize = int(block0.Size % blockSize)
		}

		bs := make([]byte, bsize)
		if rn, err := fp.ReadAt(bs, int64(n)*blockSize); err != nil {
			return kv2.NewObjectResultClientError(err)
		} else if rn != bsize {
			return kv2.NewObjectResultClientError(errors.New("io error"))
		} else {

			mpBlock := kv2.FileObjectBlock{
				Path:  block0.Path,
				Size:  block0.Size,
				Attrs: block0.Attrs,
				Num:   n,
				Data:  bs,
			}

			foPath := foFilePathBlock(block0.Path, n)

			if rs := cn.NewWriter(foPath, mpBlock).
				Commit(); !rs.OK() {
				return kv2.NewObjectResultServerError(rs.Error())
			}

			if n == num {
				hlog.Printf("info", "kvgo/fo size %d, block %d, path %s",
					block0.Size, n, foPath)
			}
		}
	}

	return kv2.NewObjectResultOK()
}

func (cn *FileObjectConn) FoFileOpen(path string) (io.ReadSeeker, error) {

	rs := cn.NewReader(foFilePathBlock(kv2.FileObjectPathEncode(path), 0)).Query()
	if !rs.OK() {
		return nil, rs.Error()
	}

	var block0 kv2.FileObjectBlock
	if err := rs.Decode(&block0); err != nil {
		return nil, errors.New("ERR decode meta : " + err.Error())
	}

	return &FoReadSeeker{
		dbt:    cn.ClientTable,
		block0: block0,
		path:   path,
		offset: 0,
	}, nil
}

func (cn *FileObjectConn) Close() error {
	return cn.Close()
}

type FoReadSeeker struct {
	dbt    kv2.ClientTable
	block0 kv2.FileObjectBlock
	blockx *kv2.FileObjectBlock
	path   string
	offset int64
}

func (fo *FoReadSeeker) Seek(offset int64, whence int) (int64, error) {

	abs := int64(0)

	switch whence {
	case 0:
		abs = offset

	case 1:
		abs = fo.offset + offset

	case 2:
		abs = offset + int64(fo.block0.Size)

	default:
		return 0, errors.New("invalid seek whence")
	}

	if abs < 0 {
		return 0, errors.New("out range of size")
	}
	fo.offset = abs

	return fo.offset, nil
}

func (fo *FoReadSeeker) Read(b []byte) (n int, err error) {

	if len(b) == 0 {
		return 0, nil
	}

	blockSize := fo.block0.BlockSize()
	if blockSize == 0 {
		return 0, errors.New("protocol error")
	}

	blockNumMax := uint32(fo.block0.Size / blockSize)
	if (fo.block0.Size % blockSize) > 0 {
		blockNumMax += 1
	}

	var (
		nDone = 0
		nLen  = len(b)
	)

	for {

		if fo.offset >= int64(fo.block0.Size) {
			return nDone, io.EOF
		}

		var (
			blockNum    = uint32(fo.offset / blockSize)
			blockOffset = int(fo.offset % blockSize)
		)

		if blockNum > blockNumMax {
			return nDone, io.EOF
		}

		if blockNum == 0 {
			fo.blockx = &fo.block0
		}

		if fo.blockx == nil || fo.blockx.Num != blockNum {

			rs := fo.dbt.NewReader(foFilePathBlock(fo.path, blockNum)).Query()
			if !rs.OK() {
				return 0, errors.New("io error : " + rs.Message)
			}

			var foBlock kv2.FileObjectBlock
			if err := rs.Decode(&foBlock); err != nil {
				return 0, errors.New("io error : " + err.Error())
			}

			if len(foBlock.Data) < 1 {
				return 0, errors.New("io error : invalid size")
			}

			fo.blockx = &foBlock
		}

		blockOffset_n := len(fo.blockx.Data) - blockOffset
		if blockOffset_n < 1 {
			return 0, errors.New("io error : offset")
		}
		if blockOffset_n > nLen {
			blockOffset_n = nLen
		}

		copy(b[nDone:], fo.blockx.Data[blockOffset:(blockOffset+blockOffset_n)])

		fo.offset += int64(blockOffset_n)
		nDone += blockOffset_n
		nLen -= blockOffset_n

		if nLen < 1 {
			break
		}
	}

	return nDone, nil
}
