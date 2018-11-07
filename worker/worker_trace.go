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

package worker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lynkdb/iomix/skv"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

type traceCommandEntry struct {
	Command   string
	CleanFile string
	Done      bool
}

const (
	perfCmdRecord = "perf record -F 99 -g -p %d -o %s.data -- sleep %d"
	perfCmdUnfold = "perf script -f -i %s.data > %s.unfold"
	perfCmdJson   = "%s/bin/burn convert --output=%s.json %s.unfold"
	perfCmdSvg    = "%s/deps/FlameGraph/stackcollapse-perf.pl %s.unfold | %s/deps/FlameGraph/flamegraph.pl --title=' ' > %s.svg"
)

func projActionDyTraceCommands(pid int32, perfPrefix string, time_d uint32) []*traceCommandEntry {

	return []*traceCommandEntry{
		{
			fmt.Sprintf(perfCmdRecord, pid, perfPrefix, time_d),
			".data",
			false,
		},
		{
			fmt.Sprintf(perfCmdUnfold, perfPrefix, perfPrefix),
			".unfold",
			false,
		},
		// {
		// 	fmt.Sprintf(perfCmdJson, config.Prefix, perfPrefix, perfPrefix),
		// 	".json",
		// 	false,
		// },
		{
			fmt.Sprintf(perfCmdSvg, config.Prefix, perfPrefix, config.Prefix, perfPrefix),
			".svg",
			false,
		},
	}
}

func projActionDyTrace(proj hapi.ProjEntry, entry *hapi.ProjProcEntry) error {

	if procTraceNum > 20 {
		return nil
	}

	tn := uint32(time.Now().Unix())

	if entry.Process == nil || entry.Tracing != nil {
		return nil
	}

	var (
		opNext = false
		timer  = &hapi.ProjTraceOptionTimer{
			Interval: hapi.ProjTraceTimeIntervalDef,
			Duration: hapi.ProjTraceTimeDurationDef,
		}
	)

	if proj.TraceOptions != nil {

		if proj.TraceOptions.Overload != nil {
			if cpup, err := entry.Process.CPUPercent(); err == nil {
				if uint32(cpup) > proj.TraceOptions.Overload.Cpu {
					opNext = true
				}
			}
			timer.Interval = proj.TraceOptions.Overload.Interval
			timer.Duration = proj.TraceOptions.Overload.Duration

		} else if proj.TraceOptions.FixTimer != nil {
			timer = proj.TraceOptions.FixTimer
			opNext = true
		}
	} else {
		opNext = true
	}

	if entry.OpAction != hapi.ProjProcEntryOpTraceForce &&
		((entry.Traced+timer.Interval) >= tn || !opNext) {
		return nil
	}

	// hlog.Printf("info", "cpu %v, tn %d, traced %d, inter %d",
	// 	opNext, tn, entry.Traced, timer.Interval)

	entry.OpAction = 0
	entry.Tracing = &hapi.ProjProcTraceEntry{
		ProjId:   entry.ProjId,
		Pid:      entry.Pid,
		Pcreated: entry.Created,
		Created:  tn,
	}

	procTraceNum += 1

	tkey := hapi.DataPathProjProcTraceEntry(
		entry.ProjId, entry.Created, uint32(entry.Pid),
		entry.Tracing.Created,
	)

	data.Data.KvProgPut(
		tkey,
		skv.NewKvEntry(entry.Tracing),
		&skv.KvProgWriteOptions{
			Expired: hapi.DataExpired(),
			Actions: skv.KvProgOpFoldMeta,
		},
	)

	timer.Fix()

	go func(proj_id string, entry *hapi.ProjProcEntry, tkey skv.KvProgKey,
		timer_d uint32) {

		var (
			perfId = fmt.Sprintf("perf.%d.%d.%d",
				entry.Tracing.Created, entry.Created, entry.Pid)
			perfPrefix = fmt.Sprintf("%s/var/tmp/%s", config.Prefix, perfId)
			out        []byte
			err        error
		)

		// hlog.Printf("debug", "projActionDyTrace %s", perfId)

		cmds := projActionDyTraceCommands(entry.Pid, perfPrefix, timer_d)

		json_ok := true // FlameGraph to JSON in not stable right now

		for _, cmd := range cmds {
			out, err = exec.Command("/bin/bash", "-c", cmd.Command).Output()
			cmd.Done = true
			if cmd.CleanFile == ".data" {
				if fst, err := os.Stat(perfPrefix + cmd.CleanFile); err == nil {
					entry.Tracing.PerfSize = uint32(fst.Size())
				}
			}
			if err != nil {
				hlog.Printf("warn", "failed to exec %s, err %s, out %s, lfile %d, cmd %s",
					perfId, err.Error(), string(out), entry.Tracing.PerfSize, cmd.Command)
				if cmd.CleanFile == ".json" {
					json_ok, err = false, nil
				} else {
					break
				}
			}
			hlog.Printf("debug", "OK lfile %d, cmd %s", entry.Tracing.PerfSize, cmd.Command)
		}

		if err == nil {

			if false && json_ok {
				var gitem hapi.FlameGraphBurnProfile
				if err = json.DecodeFile(perfPrefix+".json", &gitem); err == nil {
					entry.Tracing.GraphBurn = &gitem
				}
			}

			if gfp, err := os.Open(perfPrefix + ".svg"); err == nil {

				fsts, _ := gfp.Stat()
				if fsts.Size() > 100 && fsts.Size() < 8*hapi.MB {
					if bs, err := ioutil.ReadAll(gfp); err == nil {
						entry.Tracing.GraphOnCPU = string(bs)
					}
				} else {
					hlog.Printf("error", "svg size %d, %s", fsts.Size(), perfPrefix+".svg")
				}

				gfp.Close()
			}
		}

		entry.Tracing.Updated = uint32(time.Now().Unix())
		entry.Traced = uint32(time.Now().Unix())

		if entry.Tracing.GraphOnCPU != "" {
			hlog.Printf("debug", "trace %s in %d s", perfId,
				(entry.Tracing.Updated - entry.Tracing.Created),
			)
		} else {
			hlog.Printf("error", "ERR trace %s", perfId)
			entry.Tracing.PerfSize = 0
		}

		data.Data.KvProgPut(
			tkey,
			skv.NewKvEntry(entry.Tracing),
			&skv.KvProgWriteOptions{
				Expired: hapi.DataExpired(),
				Actions: skv.KvProgOpFoldMeta,
			},
		)

		entry.Tracing = nil

		// projProcEntrySync(proj_id, entry)

		procTraceNum -= 1

		for _, cmd := range cmds {
			if cmd.Done && cmd.CleanFile != "" {
				os.Remove(perfPrefix + cmd.CleanFile)
			}
		}

	}(proj.Id, entry, tkey, timer.Duration)

	return nil
}
