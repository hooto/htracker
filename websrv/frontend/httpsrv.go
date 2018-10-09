package frontend

import (
	"github.com/hooto/httpsrv"

	"github.com/hooto/htracker/config"
)

func NewModule() httpsrv.Module {

	module := httpsrv.NewModule("frontend")

	module.RouteSet(httpsrv.Route{
		Type:       httpsrv.RouteTypeStatic,
		Path:       "~",
		StaticPath: config.Prefix + "/webui",
	})

	module.ControllerRegister(new(Index))

	return module
}
