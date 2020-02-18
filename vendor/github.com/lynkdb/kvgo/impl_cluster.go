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
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/iam/iamauth"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"

	"github.com/lynkdb/iomix/sko"
)

type ServiceImpl struct {
	server     *grpc.Server
	db         *Conn
	prepares   map[string]*sko.ObjectWriter
	proposalMu sync.RWMutex
	sock       net.Listener
}

var (
	grpcMsgByteMax  = 12 * 1024 * 1024
	grpcClientConns = map[string]*grpc.ClientConn{}
	grpcClientMu    sync.Mutex
)

func (cn *Conn) clusterStart() error {

	if nCap := len(cn.opts.Cluster.Masters); nCap > 0 {

		if nCap > sko.ObjectClusterNodeMax {
			return errors.New("Deny of sko.ObjectClusterNodeMax")
		}

		var (
			masters = []ConfigClusterMaster{}
			addrs   = map[string]bool{}
		)

		for _, v := range cn.opts.Cluster.Masters {

			host, port, err := net.SplitHostPort(v.Addr)
			if err != nil {
				return err
			}
			if _, ok := addrs[host+":"+port]; ok {
				hlog.Printf("warn", "Duplicate host:port (%s:%s) setting", host, port)
				continue
			}

			v.Addr = host + ":" + port

			if err := cn.authKeySet(v.Addr, v.AuthSecretKey); err != nil {
				return err
			}

			masters = append(masters, v)
		}

		cn.opts.Cluster.Masters = masters
	}

	if cn.opts.Server.Bind != "" && !cn.opts.ClientConnectEnable {

		if err := cn.authKeySetup(cn.serverKey, cn.opts.Server.AuthSecretKey); err != nil {
			return err
		}

		host, port, err := net.SplitHostPort(cn.opts.Server.Bind)
		if err != nil {
			return err
		}

		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return err
		}

		cn.opts.Server.Bind = host + ":" + port

		server := grpc.NewServer(
			grpc.MaxMsgSize(grpcMsgByteMax),
			grpc.MaxSendMsgSize(grpcMsgByteMax),
			grpc.MaxRecvMsgSize(grpcMsgByteMax),
		)

		go server.Serve(lis)

		cn.cluster = &ServiceImpl{
			sock:     lis,
			server:   server,
			db:       cn,
			prepares: map[string]*sko.ObjectWriter{},
		}

		sko.RegisterObjectServer(server, cn.cluster)
	} else {
		cn.cluster = &ServiceImpl{
			db: cn,
		}
	}

	if len(cn.opts.Cluster.Masters) > 0 && !cn.opts.ClientConnectEnable {
		go cn.workerClusterReplica()
	}

	return nil
}

func clientConn(addr string, key *iamauth.AuthKey) (*grpc.ClientConn, error) {

	if key == nil {
		return nil, errors.New("not auth key setup")
	}

	grpcClientMu.Lock()
	defer grpcClientMu.Unlock()

	if c, ok := grpcClientConns[addr]; ok {
		return c, nil
	}

	c, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(newAppCredential(key)),
		grpc.WithMaxMsgSize(grpcMsgByteMax),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMsgByteMax)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMsgByteMax)),
	)
	if err != nil {
		return nil, err
	}

	grpcClientConns[addr] = c

	return c, nil
}

func (it *ServiceImpl) Query(ctx context.Context,
	or *sko.ObjectReader) (*sko.ObjectResult, error) {
	if ctx != nil {
		if err := appAuthValid(ctx, it.db.serverKey); err != nil {
			return sko.NewObjectResultClientError(err), nil
		}
	}
	return it.db.Query(or), nil
}

type pQueItem struct {
	Log uint64
	Inc uint64
}

func (it *ServiceImpl) Commit(ctx context.Context,
	rr *sko.ObjectWriter) (*sko.ObjectResult, error) {

	if ctx != nil {
		if err := appAuthValid(ctx, it.db.serverKey); err != nil {
			return sko.NewObjectResultClientError(err), nil
		}
	}

	if len(it.db.opts.Cluster.Masters) == 0 {
		return it.db.Commit(rr), nil
	}

	if err := rr.CommitValid(); err != nil {
		return sko.NewObjectResultClientError(err), nil
	}

	meta, err := it.db.objectMetaGet(rr)
	if meta == nil && err != nil {
		return sko.NewObjectResultServerError(err), nil
	}

	if meta == nil {

		if sko.AttrAllow(rr.Mode, sko.ObjectWriterModeDelete) {
			return sko.NewObjectResultOK(), nil
		}

	} else {

		if sko.AttrAllow(rr.Mode, sko.ObjectWriterModeCreate) ||
			(rr.Meta.Expired == meta.Expired && meta.DataCheck == rr.Meta.DataCheck) {
			rs := sko.NewObjectResultOK()
			rs.Meta = &sko.ObjectMeta{
				Version: meta.Version,
				IncrId:  meta.IncrId,
				Created: meta.Created,
				Updated: meta.Updated,
			}
			return rs, nil
		}

		if rr.PrevVersion > 0 && rr.PrevVersion != meta.Version {
			return sko.NewObjectResultClientError(errors.New("invalid prev-version")), nil
		}

		if rr.PrevDataCheck > 0 && rr.PrevDataCheck != meta.DataCheck {
			return sko.NewObjectResultClientError(errors.New("invalid prev-data-check")), nil
		}

		if meta.IncrId > 0 {
			rr.Meta.IncrId = meta.IncrId
		}

		if meta.Created > 0 {
			rr.Meta.Created = meta.Created
		}
	}

	if rr.Meta.Created == 0 {
		rr.Meta.Created = rr.Meta.Updated
	}

	var (
		nCap = len(it.db.opts.Cluster.Masters)
		pNum = 0
		pLog = uint64(0)
		pInc = uint64(0)
		pQue = make(chan pQueItem, nCap+1)
		pTTL = time.Millisecond * time.Duration(objAcceptTTL)
	)

	for _, v := range it.db.opts.Cluster.Masters {

		conn, err := clientConn(v.Addr, it.db.authKey(v.Addr))
		if err != nil {
			continue
		}

		go func(conn *grpc.ClientConn, rr *sko.ObjectWriter) {
			ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
			defer fc()

			rs, err := sko.NewObjectClient(conn).Prepare(ctx, rr)
			if err == nil && rs.Meta != nil && rs.Meta.Version > 0 {
				pQue <- pQueItem{
					Log: rs.Meta.Version,
					Inc: rs.Meta.IncrId,
				}
			} else {
				pQue <- pQueItem{
					Log: 0,
				}
			}
		}(conn, rr)
	}

	for {

		select {
		case v := <-pQue:
			if v.Log > 0 {
				pNum += 1
				if v.Log > pLog {
					pLog = v.Log
				}
				if v.Inc > pInc {
					pInc = v.Inc
				}
			}

		case <-time.After(pTTL):
			pTTL = -1
		}

		if (pNum*2) > nCap || pTTL == -1 {
			if pNum < nCap && pTTL > 0 {
				pTTL = time.Millisecond * 10
				continue
			}
			break
		}
	}

	if (pNum * 2) <= nCap {
		return nil, fmt.Errorf("p1 fail %d/%d", pNum, nCap)
	}

	pNum = 0
	pTTL = time.Millisecond * time.Duration(objAcceptTTL)
	pQue2 := make(chan uint64, nCap+1)

	rr2 := sko.NewObjectWriter(rr.Meta.Key, nil)
	rr2.Meta.Version = pLog
	rr2.Meta.IncrId = pInc

	for _, v := range it.db.opts.Cluster.Masters {

		conn, err := clientConn(v.Addr, it.db.authKey(v.Addr))
		if err != nil {
			continue
		}

		go func(conn *grpc.ClientConn, rr *sko.ObjectWriter) {
			ctx, fc := context.WithTimeout(context.Background(), time.Second*3)
			defer fc()

			rs, err := sko.NewObjectClient(conn).Accept(ctx, rr2)
			if err == nil && rs.Meta != nil && rs.Meta.Version == pLog {
				pQue2 <- 1
			} else {
				pQue2 <- 0
			}
		}(conn, rr)
	}

	for {

		select {
		case v := <-pQue2:
			if v == 1 {
				pNum += 1
			}

		case <-time.After(pTTL):
			pTTL = -1
		}

		if (pNum*2) > nCap || pTTL == -1 {
			if pNum < nCap && pTTL > 0 {
				pTTL = time.Millisecond * 10
				continue
			}
			break
		}
	}

	if (pNum * 2) <= nCap {
		return nil, fmt.Errorf("p2 fail %d/%d", pNum, nCap)
	}

	rs := sko.NewObjectResultOK()
	rs.Meta = &sko.ObjectMeta{
		Version: pLog,
		IncrId:  pInc,
		Updated: rr.Meta.Updated,
	}

	return rs, nil
}

func (it *ServiceImpl) Prepare(ctx context.Context,
	or *sko.ObjectWriter) (*sko.ObjectResult, error) {

	if err := appAuthValid(ctx, it.db.serverKey); err != nil {
		return sko.NewObjectResultClientError(err), nil
	}

	it.proposalMu.Lock()
	defer it.proposalMu.Unlock()

	tn := uint64(time.Now().UnixNano() / 1e6)

	if len(it.prepares) > 10 {
		dels := []string{}
		for k, v := range it.prepares {
			if (v.ProposalExpired + objAcceptTTL) < tn {
				dels = append(dels, k)
			}
		}
		for _, k := range dels {
			delete(it.prepares, k)
		}
	}

	p, ok := it.prepares[string(or.Meta.Key)]
	if ok && (p.ProposalExpired+objAcceptTTL) > tn {
		return nil, errors.New("deny")
	}

	pLog, err := it.db.objectLogVersionSet(1, 0)
	if err != nil {
		return nil, err
	}

	pInc := or.Meta.IncrId
	if or.IncrNamespace != "" && or.Meta.IncrId == 0 {
		pInc, err = it.db.objectIncrSet(or.IncrNamespace, 1, 0)
		if err != nil {
			return nil, err
		}
	}

	or.ProposalExpired = tn + objAcceptTTL

	it.prepares[string(or.Meta.Key)] = or

	rs := sko.NewObjectResultOK()
	rs.Meta = &sko.ObjectMeta{
		Version: pLog,
		IncrId:  pInc,
	}
	return rs, nil
}

func (it *ServiceImpl) Accept(ctx context.Context,
	rr2 *sko.ObjectWriter) (*sko.ObjectResult, error) {

	if err := appAuthValid(ctx, it.db.serverKey); err != nil {
		return sko.NewObjectResultClientError(err), nil
	}

	it.proposalMu.Lock()
	defer it.proposalMu.Unlock()

	if rr2.Meta == nil {
		return nil, errors.New("invalid request")
	}

	var (
		tn   = uint64(time.Now().UnixNano() / 1e6)
		cLog = rr2.Meta.Version
		cInc = rr2.Meta.IncrId
	)

	rr, ok := it.prepares[string(rr2.Meta.Key)]
	if !ok || (rr.ProposalExpired+objAcceptTTL) < tn {
		return nil, errors.New("deny")
	}

	if rr.Meta.Version > cLog {
		return nil, errors.New("invalid version")
	}

	it.db.objectLogVersionSet(0, cLog)
	if rr.IncrNamespace != "" && cInc > 0 {
		it.db.objectIncrSet(rr.IncrNamespace, 0, cInc)
	}

	it.db.mu.Lock()
	defer it.db.mu.Unlock()

	meta, err := it.db.objectMetaGet(rr)
	if meta == nil && err != nil {
		return nil, err
	}
	if meta != nil && meta.Version > cLog {
		return nil, errors.New("invalid version")
	}

	rr.Meta.Version = cLog
	rr.Meta.IncrId = cInc

	if sko.AttrAllow(rr.Mode, sko.ObjectWriterModeDelete) {

		rr.Meta.Attrs = sko.ObjectMetaAttrDelete

		if bsMeta, err := rr.MetaEncode(); err == nil {

			batch := new(leveldb.Batch)

			if meta != nil {
				batch.Delete(keyEncode(nsKeyMeta, rr.Meta.Key))
				batch.Delete(keyEncode(nsKeyData, rr.Meta.Key))
				batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
			}

			batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)

			err = it.db.db.Write(batch, nil)
		}

	} else {

		if bsMeta, bsData, err := rr.PutEncode(); err == nil {

			batch := new(leveldb.Batch)

			batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsMeta)
			batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)

			if rr.Meta.Expired > 0 {
				batch.Put(keyExpireEncode(nsKeyTtl, rr.Meta.Expired, rr.Meta.Key), bsMeta)
			}

			if meta != nil {
				if meta.Version < cLog {
					batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
				}
				if meta.Expired > 0 && meta.Expired != rr.Meta.Expired {
					batch.Delete(keyExpireEncode(nsKeyTtl, meta.Expired, rr.Meta.Key))
				}
			}

			err = it.db.db.Write(batch, nil)
		}
	}

	if err != nil {
		return nil, err
	}

	delete(it.prepares, string(rr2.Meta.Key))

	rs := sko.NewObjectResultOK()
	rs.Meta = &sko.ObjectMeta{
		Version: cLog,
		IncrId:  cInc,
	}

	return rs, nil
}
