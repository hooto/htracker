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
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"strings"

	"github.com/hooto/hauth/go/hauth/v1"
)

type Config struct {

	// Storage Settings
	Storage ConfigStorage `toml:"storage" json:"storage" desc:"Storage Settings"`

	// Server Settings
	Server ConfigServer `toml:"server" json:"server" desc:"Server Settings"`

	// Performance Settings
	Performance ConfigPerformance `toml:"performance" json:"performance" desc:"Performance Settings"`

	// Feature Settings
	Feature ConfigFeature `toml:"feature" json:"feature"`

	// Cluster Settings
	Cluster ConfigCluster `toml:"cluster" json:"cluster" desc:"Cluster Settings"`

	// Client Settings
	ClientConnectEnable bool `toml:"-" json:"-"`

	// Client Keys
	// ClientAccessKeys []*hauth.AccessKey `toml:"client_access_keys" json:"client_access_keys`
}

type ConfigStorage struct {
	DataDirectory string `toml:"data_directory" json:"data_directory"`
}

type ConfigTLSCertificate struct {
	ServerKeyFile  string `toml:"server_key_file" json:"server_key_file"`
	ServerKeyData  string `toml:"server_key_data" json:"server_key_data"`
	ServerCertFile string `toml:"server_cert_file" json:"server_cert_file"`
	ServerCertData string `toml:"server_cert_data" json:"server_cert_data"`
}

type ConfigServer struct {
	Bind        string                `toml:"bind" json:"bind"`
	AccessKey   *hauth.AccessKey      `toml:"access_key" json:"access_key"`
	AuthTLSCert *ConfigTLSCertificate `toml:"auth_tls_cert" json:"auth_tls_cert"`
}

type ConfigPerformance struct {
	WriteBufferSize int `toml:"write_buffer_size" json:"write_buffer_size" desc:"in MiB, default to 8"`
	BlockCacheSize  int `toml:"block_cache_size" json:"block_cache_size" desc:"in MiB, default to 32"`
	MaxTableSize    int `toml:"max_table_size" json:"max_table_size" desc:"in MiB, default to 8"`
	MaxOpenFiles    int `toml:"max_open_files" json:"max_open_files" desc:"default to 500"`
}

type ConfigFeature struct {
	WriteMetaDisable  bool   `toml:"write_meta_disable" json:"write_meta_disable"`
	WriteLogDisable   bool   `toml:"write_log_disable" json:"write_log_disable"`
	TableCompressName string `toml:"table_compress_name" json:"table_compress_name"`
}

type ConfigCluster struct {
	//
	MainNodes []*ClientConfig `toml:"main_nodes" json:"main_nodes"`

	// Replica-Of nodes settings
	ReplicaOfNodes []*ConfigReplicaOfNode `toml:"replica_of_nodes" json:"replica_of_nodes" desc:"Replica-Of nodes settings"`
}

type ConfigReplicaOfNode struct {
	*ClientConfig
	TableMaps []*ConfigReplicaTableMap `toml:"table_maps" json:"table_maps"`
}

type ConfigReplicaTableMap struct {
	From string `toml:"from" json:"from"`
	To   string `toml:"to" json:"to"`
}

func (it *ConfigCluster) Master(addr string) *ClientConfig {

	for _, v := range it.MainNodes {
		if addr == v.Addr {
			return v
		}
	}
	return nil
}

func (it *ConfigCluster) randMainNodes(cap int) []*ClientConfig {

	var (
		ls     = []*ClientConfig{}
		offset = rand.Intn(len(it.MainNodes))
	)

	for i := offset; i < len(it.MainNodes) && len(ls) <= cap; i++ {
		ls = append(ls, it.MainNodes[i])
	}
	for i := 0; i < offset && len(ls) <= cap; i++ {
		ls = append(ls, it.MainNodes[i])
	}

	return ls
}

func (it *Config) Valid() error {

	if it.ClientConnectEnable {
		if len(it.Cluster.MainNodes) < 1 {
			return errors.New("no cluster/main_nodes setup")
		}
	}

	return nil
}

func NewConfig(dir string) *Config {
	return &Config{
		Storage: ConfigStorage{
			DataDirectory: filepath.Clean(dir),
		},
	}
}

func (it *Config) Reset() *Config {

	if it.Performance.WriteBufferSize < 4 {
		it.Performance.WriteBufferSize = 4
	} else if it.Performance.WriteBufferSize > 128 {
		it.Performance.WriteBufferSize = 128
	}

	if it.Performance.BlockCacheSize < 8 {
		it.Performance.BlockCacheSize = 8
	} else if it.Performance.BlockCacheSize > 4096 {
		it.Performance.BlockCacheSize = 4096
	}

	if it.Performance.MaxTableSize < 8 {
		it.Performance.MaxTableSize = 8
	} else if it.Performance.MaxTableSize > 64 {
		it.Performance.MaxTableSize = 64
	}

	if it.Performance.MaxOpenFiles < 500 {
		it.Performance.MaxOpenFiles = 500
	} else if it.Performance.MaxOpenFiles > 10000 {
		it.Performance.MaxOpenFiles = 10000
	}

	if it.Feature.TableCompressName != "none" {
		it.Feature.TableCompressName = "snappy"
	}

	if it.Server.Bind != "" && it.Server.AccessKey == nil {
		it.Server.AccessKey = NewSystemAccessKey()
	}

	if it.Server.AuthTLSCert != nil {

		if it.Server.AuthTLSCert.ServerKeyFile != "" &&
			it.Server.AuthTLSCert.ServerKeyData == "" {
			if bs, err := ioutil.ReadFile(it.Server.AuthTLSCert.ServerKeyFile); err == nil {
				it.Server.AuthTLSCert.ServerKeyData = strings.TrimSpace(string(bs))
			}
		}

		if it.Server.AuthTLSCert.ServerCertFile != "" &&
			it.Server.AuthTLSCert.ServerCertData == "" {
			if bs, err := ioutil.ReadFile(it.Server.AuthTLSCert.ServerCertFile); err == nil {
				it.Server.AuthTLSCert.ServerCertData = strings.TrimSpace(string(bs))
			}
		}
	}

	return it
}
