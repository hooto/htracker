// Copyright 2019 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package iamauth

import (
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	appHttpHeaderKey       = "x-iam-auth"
	appTokenVersion2       = "2"    // sha256
	appAuthTimeRange int64 = 600000 // milliseconds
)

type AppCredential struct {
	key     *AuthKey
	payload *AppPayload
}

type AppValidator struct {
	AppPayload
	version string
	payload []byte
	sign    string
}

func NewAppCredential(k *AuthKey) *AppCredential {
	return &AppCredential{
		key: k,
	}
}

func (it *AppCredential) SignToken(data []byte) string {

	if it.payload == nil {
		it.payload = &AppPayload{
			AccessKey: it.key.AccessKey,
			User:      it.key.User,
			Created:   time.Now().UnixNano() / 1e6,
		}
	} else {
		it.payload.Created = time.Now().UnixNano() / 1e6
	}

	pbs, _ := proto.Marshal(it.payload)

	return appTokenVersion2 + "." +
		base64Std.EncodeToString(pbs) + "." +
		appSign(appTokenVersion2, pbs, data, it.key.SecretKey)
}

func (it *AppCredential) SignHttpToken(r *http.Request, data []byte) {
	r.Header.Set(appHttpHeaderKey, it.SignToken(data))
}

func AppValidWithHttpRequest(r *http.Request, data []byte, key *AuthKey) (*AppValidator, error) {
	return AppValid(r.Header.Get(appHttpHeaderKey), data, key)
}

func AppValid(token string, data []byte, key *AuthKey) (*AppValidator, error) {
	v, err := NewAppValidator(token)
	if err != nil {
		return nil, err
	}
	return v, v.SignValid(data, key)
}

func NewAppValidatorWithHttpRequest(r *http.Request) (*AppValidator, error) {
	return NewAppValidator(r.Header.Get(appHttpHeaderKey))
}

func NewAppValidator(token string) (*AppValidator, error) {

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

			if len(pv.AccessKey) < 4 {
				return nil, errors.New("payload/access_key not found")
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

			return &pv, nil
		}
	}

	return nil, errors.New("access token not found")
}

func (it *AppValidator) SignValid(data []byte, key *AuthKey) error {

	if appSign(it.version, it.payload, data, key.SecretKey) == it.sign {
		return nil
	}

	return errors.New("sign denied")
}

func appSign(version string, payload, data []byte, secretKey string) string {

	hs := sha256.New()
	hs.Write(payload)
	hs.Write(data)
	hs.Write([]byte(secretKey))

	return base64Std.EncodeToString(hs.Sum(nil))
}
