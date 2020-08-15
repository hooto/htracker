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
	"time"

	"github.com/hooto/hlang4g/hlang"
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
		time.Sleep(200e6)
	}()

	if err := config.Setup(version, release); err != nil {
		hlog.Printf("fatal", "Config/Setup %s", err.Error())
		fmt.Println("Fatal Config/Setup :", err)
		os.Exit(1)
	}

	if err := data.Setup(); err != nil {
		hlog.Printf("fatal", "data/Setup %s", err.Error())
		fmt.Println("Fatal data/Setup :", err)
		os.Exit(1)
	}

	go worker.Start()

	var (
		quit = make(chan os.Signal, 2)
		hs   = httpsrv.NewService()
		v1m  = v1.NewModule()
	)

	// i18n
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/en.json", true)
	hlang.StdLangFeed.LoadMessages(config.Prefix+"/i18n/zh-CN.json", true)
	if hlang.StdLangFeed.Init() {
		v1m.ControllerRegister(new(hlang.Langsrv))
	}

	if config.Config.HttpBasepath != "" {
		hs.Config.UrlBasePath = config.Config.HttpBasepath
	}

	// register module to httpsrv
	hs.ModuleRegister("/htracker/v1", v1m)
	hs.ModuleRegister("/htracker", frontend.NewModule())
	hs.ModuleRegister("/", frontend.NewModule())

	// start
	hs.Config.HttpPort = config.Config.HttpPort
	go func() {
		if err := hs.Start(); err != nil {
			hlog.Printf("fatal", "Server/Start Fatal: %s", err.Error())
			fmt.Println("Fatal http/start :", err)
		}
		quit <- syscall.SIGQUIT
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
