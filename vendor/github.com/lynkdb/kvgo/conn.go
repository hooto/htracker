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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/iam/iamauth"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/lynkdb/iomix/connect"
)

var (
	connMu sync.Mutex
	conns  = map[string]*Conn{}
)

type Conn struct {
	instId     string
	db         *leveldb.DB
	opts       *Config
	clients    int
	mu         sync.RWMutex
	logMu      sync.RWMutex
	logOffset  uint64
	logCutset  uint64
	incrMu     sync.RWMutex
	incrOffset uint64
	incrCutset uint64
	cluster    *ServiceImpl
	serverKey  *iamauth.AuthKey
	keyMu      sync.RWMutex
	keys       map[string]*iamauth.AuthKey
}

func Open(args ...interface{}) (*Conn, error) {

	if len(args) < 1 {
		return nil, errors.New("no config setup")
	}

	connMu.Lock()
	defer connMu.Unlock()

	var (
		cn = &Conn{
			clients:    1,
			logOffset:  0,
			logCutset:  0,
			incrOffset: 0,
			incrCutset: 0,
			serverKey:  authKeyDefault(),
			keys:       map[string]*iamauth.AuthKey{},
			opts:       &Config{},
		}
		err error
	)

	for _, cfg := range args {

		switch cfg.(type) {

		case Config:
			c := cfg.(Config)
			cn.opts = &c

		case *Config:
			cn.opts = cfg.(*Config)

		case ConfigStorage:
			c := cfg.(ConfigStorage)
			cn.opts.Storage = c

		case ConfigServer:
			cn.opts.Server = cfg.(ConfigServer)

		case ConfigPerformance:
			cn.opts.Performance = cfg.(ConfigPerformance)

		case ConfigFeature:
			cn.opts.Feature = cfg.(ConfigFeature)

		case ConfigCluster:
			cn.opts.Cluster = cfg.(ConfigCluster)

		case connect.ConnOptions:
			if cn.opts, err = configParse(cfg.(connect.ConnOptions)); err != nil {
				return nil, err
			}

		default:
			return nil, errors.New("invalid config")
		}
	}

	cn.opts.reset()

	if err := cn.opts.Valid(); err != nil {
		return nil, err
	}

	if cn.opts.Storage.DataDirectory == "" {
		cn.opts.ClientConnectEnable = true
	}

	if cn.opts.ClientConnectEnable {

		if err := cn.clusterStart(); err != nil {
			cn.Close()
			return nil, err
		}
		hlog.Printf("info", "kvgo client connected")
		return cn, nil
	}

	if pconn, ok := conns[cn.opts.Storage.DataDirectory]; ok {
		pconn.clients++
		return pconn, nil
	}

	if cn.opts.Storage.DataDirectory != "" {

		dir := filepath.Clean(fmt.Sprintf("%s/%d_%d_%d", cn.opts.Storage.DataDirectory, 10, 0, 0))
		if err := os.MkdirAll(dir, 0750); err != nil {
			return cn, err
		}

		ldbCfg := &opt.Options{
			WriteBuffer:            cn.opts.Performance.WriteBufferSize * opt.MiB,
			BlockCacheCapacity:     cn.opts.Performance.BlockCacheSize * opt.MiB,
			CompactionTableSize:    cn.opts.Performance.MaxTableSize * opt.MiB,
			OpenFilesCacheCapacity: cn.opts.Performance.MaxOpenFiles,
			Filter:                 filter.NewBloomFilter(10),
		}
		if cn.opts.Feature.TableCompressName == "snappy" {
			ldbCfg.Compression = opt.SnappyCompression
		} else {
			ldbCfg.Compression = opt.NoCompression
		}

		if cn.db, err = leveldb.OpenFile(dir, ldbCfg); err != nil {
			return nil, err
		}

		if bs, err := cn.db.Get(keySysInstanceId, nil); err == nil {
			cn.instId = string(bs)
		} else if err.Error() == ldbNotFound {
			cn.instId = randHexString(16)
			if err := cn.db.Put(keySysInstanceId, []byte(cn.instId), nil); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if err := cn.clusterStart(); err != nil {

		if cn.db != nil {
			cn.db.Close()
		}

		return nil, err
	}

	go cn.workerLocal()

	hlog.Printf("info", "kvgo %s started", cn.instId)

	conns[cn.opts.Storage.DataDirectory] = cn

	return cn, nil
}

func (cn *Conn) Close() error {

	connMu.Lock()
	defer connMu.Unlock()

	if pconn, ok := conns[cn.opts.Storage.DataDirectory]; ok {

		if pconn.clients > 1 {
			pconn.clients--
			return nil
		}
	}

	if cn.cluster != nil && cn.cluster.sock != nil {
		cn.cluster.sock.Close()
	}

	if cn.db != nil {
		cn.db.Close()
	}

	delete(conns, cn.opts.Storage.DataDirectory)

	return nil
}
