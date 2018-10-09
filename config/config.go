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
	Prefix      = "/opt/hooto/htracker"
	Version     = "0.1.0"
	VersionHash = Version // TODO
	err         error
	Config      ConfigCommon
)

type ConfigCommon struct {
	InstanceId string              `json:"instance_id"`
	Data       connect.ConnOptions `json:"data"`
}

func Setup() error {

	prefix := ""

	if v, ok := hflag.Value("prefix"); ok {
		prefix = v.String()
	}

	if prefix == "" {
		prefix, _ = filepath.Abs(filepath.Dir(os.Args[0]) + "/")
		if strings.HasSuffix(prefix, "/bin") {
			prefix = prefix[:len(prefix)-4]
		}
	}

	if prefix != "" {
		Prefix = filepath.Clean(prefix)
	}

	hlog.Printf("warn", "setup prefix %s", Prefix)

	file := Prefix + "/etc/config.json"
	if err := json.DecodeFile(file, &Config); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	Config.Data = connect.ConnOptions{
		Name:      "htracker_data",
		Connector: "iomix/skv/Connector",
		Driver:    types.NewNameIdentifier("lynkdb/kvgo"),
	}
	data_dir := Prefix + "/var/" + string(Config.Data.Name)
	Config.Data.SetValue("data_dir", data_dir)

	if err = os.MkdirAll(data_dir, 0755); err != nil {
		return err
	}

	hlog.Printf("warn", "setup data dir %s", data_dir)

	return nil
}

func Sync() error {
	return json.EncodeToFile(Config, Prefix+"/etc/config.json", "  ")
}
