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
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/lessos/lessgo/types"
	ps_cpu "github.com/shirou/gopsutil/cpu"
	ps_disk "github.com/shirou/gopsutil/disk"
	ps_host "github.com/shirou/gopsutil/host"
	ps_mem "github.com/shirou/gopsutil/mem"
	ps_net "github.com/shirou/gopsutil/net"
	"github.com/sysinner/incore/inutils"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
)

var (
	stats_podrep_names = []string{
		"ram/us", "ram/cc",
		"net/rs", "net/ws",
		"cpu/us",
		"fs/rn", "fs/rs", "fs/wn", "fs/ws",
	}
	stats_host_names = []string{
		"ram/us", "ram/cc",
		"net/rs", "net/ws",
		"cpu/p", "cpu/sys", "cpu/user",
		"fs/sp/rn", "fs/sp/rs", "fs/sp/wn", "fs/sp/ws",
	}
	sync_vols_last int64 = 0
)

func sysRefresh() {

	if status.Host.Spec.Platform == nil {

		os, arch, _ := inutils.ResSysHostEnvDistArch()

		status.Host.Spec.Platform = &hapi.ResPlatform{
			Os:     os,
			Arch:   arch,
			Kernel: inutils.ResSysHostKernel(),
		}
	}

	if status.Host.Spec.Capacity == nil {

		vm, _ := ps_mem.VirtualMemory()

		status.Host.Spec.Capacity = &hapi.ResHostResource{
			Cpu: uint64(runtime.NumCPU()),
			Mem: vm.Total,
		}
	}

	tn := time.Now()

	if len(status.Host.Status.Volumes) == 0 ||
		tn.Unix()-sync_vols_last > 600 {

		var (
			devs, _ = ps_disk.Partitions(false)
			vols    = []*hapi.ResHostVolume{}
		)

		sort.Slice(devs, func(i, j int) bool {
			if strings.Compare(devs[i].Device+devs[i].Mountpoint, devs[j].Device+devs[j].Device) < 0 {
				return true
			}
			return false
		})

		ars := types.ArrayString{}
		for _, dev := range devs {

			if ars.Has(dev.Device) {
				continue
			}
			ars.Set(dev.Device)

			if !strings.HasPrefix(dev.Device, "/dev/") ||
				strings.HasPrefix(dev.Mountpoint, "/boot") ||
				strings.Contains(dev.Mountpoint, "/devicemapper/mnt/") ||
				strings.HasPrefix(dev.Mountpoint, "/snap") {
				continue
			}

			if st, err := ps_disk.Usage(dev.Mountpoint); err == nil {
				vols = append(vols, &hapi.ResHostVolume{
					Name:  dev.Mountpoint,
					Total: st.Total,
					Used:  st.Used,
				})
			}
		}

		if len(vols) > 0 {

			sort.Slice(vols, func(i, j int) bool {
				if strings.Compare(vols[i].Name, vols[j].Name) < 0 {
					return true
				}
				return false
			})

			status.Host.Status.Volumes = vols
		}

		sync_vols_last = tn.Unix()
	}

	if status.Host.Status.Uptime < 1 {
		if tu, _ := ps_host.Uptime(); tu > 0 {
			status.Host.Status.Uptime = uint32(tn.Unix()) - uint32(tu)
		}
	}

	host_stats_refresh()
}

var (
	host_stats = hapi.NewPbStatsSampleFeed(hapi.BoxStatsSampleCycle)
)

func host_stats_refresh() {

	timo := uint32(time.Now().Unix())

	// RAM
	vm, _ := ps_mem.VirtualMemory()
	host_stats.SampleSync("ram/us", timo, int64(vm.Used))
	host_stats.SampleSync("ram/cc", timo, int64(vm.Cached))

	// Networks
	nio, _ := ps_net.IOCounters(false)
	if len(nio) > 0 {
		host_stats.SampleSync("net/rs", timo, int64(nio[0].BytesRecv))
		host_stats.SampleSync("net/ws", timo, int64(nio[0].BytesSent))
	}

	// CPU
	// cio, _ := ps_cpu.Times(false)
	cio, _ := ps_cpu.Percent(10e9, false)
	if len(cio) > 0 {
		// host_stats.SampleSync("cpu/sys", timo, int64(cio[0].User*float64(1e9)))
		// host_stats.SampleSync("cpu/user", timo, int64(cio[0].System*float64(1e9)))
		host_stats.SampleSync("cpu/p", timo, int64(cio[0]*100))
	}

	// Storage IO
	devs, _ := ps_disk.Partitions(false)
	if dev_name := disk_dev_name(devs, "/opt/"); dev_name != "" {
		if diom, err := ps_disk.IOCounters(dev_name); err == nil {
			if dio, ok := diom[dev_name]; ok {
				host_stats.SampleSync("fs/sp/rn", timo, int64(dio.ReadCount))
				host_stats.SampleSync("fs/sp/rs", timo, int64(dio.ReadBytes))
				host_stats.SampleSync("fs/sp/wn", timo, int64(dio.WriteCount))
				host_stats.SampleSync("fs/sp/ws", timo, int64(dio.WriteBytes))
			}
		}
	}

	//
	var (
		feed            = hapi.NewPbStatsSampleFeed(hapi.BoxStatsLogCycle)
		ec_time  uint32 = 0
		ec_value int64  = 0
	)
	for _, name := range stats_host_names {
		if ec_time, ec_value = host_stats.Extract(name, hapi.BoxStatsLogCycle, ec_time); ec_value >= 0 {
			feed.SampleSync(name, ec_time, ec_value)
		}
	}

	if len(feed.Items) < 1 {
		return
	}

	arrs := hapi.NewPbStatsIndexList(600, 60)
	for _, v := range feed.Items {
		for _, v2 := range v.Items {
			arrs.Sync(v.Name, v2.Time, v2.Value)
		}
	}

	for _, v := range arrs.Items {
		pk := hapi.DataSysHostStats(v.Time)

		var stats_index hapi.PbStatsIndexFeed
		if rs := data.Data.NewReader(pk).Query(); rs.OK() {
			rs.Decode(&stats_index)
			if stats_index.Time < 1 {
				continue
			}
		}

		stats_index.Time = v.Time
		for _, entry := range v.Items {
			for _, sv := range entry.Items {
				stats_index.Sync(entry.Name, sv.Time, sv.Value)
			}
		}

		if len(stats_index.Items) > 0 {
			data.Data.NewWriter(pk, stats_index).
				ExpireSet(hapi.DataExpired).Commit()
		}
	}
}

func disk_dev_name(pls []ps_disk.PartitionStat, path string) string {

	path = filepath.Clean(path)

	for {

		for _, v := range pls {
			if path == v.Mountpoint {
				if strings.HasPrefix(v.Device, "/dev/") {
					return v.Device[5:]
				}
				return ""
			}
		}

		if i := strings.LastIndex(path, "/"); i > 0 {
			path = path[:i]
		} else if len(path) > 1 && path[0] == '/' {
			path = "/"
		} else {
			break
		}
	}

	return ""
}
