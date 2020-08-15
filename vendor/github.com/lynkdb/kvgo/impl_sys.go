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
	"strings"

	hauth "github.com/hooto/hauth/go/hauth/v1"
	kv2 "github.com/lynkdb/kvspec/go/kvspec/v2"
)

func (cn *Conn) SysCmd(rr *kv2.SysCmdRequest) *kv2.ObjectResult {

	if len(cn.opts.Cluster.MainNodes) > 0 {

		if cn.opts.ClientConnectEnable {
			return cn.sysCmdRemote(rr)
		}

		rs, err := cn.public.SysCmd(nil, rr)
		if err != nil {
			return kv2.NewObjectResultServerError(err)
		}
		return rs
	}

	return cn.sysCmdLocal(nil, rr)
}

func (cn *Conn) sysCmdLocal(av *hauth.AppValidator, rr *kv2.SysCmdRequest) *kv2.ObjectResult {

	var rs *kv2.ObjectResult

	switch rr.Method {

	case "TableSet":

		var req2 kv2.TableSetRequest

		if err := kv2.StdProto.Decode(rr.Body, &req2); err != nil {
			return kv2.NewObjectResultClientError(err)
		}

		if !kv2.TableNameReg.MatchString(req2.Name) {
			return kv2.NewObjectResultClientError(errors.New("invalid table name"))
		}

		if av != nil {
			if err := av.Allow(authPermTableWrite,
				hauth.NewScopeFilter(AuthScopeTable, req2.Name)); err != nil {
				return kv2.NewObjectResultAccessDenied(fmt.Sprintf("table (%s) %s", req2.Name, err.Error()))
			}
		}

		rr2 := kv2.NewObjectWriter(nsSysTable(req2.Name), &kv2.TableItem{
			Name: req2.Name,
			Desc: req2.Desc,
		}).IncrNamespaceSet(sysTableIncrNS).
			TableNameSet(sysTableName)

		tdb := cn.tabledb(req2.Name)
		if tdb == nil {
			rr2.ModeCreateSet(true)

		}

		rs = cn.Commit(rr2)
		if rs.OK() {
			if tdb == nil && rs.Meta.IncrId > 0 {
				cn.dbTableSetup(req2.Name, uint32(rs.Meta.IncrId))
			}
		}

	case "TableList":

		rr2 := kv2.NewObjectReader(nil).
			TableNameSet(sysTableName).
			KeyRangeSet(nsSysTable(""), append(nsSysTable(""), 0xff)).
			LimitNumSet(1000)

		if rs = cn.Query(rr2); rs.OK() {

			rr2.KeyRangeSet(nsSysTableStatus(""), append(nsSysTableStatus(""), 0xff))

			if rs2 := cn.Query(rr2); rs2.OK() {

				statuses := map[string]*kv2.TableStatus{}
				for _, v := range rs2.Items {
					var item kv2.TableStatus
					if err := v.DataValue().Decode(&item, nil); err == nil {
						statuses[item.Name] = &item
					}
				}

				for _, v := range rs.Items {
					var item kv2.TableItem
					if err := v.DataValue().Decode(&item, nil); err == nil {
						if st := statuses[item.Name]; st != nil {
							item.Status = st
							v.DataValueSet(item, nil)
						}
					}
				}
			}
		}

	case "AccessKeySet":

		key := new(hauth.AccessKey)
		if err := kv2.StdProto.Decode(rr.Body, key); err != nil {
			return kv2.NewObjectResultClientError(err)
		}

		{
			roles := []string{}
			reqRoles := key.Roles
			for _, v2 := range defaultRoles {
				for _, v3 := range reqRoles {
					if v3 == v2.Name {
						roles = append(roles, v2.Name)
						break
					}
				}
			}
			if len(roles) == 0 {
				return kv2.NewObjectResultClientError(errors.New("invalid roles settings"))
			}
			key.Roles = roles
		}

		{
			scopes := []*hauth.ScopeFilter{}
			for _, v2 := range defaultScopes {
				for _, v3 := range key.Scopes {
					if v3.Name == v2 {

						v3.Value = strings.TrimSpace(v3.Value)
						if v3.Value == "" {
							return kv2.NewObjectResultClientError(errors.New("scope value not found"))
						}
						scopes = append(scopes, v3)
						break
					}
				}
			}
			key.Scopes = scopes
		}

		var pkey *hauth.AccessKey

		if key.Id == "" {
			pkey = hauth.NewAccessKey()
		} else {
			pkey = cn.keyMgr.KeyGet(key.Id)
			if pkey == nil {
				return kv2.NewObjectResultClientError(errors.New("access key not found"))
			}
		}

		pkey.Roles = key.Roles
		pkey.Scopes = key.Scopes

		key = pkey

		if !hauth.AccessKeyIdReg.MatchString(key.Id) {
			return kv2.NewObjectResultClientError(errors.New("invalid access key id"))
		}

		rr2 := kv2.NewObjectWriter(nsSysAccessKey(key.Id), key).
			TableNameSet(sysTableName)

		tdb := cn.tabledb(sysTableName)
		if tdb != nil {
			rs = cn.Commit(rr2)
			if rs.OK() {
				cn.keyMgr.KeySet(key)
			}
		}

	case "AccessKeyList":

		keyId := ""
		if len(rr.Body) > 0 {
			keyId = string(rr.Body)
			if !hauth.AccessKeyIdReg.MatchString(keyId) {
				return kv2.NewObjectResultClientError(errors.New("invalid access key id"))
			}
		}

		rr2 := kv2.NewObjectReader(nil).
			TableNameSet(sysTableName).
			KeyRangeSet(nsSysAccessKey(""), append(nsSysAccessKey(""), 0xff)).
			LimitNumSet(1000)

		if rs2 := cn.Query(rr2); rs2.OK() {

			rs = kv2.NewObjectResultOK()

			for _, v := range rs2.Items {
				var item hauth.AccessKey
				if err := v.DataValue().Decode(&item, nil); err == nil {
					if keyId != "" {
						if keyId == item.Id {
							v.DataValueSet(item, nil)
							rs.Items = append(rs.Items, v)
							break
						}
					} else {
						if n := len(item.Secret); n > 8 {
							item.Secret = item.Secret[:4] + strings.Repeat("*", 8) + item.Secret[n-4:]
						} else {
							item.Secret = strings.Repeat("*", 8)
						}
						v.DataValueSet(item, nil)
						rs.Items = append(rs.Items, v)
					}
				}
			}
		} else {
			rs = rs2
		}

	default:
		rs = kv2.NewObjectResultClientError(errors.New("cmd not found"))
	}

	if rs == nil {
		rs = kv2.NewObjectResultClientError(errors.New("cmd not found"))
	}

	return rs
}

func (cn *Conn) sysCmdRemote(rr *kv2.SysCmdRequest) *kv2.ObjectResult {

	mainNodes := cn.opts.Cluster.randMainNodes(3)
	if len(mainNodes) < 1 {
		return kv2.NewObjectResultClientError(errors.New("no master found"))
	}

	for _, v := range mainNodes {

		c, err := v.NewClient()
		if err != nil {
			continue
		}

		if rs := c.Connector().SysCmd(rr); rs.OK() {
			return rs
		}
	}

	return kv2.NewObjectResultServerError(errors.New("no cluster nodes"))
}
