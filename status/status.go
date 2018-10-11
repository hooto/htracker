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

package status

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/types"
	"github.com/shirou/gopsutil/process"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/hapi"
)

var (
	ProcList      hapi.ProjProcList
	mu            sync.Mutex
	procUpdated   uint32 = 0
	procPending          = false
	procSkipCmds         = []string{}
	procSkipNames        = types.ArrayString([]string{
		"perf",
		"burn",
		"sh",
		"bash",
	})
)

func init() {
	procSkipCmds = []string{
		"perf record",
		"perf script",
		fmt.Sprintf("%s/deps/", config.Prefix),
	}
}

func procSkipCmd(cmd string) bool {
	for _, sv := range procSkipCmds {
		if strings.Contains(cmd, sv) {
			return true
		}
	}
	return false
}

func ProcListRefresh() error {

	mu.Lock()
	if procPending {
		mu.Unlock()
		if len(ProcList.Items) < 1 {
			return fmt.Errorf("pending")
		}
		return nil
	}
	procPending = true
	mu.Unlock()

	defer func() {
		procUpdated = uint32(time.Now().Unix())
		procPending = false
	}()

	tn := time.Now()
	if uint32(tn.Unix())-procUpdated < 10 {
		return nil
	}

	plist, err := process.Processes()
	if err != nil {
		return err
	}

	var ok = false

	for _, v := range ProcList.Items {

		for _, v2 := range plist {
			if v.Pid == v2.Pid {
				ok = true
				break
			}
		}

		if !ok {
			ProcList.Del(v.Pid, v.Created)
		}

		ok = false
	}

	for _, p := range plist {

		if p.Pid < 300 {
			continue
		}

		var (
			name, _    = p.Name()
			cmd, _     = p.Cmdline()
			created, _ = p.CreateTime()
		)

		if procSkipNames.Has(name) || procSkipCmd(cmd) {
			hlog.Printf("debug", "Proc Skip %s, CMD %s", name, cmd)
			continue
		}

		ProcList.Set(&hapi.ProjProcEntry{
			Pid:     p.Pid,
			Created: uint32(created / 1e3),
			Name:    name,
			Cmd:     cmd,
			Process: p,
		})
	}

	hlog.Printf("debug", "Proc Refresh %d in %v",
		len(ProcList.Items), time.Since(tn))

	return nil
}
