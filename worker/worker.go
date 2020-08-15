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
	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
)

func Start() error {

	var (
		offset = hapi.DataPathProjActiveEntry("z")
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

		rs := data.Data.NewReader(nil).KeyRangeSet([]byte(offset), []byte(cutset)).
			ModeRevRangeSet(true).LimitNumSet(int64(limit)).Query()
		if !rs.OK() {
			continue
		}

		for _, entry := range rs.Items {
			var set hapi.ProjEntry
			if err := entry.Decode(&set); err == nil {
				if err = projAction(set); err != nil {
					hlog.Printf("error", "proj/action %s, err %s",
						set.Id, err.Error())
				} else {
					// TODO
				}
			}
		}
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

	rs := data.Data.NewReader(nil).KeyRangeSet(offset, cutset).
		ModeRevRangeSet(true).LimitNumSet(1000).Query()
	if rs.OK() {
		item.ProcNum = 0
		for _, v := range rs.Items {
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
		data.Data.NewWriter([]byte(pkey), item).Commit()
	}

	hlog.Printf("debug", "proj/action hit %d", len(pids))

	return nil
}

func projProcEntrySync(proj_id string, entry *hapi.ProjProcEntry) error {

	pkey := hapi.DataPathProjProcHitEntry(
		proj_id, entry.Created, uint32(entry.Pid))

	var rs *kv2.ObjectResult

	if entry.Traced < 1 && entry.Exited < 1 {

		rs = data.Data.NewReader(pkey).Query()
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

		// Actions: skv.KvProgOpFoldMeta,
		rs = data.Data.NewWriter(pkeyt, entry).ExpireSet(hapi.DataExpired).Commit()

		hlog.Printf("debug", "Project/Process Exit %s, %d", entry.ProjId, entry.Pid)
		if !rs.OK() {
			return errors.New("database error")
		}
	}

	if entry.Exited > 0 {
		rs = data.Data.NewWriter(pkey, nil).ModeDeleteSet(true).Commit()
	} else {
		entry.ProjId = proj_id
		rs = data.Data.NewWriter(pkey, entry).ExpireSet(hapi.DataExpired).Commit()
		// Actions: skv.KvProgOpFoldMeta,

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

	rs := data.Data.NewReader(nil).KeyRangeSet(offset, cutset).
		ModeRevRangeSet(true).LimitNumSet(1000).Query()
	if rs.OK() {

		hlog.Printf("warn", "Project/Remove %s, N %d", item.Id, len(rs.Items))
		for _, vset := range rs.Items {

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
		rs = data.Data.NewWriter([]byte(key), nil).ModeDeleteSet(true).Commit()
	} else {

		item.Closed = tn

		rs = data.Data.NewWriter([]byte(key_history), item).Commit()
		if rs.OK() {
			rs = data.Data.NewWriter([]byte(key), nil).ModeDeleteSet(true).Commit()
		}
	}

	if !rs.OK() {
		return errors.New("Database Error")
	}

	return nil
}
