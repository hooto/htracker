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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hooto/hauth/go/hauth/v1"
	"github.com/hooto/hflag4g/hflag"
	"github.com/hooto/hlog4g/hlog"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

var (
	connMu sync.Mutex
	conns  = map[string]*Conn{}
)

type dbTableIncrSet struct {
	offset uint64
	cutset uint64
}

type dbTable struct {
	instId    string
	tableId   uint32
	tableName string
	db        *leveldb.DB
	incrSets  map[string]*dbTableIncrSet
}

type Conn struct {
	mu                   sync.RWMutex
	dbmu                 sync.Mutex
	dbSys                *leveldb.DB
	tables               map[string]*dbTable
	opts                 *Config
	clients              int
	logMu                sync.RWMutex
	logOffset            uint64
	logCutset            uint64
	incrMu               sync.RWMutex
	incrOffset           uint64
	incrCutset           uint64
	client               *kv2.PublicClient
	public               *PublicServiceImpl
	internal             *InternalServiceImpl
	keyMgr               *hauth.AccessKeyManager
	close                bool
	tableName            string
	workmu               sync.Mutex
	workerLocalRunning   bool
	uptime               int64
	workerTableRefreshed int64
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
			keyMgr:     hauth.NewAccessKeyManager(),
			tables:     map[string]*dbTable{},
			opts:       &Config{},
			uptime:     time.Now().Unix(),
		}
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

		default:
			return nil, errors.New("invalid config")
		}
	}

	cn.opts.Reset()

	if err := cn.opts.Valid(); err != nil {
		return nil, err
	}

	if cn.opts.Storage.DataDirectory == "" {
		cn.opts.ClientConnectEnable = true
	}

	if cn.opts.ClientConnectEnable {

		if err := cn.serviceStart(); err != nil {
			cn.closeForce()
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

		if err := cn.dbSysSetup(); err != nil {
			hlog.Printf("error", "kvgo db-meta setup error %s", err.Error())
			return nil, err
		}

		if err := cn.dbTableListSetup(); err != nil {
			hlog.Printf("error", "kvgo db-table setup error %s", err.Error())
			return nil, err
		}
	}

	if err := cn.serviceStart(); err != nil {
		cn.closeForce()
		return nil, err
	}

	go cn.workerLocal()

	hlog.Printf("info", "kvgo started (%s)", cn.opts.Storage.DataDirectory)

	conns[cn.opts.Storage.DataDirectory] = cn

	time.Sleep(500e6)

	return cn, nil
}

func (it *Conn) NewClient() (kv2.Client, error) {
	return kv2.NewClient(it)
}

func (cn *Conn) tabledb(name string) *dbTable {
	if name == "" {
		name = "main"
	}
	dt := cn.tables[name]
	if dt != nil && dt.db != nil {
		return dt
	}
	return nil
}

func (cn *Conn) dbSetup(dir string, opts *opt.Options) (*dbTable, error) {

	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, err
	}

	if false {
		opts.CompactionL0Trigger = 8
		opts.WriteL0PauseTrigger = 24
		opts.WriteL0SlowdownTrigger = 16
	}

	db, err := leveldb.OpenFile(dir, opts)
	if err != nil {
		return nil, err
	}

	if _, ok := hflag.ValueOK("db-ns-stats"); ok {

		for _, v := range []uint8{
			nsKeySys,
			nsKeyMeta,
			nsKeyData,
			nsKeyLog,
			nsKeyTtl,
		} {

			if strings.HasSuffix(dir, "/sys") {
				continue
			}

			iter := db.NewIterator(&util.Range{
				Start: []byte{v},
				Limit: []byte{v, 0xff},
			}, nil)
			defer iter.Release()

			num := 0

			for iter.Next() {
				num += 1
			}

			if num == 0 {
				continue
			}

			hlog.Printf("info", "db-ns-stats table %s, ns %d, num %d",
				dir, v, num)

			if v == nsKeyLog {

				if iter.Prev() {
					meta, err := kv2.ObjectMetaDecode(bytesClone(iter.Value()))
					if err == nil {
						hlog.Printf("info", "db-ns-stats table %s, ns %d, log-id %d",
							dir, v, meta.Version)
					}
				}
			}
		}
	}

	dt := &dbTable{
		db:       db,
		incrSets: map[string]*dbTableIncrSet{},
	}

	bs, err := dt.db.Get(keySysInstanceId, nil)
	if err == nil {
		dt.instId = string(bs)
	} else if err.Error() == ldbNotFound {
		dt.instId = randHexString(16)
		err = dt.db.Put(keySysInstanceId, []byte(dt.instId), nil)
	}

	if err != nil {
		dt.Close()
		dt = nil
	}

	return dt, err
}

func (cn *Conn) dbSysSetup() error {

	var (
		dir  = filepath.Clean(fmt.Sprintf("%s/%s", cn.opts.Storage.DataDirectory, sysTableName))
		opts = &opt.Options{
			WriteBuffer:            2 * opt.MiB,
			BlockCacheCapacity:     2 * opt.MiB,
			CompactionTableSize:    2 * opt.MiB,
			OpenFilesCacheCapacity: 10,
			Filter:                 filter.NewBloomFilter(10),
			Compression:            opt.NoCompression,
		}
	)

	dt, err := cn.dbSetup(dir, opts)
	if err != nil {
		return err
	}

	cn.dbSys = dt.db
	cn.tables[sysTableName] = &dbTable{
		tableId:   0,
		tableName: sysTableName,
		db:        dt.db,
		incrSets:  map[string]*dbTableIncrSet{},
	}

	if cn.opts.Server.Bind != "" {

		rr2 := kv2.NewObjectReader(nil).
			TableNameSet(sysTableName).
			KeyRangeSet(nsSysAccessKey(""), append(nsSysAccessKey(""), 0xff)).
			LimitNumSet(1000)

		if rs := cn.objectLocalQuery(rr2); rs.OK() {
			for _, v := range rs.Items {
				var key hauth.AccessKey
				if err := v.DataValue().Decode(&key, nil); err == nil {
					cn.keyMgr.KeySet(&key)
				}
			}
			hlog.Printf("info", "server load access keys %d", len(rs.Items))
		}

		if cn.opts.Server.AccessKey != nil &&
			len(cn.opts.Server.AccessKey.Secret) > 20 {
			key := cn.opts.Server.AccessKey
			if pkey := cn.keyMgr.KeyGet(key.Id); pkey == nil || key.Secret != pkey.Secret {

				rootKey := NewSystemAccessKey()
				key.Roles = rootKey.Roles
				key.Scopes = rootKey.Scopes

				rr2 := kv2.NewObjectWriter(nsSysAccessKey(key.Id), key).
					TableNameSet(sysTableName)
				tdb := cn.tabledb(sysTableName)
				if tdb != nil {
					cn.commitLocal(rr2, 0)
					cn.keyMgr.KeySet(key)
					hlog.Printf("warn", "server force rewrite root access key")
				}
			}
		}

		for _, role := range defaultRoles {
			cn.keyMgr.RoleSet(role)
		}
	}

	return nil
}

func (cn *Conn) dbTableListSetup() error {

	tables := map[string]*dbTable{
		"main": {
			tableId:   10,
			tableName: "main",
			incrSets:  map[string]*dbTableIncrSet{},
		},
	}

	for _, t := range tables {

		k := nsSysTable(t.tableName)

		if _, err := cn.dbSys.Get(keyEncode(nsKeyData, k), nil); err != nil {

			if err.Error() == ldbNotFound {

				obj := kv2.NewObjectWriter(k, &kv2.TableItem{
					Name: t.tableName,
				}).IncrNamespaceSet(sysTableIncrNS)

				obj.Meta.IncrId = uint64(t.tableId)
				obj.TableName = sysTableName

				rs := cn.commitLocal(obj, 0)
				if !rs.OK() {
					return errors.New(rs.Message)
				}

				hlog.Printf("info", "init db %s table ok", sysTableName)

			} else if err.Error() != ldbNotFound {
				return err
			}
		}
	}

	var (
		offset = keyEncode(nsKeyData, nsSysTable(""))
		cutset = keyEncode(nsKeyData, nsSysTable(""))
		values = [][]byte{}
	)
	cutset = append(cutset, 0xff)

	iter := cn.dbSys.NewIterator(&util.Range{
		Start: offset,
		Limit: cutset,
	}, nil)
	defer iter.Release()

	for iter.Next() {

		if bytes.Compare(iter.Key(), offset) <= 0 {
			continue
		}

		if bytes.Compare(iter.Key(), cutset) >= 0 {
			break
		}

		values = append(values, bytesClone(iter.Value()))
	}

	if iter.Error() != nil {
		return iter.Error()
	}

	for _, bs := range values {

		item, err := kv2.ObjectItemDecode(bs)
		if err != nil {
			return err
		}

		var tb kv2.TableItem
		if err = item.DataValue().Decode(&tb, nil); err != nil {
			return err
		}

		if tables[tb.Name] != nil &&
			uint64(tables[tb.Name].tableId) != item.Meta.IncrId {
			return fmt.Errorf("table name (%s) conflict", tb.Name)
		}

		if tb.Name == sysTableName {
			continue
		}

		tables[tb.Name] = &dbTable{
			tableId:   uint32(item.Meta.IncrId),
			tableName: tb.Name,
			incrSets:  map[string]*dbTableIncrSet{},
		}
	}

	for _, t := range tables {

		if err := cn.dbTableSetup(t.tableName, t.tableId); err != nil {
			return err
		}

		hlog.Printf("info", "kvgo table %s (%d) started", t.tableName, t.tableId)
	}

	return nil
}

func (cn *Conn) dbTableSetup(tableName string, tableId uint32) error {

	cn.dbmu.Lock()
	defer cn.dbmu.Unlock()

	tdb := cn.tabledb(tableName)
	if tdb != nil {
		return nil
	}

	dir := filepath.Clean(fmt.Sprintf("%s/%d_%d_%d", cn.opts.Storage.DataDirectory,
		tableId, 0, 0))

	ldbOpts := &opt.Options{
		WriteBuffer:            cn.opts.Performance.WriteBufferSize * opt.MiB,
		BlockCacheCapacity:     cn.opts.Performance.BlockCacheSize * opt.MiB,
		CompactionTableSize:    cn.opts.Performance.MaxTableSize * opt.MiB,
		OpenFilesCacheCapacity: cn.opts.Performance.MaxOpenFiles,
		Filter:                 filter.NewBloomFilter(10),
	}

	if cn.opts.Feature.TableCompressName == "snappy" {
		ldbOpts.Compression = opt.SnappyCompression
	} else {
		ldbOpts.Compression = opt.NoCompression
	}

	dt, err := cn.dbSetup(dir, ldbOpts)
	if err != nil {
		return err
	}

	cn.tables[tableName] = &dbTable{
		tableId:   tableId,
		tableName: tableName,
		incrSets:  map[string]*dbTableIncrSet{},
		db:        dt.db,
	}

	return nil
}

func (cn *Conn) OptionApply(opts ...kv2.ClientOption) {
	// TODO
}

func (cn *Conn) Close() error {

	connMu.Lock()
	defer connMu.Unlock()

	cn.close = true

	return cn.closeForce()
}

func (cn *Conn) closeForce() error {

	if pconn, ok := conns[cn.opts.Storage.DataDirectory]; ok {

		if pconn.clients > 1 {
			pconn.clients--
			return nil
		}
	}

	if cn.public != nil && cn.public.sock != nil {
		cn.public.sock.Close()
	}

	for _, tdb := range cn.tables {
		tdb.Close()
	}

	if cn.dbSys != nil {
		// cn.dbSys.Close()
	}

	delete(conns, cn.opts.Storage.DataDirectory)

	return nil
}

func (it *dbTable) Close() error {

	if it.db == nil {
		return nil
	}

	for ns, incrSet := range it.incrSets {

		if incrSet.cutset <= incrSet.offset {
			continue
		}

		incrSet.cutset = incrSet.offset

		if err := it.db.Put(keySysIncrCutset(ns),
			[]byte(strconv.FormatUint(incrSet.cutset, 10)), nil); err != nil {
			hlog.Printf("info", "db error %s", err.Error())
		} else {
			hlog.Printf("info", "db incr set sync offset %d", incrSet.offset)
		}
	}

	it.db.Close()
	it.db = nil

	return nil
}

type connTable struct {
	*Conn
	tableName string
}

func (it *Conn) OpenTable(tableName string) kv2.ClientTable {
	return &connTable{
		Conn:      it,
		tableName: tableName,
	}
}

func (it *connTable) NewBatch() *kv2.ClientBatch {
	return kv2.NewClientBatch(it, it.tableName)
}
