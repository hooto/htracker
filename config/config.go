// Copyright 2018 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hooto/hconf4g/hconf"
	"github.com/hooto/hflag4g/hflag"
	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/kvgo"
)

var (
	Prefix      = "/opt/hooto/tracker"
	Version     = "0.1.10"
	Release     = "1"
	VersionHash = Version // TODO
	err         error
	Config      ConfigCommon
)

type ConfigCommon struct {
	filepath     string
	HttpPort     uint16             `json:"http_port" toml:"http_port"`
	HttpBasepath string             `json:"http_basepath,omitempty" toml:"http_basepath,omitempty"`
	RunMode      string             `json:"run_mode,omitempty" toml:"run_mode,omitempty"`
	Auth         string             `json:"auth" toml:"auth"`
	DataStorage  kvgo.ConfigStorage `json:"data_storage" toml:"data_storage"`
}

func Setup(version, release string) error {

	if version != "" {
		ver := types.Version(version)
		if ver.Valid() {
			Version = ver.String()
		}
	}

	if release != "" {
		rel := types.Version(release)
		if rel.Valid() {
			Release = rel.String()
		}
	}

	VersionHash = Version + "-" + Release

	hlog.Printf("info", "version %s, release %s", Version, Release)

	prefix := hflag.Value("prefix").String()

	if prefix == "" {
		prefix, _ = filepath.Abs(filepath.Dir(os.Args[0]) + "/")
		if strings.HasSuffix(prefix, "/bin") {
			prefix = prefix[:len(prefix)-4]
		}
	}

	if prefix != "" {
		Prefix = filepath.Clean(prefix)
	}

	hlog.Printf("info", "setup prefix %s", Prefix)

	if err := hconf.DecodeFromFile(&Config, Prefix+"/etc/main.conf"); err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		//
		if err := json.DecodeFile(Prefix+"/etc/config.json", &Config); err != nil &&
			!os.IsNotExist(err) {
			return err
		}
	}

	Config.filepath = Prefix + "/etc/main.conf"

	Config.DataStorage.DataDirectory = Prefix + "/var/db_local"

	if Config.HttpPort < 1 {
		Config.HttpPort = 9520
	}

	if Config.HttpBasepath != "" {
		Config.HttpBasepath = filepath.Clean(Config.HttpBasepath)
		if Config.HttpBasepath == "." ||
			Config.HttpBasepath == ".." ||
			Config.HttpBasepath == "/" {
			Config.HttpBasepath = ""
		}
		hlog.Printf("info", "setup http_prefix %s", Config.HttpBasepath)
	}

	return Sync()
}

func Sync() error {
	if Config.filepath != "" {
		return hconf.EncodeToFile(Config, Config.filepath)
	}
	return nil
}
