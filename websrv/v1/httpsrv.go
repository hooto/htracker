package v1

import (
	"github.com/hooto/httpsrv"
)

func NewModule() httpsrv.Module {

	module := httpsrv.NewModule("api_v1")

	module.ControllerRegister(new(Process))
	module.ControllerRegister(new(Tracer))

	return module
}
