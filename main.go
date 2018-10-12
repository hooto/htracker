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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/websrv/frontend"
	"github.com/hooto/htracker/websrv/v1"
	"github.com/hooto/htracker/worker"
)

var (
	version = ""
	release = ""
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			hlog.Printf("fatal", "Server/Panic %s", err)
		}
		hlog.Flush()
	}()

	if err := config.Setup(version, release); err != nil {
		fmt.Println(err)
		return
	}

	if err := data.Setup(); err != nil {
		fmt.Println(err)
		return
	}

	go worker.Start()

	var (
		hs   = httpsrv.NewService()
		quit = make(chan os.Signal, 2)
	)

	// register module to httpsrv
	hs.ModuleRegister("/htracker/v1", v1.NewModule())
	hs.ModuleRegister("/htracker", frontend.NewModule())
	hs.ModuleRegister("/", frontend.NewModule())

	// start
	hs.Config.HttpPort = config.Config.HttpPort
	go func() {
		if err := hs.Start(); err != nil {
			hlog.Printf("fatal", "Server/Start Fatal: %s", err.Error())
			quit <- syscall.SIGQUIT
		}
	}()

	//
	signal.Notify(quit,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL)
	sg := <-quit

	data.Data.Close()

	hlog.Printf("warn", "Signal Quit: %s", sg.String())
}
