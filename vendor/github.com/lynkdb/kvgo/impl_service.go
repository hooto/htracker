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
	"crypto/tls"
	"errors"
	"net"

	"github.com/hooto/hlog4g/hlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func (cn *Conn) serviceStart() error {

	if nCap := len(cn.opts.Cluster.MainNodes); nCap > 0 {

		if nCap > kv2.ObjectClusterNodeMax {
			return errors.New("Deny of kv2.ObjectClusterNodeMax")
		}

		var (
			masters = []*ClientConfig{}
			addrs   = map[string]bool{}
		)

		for _, v := range cn.opts.Cluster.MainNodes {

			host, port, err := net.SplitHostPort(v.Addr)
			if err != nil {
				return err
			}

			if _, ok := addrs[host+":"+port]; ok {
				hlog.Printf("warn", "Duplicate host:port (%s:%s) setting", host, port)
				continue
			}

			v.Addr = host + ":" + port

			if err := cn.keyMgr.KeySet(v.AccessKey); err != nil {
				return err
			}

			masters = append(masters, v)
		}

		cn.opts.Cluster.MainNodes = masters
	}

	if cn.opts.Server.Bind != "" && !cn.opts.ClientConnectEnable {

		if cn.opts.Server.AccessKey == nil {
			return errors.New("no [server.access_key] setup")
		}

		cn.keyMgr.KeySet(cn.opts.Server.AccessKey)

		host, port, err := net.SplitHostPort(cn.opts.Server.Bind)
		if err != nil {
			return err
		}

		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return err
		}
		hlog.Printf("info", "server bind %s:%s", host, port)

		cn.opts.Server.Bind = host + ":" + port

		serverOptions := []grpc.ServerOption{
			grpc.MaxMsgSize(grpcMsgByteMax),
			grpc.MaxSendMsgSize(grpcMsgByteMax),
			grpc.MaxRecvMsgSize(grpcMsgByteMax),
		}

		if cn.opts.Server.AuthTLSCert != nil {

			cert, err := tls.X509KeyPair(
				[]byte(cn.opts.Server.AuthTLSCert.ServerCertData),
				[]byte(cn.opts.Server.AuthTLSCert.ServerKeyData))
			if err != nil {
				return err
			}

			certs := credentials.NewServerTLSFromCert(&cert)

			serverOptions = append(serverOptions, grpc.Creds(certs))
		}

		server := grpc.NewServer(serverOptions...)

		go server.Serve(lis)

		cn.public = &PublicServiceImpl{
			sock:     lis,
			server:   server,
			db:       cn,
			prepares: map[string]*kv2.ObjectWriter{},
		}

		cn.internal = &InternalServiceImpl{
			sock:     lis,
			server:   server,
			db:       cn,
			prepares: map[string]*kv2.ObjectWriter{},
		}

		kv2.RegisterPublicServer(server, cn.public)
		kv2.RegisterInternalServer(server, cn.internal)

	} else {
		cn.public = &PublicServiceImpl{
			db: cn,
		}
	}

	return nil
}
