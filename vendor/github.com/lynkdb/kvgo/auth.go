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
	"context"
	"errors"

	"github.com/hooto/iam/iamauth"
	"google.golang.org/grpc/credentials"
)

func authKeyDefault() *iamauth.AuthKey {
	return &iamauth.AuthKey{
		AccessKey: "kvgo",
		User:      "root",
	}
}

func (it *Conn) authKey(ak string) *iamauth.AuthKey {
	it.keyMu.RLock()
	defer it.keyMu.RUnlock()
	k, ok := it.keys[ak]
	if ok {
		return k
	}
	return authKeyDefault()
}

func (it *Conn) authKeySet(ak, sk string) error {

	if len(sk) < 20 {
		return errors.New("invalid secret_key")
	}

	it.keyMu.Lock()
	defer it.keyMu.Unlock()

	it.keys[ak] = &iamauth.AuthKey{
		AccessKey: ak,
		SecretKey: sk,
	}

	return nil
}

func (it *Conn) authKeySetup(ak *iamauth.AuthKey, secretKey string) error {

	if len(secretKey) < 20 {
		return errors.New("invalid secret_key")
	}

	ak.SecretKey = secretKey

	return nil
}

func newAppCredential(key *iamauth.AuthKey) credentials.PerRPCCredentials {
	return iamauth.NewGrpcAppCredential(key)
}

func appAuthValid(ctx context.Context, key *iamauth.AuthKey) error {
	if key == nil {
		return errors.New("not auth key setup")
	}
	return iamauth.GrpcAppCredentialValid(ctx, key)
}
