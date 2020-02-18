package hlog

import (
	"fmt"
	"sync"
	"time"
)

var (
	sitMu         sync.RWMutex
	statusInTimes = map[string]*statusInTimeItem{}
)

type statusInTimeItemStatusItem struct {
	status string
	count  int64
}

type statusInTimeItem struct {
	mu         sync.Mutex
	statuses   []*statusInTimeItemStatusItem
	timeCycle  int64
	timeOffset int64
}

func statusTimeItem(name string) *statusInTimeItem {

	sitMu.Lock()
	defer sitMu.Unlock()

	v, ok := statusInTimes[name]
	if ok {
		return v
	}

	return nil
}

func ExpStInit(name string, timeCycle int64, statuses []string) {

	sitMu.Lock()
	defer sitMu.Unlock()

	if timeCycle > 600 {
		timeCycle = 600
	} else if timeCycle < 60 {
		timeCycle = 60
	}

	statusInTimes[name] = &statusInTimeItem{
		timeCycle:  timeCycle,
		timeOffset: time.Now().Unix(),
	}
	for _, v := range statuses {
		statusInTimes[name].statuses = append(statusInTimes[name].statuses, &statusInTimeItemStatusItem{
			status: v,
		})
	}
}

func ExpStPrint(name string, status string) {

	item := statusTimeItem(name)
	if item == nil {
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
			logMsg := fmt.Sprintf("SIT %s (%ds)", name, item.timeCycle)
			for _, v := range item.statuses {
				logMsg += fmt.Sprintf(", %s %d %5.2f%%", v.status, v.count, float64(100*v.count)/float64(n))
				v.count = 0
			}
			newEntry(printDefault, "info", "", logMsg)
		}
		item.timeOffset = tn
	}

	for _, v := range item.statuses {
		if v.status == status {
			v.count += 1
			break
		}
	}
}
