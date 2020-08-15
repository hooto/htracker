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
	"errors"
	"fmt"

	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func (cn *Conn) BatchCommit(rr *kv2.BatchRequest) *kv2.BatchResult {

	if len(cn.opts.Cluster.MainNodes) > 0 {

		if cn.opts.ClientConnectEnable {
			return cn.batchCommitRemote(rr)
		}

		rs, err := cn.public.BatchCommit(nil, rr)
		if err != nil {
			return rr.NewResult(kv2.ResultClientError, err.Error())
		}
		return rs
	}

	return cn.batchCommitLocal(rr)
}

func (cn *Conn) batchCommitLocal(rr *kv2.BatchRequest) *kv2.BatchResult {

	if len(rr.Items) == 0 {
		return rr.NewResult(kv2.ResultOK, "")
	}

	var (
		rs = rr.NewResult(0, "")
		ok = 0
	)

	for i, v := range rr.Items {

		var rs2 *kv2.ObjectResult

		if v.Reader != nil {

			if v.Reader.TableName == "" {
				v.Reader.TableName = rr.TableName
			}
			rs2 = cn.objectLocalQuery(v.Reader)

		} else if v.Writer != nil {

			if v.Writer.TableName == "" {
				v.Writer.TableName = rr.TableName
			}
			rs2 = cn.commitLocal(v.Writer, 0)

		} else {
			rs2 = kv2.NewObjectResultClientError(errors.New("no reader/writer commit"))
		}

		rs.Items = append(rs.Items, rs2)

		if rs2.OK() {
			ok += 1
		} else {
			if rs.Message != "" {
				rs.Message += ", "
			}
			rs.Message += fmt.Sprintf("b%d %s ", i, rs2.Message)
		}
	}

	if ok == len(rs.Items) {
		rs.Status = kv2.ResultOK
	}

	return rs
}

func (cn *Conn) batchCommitRemote(rr *kv2.BatchRequest) *kv2.BatchResult {

	mainNodes := cn.opts.Cluster.randMainNodes(3)

	for _, v := range mainNodes {

		c, err := v.NewClient()
		if err != nil {
			continue
		}

		if rs := c.Connector().BatchCommit(rr); rs.OK() {
			return rs
		}
	}

	return rr.NewResult(kv2.ResultClientError, "no master found")
}
