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

	"github.com/lynkdb/iomix/skv"
)

func DataPathTracerEntry(id string) string {
	return fmt.Sprintf("tracer/active/%s", id)
}

func DataPathTracerEntryHistory(id string) string {
	return fmt.Sprintf("tracer/hist/%s", id)
}

func DataPathTracerProcessEntry(
	tracer_id string,
	ptime uint32, pid uint32) skv.KvProgKey {

	if ptime < 1 {
		return skv.NewKvProgKey(
			"tracer:proc",
			tracer_id,
			"")
	}

	return skv.NewKvProgKey(
		"tracer:proc",
		tracer_id,
		Uint32ToHexString(ptime)+Uint32ToHexString(pid))
}

func DataPathTracerProcessStatsEntry(
	ptime uint32, pid uint32,
	created uint32) skv.KvProgKey {

	return skv.NewKvProgKey(
		"pstats",
		Uint32ToHexString(ptime)+Uint32ToHexString(pid),
		Uint32ToHexString(created))
}

func DataPathTracerProcessTraceEntry(
	tracer_id string,
	ptime uint32, pid uint32,
	created uint32) skv.KvProgKey {

	if created < 1 {
		return skv.NewKvProgKey(
			"ptrace",
			tracer_id,
			Uint32ToHexString(ptime)+Uint32ToHexString(pid),
			"")
	}

	return skv.NewKvProgKey(
		"ptrace",
		tracer_id,
		Uint32ToHexString(ptime)+Uint32ToHexString(pid),
		Uint32ToHexString(created))
}
