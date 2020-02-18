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
	"math/rand"
	"path/filepath"

	"github.com/lynkdb/iomix/connect"
)

type Config struct {

	// Storage Settings
	Storage ConfigStorage `toml:"storage" json:"storage"`

	// Server Settings
	Server ConfigServer `toml:"server" json:"server"`

	// Performance Settings
	Performance ConfigPerformance `toml:"performance" json:"performance"`

	// Feature Settings
	Feature ConfigFeature `toml:"feature" json:"feature"`

	// Cluster Settings
	Cluster ConfigCluster `toml:"cluster" json:"cluster"`

	// Client Settings
	ClientConnectEnable bool `toml:"-" json:"-"`
}

type ConfigStorage struct {
	DataDirectory string `toml:"data_directory" json:"data_directory"`
}

type ConfigServer struct {
	Bind          string `toml:"bind" json:"bind"`
	AuthSecretKey string `toml:"auth_secret_key" json:"auth_secret_key"`
}

type ConfigPerformance struct {
	WriteBufferSize int `toml:"write_buffer_size" json:"write_buffer_size"`
	BlockCacheSize  int `toml:"block_cache_size" json:"block_cache_size"`
	MaxTableSize    int `toml:"max_table_size" json:"max_table_size"`
	MaxOpenFiles    int `toml:"max_open_files" json:"max_open_files"`
}

type ConfigFeature struct {
	WriteMetaDisable  bool   `toml:"write_meta_disable" json:"write_meta_disable"`
	WriteLogDisable   bool   `toml:"write_log_disable" json:"write_log_disable"`
	TableCompressName string `toml:"table_compress_name" json:"table_compress_name"`
}

type ConfigCluster struct {
	Masters []ConfigClusterMaster `toml:"masters" json:"masters"`
}

type ConfigClusterMaster struct {
	Addr          string `toml:"addr" json:"addr"`
	AuthSecretKey string `toml:"auth_secret_key" json:"auth_secret_key"`
}

func (it *ConfigCluster) masterAddrs(cap int) []string {

	var (
		addrs  = []string{}
		offset = rand.Intn(len(it.Masters))
	)

	for i := offset; i < len(it.Masters) && len(addrs) <= cap; i++ {
		addrs = append(addrs, it.Masters[i].Addr)
	}
	for i := 0; i < offset && len(addrs) <= cap; i++ {
		addrs = append(addrs, it.Masters[i].Addr)
	}

	return addrs
}

func (it *Config) Valid() error {

	if it.ClientConnectEnable {
		if len(it.Cluster.Masters) < 1 {
			return errors.New("no cluster/masters setup")
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

func (it *Config) reset() *Config {

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

	if it.Feature.TableCompressName != "snappy" {
		it.Feature.TableCompressName = "none"
	}

	return it
}

func configParse(opts connect.ConnOptions) (*Config, error) {

	cfg := &Config{}

	// Storage Settings
	{
		if v, ok := opts.Items.Get("storage/data_directory"); ok {
			cfg.Storage.DataDirectory = filepath.Clean(v.String())
		} else if v, ok := opts.Items.Get("data_dir"); ok {
			cfg.Storage.DataDirectory = filepath.Clean(v.String())
		} else {
			return nil, errors.New("No storage/data_directory Found")
		}
	}

	// Server Settings
	{
		if v, ok := opts.Items.Get("server/bind"); ok {
			cfg.Server.Bind = v.String()
		}
	}

	// Performance Settings
	{
		if v, ok := opts.Items.Get("performance/write_buffer_size"); ok {
			cfg.Performance.WriteBufferSize = v.Int()
		}

		if v, ok := opts.Items.Get("performance/block_cache_size"); ok {
			cfg.Performance.BlockCacheSize = v.Int()
		}

		if v, ok := opts.Items.Get("performance/max_open_files"); ok {
			cfg.Performance.MaxOpenFiles = v.Int()
		}

		if v, ok := opts.Items.Get("performance/max_table_size"); ok {
			cfg.Performance.MaxTableSize = v.Int()
		}
	}

	// Feature Settings
	{
		if v, ok := opts.Items.Get("feature/write_meta_disable"); ok && v.String() == "true" {
			cfg.Feature.WriteMetaDisable = true
		}

		if v, ok := opts.Items.Get("feature/write_log_disable"); ok && v.String() == "true" {
			cfg.Feature.WriteLogDisable = true
		}
	}

	// Cluster Settings
	{
	}

	return cfg.reset(), nil
}
