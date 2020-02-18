// Copyright 2015 lynkdb Authors, All rights reserved.
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

package connect // import "github.com/lynkdb/iomix/connect"

import (
	"sync"

	"github.com/lessos/lessgo/types"
)

var (
	conn_mu sync.Mutex
)

//
type ConnOptions struct {
	//
	Name types.NameIdentifier `json:"name"`

	// Connector is the interface that must be implemented by a database/storage driver.
	// example:
	// 	iomix/fs/Connector
	// 	iomix/skv/Connector
	Connector types.NameIdentifier `json:"connector"`

	// Driver ia a database/storage package that implemented the Connector interface.
	Driver types.NameIdentifier `json:"driver"`

	// Data dynamic link library
	DriverPlugin types.NameIdentifier `json:"driver_plugin,omitempty"`

	// Items defines configurations used by driver
	Items types.Labels `json:"items"`
}

func (o *ConnOptions) Value(name string) string {

	if v, ok := o.Items.Get(name); ok {
		return v.String()
	}

	return ""
}

func (o *ConnOptions) SetValue(name, value string) {
	o.Items.Set(name, value)
}

//
type MultiConnOptions []*ConnOptions

func (mo *MultiConnOptions) SetOptions(o ConnOptions) error {

	conn_mu.Lock()
	defer conn_mu.Unlock()

	for _, prev := range *mo {

		if prev.Name == o.Name {

			if len(o.Driver) > 0 && prev.Driver != o.Driver {
				prev.Driver = o.Driver
			}

			if len(o.Connector) > 0 && prev.Connector != o.Connector {
				prev.Connector = o.Connector
			}

			for _, v := range o.Items {
				prev.Items.Set(v.Name, v.Value)
			}

			return nil
		}
	}

	*mo = append(*mo, &o)

	return nil
}

func (mo *MultiConnOptions) Options(name types.NameIdentifier) *ConnOptions {

	conn_mu.Lock()
	defer conn_mu.Unlock()

	for _, prev := range *mo {

		if prev.Name == name {
			return prev
		}
	}

	return nil
}
