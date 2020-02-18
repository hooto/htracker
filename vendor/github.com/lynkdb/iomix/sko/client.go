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

package sko // import "github.com/lynkdb/iomix/sko"

// Client Connector APIs
type ClientConnector interface {
	Close() error
	Query(rq *ObjectReader) *ObjectResult
	Commit(rq *ObjectWriter) *ObjectResult
	NewReader(key []byte) *ClientReader
	NewWriter(key []byte, value interface{}) *ClientWriter
}

type ClientReader struct {
	reader    *ObjectReader
	connector ClientConnector
}

func NewClientReader(cn ClientConnector, key []byte) *ClientReader {
	return &ClientReader{
		reader:    NewObjectReader(key),
		connector: cn,
	}
}

func (it *ClientReader) Query() *ObjectResult {
	return it.connector.Query(it.reader)
}

func (it *ClientReader) KeySet(key []byte) *ClientReader {
	it.reader.KeySet(key)
	return it
}

func (it *ClientReader) KeyRangeSet(keyOffset, keyCutset []byte) *ClientReader {
	it.reader.KeyRangeSet(keyOffset, keyCutset)
	return it
}

func (it *ClientReader) LimitNumSet(num int64) *ClientReader {
	it.reader.LimitNumSet(num)
	return it
}

func (it *ClientReader) ModeRevRangeSet(v bool) *ClientReader {
	it.reader.ModeRevRangeSet(v)
	return it
}

func (it *ClientReader) LogOffsetSet(logOffset uint64) *ClientReader {
	it.LogOffsetSet(logOffset)
	return it
}

type ClientWriter struct {
	writer    *ObjectWriter
	connector ClientConnector
}

func NewClientWriter(cn ClientConnector, key []byte, value interface{}) *ClientWriter {
	w := &ClientWriter{
		writer:    NewObjectWriter(key, value),
		connector: cn,
	}
	return w.DataValueSet(value, nil)
}

func (it *ClientWriter) Commit() *ObjectResult {
	return it.connector.Commit(it.writer)
}

func (it *ClientWriter) ModeCreateSet(v bool) *ClientWriter {
	it.writer.ModeCreateSet(v)
	return it
}

func (it *ClientWriter) IncrNamespaceSet(ns string) *ClientWriter {
	it.writer.IncrNamespaceSet(ns)
	return it
}

func (it *ClientWriter) ModeDeleteSet(v bool) *ClientWriter {
	it.writer.ModeDeleteSet(v)
	return it
}

func (it *ClientWriter) ExpireSet(v int64) *ClientWriter {
	it.writer.ExpireSet(v)
	return it
}

func (it *ClientWriter) PrevDataCheckSet(v interface{}) *ClientWriter {
	it.writer.PrevDataCheckSet(v)
	return it
}

func (it *ClientWriter) DataValueSet(
	value interface{}, codec DataValueCodec) *ClientWriter {
	it.writer.DataValueSet(value, codec)
	return it
}

type ClientObjectItem struct {
	Key   []byte
	Value interface{}
}
