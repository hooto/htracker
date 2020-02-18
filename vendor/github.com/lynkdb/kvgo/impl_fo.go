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

	"github.com/lynkdb/iomix/sko"
)

func foFilePathBlock(path string, n uint32) []byte {
	return []byte(path + ":" + sko.Uint32ToHexString(n))
}

func (cn *Conn) FoFilePut(src_path, dst_path string) *sko.ObjectResult {

	fp, err := os.Open(src_path)
	if err != nil {
		return sko.NewObjectResultClientError(err)
	}

	st, err := fp.Stat()
	if err != nil {
		return sko.NewObjectResultClientError(err)
	}

	if st.Size() < 1 {
		return sko.NewObjectResultClientError(errors.New("invalid file size"))
	}

	var (
		block0    = sko.NewFileObjectBlock(dst_path, st.Size(), 0, nil)
		blockSize = int64(0)
	)

	if ors := cn.NewReader(foFilePathBlock(block0.Path, 0)).Query(); ors.OK() {

		var prev sko.FileObjectBlock

		if err := ors.Decode(&prev); err != nil {
			return sko.NewObjectResultClientError(err)
		}

		if prev.Size != st.Size() {
			return sko.NewObjectResultClientError(errors.New("protocol error"))
		}

		block0.Attrs = prev.Attrs
	}

	if block0.AttrAllow(sko.FileObjectBlockAttrBlockSize4) {
		blockSize = sko.FileObjectBlockSize4
	}

	if blockSize == 0 {
		return sko.NewObjectResultClientError(errors.New("protocol error"))
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
			return sko.NewObjectResultClientError(err)
		} else if rn != bsize {
			return sko.NewObjectResultClientError(errors.New("io error"))
		} else {

			mpBlock := sko.FileObjectBlock{
				Path:  block0.Path,
				Size:  block0.Size,
				Attrs: block0.Attrs,
				Num:   n,
				Data:  bs,
			}

			if rs := cn.NewWriter(foFilePathBlock(block0.Path, n), mpBlock).
				Commit(); !rs.OK() {
				return sko.NewObjectResultServerError(rs.Error())
			}

		}
	}

	return sko.NewObjectResultOK()
}

func (cn *Conn) FoFileOpen(path string) (io.ReadSeeker, error) {

	rs := cn.NewReader(foFilePathBlock(path, 0)).Query()
	if !rs.OK() {
		return nil, rs.Error()
	}

	var block0 sko.FileObjectBlock
	if err := rs.Decode(&block0); err != nil {
		return nil, errors.New("ER decode meta : " + err.Error())
	}

	return &FoReadSeeker{
		conn:   cn,
		block0: block0,
		path:   path,
		offset: 0,
	}, nil
}

type FoReadSeeker struct {
	conn   *Conn
	block0 sko.FileObjectBlock
	blockx *sko.FileObjectBlock
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

	blockSize := int64(0)
	if fo.block0.AttrAllow(sko.FileObjectBlockAttrBlockSize4) {
		blockSize = sko.FileObjectBlockSize4
	}
	if blockSize == 0 {
		return 0, errors.New("protocol error")
	}

	blk_num_max := uint32(fo.block0.Size / blockSize)
	if (fo.block0.Size % blockSize) > 0 {
		blk_num_max += 1
	}

	var (
		n_done = 0
		n_len  = len(b)
	)

	for {

		if fo.offset >= int64(fo.block0.Size) {
			return n_done, io.EOF
		}

		var (
			blk_num = uint32(fo.offset / blockSize)
			blk_off = int(fo.offset % blockSize)
		)

		if blk_num > blk_num_max {
			return n_done, io.EOF
		}

		if blk_num == 0 {
			fo.blockx = &fo.block0
		}

		if fo.blockx == nil || fo.blockx.Num != blk_num {

			rs := fo.conn.NewReader(foFilePathBlock(fo.path, blk_num)).Query()
			if !rs.OK() {

				return 0, errors.New("io error")
			}

			var foBlock sko.FileObjectBlock
			if err := rs.Decode(&foBlock); err != nil {
				return 0, errors.New("io error")
			}

			if len(foBlock.Data) < 1 {
				return 0, errors.New("io error")
			}

			fo.blockx = &foBlock
		}

		blk_off_n := len(fo.blockx.Data) - blk_off
		if blk_off_n < 1 {
			return 0, errors.New("offset error")
		}
		if blk_off_n > n_len {
			blk_off_n = n_len
		}

		copy(b[n_done:], fo.blockx.Data[blk_off:(blk_off+blk_off_n)])

		fo.offset += int64(blk_off_n)
		n_done += blk_off_n
		n_len -= blk_off_n

		if n_len < 1 {
			break
		}
	}

	return n_done, nil
}
