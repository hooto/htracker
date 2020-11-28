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

package injob

import (
	"time"
)

type Context struct {
	daemon *Daemon
}

func (it *Context) ConditionRefresh(name string) *Context {
	if it.daemon != nil {
		it.daemon.conditionSet(name, time.Now().UnixNano()/1e6)
	}
	return it
}

func (it *Context) ConditionSet(name string, v int64) *Context {
	if it.daemon != nil {
		it.daemon.conditionSet(name, v)
	}
	return it
}

func (it *Context) ConditionDel(name string) *Context {
	if it.daemon != nil {
		it.daemon.conditionDel(name)
	}
	return it
}
