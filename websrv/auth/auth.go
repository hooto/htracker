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

package auth

import (
	"net/http"
	"time"

	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/crypto/phash"
	"github.com/lessos/lessgo/types"

	"github.com/hooto/htracker/config"
	"github.com/hooto/htracker/data"
	"github.com/hooto/htracker/hapi"
)

const (
	AccessTokenKey         = "ht_token"
	AuthSessionInit uint32 = 1 << 2
)

var (
	AuthErrUnAuth   = types.NewTypeErrorMeta("401", "Unauthorized")
	AuthErrInitAuth = AuthSession{
		TypeMeta: types.TypeMeta{
			Kind: "AuthSession",
		},
		Action: AuthSessionInit,
	}
)

type AuthSession struct {
	types.TypeMeta `json:",inline"`
	User           string `json:"user"`
	AccessToken    string `json:"access_token"`
	Action         uint32 `json:"action"`
}

type AuthLogin struct {
	Auth        string `json:"auth"`
	AuthConfirm string `json:"auth_confirm"`
}

func AuthSessionInstance(s *httpsrv.Session) *AuthSession {

	token := s.Get(AccessTokenKey)

	if rs := data.Data.NewReader(hapi.DataPathUserSessionEntry(token)).Query(); rs.OK() {
		var set AuthSession
		if err := rs.Decode(&set); err == nil && set.AccessToken == token {
			return &set
		}
	}

	return nil
}

type Auth struct {
	*httpsrv.Controller
}

func (c Auth) SessionAction() {

	if config.Config.Auth == "" {
		c.RenderJson(AuthErrInitAuth)
	} else if sess := AuthSessionInstance(c.Session); sess != nil {
		c.RenderJson(sess)
	} else {
		c.RenderJson(AuthErrUnAuth)
	}
}

func (c Auth) LoginAction() {
	var set AuthSession
	defer c.RenderJson(&set)

	var req AuthLogin

	if err := c.Request.JsonDecode(&req); err != nil {
		set.Error = types.NewErrorMeta("400", "Invalid Request "+err.Error())
		return
	}

	if req.Auth == "" {
		set.Error = types.NewErrorMeta("400", "Password Can Not be Empty")
		return
	}

	if config.Config.Auth == "" {
		if req.Auth != req.AuthConfirm {
			set.Error = types.NewErrorMeta("400", "Invalid Request")
			return
		}

		config.Config.Auth, _ = phash.Generate(req.Auth)

		if err := config.Flush(); err != nil {
			set.Error = types.NewErrorMeta("500", "Server Error "+err.Error())
			return
		}

	} else {

		if !phash.Verify(req.Auth, config.Config.Auth) {
			set.Error = types.NewErrorMeta("400", "Invalid Password")
			return
		}
	}

	set.User = "admin"
	set.AccessToken = hapi.ObjectId(uint32(time.Now().Unix()), 16)

	if rs := data.Data.NewWriter(hapi.DataPathUserSessionEntry(set.AccessToken), set).
		ExpireSet(86400 * 10 * 1000).Commit(); !rs.OK() {
		set.Error = types.NewErrorMeta("500", "Server Error (database)")
		return
	}

	http.SetCookie(c.Response.Out, &http.Cookie{
		Name:     AccessTokenKey,
		Value:    set.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(time.Hour * 240),
	})

	set.Kind = "AuthSession"
}

func (c Auth) SignOutAction() {

	http.SetCookie(c.Response.Out, &http.Cookie{
		Name:    AccessTokenKey,
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-86400),
	})

	c.Redirect("/htracker/")
}
