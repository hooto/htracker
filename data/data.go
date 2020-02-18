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

package data

import (
	"github.com/lynkdb/iomix/sko"
	"github.com/lynkdb/kvgo"

	"github.com/hooto/htracker/config"
)

var (
	Data sko.ClientConnector
	err  error
)

func Setup() error {

	if Data, err = kvgo.Open(config.Config.DataStorage); err != nil {
		return err
	}

	return nil
}
