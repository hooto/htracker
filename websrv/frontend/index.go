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

package frontend

import (
	"github.com/hooto/httpsrv"

	"github.com/hooto/htracker/config"
)

type Index struct {
	*httpsrv.Controller
}

func (c Index) IndexAction() {

	c.Response.Out.Header().Set("Cache-Control", "no-cache")

	c.RenderString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
  <title>hooto Tracker | Open Source Application Performance Management</title>
  <script src="/htracker/~/lessui/js/sea.js?v=` + config.VersionHash + `"></script>
  <script src="/htracker/~/htracker/js/main.js?v=` + config.VersionHash + `"></script>
  <link rel="stylesheet" href="/htracker/~/htracker/css/main.css?v=` + config.VersionHash + `" type="text/css">
  <link rel="shortcut icon" type="image/x-icon" href="/htracker/~/htracker/img/ht-tab-dark.png">
  <script type="text/javascript">
    htracker.version = "` + config.VersionHash + `";
	window.onload_hooks = [];
    window.onload = htracker.Boot();
  </script>
</head>
<body id="body-content">
</body>
</html>
`)

}
