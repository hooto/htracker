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
	"net"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/grpc"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

type InternalServiceImpl struct {
	kv2.UnimplementedInternalServer
	server     *grpc.Server
	db         *Conn
	prepares   map[string]*kv2.ObjectWriter
	proposalMu sync.RWMutex
	sock       net.Listener
}

func (it *InternalServiceImpl) Prepare(ctx context.Context,
	or *kv2.ObjectWriter) (*kv2.ObjectResult, error) {

	if err := appAuthValid(ctx, it.db.keyMgr); err != nil {
		return kv2.NewObjectResultClientError(err), nil
	}

	tdb := it.db.tabledb(or.TableName)
	if tdb == nil {
		return kv2.NewObjectResultClientError(errors.New("table not found")), nil
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

	pLog, err := tdb.objectLogVersionSet(1, 0, tn)
	if err != nil {
		return nil, err
	}

	pInc := or.Meta.IncrId
	if or.IncrNamespace != "" && or.Meta.IncrId == 0 {
		pInc, err = tdb.objectIncrSet(or.IncrNamespace, 1, 0)
		if err != nil {
			return nil, err
		}
	}

	or.ProposalExpired = tn + objAcceptTTL

	it.prepares[string(or.Meta.Key)] = or

	rs := kv2.NewObjectResultOK()
	rs.Meta = &kv2.ObjectMeta{
		Version: pLog,
		IncrId:  pInc,
	}
	return rs, nil
}

func (it *InternalServiceImpl) Accept(ctx context.Context,
	rr2 *kv2.ObjectWriter) (*kv2.ObjectResult, error) {

	if err := appAuthValid(ctx, it.db.keyMgr); err != nil {
		return kv2.NewObjectResultClientError(err), nil
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

	tdb := it.db.tabledb(rr.TableName)
	if tdb == nil {
		return nil, errors.New("table not found")
	}

	tdb.objectLogVersionSet(0, cLog, tn)
	if rr.IncrNamespace != "" && cInc > 0 {
		tdb.objectIncrSet(rr.IncrNamespace, 0, cInc)
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

	if rr.Meta.Updated < 1 {
		rr.Meta.Updated = tn
	}

	if rr.Meta.Created < 1 {
		rr.Meta.Created = tn
	}

	rr.Meta.Version = cLog
	rr.Meta.IncrId = cInc

	if kv2.AttrAllow(rr.Mode, kv2.ObjectWriterModeDelete) {

		rr.Meta.Attrs = kv2.ObjectMetaAttrDelete

		if bsMeta, err := rr.MetaEncode(); err == nil {

			batch := new(leveldb.Batch)

			if meta != nil {
				batch.Delete(keyEncode(nsKeyMeta, rr.Meta.Key))
				batch.Delete(keyEncode(nsKeyData, rr.Meta.Key))
				batch.Delete(keyEncode(nsKeyLog, uint64ToBytes(meta.Version)))
			}

			batch.Put(keyEncode(nsKeyLog, uint64ToBytes(cLog)), bsMeta)

			err = tdb.db.Write(batch, nil)
		}

	} else {

		if bsMeta, bsData, err := rr.PutEncode(); err == nil {

			batch := new(leveldb.Batch)

			if kv2.AttrAllow(rr.Meta.Attrs, kv2.ObjectMetaAttrDataOff) {
				batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsData)
			} else if kv2.AttrAllow(rr.Meta.Attrs, kv2.ObjectMetaAttrMetaOff) {
				batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			} else {
				batch.Put(keyEncode(nsKeyMeta, rr.Meta.Key), bsMeta)
				batch.Put(keyEncode(nsKeyData, rr.Meta.Key), bsData)
			}

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

			err = tdb.db.Write(batch, nil)
			if err == nil {
				tdb.objectLogFree(cLog)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	delete(it.prepares, string(rr2.Meta.Key))

	rs := kv2.NewObjectResultOK()
	rs.Meta = &kv2.ObjectMeta{
		Version: cLog,
		IncrId:  cInc,
	}

	return rs, nil
}
