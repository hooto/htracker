package hlog

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	ditMu            sync.RWMutex
	durationInCycles = map[string]*durationInCycleItem{}
)

type durationInCycleItemStatusItem struct {
	dur   int64
	count int64
}

type durationInCycleItem struct {
	mu         sync.Mutex
	statuses   []*durationInCycleItemStatusItem
	timeCycle  int64
	timeOffset int64
}

func durationInCycleItemGet(name string) *durationInCycleItem {

	ditMu.Lock()
	defer ditMu.Unlock()

	v, ok := durationInCycles[name]
	if ok {
		return v
	}

	return nil
}

func ExpDtInit(name string, timeCycle int64, durations []int64) {

	ditMu.Lock()
	defer ditMu.Unlock()

	if timeCycle > 600 {
		timeCycle = 600
	} else if timeCycle < 60 {
		timeCycle = 60
	}

	durationInCycles[name] = &durationInCycleItem{
		timeCycle:  timeCycle,
		timeOffset: time.Now().Unix(),
	}

	for _, v := range durations {
		durationInCycles[name].statuses = append(durationInCycles[name].statuses, &durationInCycleItemStatusItem{
			dur: v,
		})
	}

	sort.Slice(durationInCycles[name].statuses, func(i, j int) bool {
		return durationInCycles[name].statuses[i].dur < durationInCycles[name].statuses[i].dur
	})
}

func ExpDtPrint(name string, dur int64) {

	if dur < 0 {
		return
	}

	item := durationInCycleItemGet(name)
	if item == nil || len(item.statuses) < 1 {
		return
	}

	item.mu.Lock()
	defer item.mu.Unlock()

	tn := time.Now().Unix()

	if tn-item.timeCycle >= item.timeOffset {
		n := int64(0)
		for _, v := range item.statuses {
			n += v.count
		}
		if n > 0 {
			logMsg := fmt.Sprintf("DT %s (%ds)", name, item.timeCycle)
			for _, v := range item.statuses {
				logMsg += fmt.Sprintf(", ~%dms %d %5.2f%%", v.dur, v.count, float64(100*v.count)/float64(n))
				v.count = 0
			}
			newEntry(printDefault, "info", "", logMsg)
		}
		item.timeOffset = tn
	}

	for _, v := range item.statuses {
		if dur <= v.dur {
			v.count += 1
			name = ""
			break
		}
	}

	if name != "" {
		item.statuses[len(item.statuses)-1].count += 1
	}
}
