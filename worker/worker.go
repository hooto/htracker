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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lynkdb/iomix/skv"
	"github.com/shirou/gopsutil/process"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

func Start() error {

	var (
		offset = hapi.DataPathTracerEntry("Z")
		cutset = hapi.DataPathTracerEntry("")
		limit  = 100
	)

	for {
		time.Sleep(10e9)

		rs := data.Data.KvRevScan([]byte(offset), []byte(cutset), limit)
		if !rs.OK() {
			continue
		}

		rs.KvEach(func(entry *skv.ResultEntry) int {
			var set hapi.TracerEntry
			if err := entry.Decode(&set); err == nil {
				if err = tracerAction(set); err != nil {
					hlog.Printf("error", "tracer/action %s, err %s",
						set.Id, err.Error())
				} else {
					// TODO
				}
			}
			return 0
		})
	}

	return nil
}

var (
	procStatsList hapi.TracerProcessList
	procTraceNum  = 0
)

func tracerAction(item hapi.TracerEntry) error {

	hlog.Printf("debug", "tracer/action %s",
		item.Id)

	ps := []*hapi.TracerProcessEntry{}

	if item.Filter.ProcId > 0 {

		p, err := process.NewProcess(item.Filter.ProcId)
		if err != nil {
			return tracerRemove(item)
		}

		created, _ := p.CreateTime()
		if uint32(created/1e3) != item.Filter.ProcCreated {
			return tracerRemove(item)
		}

		cmd, _ := p.Cmdline()

		ps = append(ps, &hapi.TracerProcessEntry{
			Tid:     item.Id,
			Pid:     p.Pid,
			Created: uint32(created / 1e3),
			Cmd:     cmd,
			Process: p,
		})
	} else if item.Filter.ProcName != "" {

		pls, _ := process.Processes()
		for _, p := range pls {

			if p.Pid < 300 {
				continue
			}

			name, _ := p.Name()
			if name != item.Filter.ProcName {
				continue
			}

			var (
				created, _ = p.CreateTime()
				cmd, _     = p.Cmdline()
			)

			ps = append(ps, &hapi.TracerProcessEntry{
				Tid:     item.Id,
				Pid:     p.Pid,
				Created: uint32(created / 1e3),
				Cmd:     cmd,
				Process: p,
			})
		}
	}

	for _, p := range ps {

		if p.Process == nil {
			continue
		}

		entry := procStatsList.Entry(p.Pid, p.Created)
		if entry == nil {

			entry = &hapi.TracerProcessEntry{
				Tid:             p.Tid,
				Pid:             p.Pid,
				Created:         p.Created,
				Cmd:             p.Cmd,
				StatsSampleFeed: hapi.NewPbStatsSampleFeed(20),
				Process:         p.Process,
			}

			procStatsList.Items = append(procStatsList.Items, entry)
		}

		if err := tracerActionStats(entry); err != nil {
			continue
		}

		tracerActionDyTrace(entry)
	}

	return nil
}

func tracerActionStats(entry *hapi.TracerProcessEntry) error {

	var (
		tn      = uint32(time.Now().Unix())
		updated = uint32(0)
	)

	// CPU
	if cpup, err := entry.Process.CPUPercent(); err == nil {
		entry.StatsSampleFeed.SampleSync("cpu/p", tn, int64(cpup*10000))
	}

	// Net
	if netConns, err := entry.Process.Connections(); err == nil {
		entry.StatsSampleFeed.SampleSync("net/c", tn, int64(len(netConns)))
	}
	if nioc, err := entry.Process.NetIOCounters(false); err == nil && len(nioc) == 1 {
		entry.StatsSampleFeed.SampleSync("net/rc", tn, int64(nioc[0].PacketsRecv))
		entry.StatsSampleFeed.SampleSync("net/rb", tn, int64(nioc[0].BytesRecv))
		entry.StatsSampleFeed.SampleSync("net/re", tn, int64(nioc[0].Errin))
		entry.StatsSampleFeed.SampleSync("net/wc", tn, int64(nioc[0].PacketsSent))
		entry.StatsSampleFeed.SampleSync("net/wb", tn, int64(nioc[0].BytesSent))
		entry.StatsSampleFeed.SampleSync("net/we", tn, int64(nioc[0].Errout))
	}

	// IO
	if ioc, err := entry.Process.IOCounters(); err == nil {
		entry.StatsSampleFeed.SampleSync("io/rc", tn, int64(ioc.ReadCount))
		entry.StatsSampleFeed.SampleSync("io/rb", tn, int64(ioc.ReadBytes))
		entry.StatsSampleFeed.SampleSync("io/wc", tn, int64(ioc.WriteCount))
		entry.StatsSampleFeed.SampleSync("io/wb", tn, int64(ioc.WriteBytes))
	}

	// IO/Fds
	if nfd, err := entry.Process.NumFDs(); err == nil {
		entry.StatsSampleFeed.SampleSync("io/fd", tn, int64(nfd))
	}
	if ntd, err := entry.Process.NumThreads(); err == nil {
		entry.StatsSampleFeed.SampleSync("io/td", tn, int64(ntd))
	}

	// Memory
	if mis, err := entry.Process.MemoryInfo(); err == nil {
		entry.StatsSampleFeed.SampleSync("mem/rss", tn, int64(mis.RSS))
		entry.StatsSampleFeed.SampleSync("mem/vms", tn, int64(mis.VMS))
		entry.StatsSampleFeed.SampleSync("mem/data", tn, int64(mis.Data))
	}

	// hapi.ObjPrint("entry.StatsSampleFeed", entry.StatsSampleFeed)

	var (
		feed            = hapi.NewPbStatsSampleFeed(60)
		ec_time  uint32 = 0
		ec_value int64  = 0
	)
	for _, name := range []string{
		"cpu/p",
		"net/c",
		"net/rc",
		"net/rb",
		"net/re",
		"net/wc",
		"net/wb",
		"net/we",
		"io/rc",
		"io/rb",
		"io/wc",
		"io/wb",
		"io/fd",
		"io/td",
		"mem/rss",
		"mem/vms",
		"mem/data",
	} {

		if ec_time, ec_value = entry.StatsSampleFeed.Extract(name, 60, ec_time); ec_value >= 0 {
			feed.SampleSync(name, ec_time, ec_value)
		}
	}

	if len(feed.Items) < 1 {
		return nil
	}

	arrs := hapi.NewPbStatsIndexList(600, 60)
	for _, v := range feed.Items {
		for _, v2 := range v.Items {
			arrs.Sync(v.Name, v2.Time, v2.Value)
		}
	}

	// hapi.ObjPrint("entry.StatsIndex", arrs)

	for _, v := range arrs.Items {

		pk := hapi.DataPathTracerProcessStatsEntry(
			entry.Created, uint32(entry.Pid),
			v.Time)

		var prev hapi.PbStatsIndexFeed
		if rs := data.Data.KvProgGet(pk); rs.OK() {
			rs.Decode(&prev)
			if prev.Time < 1 {
				continue
			}
		}

		prev.Time = v.Time
		for _, entry := range v.Items {
			for _, sv := range entry.Items {
				prev.Sync(entry.Name, sv.Time, sv.Value)
			}
		}

		if len(prev.Items) > 0 {
			data.Data.KvProgPut(
				pk,
				skv.NewKvEntry(prev),
				&skv.KvProgWriteOptions{
					Expired: uint64(time.Now().AddDate(0, 0, 10).UnixNano()),
					Actions: skv.KvProgOpFoldMeta,
				},
			)
			// hapi.ObjPrint(pk, prev)
			updated = tn
		}
	}

	if updated > 0 {
		entry.Updated = updated
		tracerActionEntrySync(entry)
		// hapi.ObjPrint(pkey, entry)
	}

	return nil
}

func tracerActionEntrySync(entry *hapi.TracerProcessEntry) error {

	pkey := hapi.DataPathTracerProcessEntry(
		entry.Tid, entry.Created, uint32(entry.Pid))

	data.Data.KvProgPut(
		pkey,
		skv.NewKvEntry(entry),
		&skv.KvProgWriteOptions{
			Expired: uint64(time.Now().AddDate(0, 0, 10).UnixNano()),
			Actions: skv.KvProgOpFoldMeta,
		},
	)

	return nil
}

func tracerActionDyTrace(entry *hapi.TracerProcessEntry) error {

	if procTraceNum > 20 {
		return nil
	}

	tn := uint32(time.Now().Unix())

	if entry.Tracing != nil ||
		(tn-entry.Traced) < 1800 {
		return nil
	}

	entry.Tracing = &hapi.TracerProcessTraceEntry{
		Created: tn,
	}

	procTraceNum += 1

	go func(entry *hapi.TracerProcessEntry) {

		var (
			perf_id = fmt.Sprintf("perf.%d.%d.%d",
				entry.Tracing.Created, entry.Created, entry.Pid)
			perf_tmp   = fmt.Sprintf("%s/var/tmp/%s", config.Prefix, perf_id)
			perf_oncpu = fmt.Sprintf("%s/var/tmp/%s.svg", config.Prefix, perf_id)
			perf_json  = fmt.Sprintf("%s/var/tmp/%s.json", config.Prefix, perf_id)
			out        []byte
			err        error
			time_in    = 30
		)

		cmds := []string{
			fmt.Sprintf("perf record -F 99 -g -p %d -o %s.data -- sleep %d",
				entry.Pid, perf_tmp, time_in),
			fmt.Sprintf("perf script -f -i %s.data > %s.unfold",
				perf_tmp, perf_tmp),
			fmt.Sprintf("%s/bin/burn convert --output=%s %s.unfold",
				config.Prefix, perf_json, perf_tmp),
			fmt.Sprintf("%s/deps/FlameGraph/stackcollapse-perf.pl %s.unfold | %s/deps/FlameGraph/flamegraph.pl --title=' ' > %s",
				config.Prefix, perf_tmp,
				config.Prefix, perf_oncpu,
			),
		}

		for _, cmd := range cmds {
			out, err = exec.Command("/bin/sh", "-c", cmd).Output()
			if err != nil {
				hlog.Printf("error", "failed to trace %s, err %s, out %s,cmd %s",
					perf_id, err.Error(), string(out), cmd)
				break
			}
			hlog.Printf("debug", "OK %s", cmd)
		}

		if err == nil {

			if fst, err := os.Stat(perf_tmp + ".data"); err == nil {
				entry.Tracing.PerfSize = uint32(fst.Size())
			}

			var gitem hapi.FlameGraphBurnProfile
			if err = json.DecodeFile(perf_json, &gitem); err == nil {

				entry.Tracing.GraphBurn = &gitem

				gfp, err := os.Open(perf_oncpu)
				if err == nil {

					fsts, _ := gfp.Stat()
					if fsts.Size() > 100 && fsts.Size() < 8*hapi.MB {
						if bs, err := ioutil.ReadAll(gfp); err == nil {
							entry.Tracing.GraphOnCPU = string(bs)
							// os.Remove(perf_oncpu)
						}
					} else {
						hlog.Printf("error", "svg size %d, %s", fsts.Size(), perf_oncpu)
					}

					gfp.Close()
				}

				entry.Tracing.Tid = entry.Tid
				entry.Tracing.Pid = entry.Pid
				entry.Tracing.Pcreated = entry.Created

				entry.Tracing.Updated = uint32(time.Now().Unix())

				pkey := hapi.DataPathTracerProcessTraceEntry(
					entry.Tid, entry.Created, uint32(entry.Pid),
					entry.Tracing.Created,
				)

				data.Data.KvProgPut(
					pkey,
					skv.NewKvEntry(entry.Tracing),
					&skv.KvProgWriteOptions{
						Expired: uint64(time.Now().AddDate(0, 0, 10).UnixNano()),
						Actions: skv.KvProgOpFoldMeta,
					},
				)

				hlog.Printf("info", "trace %s in %d s", perf_id,
					entry.Tracing.Updated-entry.Tracing.Created,
				)

				entry.Traced = uint32(time.Now().Unix())
			}
		}

		entry.Tracing = nil

		tracerActionEntrySync(entry)

		procTraceNum -= 1

		os.Remove(perf_tmp + ".data")
		os.Remove(perf_tmp + ".unfold")
		os.Remove(perf_tmp + ".json")
		os.Remove(perf_tmp + ".svg")

	}(entry)

	return nil
}

func tracerRemove(item hapi.TracerEntry) error {

	var (
		key         = hapi.DataPathTracerEntry(item.Id)
		key_history = hapi.DataPathTracerEntryHistory(item.Id)
	)
	item.Closed = uint32(time.Now().Unix())

	rs := data.Data.KvPut([]byte(key_history), item, nil)
	if rs.OK() {
		rs = data.Data.KvDel([]byte(key))
	}

	if !rs.OK() {
		return errors.New("Database Error")
	}

	return nil
}
