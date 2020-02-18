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
	"time"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

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
		if rs := data.Data.NewReader(pk).Query(); rs.OK() {
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
			data.Data.NewWriter(pk, prev).ExpireSet(hapi.DataExpired).Commit()
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
