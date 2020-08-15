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
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	appHttpHeaderKey       = "x-hooto-auth"
	appTokenVersion2       = "2"    // sha256
	appAuthTimeRange int64 = 600000 // milliseconds
)

type AppCredential struct {
	key     *AccessKey
	payload *AppPayload
}

type AppValidator struct {
	AppPayload
	mu      sync.Mutex
	version string
	payload []byte
	sign    string
	keyMgr  *AccessKeyManager
	key     *AccessKey
	roles   []*accessKeyManagerRole
	scopes  map[string]string
}

func NewAppCredential(k *AccessKey) *AppCredential {
	return &AppCredential{
		key: k,
	}
}

func (it *AppCredential) SignToken(data []byte) string {

	if it.payload == nil {
		it.payload = &AppPayload{
			Id:      it.key.Id,
			User:    it.key.User,
			Created: time.Now().UnixNano() / 1e6,
		}
	} else {
		it.payload.Created = time.Now().UnixNano() / 1e6
	}

	pbs, _ := proto.Marshal(it.payload)

	return appTokenVersion2 + "." +
		base64Std.EncodeToString(pbs) + "." +
		appSign(appTokenVersion2, pbs, data, it.key.Secret)
}

func (it *AppCredential) SignHttpToken(r *http.Request, data []byte) {
	r.Header.Set(appHttpHeaderKey, it.SignToken(data))
}

func AppValidWithHttpRequest(r *http.Request, data []byte, keyMgr *AccessKeyManager) (*AppValidator, error) {
	return AppValid(r.Header.Get(appHttpHeaderKey), data, keyMgr)
}

func AppValid(token string, data []byte, keyMgr *AccessKeyManager) (*AppValidator, error) {
	v, err := NewAppValidator(token, keyMgr)
	if err != nil {
		return nil, err
	}
	return v, v.SignValid(data)
}

func NewAppValidatorWithHttpRequest(r *http.Request, keyMgr *AccessKeyManager) (*AppValidator, error) {
	return NewAppValidator(r.Header.Get(appHttpHeaderKey), keyMgr)
}

func NewAppValidator(token string, keyMgr *AccessKeyManager) (*AppValidator, error) {

	var (
		n1 = strings.IndexByte(token, '.')
		n2 = strings.LastIndexByte(token, '.')
	)

	if 0 < n1 && n1 < n2 && (n2+2) < len(token) {

		switch token[:n1] {

		case appTokenVersion2:

			//
			var pv AppValidator
			bs, err := base64Std.DecodeString(base64nopad(token[n1+1 : n2]))
			if err == nil {
				err = proto.Unmarshal(bs, &pv.AppPayload)
			}
			if err != nil {
				return nil, errors.New("invalid payload data, err " + err.Error())
			}

			if len(pv.Id) < 4 {
				return nil, errors.New("payload/access_key id not found")
			}

			if pv.Created < 1000000000 {
				return nil, errors.New("invalid request time")
			}

			//
			tr := (time.Now().UnixNano() / 1e6) - pv.Created
			if tr > appAuthTimeRange || tr < -appAuthTimeRange {
				return nil, errors.New("invalid request time")
			}

			pv.version = appTokenVersion2
			pv.payload = bs
			pv.sign = token[n2+1:]

			pv.keyMgr = keyMgr

			return &pv, nil
		}
	}

	return nil, errors.New("access_token not found")
}

func (it *AppValidator) SignValid(data []byte) error {

	if it.keyMgr == nil {
		return errors.New("sign denied : no KeyManager found")
	}

	if it.key == nil {
		it.key = it.keyMgr.KeyGet(it.AppPayload.Id)
		if it.key == nil {
			return errors.New("sign denied : no AccessKey found")
		}
	}

	if appSign(it.version, it.payload, data, it.key.Secret) == it.sign {
		return nil
	}

	return errors.New("sign denied")
}

func (it *AppValidator) Allow(args ...interface{}) error {

	if len(args) == 0 {
		return errors.New("args not found")
	}

	if it.key == nil || it.keyMgr == nil {
		return errors.New("access_key not found")
	}

	var (
		permissions = map[string]bool{}
		scopes      = []*ScopeFilter{}
	)

	for _, arg := range args {

		switch arg.(type) {

		case string:
			permissions[arg.(string)] = false

		case *ScopeFilter:
			scopes = append(scopes, arg.(*ScopeFilter))

		default:
			return errors.New("invalid args type")
		}
	}

	if len(scopes) > 0 {

		if len(it.key.Scopes) < 1 {
			return errors.New("access_key/scopes not found")
		}

		if it.scopes == nil || len(it.scopes) == 0 {
			it.scopes = map[string]string{}
			for _, v := range it.key.Scopes {
				it.scopes[v.Name] = "," + v.Value + ","
			}
		}

		for _, scope := range scopes {
			if p, ok := it.scopes[scope.Name]; !ok || len(p) == 0 {
				return errors.New("access_key/scopes/name=" + scope.Name + " not match")
			} else if p != ",*," && !strings.Contains(p, ","+scope.Value+",") {
				return errors.New("access_key/scopes/value=" + scope.Value + " not match")
			}
		}
	}

	if len(permissions) == 0 {
		return nil
	}

	if len(it.roles) == 0 {
		it.roles = it.keyMgr.keyRoles(it.key)
		if len(it.roles) == 0 {
			return errors.New("access_key/roles not found")
		}
	}

	for permission, _ := range permissions {
		hit := false
		for _, role := range it.roles {
			if _, hit = role.permissions[permission]; hit {
				break
			}
		}
		if !hit {
			return errors.New("permission=" + permission + " not allow")
		}
	}

	return nil
}

func appSign(version string, payload, data []byte, secretKey string) string {

	hs := sha256.New()
	hs.Write(payload)
	hs.Write(data)
	hs.Write([]byte(secretKey))

	return base64Std.EncodeToString(hs.Sum(nil))
}
