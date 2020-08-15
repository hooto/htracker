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

package kvgo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hooto/hauth/go/hauth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

var (
	grpcClientConns = map[string]*grpc.ClientConn{}
	grpcClientMu    sync.Mutex
)

type ClientConfig struct {
	Addr        string                `toml:"addr" json:"addr"`
	AccessKey   *hauth.AccessKey      `toml:"access_key" json:"access_key"`
	AuthTLSCert *ConfigTLSCertificate `toml:"auth_tls_cert" json:"auth_tls_cert"`
	Options     *kv2.ClientOptions    `toml:"options,omitempty" json:"options,omitempty"`
	c           kv2.Client            `toml:"-" json:"-"`
	cc          *ClientConnector      `toml:"-" json:"-"`
}

type ClientConnector struct {
	cfg  *ClientConfig
	conn *grpc.ClientConn
	err  error
}

func (it *ClientConnector) reconnect(retry bool) error {

	if retry || (it.err != nil && it.err == grpc.ErrClientConnClosing) {
		if it.conn != nil {
			it.conn.Close()
		}
		it.err = nil
		it.conn = nil
	}

	if it.conn == nil {
		it.conn, it.err = clientConn(it.cfg.Addr, it.cfg.AccessKey, it.cfg.AuthTLSCert, true)
		if it.err != nil {
			return it.err
		}
	}

	return nil
}

func (it *ClientConfig) NewClient() (kv2.Client, error) {

	if it.c == nil {

		var (
			cc     = &ClientConnector{}
			c, err = kv2.NewClient(cc)
		)

		if err != nil {
			return nil, err
		}

		if it.Options == nil {
			it.Options = kv2.DefaultClientOptions()
		}

		c.OptionApply(it.Options)
		cc.cfg = it

		it.c = c
		it.cc = cc
	}

	return it.c, nil
}

func (it *ClientConnector) timeout() time.Duration {
	return time.Millisecond * time.Duration(it.cfg.Options.Timeout)
}

func (it *ClientConnector) Query(req *kv2.ObjectReader) *kv2.ObjectResult {

	if err := it.reconnect(false); err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	ctx, fc := context.WithTimeout(context.Background(), it.timeout())
	defer fc()

	rs, err := kv2.NewPublicClient(it.conn).Query(ctx, req)
	if err != nil {
		if err = it.reconnect(true); err == nil {
			rs, err = kv2.NewPublicClient(it.conn).Query(ctx, req)
		}
	}

	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	return rs
}

func (it *ClientConnector) Commit(req *kv2.ObjectWriter) *kv2.ObjectResult {

	if err := it.reconnect(false); err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	ctx, fc := context.WithTimeout(context.Background(), it.timeout())
	defer fc()

	rs, err := kv2.NewPublicClient(it.conn).Commit(ctx, req)
	if err != nil {
		if err = it.reconnect(true); err == nil {
			rs, err = kv2.NewPublicClient(it.conn).Commit(ctx, req)
		}
	}

	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	return rs
}

func (it *ClientConnector) BatchCommit(req *kv2.BatchRequest) *kv2.BatchResult {

	if err := it.reconnect(false); err != nil {
		return req.NewResult(kv2.ResultClientError, err.Error())
	}

	ctx, fc := context.WithTimeout(context.Background(), it.timeout())
	defer fc()

	rs, err := kv2.NewPublicClient(it.conn).BatchCommit(ctx, req)
	if err != nil {
		if err = it.reconnect(true); err == nil {
			rs, err = kv2.NewPublicClient(it.conn).BatchCommit(ctx, req)
		}
	}

	if err != nil {
		return req.NewResult(kv2.ResultClientError, err.Error())
	}

	return rs
}

func (it *ClientConnector) SysCmd(req *kv2.SysCmdRequest) *kv2.ObjectResult {

	if err := it.reconnect(false); err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	ctx, fc := context.WithTimeout(context.Background(), it.timeout())
	defer fc()

	rs, err := kv2.NewPublicClient(it.conn).SysCmd(ctx, req)
	if err != nil {
		if err = it.reconnect(true); err == nil {
			rs, err = kv2.NewPublicClient(it.conn).SysCmd(ctx, req)
		}
	}

	if err != nil {
		return kv2.NewObjectResultClientError(err)
	}

	return rs
}

func (it *ClientConnector) Close() error {
	if it.conn != nil {
		return it.conn.Close()
	}
	return nil
}

func clientConn(addr string,
	key *hauth.AccessKey, cert *ConfigTLSCertificate,
	forceNew bool) (*grpc.ClientConn, error) {

	if key == nil {
		return nil, errors.New("not auth key setup")
	}

	ck := fmt.Sprintf("%s.%s", addr, key.Id)

	grpcClientMu.Lock()
	defer grpcClientMu.Unlock()

	if c, ok := grpcClientConns[ck]; ok {
		if forceNew {
			c.Close()
			delete(grpcClientConns, ck)
		} else {
			return c, nil
		}
	}

	dialOptions := []grpc.DialOption{
		grpc.WithPerRPCCredentials(newAppCredential(key)),
		grpc.WithMaxMsgSize(grpcMsgByteMax),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMsgByteMax)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMsgByteMax)),
	}

	if cert == nil {

		dialOptions = append(dialOptions, grpc.WithInsecure())

	} else {

		block, _ := pem.Decode([]byte(cert.ServerCertData))
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, errors.New("failed to decode CERTIFICATE")
		}

		crt, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.New("failed to parse cert : " + err.Error())
		}

		certPool := x509.NewCertPool()
		certPool.AddCert(crt)

		// creds := credentials.NewClientTLSFromCert(certPool, addr)
		creds := credentials.NewTLS(&tls.Config{
			ServerName: crt.Subject.CommonName,
			RootCAs:    certPool,
		})

		dialOptions = append(dialOptions, grpc.WithTransportCredentials(creds))
	}

	c, err := grpc.Dial(addr, dialOptions...)
	if err != nil {
		return nil, err
	}

	grpcClientConns[ck] = c

	return c, nil
}
