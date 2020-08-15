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
	"context"
	"errors"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type GrpcAppCredential struct {
	ac *AppCredential
}

func NewGrpcAppCredential(k *AccessKey) credentials.PerRPCCredentials {
	return GrpcAppCredential{
		ac: NewAppCredential(k),
	}
}

func (s GrpcAppCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		appHttpHeaderKey: s.ac.SignToken(nil),
	}, nil
}

func (s GrpcAppCredential) RequireTransportSecurity() bool {
	return false
}

func GrpcAppValidator(ctx context.Context, keyMgr *AccessKeyManager) (*AppValidator, error) {

	if ctx == nil {
		return nil, errors.New("no auth token found")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md) < 1 {
		return nil, errors.New("no auth token found")
	}

	//
	t, ok := md[appHttpHeaderKey]
	if !ok || len(t) == 0 {
		return nil, errors.New("no auth token found")
	}

	av, err := NewAppValidator(t[0], keyMgr)
	if err != nil {
		return nil, err
	}

	return av, nil
}

func GrpcAppCredentialValid(ctx context.Context, keyMgr *AccessKeyManager) error {

	if keyMgr == nil {
		return errors.New("no AccessKeyManager found")
	}

	av, err := GrpcAppValidator(ctx, keyMgr)
	if err != nil {
		return err
	}

	return av.SignValid(nil)
}
