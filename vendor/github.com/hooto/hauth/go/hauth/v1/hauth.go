// Copyright 2020 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package hauth

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"sync"

	"github.com/lessos/lessgo/crypto/idhash"
)

var (
	AccessKeyIdReg = regexp.MustCompile("^[0-9a-z]{1}[a-z0-9_]{3,31}$")
)

var (
	authKeyDefault = &AccessKey{
		Id:     "00000000",
		Secret: idhash.RandBase64String(40),
	}
	base64Std = base64.StdEncoding.WithPadding(base64.NoPadding)
	base64Url = base64.URLEncoding.WithPadding(base64.NoPadding)
)

type AccessKeyManager struct {
	mu    sync.RWMutex
	items map[string]*AccessKey
	roles map[string]*accessKeyManagerRole
}

type accessKeyManagerRole struct {
	permissions map[string]bool
}

func NewAccessKeyManager() *AccessKeyManager {
	return &AccessKeyManager{
		items: map[string]*AccessKey{},
		roles: map[string]*accessKeyManagerRole{},
	}
}

func (it *AccessKeyManager) KeySet(k *AccessKey) error {

	it.mu.Lock()
	defer it.mu.Unlock()

	if ak, ok := it.items[k.Id]; !ok || k != ak {
		it.items[k.Id] = k
	}

	return nil
}

func (it *AccessKeyManager) KeyGet(id string) *AccessKey {

	it.mu.RLock()
	defer it.mu.RUnlock()

	key, ok := it.items[id]
	if ok {
		return key
	}
	return nil
}

func (it *AccessKeyManager) KeyRand() *AccessKey {

	it.mu.RLock()
	defer it.mu.RUnlock()

	for _, key := range it.items {
		return key
	}

	return authKeyDefault
}

func (it *AccessKeyManager) RoleSet(r *Role) *AccessKeyManager {

	it.mu.Lock()
	defer it.mu.Unlock()

	role, ok := it.roles[r.Name]
	if !ok {
		role = &accessKeyManagerRole{
			permissions: map[string]bool{},
		}
		it.roles[r.Name] = role
	}

	for _, p := range r.Permissions {
		role.permissions[p] = true
	}

	return it
}

func (it *AccessKeyManager) keyRoles(key *AccessKey) []*accessKeyManagerRole {

	it.mu.RLock()
	defer it.mu.RUnlock()

	if len(key.Roles) == 0 {
		return nil
	}

	roles := []*accessKeyManagerRole{}

	for _, r := range key.Roles {
		if role, ok := it.roles[r]; ok {
			roles = append(roles, role)
		}
	}

	return roles
}

func base64nopad(s string) string {
	if i := strings.IndexByte(s, '='); i > 0 {
		return s[:i]
	}
	return s
}

func base64pad(s string) string {
	if n := len(s) % 4; n > 0 {
		s += strings.Repeat("=", 4-n)
	}
	return s
}

func NewAccessKey() *AccessKey {
	return &AccessKey{
		Id:     idhash.RandHexString(16),
		Secret: idhash.RandBase64String(40),
	}
}

type fixAccessKey AccessKey

func (it *AccessKey) redec() {

	if it.Id == "" && it.AccessKey != "" {
		it.Id = it.AccessKey
	} else {
		it.AccessKey = it.Id
	}

	if it.Secret == "" && it.SecretKey != "" {
		it.Secret = it.SecretKey
	} else {
		it.SecretKey = it.Secret
	}
}

func (it *AccessKey) reenc() {
	if it.Id != "" {
		it.AccessKey = ""
	}
	if it.Secret != "" {
		it.SecretKey = ""
	}
}

func (it *AccessKey) UnmarshalTOML(p interface{}) error {

	data, ok := p.(map[string]interface{})
	if ok {

		for k, v := range data {
			switch k {
			case "id":
				it.Id = v.(string)

			case "access_key":
				if it.Id == "" {
					it.Id = v.(string)
				}

			case "secret":
				it.Secret = v.(string)

			case "secret_key":
				if it.Secret == "" {
					it.Secret = v.(string)
				}

			case "user":
				it.User = v.(string)

			case "roles":
				if v2, ok := v.([]interface{}); ok {
					for _, v3 := range v2 {
						it.Roles = append(it.Roles, v3.(string))
					}
				}

			case "scopes":
				if v2, ok := v.([]map[string]interface{}); ok {
					for _, v3 := range v2 {
						var (
							name, ok1  = v3["name"]
							value, ok2 = v3["value"]
						)
						if ok1 && ok2 {
							it.Scopes = append(it.Scopes, &ScopeFilter{
								Name:  name.(string),
								Value: value.(string),
							})
						}
					}
				}
			}
		}

		it.redec()
	}

	return nil
}

func (it *AccessKey) UnmarshalJSON(b []byte) error {
	var it2 fixAccessKey
	if err := json.Unmarshal(b, &it2); err != nil {
		return err
	}
	*it = AccessKey(it2)
	it.redec()
	return nil
}

func (it AccessKey) MarshalJSON() ([]byte, error) {
	it.redec()
	it.reenc()
	return json.Marshal(fixAccessKey(it))
}

func NewScopeFilter(name, value string) *ScopeFilter {
	return &ScopeFilter{
		Name:  name,
		Value: value,
	}
}
