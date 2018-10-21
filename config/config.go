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

	"github.com/hooto/hflag4g/hflag"
	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/connect"
)

var (
	Prefix      = "/opt/hooto/tracker"
	Version     = "0.1.5"
	Release     = "1"
	VersionHash = Version // TODO
	err         error
	Config      ConfigCommon
)

type ConfigCommon struct {
	HttpPort uint16              `json:"http_port"`
	RunMode  string              `json:"run_mode,omitempty"`
	Auth     string              `json:"auth"`
	Data     connect.ConnOptions `json:"data"`
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

	file := Prefix + "/etc/config.json"
	if err := json.DecodeFile(file, &Config); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	Config.Data = connect.ConnOptions{
		Name:      "tracker_db",
		Connector: "iomix/skv/Connector",
		Driver:    types.NewNameIdentifier("lynkdb/kvgo"),
	}
	data_dir := Prefix + "/var/" + string(Config.Data.Name)
	Config.Data.SetValue("data_dir", data_dir)

	if err = os.MkdirAll(data_dir, 0755); err != nil {
		return err
	}

	if Config.HttpPort < 1 {
		Config.HttpPort = 9520
	}

	hlog.Printf("info", "setup data dir %s", data_dir)

	return Sync()
}

func Sync() error {
	return json.EncodeToFile(Config, Prefix+"/etc/config.json", "  ")
}
