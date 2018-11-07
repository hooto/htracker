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
	"strings"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
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
			time.Sleep(3e9)
		} else {
			inited = true
		}

		if err := status.ProcListRefresh(); err != nil {
			hlog.Printf("warn", "proj/proc/refresh err %s", err.Error())
			continue
		}

		sysAction()

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

		projActionDyTrace(item, entry)
	}

	var (
		offset = hapi.DataPathProjProcHitEntry(item.Id, 0, 0)
		cutset = hapi.DataPathProjProcHitEntry(item.Id, 1, 0)
		tn     = uint32(time.Now().Unix())
	)

	rs := data.Data.KvProgRevScan(offset, cutset, 1000)
	if rs.OK() {
		rss := rs.KvList()
		item.ProcNum = 0
		for _, v := range rss {
			var set hapi.ProjProcEntry
			if err := v.Decode(&set); err != nil {
				continue
			}
			if !pids.Has(uint32(set.Pid)) {
				set.Exited = tn
			} else {
				entry := status.ProcList.Entry(int32(set.Pid), 0)
				if entry != nil && entry.Created != set.Created {
					set.Exited = tn
				}
			}
			if set.Exited > 0 {
				projProcEntrySync(item.Id, &set)
				continue
			}

			item.ProcNum += 1
		}
		pkey := hapi.DataPathProjActiveEntry(item.Id)
		data.Data.KvPut([]byte(pkey), item, nil)
	}

	hlog.Printf("debug", "proj/action hit %d", len(pids))

	return nil
}

func projProcEntrySync(proj_id string, entry *hapi.ProjProcEntry) error {

	var rs skv.Result

	pkey := hapi.DataPathProjProcHitEntry(
		proj_id, entry.Created, uint32(entry.Pid))

	if entry.Traced < 1 && entry.Exited < 1 {

		rs = data.Data.KvProgGet(pkey)
		if rs.OK() {
			var prev hapi.ProjProcEntry
			if err := rs.Decode(&prev); err == nil && prev.Traced > 0 {
				entry.Traced = prev.Traced
			}
		}
	}

	if entry.Exited > 0 {
		pkeyt := hapi.DataPathProjProcExitEntry(
			proj_id, entry.Created, uint32(entry.Pid))
		entry.ProjId = proj_id

		rs = data.Data.KvProgPut(
			pkeyt,
			skv.NewKvEntry(entry),
			&skv.KvProgWriteOptions{
				Expired: hapi.DataExpired(),
				Actions: skv.KvProgOpFoldMeta,
			},
		)
		hlog.Printf("debug", "Project/Process Exit %s, %d", entry.ProjId, entry.Pid)
		if !rs.OK() {
			return errors.New("database error")
		}
	}

	if entry.Exited > 0 {
		rs = data.Data.KvProgDel(pkey, nil)
	} else {
		entry.ProjId = proj_id
		rs = data.Data.KvProgPut(
			pkey,
			skv.NewKvEntry(entry),
			&skv.KvProgWriteOptions{
				Expired: hapi.DataExpired(),
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

		rs = data.Data.KvPut([]byte(key_history), item, nil)
		if rs.OK() {
			rs = data.Data.KvDel([]byte(key))
		}
	}

	if !rs.OK() {
		return errors.New("Database Error")
	}

	return nil
}
