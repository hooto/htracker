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
	"github.com/golang/protobuf/proto"
)

// Client Connector APIs
type ClientConnector interface {
	Query(req *ObjectReader) *ObjectResult
	Commit(req *ObjectWriter) *ObjectResult
	BatchCommit(req *BatchRequest) *BatchResult
	SysCmd(req *SysCmdRequest) *ObjectResult
	Close() error
}

// Client APIs
type Client interface {
	NewReader(keys ...[]byte) *ClientReader
	NewWriter(key []byte, value interface{}, opts ...interface{}) *ClientWriter
	OpenTable(tableName string) ClientTable
	OptionApply(opts ...ClientOption)
	Connector() ClientConnector
	Close() error
}

// Client Table APIs
type ClientTable interface {
	NewReader(keys ...[]byte) *ClientReader
	NewWriter(key []byte, value interface{}, opts ...interface{}) *ClientWriter
	NewBatch() *ClientBatch
}

type ClientReader struct {
	*ObjectReader
	cc ClientConnector
}

func NewClientReader(cn ClientConnector, keys ...[]byte) *ClientReader {
	return &ClientReader{
		ObjectReader: NewObjectReader(keys...),
		cc:           cn,
	}
}

func (it *ClientReader) Query() *ObjectResult {
	rs := it.cc.Query(it.ObjectReader)
	if rs.Meta == nil && len(rs.Items) > 0 {
		rs.Meta = rs.Items[0].Meta
	}
	return rs
}

func (it *ClientReader) TableNameSet(tableName string) *ClientReader {
	it.ObjectReader.TableNameSet(tableName)
	return it
}

func (it *ClientReader) KeySet(key []byte) *ClientReader {
	it.ObjectReader.KeySet(key)
	return it
}

func (it *ClientReader) KeyRangeSet(keyOffset, keyCutset []byte) *ClientReader {
	it.ObjectReader.KeyRangeSet(keyOffset, keyCutset)
	return it
}

func (it *ClientReader) LimitNumSet(num int64) *ClientReader {
	it.ObjectReader.LimitNumSet(num)
	return it
}

func (it *ClientReader) ModeRevRangeSet(v bool) *ClientReader {
	it.ObjectReader.ModeRevRangeSet(v)
	return it
}

func (it *ClientReader) LogOffsetSet(logOffset uint64) *ClientReader {
	it.LogOffsetSet(logOffset)
	return it
}

func (it *ClientReader) AttrSet(v uint64) *ClientReader {
	it.ObjectReader.Attrs |= v
	return it
}

type ClientWriter struct {
	*ObjectWriter
	cc ClientConnector
}

func NewClientWriter(cn ClientConnector, key []byte, value interface{},
	opts ...interface{}) *ClientWriter {
	return &ClientWriter{
		ObjectWriter: NewObjectWriter(key, value, opts...),
		cc:           cn,
	}
}

func (it *ClientWriter) Commit() *ObjectResult {
	return it.cc.Commit(it.ObjectWriter)
}

func (it *ClientWriter) TableNameSet(tableName string) *ClientWriter {
	it.ObjectWriter.TableNameSet(tableName)
	return it
}

func (it *ClientWriter) ModeCreateSet(v bool) *ClientWriter {
	it.ObjectWriter.ModeCreateSet(v)
	return it
}

func (it *ClientWriter) IncrNamespaceSet(ns string) *ClientWriter {
	it.ObjectWriter.IncrNamespaceSet(ns)
	return it
}

func (it *ClientWriter) ModeDeleteSet(v bool) *ClientWriter {
	it.ObjectWriter.ModeDeleteSet(v)
	return it
}

func (it *ClientWriter) ExpireSet(v int64) *ClientWriter {
	it.ObjectWriter.ExpireSet(v)
	return it
}

func (it *ClientWriter) PrevDataCheckSet(
	value interface{}, codec DataValueCodec) *ClientWriter {
	it.ObjectWriter.PrevDataCheckSet(value, codec)
	return it
}

func (it *ClientWriter) DataValueSet(
	value interface{}, codec DataValueCodec) *ClientWriter {
	it.ObjectWriter.DataValueSet(value, codec)
	return it
}

func (it *ClientWriter) AttrSet(v uint64) *ClientWriter {
	it.ObjectWriter.Meta.Attrs |= v
	return it
}

type ClientObjectItem struct {
	Key   []byte
	Value interface{}
}

// Batch APIs
type ClientBatch struct {
	*BatchRequest
	cc ClientConnector
}

func NewClientBatch(cc ClientConnector, tableName string) *ClientBatch {
	return &ClientBatch{
		BatchRequest: NewBatchRequest(tableName),
		cc:           cc,
	}
}

func (it *ClientBatch) Commit() *BatchResult {
	return it.cc.BatchCommit(it.BatchRequest)
}

type client struct {
	cc   ClientConnector
	opts *ClientOptions
}

func NewClient(cc ClientConnector, opts ...ClientOption) (Client, error) {

	c := &client{
		cc:   cc,
		opts: DefaultClientOptions(),
	}

	c.OptionApply(opts...)

	return c, nil
}

func (it *client) OptionApply(opts ...ClientOption) {
	for _, opt := range opts {
		opt.Apply(it.opts)
	}
}

func (it *client) NewReader(keys ...[]byte) *ClientReader {
	return NewClientReader(it.cc, keys...)
}

func (it *client) NewWriter(key []byte, value interface{}, opts ...interface{}) *ClientWriter {
	return NewClientWriter(it.cc, key, value, opts...)
}

func (it *client) OpenTable(tableName string) ClientTable {
	return NewClientTable(it, tableName)
}

func (it *client) SysCmd(req *SysCmdRequest) *ObjectResult {
	return it.cc.SysCmd(req)
}

func (it *client) Connector() ClientConnector {
	return it.cc
}

func (it *client) Close() error {
	if it.cc != nil {
		it.cc.Close()
	}
	return nil
}

type clientTable struct {
	client    Client
	tableName string
}

func NewClientTable(c Client, tableName string) ClientTable {
	return &clientTable{
		client:    c,
		tableName: tableName,
	}
}

func (it *clientTable) NewReader(keys ...[]byte) *ClientReader {
	return NewClientReader(it.client.Connector(), keys...).TableNameSet(it.tableName)
}

func (it *clientTable) NewWriter(key []byte, value interface{}, opts ...interface{}) *ClientWriter {
	return NewClientWriter(it.client.Connector(), key, value, opts...).TableNameSet(it.tableName)
}

func (it *clientTable) NewBatch() *ClientBatch {
	return NewClientBatch(it.client.Connector(), it.tableName)
}

func NewSysCmdRequest(method string, msg proto.Message) *SysCmdRequest {
	req := &SysCmdRequest{
		Method: method,
	}
	if msg != nil {
		req.Body, _ = StdProto.Encode(msg)
	}
	return req
}
