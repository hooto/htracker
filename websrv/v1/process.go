package v1

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"

	"github.com/shirou/gopsutil/process"
)

type Process struct {
	*httpsrv.Controller
}

var (
	plist hapi.ProcessList
)

func (c Process) EntryAction() {

	var set hapi.ProcessEntry
	defer c.RenderJson(&set)

	var (
		pid = int32(c.Params.Int64("pid"))
	)

	p, err := process.NewProcess(pid)
	if err != nil {
		set.Error = types.NewErrorMeta("400", "Process Not Found "+err.Error())
		return
	}

	var (
		name, _    = p.Name()
		cmd, _     = p.Cmdline()
		user, _    = p.Username()
		created, _ = p.CreateTime()
		cpup, _    = p.CPUPercent()
		status, _  = p.Status()
	)

	set = hapi.ProcessEntry{
		Pid:     p.Pid,
		Created: uint32(created / 1e3),
		Name:    name,
		Cmd:     cmd,
		CpuP:    hapi.Float64Round(cpup, 2),
		User:    user,
		Status:  status,
	}

	if memi, _ := p.MemoryInfo(); memi != nil {
		set.MemRss = int64(memi.RSS)
	}

	set.Kind = "ProcessEntry"
}

func (c Process) ListAction() {

	var sets hapi.ProcessList
	defer c.RenderJson(&sets)

	var (
		q           = c.Params.Get("q")
		sort_by     = c.Params.Get("sort_by")
		limit       = int(c.Params.Int64("limit"))
		filter_user = c.Params.Get("filter_user")
		stats_start = time.Now()
		tn          = uint32(stats_start.Unix())
	)

	if sort_by != "cpu" && sort_by != "mem" {
		sort_by = "cpu"
	}

	if limit < 10 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if (tn-plist.Updated) < 3 && len(plist.Items) > 0 {
		sets = plist
		return
	}

	pls, err := process.Processes()
	if err != nil {
		return
	}
	for _, p := range pls {

		if p.Pid < 300 {
			continue
		}

		var (
			name, _ = p.Name()
			cmd, _  = p.Cmdline()
			user, _ = p.Username()
		)

		if len(q) > 0 &&
			!strings.Contains(name, q) &&
			!strings.Contains(cmd, q) {
			continue
		}

		if filter_user != "" &&
			filter_user != user {
			continue
		}

		var (
			created, _ = p.CreateTime()
			cpup, _    = p.CPUPercent()
			status, _  = p.Status()
		)

		set := &hapi.ProcessEntry{
			Pid:     p.Pid,
			Created: uint32(created / 1e3),
			Name:    name,
			Cmd:     cmd,
			CpuP:    hapi.Float64Round(cpup, 2),
			User:    user,
			Status:  status,
		}

		if memi, _ := p.MemoryInfo(); memi != nil {
			set.MemRss = int64(memi.RSS)
		}

		sets.Items = append(sets.Items, set)

		pcn := 0
		if pc, err := p.Children(); err == nil {
			pcn = len(pc)
		}

		if false {
			fmt.Println("pid", p.Pid, "pcn", pcn, "cpu", set.CpuP,
				"name", set.Name)
		}
	}

	if sort_by == "cpu" {
		sort.Slice(sets.Items, func(i, j int) bool {
			return sets.Items[i].CpuP > sets.Items[j].CpuP
		})
	} else {
		sort.Slice(sets.Items, func(i, j int) bool {
			return sets.Items[i].MemRss > sets.Items[j].MemRss
		})
	}

	if len(sets.Items) > limit {
		sets.Items = sets.Items[:limit]
	}

	sets.Num = len(sets.Items)
	sets.Updated = uint32(time.Now().Unix())

	hlog.Printf("debug", "get processes %d in %v",
		sets.Num, time.Since(stats_start))

	plist = sets
}
