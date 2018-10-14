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

const (
	Version = "0.1.1"

	ns_meta     uint8 = 4
	ns_ttl      uint8 = 5
	ns_kv       uint8 = 8
	ns_prog_ttl uint8 = 32
	ns_prog_def uint8 = 36
	ns_prog_cut uint8 = 63
	ns_prog_x   uint8 = 64

	prog_ttl_zero uint64 = 1500000000000000000
)

var (
	ns_prog_x_incr = []byte{ns_prog_x, 1}
)

// PKV
// kv    1|key                 : ns|n|meta|value

// hash  2|n|key               : ns|meta
// hash  2|n|key|field         : ns|n|meta|value

// sets  2|n|key               : ns|meta
// sets  2|n|key|field         : ns

// zset  2|n|key               : ns|meta
// zset  2|n|key|field         : ns|n|meta|value
// zset  3|n|key|n|score|field : ns

// list  2|n|key               : ns|meta
// list  2|n|key|incr          : ns|n|meta|value
