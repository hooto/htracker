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

	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
	"github.com/hooto/htracker/status"
	"github.com/hooto/htracker/websrv/auth"
)

type Sys struct {
	*httpsrv.Controller
}

func (c *Sys) Init() int {

	if config.Config.Auth == "" {
		c.RenderJson(auth.AuthErrInitAuth)
		return 1
	}

	if sess := auth.AuthSessionInstance(c.Session); sess == nil {
		c.RenderJson(auth.AuthErrUnAuth)
		return 1
	}

	return 0
}

func (c Sys) ItemAction() {
	c.RenderJson(status.Host)
}

func (c Sys) StatsFeedAction() {

	var (
		qry = c.Params.Get("qry")
		fq  hapi.TimeStatsFeedQuerySet
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
		hapi.DataSysHostStats(fq.TimeStart),
		hapi.DataSysHostStats(fq.TimeCutset+600),
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
