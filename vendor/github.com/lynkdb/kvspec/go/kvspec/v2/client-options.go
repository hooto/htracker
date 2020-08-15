// Copyright 2019 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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

package kvspec

type ClientOption interface {
	Apply(*ClientOptions)
}

type ClientOptions struct {
	// timeout in milliseconds
	Timeout ClientTimeout `toml:"timeout" json:"timeout"`
}

func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		Timeout: 10000,
	}
}

func (it *ClientOptions) Apply(opts *ClientOptions) {
	if it.Timeout > 0 {
		it.Timeout.Apply(opts)
	}
}

type ClientTimeout int64

func (it ClientTimeout) Apply(opts *ClientOptions) {
	if it < 1e3 {
		it = 1e3
	} else if it > 60e3 {
		it = 60e3
	}
	opts.Timeout = it
}
