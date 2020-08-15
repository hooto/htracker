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
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	userTokenVersion2       = "2"        // sha256
	userAuthTtlMin    int64 = 600        // seconds
	userAuthTtlMax    int64 = 86400 * 30 // seconds
)

type UserValidator struct {
	UserPayload
	version     string
	payload     string
	accessKeyId string
	sign        string
}

func NewUserPayload(id, name string, roles []uint32, groups []string, ttl int64) *UserPayload {

	if ttl < userAuthTtlMin {
		ttl = userAuthTtlMin
	} else if ttl > userAuthTtlMax {
		ttl = userAuthTtlMax
	}

	return &UserPayload{
		Id:      id,
		Name:    name,
		Roles:   roles,
		Groups:  groups,
		Expired: time.Now().Unix() + ttl,
	}
}

func (it *UserPayload) IsExpired() bool {
	return it.Expired <= time.Now().Unix()
}

func (it *UserPayload) SignToken(keyMgr *AccessKeyManager) string {

	var (
		key     = keyMgr.KeyRand()
		bs, _   = proto.Marshal(it)
		payload = base64Url.EncodeToString(bs)
	)

	return userTokenVersion2 + "." +
		payload + "." +
		key.Id + ":" + userSign(userTokenVersion2, payload, key.Secret)
}

func UserValid(token string, keyMgr *AccessKeyManager) (*UserValidator, error) {
	v, err := NewUserValidator(token)
	if err != nil {
		return nil, err
	}
	return v, v.SignValid(keyMgr)
}

func NewUserValidator(token string) (*UserValidator, error) {

	var (
		n1 = strings.IndexByte(token, '.')
		n2 = strings.LastIndexByte(token, '.')
	)

	if 0 < n1 && n1 < n2 && (n2+2) < len(token) {

		switch token[:n1] {

		case userTokenVersion2:

			//
			n2k := strings.LastIndexByte(token, ':')
			if (n2k-4) <= n2 || (n2k+4) >= len(token) {
				return nil, errors.New("invalid sign token")
			}

			vr := UserValidator{
				version: userTokenVersion2,
				payload: token[n1+1 : n2],
			}

			//
			bs, err := base64Url.DecodeString(base64nopad(vr.payload))
			if err == nil {
				err = proto.Unmarshal(bs, &vr.UserPayload)
			}
			if err != nil {
				return nil, errors.New("invalid payload data, " + err.Error())
			}

			//
			vr.accessKeyId = token[n2+1 : n2k]
			vr.sign = token[n2k+1:]

			return &vr, nil
		}
	}

	return nil, errors.New("invalid sign token")
}

func (it *UserValidator) SignValid(keyMgr *AccessKeyManager) error {

	//
	if it.IsExpired() {
		return errors.New("sign token expired")
	}

	//
	key := keyMgr.KeyGet(it.accessKeyId)
	if key == nil ||
		userSign(it.version, it.payload, key.Secret) != it.sign {
		return errors.New("sign denied")
	}

	return nil
}

func userSign(version, payload, secretKey string) string {
	hs := sha256.Sum256([]byte(payload + secretKey))
	return base64Url.EncodeToString(hs[:])
}
