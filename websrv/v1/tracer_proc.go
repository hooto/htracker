package v1

import (
	"encoding/base64"
	"strings"

	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/skv"

	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

func (c Tracer) ProcListAction() {

	var sets hapi.TracerProcessList
	defer c.RenderJson(&sets)

	var (
		tracer_id = c.Params.Get("tracer_id")
		limit     = int(c.Params.Int64("limit"))
	)

	if limit < 10 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	var (
		offset = hapi.DataPathTracerProcessEntry(tracer_id, 0, 0)
		cutset = hapi.DataPathTracerProcessEntry(tracer_id, 1, 0)
	)

	rs := data.Data.KvProgRevScan(offset, cutset, limit)
	if !rs.OK() {
		return
	}

	rs.KvEach(func(entry *skv.ResultEntry) int {
		var set hapi.TracerProcessEntry
		if err := entry.Decode(&set); err == nil {
			set.Tid = tracer_id
			sets.Items = append(sets.Items, &set)
		}
		return 0
	})

	sets.Kind = "TracerProcessList"
}

func (c Tracer) ProcStatsAction() {

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

	if rs := data.Data.KvProgScan(
		hapi.DataPathTracerProcessStatsEntry(
			proc_time,
			proc_id,
			fq.TimeStart-fq.TimeCycle-600,
		),
		hapi.DataPathTracerProcessStatsEntry(
			proc_time,
			proc_id,
			fq.TimeCutset+600,
		),
		10000,
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

		v.Items = v.Items[1:]
	}

	feed.Kind = "StatsFeed"
	c.RenderJson(feed)

}

func (c Tracer) ProcTraceListAction() {

	var (
		tracer_id = c.Params.Get("tracer_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		sets      hapi.TracerProcessTraceList
	)
	defer c.RenderJson(&sets)

	if len(tracer_id) < 10 ||
		proc_id < 10 ||
		proc_time < 100000000 {
		sets.Error = types.NewErrorMeta("400", "Bad Request")
		return
	}

	if rs := data.Data.KvProgRevScan(
		hapi.DataPathTracerProcessTraceEntry(
			tracer_id,
			proc_time,
			proc_id,
			0,
		),
		hapi.DataPathTracerProcessTraceEntry(
			tracer_id,
			proc_time,
			proc_id,
			0,
		),
		10000,
	); rs.OK() {

		ls := rs.KvList()
		for _, v := range ls {
			var item hapi.TracerProcessTraceEntry
			if err := v.Decode(&item); err == nil {
				item.GraphOnCPU = ""
				item.GraphBurn = nil
				sets.Items = append(sets.Items, &item)
			}
		}
	}

	sets.Kind = "ProcessTraceList"
}

func (c Tracer) ProcTraceGraphAction() {

	var (
		tracer_id = c.Params.Get("tracer_id")
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
		hapi.DataPathTracerProcessTraceEntry(
			tracer_id,
			proc_time,
			proc_id,
			created,
		))
	if rs.OK() {

		var item hapi.TracerProcessTraceEntry
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

func (c Tracer) ProcTraceGraphBurnAction() {

	var (
		tracer_id = c.Params.Get("tracer_id")
		proc_id   = uint32(c.Params.Int64("proc_id"))
		proc_time = uint32(c.Params.Int64("proc_time"))
		created   = uint32(c.Params.Int64("created"))
	)

	if rs := data.Data.KvProgGet(hapi.DataPathTracerProcessTraceEntry(
		tracer_id,
		proc_time,
		proc_id,
		created,
	)); rs.OK() {
		var item hapi.TracerProcessTraceEntry
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
