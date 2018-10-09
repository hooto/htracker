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
