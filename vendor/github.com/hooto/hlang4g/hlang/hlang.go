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

package hlang

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lessos/lessgo/encoding/json"
)

const (
	LocaleDefault = "en"
)

var (
	StdLangFeed = &LangList{}
)

type LangLocaleList struct {
	mu     sync.RWMutex
	Locale string             `json:"locale"`
	Items  []*LangLocaleEntry `json:"items"`
	Refer  *LangLocaleList    `json:"-"`
}

type LangLocaleEntry struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type LangList struct {
	mu    sync.RWMutex
	Items []*LangLocaleList
}

func (it *LangLocaleList) Sync(item *LangLocaleEntry, rewrite bool) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if item.Val == "" {
		item.Val = item.Key
	}

	item.Key = strings.ToLower(strings.TrimSpace(item.Key))
	if item.Key == "" {
		return
	}

	for _, v := range it.Items {

		if v.Key == item.Key {
			if v.Val != item.Val && rewrite {
				v.Val = item.Val
			}
			return
		}
	}

	it.Items = append(it.Items, &LangLocaleEntry{
		Key: item.Key,
		Val: item.Val,
	})
}

func (it *LangLocaleList) Entry(msg string) string {

	key := strings.ToLower(strings.TrimSpace(msg))

	it.mu.RLock()
	defer it.mu.RUnlock()

	for _, v := range it.Items {
		if v.Key == key {
			return v.Val
		}
	}
	return msg
}

func (it *LangList) Entry(locale string) *LangLocaleList {
	it.mu.RLock()
	defer it.mu.RUnlock()

	for _, v := range it.Items {
		if v.Locale == locale {
			return v
		}
	}

	return nil
}

func (it *LangList) Locale(locale string) *LangLocaleList {

	if entry := it.Entry(locale); entry != nil {
		if entry.Refer != nil {
			return entry.Refer
		}
		return entry
	}

	if locale != LocaleDefault {
		return it.Entry(LocaleDefault)
	}

	return nil
}

func (it *LangList) Sync(cfg LangLocaleList) *LangLocaleList {

	cfg.Locale = strings.ToLower(strings.Replace(cfg.Locale, "_", "-", 1))

	entry := it.Entry(cfg.Locale)

	it.mu.Lock()
	if entry == nil {
		entry = &LangLocaleList{
			Locale: cfg.Locale,
		}
		it.Items = append(it.Items, entry)
	}
	entry.Refer = nil
	for _, v := range cfg.Items {
		entry.Sync(v, true)
	}
	it.mu.Unlock()

	if ls := strings.Split(cfg.Locale, "-"); len(ls) == 2 {
		if base := it.Entry(ls[0]); base == nil {
			it.Items = append(it.Items, &LangLocaleList{
				Locale: ls[0],
				Refer:  entry,
			})
		}
	}

	return entry
}

func (it *LangList) Init() bool {
	if len(it.Items) < 1 {
		return false
	}

	entry := it.Entry(LocaleDefault)

	it.mu.Lock()
	defer it.mu.Unlock()

	for _, v := range it.Items {
		if v.Locale == LocaleDefault {
			continue
		}

		for _, v2 := range entry.Items {
			v.Sync(v2, false)
		}
	}

	return true
}

func (it *LangList) LoadMessages(file string, resync bool) {

	var cfg LangLocaleList
	if err := json.DecodeFile(file, &cfg); err != nil {
		hlog.Printf("error", "hooto/hlang: setup i18n err %s", err.Error())
		return
	}

	cfgSync := it.Sync(cfg)

	if resync {

		sort.Slice(cfgSync.Items, func(i, j int) bool {
			return strings.Compare(cfgSync.Items[i].Key, cfgSync.Items[j].Key) < 0
		})

		json.EncodeToFile(cfgSync, file, "  ")
	}

	hlog.Printf("info", "hooto/hlang: setup i18n %s %d",
		cfgSync.Locale, len(cfgSync.Items))
}

func (it *LangList) Translate(locale, msg string, args ...interface{}) string {

	entry := it.Locale(locale)
	if entry != nil {
		msg = entry.Entry(msg)
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	} else {
		return msg
	}
}
