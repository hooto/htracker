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
	"strings"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
)

var (
	ttlLimit int = 30 // days
)

func Start() error {

	var (
		offset = hapi.DataPathProjActiveEntry("Z")
		cutset = hapi.DataPathProjActiveEntry("")
		limit  = 100
		inited = false
	)

	for {
		if inited {
			time.Sleep(10e9)
		} else {
			inited = true
		}

		if err := status.ProcListRefresh(); err != nil {
			hlog.Printf("warn", "proj/proc/refresh err %s", err.Error())
			continue
		}

		rs := data.Data.KvRevScan([]byte(offset), []byte(cutset), limit)
		if !rs.OK() {
			continue
		}

		rs.KvEach(func(entry *skv.ResultEntry) int {
			var set hapi.ProjEntry
			if err := entry.Decode(&set); err == nil {
				if err = projAction(set); err != nil {
					hlog.Printf("error", "proj/action %s, err %s",
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
	procTraceNum = 0
)

func projAction(item hapi.ProjEntry) error {

	hlog.Printf("debug", "proj/action project %s", item.Id)

	if item.Closed > 0 {
		return projRemove(item)
	}

	pids := types.ArrayUint32{}

	if item.Closed < 1 {

		if item.Filter.ProcId > 0 {

			p := status.ProcList.Entry(item.Filter.ProcId, 0)
			if p == nil {
				return projRemove(item)
			}

			if p.Created != item.Filter.ProcCreated {
				return projRemove(item)
			}

			pids = append(pids, uint32(p.Pid))

		} else if item.Filter.ProcName != "" {

			for _, p := range status.ProcList.Items {
				if p.Name == item.Filter.ProcName {
					pids = append(pids, uint32(p.Pid))
				}
			}

		} else if item.Filter.ProcCommand != "" {

			for _, p := range status.ProcList.Items {
				if strings.Contains(p.Cmd, item.Filter.ProcCommand) {
					pids = append(pids, uint32(p.Pid))
				}
			}
		}
	}

	for _, pid := range pids {

		entry := status.ProcList.Entry(int32(pid), 0)
		if entry == nil {
			continue
		}

		if err := projActionStats(item.Id, entry); err != nil {
			continue
		}

		projActionDyTrace(item.Id, entry)
	}

	var (
		offset = hapi.DataPathProjProcHitEntry(item.Id, 0, 0)
		cutset = hapi.DataPathProjProcHitEntry(item.Id, 1, 0)
		tn     = uint32(time.Now().Unix())
	)

	rs := data.Data.KvProgRevScan(offset, cutset, 1000)
	if rs.OK() {

		rs.KvEach(func(entry *skv.ResultEntry) int {
			var set hapi.ProjProcEntry
			if err := entry.Decode(&set); err == nil {
				if !pids.Has(uint32(set.Pid)) {
					set.Exited = tn
					projProcEntrySync(item.Id, &set)
				}
			}
			return 0
		})
	}

	hlog.Printf("debug", "proj/action hit %d", len(pids))

	return nil
}

func projActionStats(proj_id string, entry *hapi.ProjProcEntry) error {

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

		pk := hapi.DataPathProjProcStatsEntry(
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
					Expired: uint64(time.Now().AddDate(0, 0, ttlLimit).UnixNano()),
					Actions: skv.KvProgOpFoldMeta,
				},
			)
			// hapi.ObjPrint(pk, prev)
			updated = tn
		}
	}

	if updated > 0 {
		entry.Updated = updated
		projProcEntrySync(proj_id, entry)
		// hapi.ObjPrint(pkey, entry)
	}

	return nil
}

func projProcEntrySync(proj_id string, entry *hapi.ProjProcEntry) error {

	var rs skv.Result

	if entry.Exited > 0 {
		pkey := hapi.DataPathProjProcExitEntry(
			proj_id, entry.Created, uint32(entry.Pid))
		entry.ProjId = proj_id

		rs = data.Data.KvProgPut(
			pkey,
			skv.NewKvEntry(entry),
			&skv.KvProgWriteOptions{
				Expired: uint64(time.Now().AddDate(0, 0, ttlLimit).UnixNano()),
				Actions: skv.KvProgOpFoldMeta,
			},
		)
		if !rs.OK() {
			return errors.New("database error")
		}
		hlog.Printf("debug", "Project/Process Exit %s, %d", entry.ProjId, entry.Pid)
	}

	pkey := hapi.DataPathProjProcHitEntry(
		proj_id, entry.Created, uint32(entry.Pid))

	if entry.Exited > 0 {
		rs = data.Data.KvProgDel(pkey, nil)
	} else {
		entry.ProjId = proj_id
		rs = data.Data.KvProgPut(
			pkey,
			skv.NewKvEntry(entry),
			&skv.KvProgWriteOptions{
				Expired: uint64(time.Now().AddDate(0, 0, ttlLimit).UnixNano()),
				Actions: skv.KvProgOpFoldMeta,
			},
		)
		hlog.Printf("debug", "Project/Process Put %s, %d", entry.ProjId, entry.Pid)
	}

	if rs.OK() {
		return nil
	}
	return errors.New("database error")
}

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

const (
	perfRecordTime     = 120
	perfRecordRangeSec = uint32(1200)
)

func projActionDyTraceCommands(pid int32, perfPrefix string) []*traceCommandEntry {

	return []*traceCommandEntry{
		{
			fmt.Sprintf(perfCmdRecord, pid, perfPrefix, perfRecordTime),
			".data",
			false,
		},
		{
			fmt.Sprintf(perfCmdUnfold, perfPrefix, perfPrefix),
			".unfold",
			false,
		},
		{
			fmt.Sprintf(perfCmdJson, config.Prefix, perfPrefix, perfPrefix),
			".json",
			false,
		},
		{
			fmt.Sprintf(perfCmdSvg, config.Prefix, perfPrefix, config.Prefix, perfPrefix),
			".svg",
			false,
		},
	}
}

func projActionDyTrace(proj_id string, entry *hapi.ProjProcEntry) error {

	if procTraceNum > 20 {
		return nil
	}

	tn := uint32(time.Now().Unix())

	if entry.Tracing != nil ||
		(tn-entry.Traced) < perfRecordRangeSec {
		return nil
	}

	entry.Tracing = &hapi.ProjProcTraceEntry{
		Created: tn,
	}

	procTraceNum += 1

	go func(proj_id string, entry *hapi.ProjProcEntry) {

		var (
			perfId = fmt.Sprintf("perf.%d.%d.%d",
				entry.Tracing.Created, entry.Created, entry.Pid)
			perfPrefix = fmt.Sprintf("%s/var/tmp/%s", config.Prefix, perfId)
			out        []byte
			err        error
		)

		// hlog.Printf("debug", "projActionDyTrace %s", perfId)

		cmds := projActionDyTraceCommands(entry.Pid, perfPrefix)

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

			if json_ok {
				var gitem hapi.FlameGraphBurnProfile
				if err = json.DecodeFile(perfPrefix+".json", &gitem); err == nil {
					entry.Tracing.GraphBurn = &gitem
				}
			}

			if err == nil {

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

			if entry.Tracing.GraphOnCPU != "" {

				entry.Tracing.ProjId = entry.ProjId
				entry.Tracing.Pid = entry.Pid
				entry.Tracing.Pcreated = entry.Created

				entry.Tracing.Updated = uint32(time.Now().Unix())

				pkey := hapi.DataPathProjProcTraceEntry(
					entry.ProjId, entry.Created, uint32(entry.Pid),
					entry.Tracing.Created,
				)

				data.Data.KvProgPut(
					pkey,
					skv.NewKvEntry(entry.Tracing),
					&skv.KvProgWriteOptions{
						Expired: uint64(time.Now().AddDate(0, 0, ttlLimit).UnixNano()),
						Actions: skv.KvProgOpFoldMeta,
					},
				)

				entry.Traced = uint32(time.Now().Unix())

				hlog.Printf("debug", "trace %s in %d s", perfId,
					entry.Tracing.Updated-entry.Tracing.Created,
				)

			} else {
				hlog.Printf("error", "ERR trace %s", perfId)
			}
		}

		entry.Tracing = nil

		projProcEntrySync(proj_id, entry)

		procTraceNum -= 1

		for _, cmd := range cmds {
			if cmd.Done && cmd.CleanFile != "" {
				os.Remove(perfPrefix + cmd.CleanFile)
			}
		}

	}(proj_id, entry)

	return nil
}

func projRemove(item hapi.ProjEntry) error {

	var (
		offset = hapi.DataPathProjProcHitEntry(item.Id, 0, 0)
		cutset = hapi.DataPathProjProcHitEntry(item.Id, 1, 0)
		tn     = uint32(time.Now().Unix())
	)

	rs := data.Data.KvProgRevScan(offset, cutset, 1000)
	if rs.OK() {

		rss := rs.KvList()

		hlog.Printf("warn", "Project/Remove %s, N %d", item.Id, len(rss))
		for _, vset := range rss {

			var set hapi.ProjProcEntry
			if err := vset.Decode(&set); err != nil {
				return err
			}

			set.Exited = tn
			if err := projProcEntrySync(item.Id, &set); err != nil {
				return err
			}
		}
	}

	var (
		key         = hapi.DataPathProjActiveEntry(item.Id)
		key_history = hapi.DataPathProjHistoryEntry(item.Id)
	)

	if item.Closed > 0 {
		rs = data.Data.KvDel([]byte(key))
	} else {

		item.Closed = tn

		rs := data.Data.KvPut([]byte(key_history), item, nil)
		if rs.OK() {
			rs = data.Data.KvDel([]byte(key))
		}
	}

	if !rs.OK() {
		return errors.New("Database Error")
	}

	return nil
}
