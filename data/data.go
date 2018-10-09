package data

import (
	"github.com/lynkdb/iomix/skv"
	"github.com/lynkdb/kvgo"

	"github.com/hooto/htracker/config"
)

var (
	Data skv.Connector
	err  error
)

func Setup() error {

	if Data, err = kvgo.Open(config.Config.Data); err != nil {
		return err
	}

	return nil
}
