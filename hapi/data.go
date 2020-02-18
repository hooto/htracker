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

package hapi

import (
	"fmt"
)

const (
	DataExpired int64 = 30 * 86400 * 1000 // days
)

func DataPathProjEntry(ptype, id string) []byte {
	return []byte(fmt.Sprintf("proj/%s/%s", ptype, id))
}

func DataPathProjActiveEntry(id string) []byte {
	return []byte(fmt.Sprintf("proj/active/%s", id))
}

func DataPathProjHistoryEntry(id string) []byte {
	return []byte(fmt.Sprintf("proj/hist/%s", id))
}

func DataPathUserSessionEntry(id string) []byte {
	return []byte("sess:" + id)
}

func DataPathProjProcEntry(
	ptype string,
	proj_id string,
	ptime uint32, pid uint32) []byte {

	if ptime < 1 {
		return []byte("proj:p:" + ptype + ":" + proj_id + ":z")
	}

	return []byte("proj:p:" + ptype + ":" + proj_id + ":" +
		Uint32ToHexString(ptime) + Uint32ToHexString(pid))
}

func DataPathProjProcHitEntry(
	proj_id string,
	ptime uint32, pid uint32) []byte {
	return DataPathProjProcEntry("hit", proj_id, ptime, pid)
}

func DataPathProjProcExitEntry(
	proj_id string,
	ptime uint32, pid uint32) []byte {
	return DataPathProjProcEntry("exit", proj_id, ptime, pid)
}

func DataPathProjProcStatsEntry(
	ptime uint32, pid uint32,
	created uint32) []byte {

	return []byte("pstats:" + Uint32ToHexString(ptime) +
		":" + Uint32ToHexString(pid) +
		":" + Uint32ToHexString(created))
}

func DataPathProjProcTraceEntry(
	proj_id string,
	ptime uint32, pid uint32,
	created uint32) []byte {

	if created < 1 {
		return []byte("ptrace:" + proj_id +
			":" + Uint32ToHexString(ptime) + Uint32ToHexString(pid) +
			":")
	}

	return []byte("ptrace:" + proj_id +
		":" + Uint32ToHexString(ptime) + Uint32ToHexString(pid) +
		":" + Uint32ToHexString(created))
}

func DataSysHostStats(timo uint32) []byte {
	return []byte("hoststats:" + Uint32ToHexString(timo))
}
