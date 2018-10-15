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
	"encoding/base64"
	"strings"
	"time"

	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

func (c Proj) ProcListAction() {

	var sets hapi.ProjProcList
	defer c.RenderJson(&sets)

	var (
		proj_id = c.Params.Get("proj_id")
		limit   = int(c.Params.Int64("limit"))
		exit    = c.Params.Get("filter_exit")
		ptype   = "hit"
	)

	if limit < 10 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if exit == "true" {
		ptype = "exit"
	}

	var (
		offset = hapi.DataPathProjProcEntry(ptype, proj_id, 0, 0)
		cutset = hapi.DataPathProjProcEntry(ptype, proj_id, 1, 0)
	)

	rs := data.Data.KvProgRevScan(offset, cutset, limit)
	if !rs.OK() {
		return
	}

	rs.KvEach(func(entry *skv.ResultEntry) int {
		var set hapi.ProjProcEntry
		if err := entry.Decode(&set); err == nil {
			set.ProjId = proj_id
			sets.Items = append(sets.Items, &set)
		}
		return 0
	})

	sets.Kind = "ProjProcList"
}

func (c Proj) ProcStatsAction() {

	var (
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		qry       = c.Params.Get("qry")
		fq        hapi.TimeStatsFeedQuerySet
	)

	if len(qry) < 10 {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	bs, err := base64.StdEncoding.DecodeString(qry)
	if err != nil || len(bs) < 10 {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	if err := json.Decode(bs, &fq); err != nil {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	if len(fq.Items) < 1 {
		return
	}

	fq.Fix()

	if fq.TimeStart >= fq.TimeCutset {
		c.RenderJson(types.NewTypeErrorMeta("400", "Bad Request"))
		return
	}

	feed := hapi.NewPbStatsSampleFeed(fq.TimeCycle)

	limit := int(fq.TimeCutset-fq.TimeStart+fq.TimeCycle) / 600
	if limit < 1 {
		limit = 1
	}
	limit += 2

	if rs := data.Data.KvProgScan(
		hapi.DataPathProjProcStatsEntry(
			proc_time,
			proc_id,
			fq.TimeStart-fq.TimeCycle-600,
		),
		hapi.DataPathProjProcStatsEntry(
			proc_time,
			proc_id,
			fq.TimeCutset+600,
		),
		limit,
	); rs.OK() {

		ls := rs.KvList()
		var ifeed hapi.PbStatsIndexFeed
		for _, v := range ls {

			if err := v.Decode(&ifeed); err != nil {
				continue
			}

			for _, ientry := range ifeed.Items {
				if fq.Get(ientry.Name) == nil {
					continue
				}
				for _, iv := range ientry.Items {
					if iv.Time <= fq.TimeCutset {
						feed.SampleSync(ientry.Name, iv.Time, iv.Value)
					}
				}
			}
		}
	}

	for _, v := range feed.Items {

		for i := fq.TimeStart; i <= fq.TimeCutset; i += fq.TimeCycle {
			v.SyncTrim(i, 0)
		}
		v.Sort()

		if len(v.Items) < 2 {
			continue
		}

		fqi := fq.Get(v.Name)
		if fqi == nil {
			continue
		}

		if fqi.Delta {
			last_value := int64(0)
			for i := len(v.Items) - 1; i > 0; i-- {

				if v.Items[i].Value <= 0 {
					if last_value > 0 {
						v.Items[i].Value = last_value
					}
				} else {
					last_value = v.Items[i].Value
				}

				if v.Items[i].Value >= v.Items[i-1].Value && v.Items[i-1].Value > 0 {
					v.Items[i].Value = v.Items[i].Value - v.Items[i-1].Value
				} else {
					v.Items[i].Value = 0
				}
			}
		}

		offset := 0
		for j, v2 := range v.Items {
			if v2.Time < fq.TimeStart {
				offset = j
			} else {
				break
			}
		}
		if offset < 1 || offset >= len(v.Items)-1 {
			offset = 1
		}

		v.Items = v.Items[offset:]
	}

	feed.Kind = "StatsFeed"
	c.RenderJson(feed)

}

func (c Proj) ProcTraceListAction() {

	var (
		proj_id     = c.Params.Get("proj_id")
		proc_id     = uint32(c.Params.Int64("proc_id"))
		proc_time   = uint32(c.Params.Int64("proc_time"))
		sets        hapi.ProjProcTraceList
		offset      = uint32(c.Params.Int64("offset"))
		offset_date = c.Params.Get("offset_date")
		cutset      = uint32(0)
		limit       = int(c.Params.Int64("limit"))
	)
	defer c.RenderJson(&sets)

	if len(proj_id) < 10 ||
		proc_id < 10 ||
		proc_time < 100000000 {
		sets.Error = types.NewErrorMeta("400", "Bad Request")
		return
	}

	if limit < 10 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if offset > 0 {
		offset -= 1
	}

	if offset_date != "" {
		if tn, err := time.Parse(offset_date, "2006-01-02"); err == nil {
			offset = uint32(tn.Unix()) + 86400
		}
	}

	if offset == 0 {
		offset = uint32(time.Now().Unix())
	}

	if rs := data.Data.KvProgRevScan(
		hapi.DataPathProjProcTraceEntry(
			proj_id,
			proc_time,
			proc_id,
			offset,
		),
		hapi.DataPathProjProcTraceEntry(
			proj_id,
			proc_time,
			proc_id,
			cutset,
		),
		limit,
	); rs.OK() {

		ls := rs.KvList()
		for _, v := range ls {
			var item hapi.ProjProcTraceEntry
			if err := v.Decode(&item); err == nil {
				item.GraphOnCPU = ""
				item.GraphBurn = nil
				sets.Items = append(sets.Items, &item)
			}
		}

		if len(ls) >= limit {
			mkey := hapi.DataPathProjProcTraceEntry(
				proj_id,
				proc_time,
				proc_id,
				0)
			if rs2 := data.Data.KvProgGet(mkey); rs2.OK() {
				if meta := rs2.Meta(); meta != nil {
					sets.Total = int64(meta.Num)
				}
			}
		}

		if sets.Total < 1 {
			sets.Total = int64(len(ls))
		}
	}

	sets.Kind = "ProcessTraceList"
}

func (c Proj) ProcTraceGraphAction() {

	var (
		proj_id   = c.Params.Get("proj_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		created   = uint32(c.Params.Int64("created"))
		svg_w     = int(c.Params.Int64("svg_w"))
		svg_h     = int(c.Params.Int64("svg_h"))
		meta_type = "image/svg+xml"
		// perf_id   = fmt.Sprintf("perf.%d.%d.%d", created, proc_time, proc_id)
	)

	if svg_w < 800 {
		svg_w = 800
	} else if svg_w > 4000 {
		svg_w = 4000
	}

	if svg_h < 400 {
		svg_w = 400
	} else if svg_w > 2000 {
		svg_w = 2000
	}

	/*
		fp, err := os.Open(abs_path)
		if err != nil {
			c.RenderError(404, "Object Not Found")
			return
		}
		defer fp.Close()
	*/

	rs := data.Data.KvProgGet(
		hapi.DataPathProjProcTraceEntry(
			proj_id,
			proc_time,
			proc_id,
			created,
		))
	if rs.OK() {

		var item hapi.ProjProcTraceEntry
		if err := rs.Decode(&item); err == nil && len(item.GraphOnCPU) > 100 {
			if n := strings.Index(item.GraphOnCPU, `<svg version=`); n > 0 {
				item.GraphOnCPU = item.GraphOnCPU[n:]
				// item.GraphOnCPU = strings.Replace(item.GraphOnCPU, `width="1200"`, `preserveAspectRatio="xMidYMid meet"`, 1)
				// item.GraphOnCPU = strings.Replace(item.GraphOnCPU, `width="1200"`, fmt.Sprintf(`width="%d"`, svg_w), 1)
				// item.GraphOnCPU = strings.Replace(item.GraphOnCPU, `height="790"`, fmt.Sprintf(`height="%d"`, svg_h), 1)
			}

			c.Response.Out.Header().Set("Content-Type", meta_type)
			c.Response.Out.Header().Set("Cache-Control", "max-age=86400")
			c.RenderString(item.GraphOnCPU)
			// http.ServeContent(c.Response.Out, c.Request.Request, perf_id+".svg", time.Now(), fp)

			return
		}
	}

	c.RenderError(404, "Object Not Found")
}

func (c Proj) ProcTraceGraphBurnAction() {

	var (
		proj_id   = c.Params.Get("proj_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		created   = uint32(c.Params.Int64("created"))
	)

	if rs := data.Data.KvProgGet(hapi.DataPathProjProcTraceEntry(
		proj_id,
		proc_time,
		proc_id,
		created,
	)); rs.OK() {
		var item hapi.ProjProcTraceEntry
		if err := rs.Decode(&item); err == nil && item.GraphBurn != nil {
			if n := strings.Index(item.GraphOnCPU, `<svg version=`); n > 0 {
				item.GraphOnCPU = item.GraphOnCPU[n:]
			}
			c.RenderJson(item)
			return
		}
	}

	c.RenderError(404, "Object Not Found")
}
