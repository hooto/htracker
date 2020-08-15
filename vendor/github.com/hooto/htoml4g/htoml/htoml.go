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

package htoml

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/hooto/htoml4g/internal/toml" // fork from "github.com/BurntSushi/toml"
)

const (
	Version = "0.9.2"
)

type EncodeOptions struct {
	Indent string
}

func Decode(obj interface{}, bs []byte) error {

	if _, err := toml.Decode(string(bs), obj); err != nil {
		return err
	}

	return nil
}

func DecodeFromFile(obj interface{}, file string) error {

	fp, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()

	bs, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}

	if _, err := toml.Decode(string(bs), obj); err != nil {
		return err
	}

	return nil
}

func Encode(obj interface{}, opts *EncodeOptions) ([]byte, error) {
	var buf bytes.Buffer
	if err := prettyEncode(obj, &buf, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func EncodeToFile(obj interface{}, file string, opts *EncodeOptions) error {

	fpo, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer fpo.Close()

	fpo.Seek(0, 0)
	fpo.Truncate(0)

	var wbuf = bufio.NewWriter(fpo)

	err = prettyEncode(obj, wbuf, opts)
	if err != nil {
		return err
	}

	return wbuf.Flush()
}

func prettyEncode(obj interface{}, bufOut io.Writer, opts *EncodeOptions) error {

	var (
		buf bytes.Buffer
		enc = toml.NewEncoder(&buf)
	)

	if opts != nil {
		enc.Indent = opts.Indent
	}

	if err := enc.Encode(obj); err != nil {
		return err
	}

	for {

		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}

		if len(line) > 8 && line[len(line)-2] == '"' && line[len(line)-3] != '"' &&
			bytes.IndexByte(line, '\n') > 2 {

			if n := bytes.Index(line, []byte(" = \"")); n > 0 {
				if nb := bytes.Index(line[n+4:len(line)-2], []byte("\\n")); nb >= 0 {
					bufOut.Write(line[:n+4])
					bufOut.Write([]byte(`""`))
					bufOut.Write(bytes.Replace(line[n+4:len(line)-2], []byte("\\n"), []byte("\n"), -1))
					bufOut.Write([]byte("\"\"\"\n"))
					continue
				}
			}
		}

		bufOut.Write(line)
	}

	return nil
}
