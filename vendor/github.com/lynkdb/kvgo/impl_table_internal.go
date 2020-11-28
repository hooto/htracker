// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package kvgo

import (
	"strconv"
	"time"

	"github.com/hooto/hlog4g/hlog"
)

func (tdb *dbTable) objectLogVersionSet(incr, set, updated uint64) (uint64, error) {

	tdb.logMu.Lock()
	defer tdb.logMu.Unlock()

	if incr == 0 && set == 0 {
		return tdb.logOffset, nil
	}

	if tdb.logCutset <= 100 {

		if bs, err := tdb.db.Get(keySysLogCutset, nil); err != nil {
			if err.Error() != ldbNotFound {
				return 0, err
			}
		} else {
			if tdb.logCutset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return 0, err
			}
			if tdb.logOffset < tdb.logCutset {
				tdb.logOffset = tdb.logCutset
			}
		}
	}

	if tdb.logOffset < 100 {
		tdb.logOffset = 100
	}

	if set > 0 && set > tdb.logOffset {
		incr += (set - tdb.logOffset)
	}

	if incr > 0 {

		if (tdb.logOffset + incr) >= tdb.logCutset {

			cutset := tdb.logOffset + incr + 100

			if n := cutset % 100; n > 0 {
				cutset += n
			}

			if err := tdb.db.Put(keySysLogCutset,
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			hlog.Printf("debug", "table %s, reset log-version to %d~%d",
				tdb.tableName, tdb.logOffset+incr, cutset)

			tdb.logCutset = cutset
		}

		tdb.logOffset += incr

		if updated > 0 {
			tdb.logLockSets[tdb.logOffset] = updated
		}
	}

	return tdb.logOffset, nil
}

func (tdb *dbTable) objectLogFree(logId uint64) {
	tdb.logMu.Lock()
	defer tdb.logMu.Unlock()
	delete(tdb.logLockSets, logId)
}

func (tdb *dbTable) objectLogDelay() uint64 {
	tdb.logMu.Lock()
	defer tdb.logMu.Unlock()
	var (
		tn    = uint64(time.Now().UnixNano() / 1e6)
		dels  = []uint64{}
		delay = tn
	)
	for k, v := range tdb.logLockSets {
		if v+3000 < tn {
			dels = append(dels, k)
		} else if v < delay {
			delay = v
		}
	}
	for _, k := range dels {
		delete(tdb.logLockSets, k)
	}
	if len(tdb.logLockSets) == 0 {
		return tn
	}
	return delay
}

func (tdb *dbTable) objectIncrSet(ns string, incr, set uint64) (uint64, error) {

	tdb.incrMu.Lock()
	defer tdb.incrMu.Unlock()

	incrSet := tdb.incrSets[ns]
	if incrSet == nil {
		incrSet = &dbTableIncrSet{
			offset: 0,
			cutset: 0,
		}
		tdb.incrSets[ns] = incrSet
	}

	if incr == 0 && set == 0 {
		return incrSet.offset, nil
	}

	if incrSet.cutset <= 100 {

		if bs, err := tdb.db.Get(keySysIncrCutset(ns), nil); err != nil {
			if err.Error() != ldbNotFound {
				return 0, err
			}
		} else {
			if incrSet.cutset, err = strconv.ParseUint(string(bs), 10, 64); err != nil {
				return 0, err
			}
			if incrSet.offset < incrSet.cutset {
				incrSet.offset = incrSet.cutset
			}
		}
	}

	if incrSet.offset < 100 {
		incrSet.offset = 100
	}

	if set > 0 && set > incrSet.offset {
		incr += (set - incrSet.offset)
	}

	if incr > 0 {

		if (incrSet.offset + incr) >= incrSet.cutset {

			cutset := incrSet.offset + incr + 100

			if err := tdb.db.Put(keySysIncrCutset(ns),
				[]byte(strconv.FormatUint(cutset, 10)), nil); err != nil {
				return 0, err
			}

			incrSet.cutset = cutset
		}

		incrSet.offset += incr
	}

	return incrSet.offset, nil
}
