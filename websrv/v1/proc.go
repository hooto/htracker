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

package v1

import (
	"sort"
	"strings"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/types"
	"github.com/shirou/gopsutil/process"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
)

var (
	plist hapi.ProcList
)

type Proc struct {
	*httpsrv.Controller
}

func (c *Proc) Init() int {

	if config.Config.Auth == "" {
		set := AuthSession{
			Action: AuthSessionInit,
		}
		set.Kind = "AuthSession"
		c.RenderJson(set)
		return 1
	}

	if sess := AuthSessionInstance(c.Session); sess == nil {
		c.Response.Out.WriteHeader(401)
		c.RenderJson(types.NewTypeErrorMeta("401", "Unauthorized"))
		return 1
	}

	return 0
}

func (c Proc) EntryAction() {

	var set hapi.ProcEntry
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

	set = hapi.ProcEntry{
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

	set.Kind = "ProcEntry"
}

func (c Proc) ListAction() {

	var sets hapi.ProcList
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

	err := status.ProcListRefresh()
	if err != nil {
		sets.Error = types.NewErrorMeta("500", "Server Pending")
		return
	}

	for _, p := range status.ProcList.Items {

		if p.Process == nil {
			continue
		}

		if len(q) > 0 &&
			!strings.Contains(p.Name, q) &&
			!strings.Contains(p.Cmd, q) {
			continue
		}

		var (
			user, _ = p.Process.Username()
		)

		if filter_user != "" &&
			filter_user != user {
			continue
		}

		var (
			cpup, _   = p.Process.CPUPercent()
			status, _ = p.Process.Status()
		)

		set := &hapi.ProcEntry{
			Pid:     p.Pid,
			Created: p.Created,
			Name:    p.Name,
			Cmd:     p.Cmd,
			CpuP:    hapi.Float64Round(cpup, 2),
			User:    user,
			Status:  status,
		}

		if memi, _ := p.Process.MemoryInfo(); memi != nil {
			set.MemRss = int64(memi.RSS)
		}

		sets.Items = append(sets.Items, set)
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

	sets.Total = len(sets.Items)

	if len(sets.Items) > limit {
		sets.Items = sets.Items[:limit]
	}
	sets.Num = len(sets.Items)

	sets.Updated = uint32(time.Now().Unix())

	hlog.Printf("debug", "get processes %d in %v",
		sets.Num, time.Since(stats_start))

	sets.Kind = "ProcList"
	plist = sets
}
