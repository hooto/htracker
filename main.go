package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/hooto/hflag4g/hflag"
	"github.com/hooto/httpsrv"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/websrv/frontend"
	"github.com/hooto/htracker/websrv/v1"
	"github.com/hooto/htracker/worker"
)

var (
	err error
)

func main() {

	if err := config.Setup(); err != nil {
		fmt.Println(err)
		return
	}

	if err := data.Setup(); err != nil {
		fmt.Println(err)
		return
	}

	go worker.Start()

	hs := httpsrv.NewService()

	// register module to httpsrv
	hs.ModuleRegister("/htracker/v1", v1.NewModule())
	hs.ModuleRegister("/htracker/~/hchart",
		httpsrv.NewStaticModule("hchart_ui", config.Prefix+"/webui/hchart/webui"))
	hs.ModuleRegister("/htracker", frontend.NewModule())
	hs.ModuleRegister("/", frontend.NewModule())

	// listening on port 8060
	hs.Config.HttpPort = 8060

	// start
	if err := hs.Start(); err != nil {
		panic(err)
	}

	fmt.Println("running")
	//
	pid, ok := hflag.Value("pid")
	if !ok {
		fmt.Println("--pid not found")
		return
	}
	if pid.Int() < 1 {
		fmt.Println("--pid not valid")
		return
	}

	//
	time_in := 30
	if v, ok := hflag.Value("time_in"); ok {
		time_in = v.Int()
		if time_in < 30 {
			time_in = 30
		} else if time_in > 3600 {
			time_in = 3600
		}
	}

	qid := fmt.Sprintf("%d", time.Now().Unix())
	os.MkdirAll("var/outputs", 0755)
	cmds := []string{
		fmt.Sprintf("perf record -F 99 -g -p %d -o var/tmp/perf.%s.data -- sleep %d",
			pid.Int(), qid, time_in),
		fmt.Sprintf("perf script -f -i var/tmp/perf.%s.data &> var/tmp/perf.%s.unfold", qid, qid),
		fmt.Sprintf("perf script -f | deps/FlameGraph/stackcollapse-perf.pl var/tmp/perf.%s.unfold | deps/FlameGraph/flamegraph.pl > var/outputs/%s.svg",
			qid, qid,
		),
	}
	for i, cmd := range cmds {
		out, err := exec.Command("/bin/sh", "-c", cmd).Output()
		if err != nil {
			fmt.Println(i, cmd, err, string(out))
			return
		}
		fmt.Println(i, cmd, "DONE")
	}

	fmt.Println("OK", fmt.Sprintf("/htracker/~/%s.svg",
		qid))

	select {}
}
