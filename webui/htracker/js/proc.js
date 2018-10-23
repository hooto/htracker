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


var htrackerProc = {
    listLimit: 30,
    listLastUpdated: 0,
    listAutoRefreshTimer: null,
    listAutoRefreshTimeRange: 10000,
    entryActive: null,
}

htrackerProc.Index = function() {
    h3tracker.KeyUpEscHook = null;
    h3tracker.ModuleNavbarLeftClean();
    htracker.Loader("#htracker-module-content", "proc/list", {
        callback: function() {
            htrackerProc.ListRefresh();
        },
    });
}

htrackerProc.ListRefresh = function() {

    var elem = document.getElementById("htracker-proc-list-box");
    if (!elem) {
        // if (htrackerProc.listAutoRefreshTimer) {
        // 	clearInterval(htrackerProc.listAutoRefreshTimer);
        // 	htrackerProc.listAutoRefreshTimer = null;
        // 	console.log("clearInterval htrackerProc.listAutoRefreshTimer");
        // }
        return;
    }
    var url = "limit=" + htrackerProc.listLimit;
    var elemq = document.getElementById("htracker-proc-list-query");
    if (elemq && elemq.value.length > 0) {
        url += ("&q=" + elemq.value);
    }

    htracker.ModuleNavbarMenuRefresh("htracker-proc-list-menus");
    htracker.OpToolsRefresh("#htracker-proc-list-optools");


    htracker.ApiCmd("proc/list?" + url, {
        callback: function(err, data) {

            if (!err && !data.error && data.updated > htrackerProc.listLastUpdated) {

                for (var i in data.items) {
                    if (!data.items[i].cpu_p) {
                        data.items[i].cpu_p = 0;
                    }
                }

                l4iTemplate.Render({
                    dstid: "htracker-proc-list-box",
                    tplid: "htracker-proc-list-box-tpl",
                    data: data,
                });
                htrackerProc.listLastUpdated = data.updated;

                var msg = l4i.T("Top %d/%d processes at %s",
                    data.num, data.total, l4i.UnixTimeFormat(data.updated, "Y-m-d H:i:s"));
                $("#htracker-proc-list-status-msg").text(msg);
            }

            if (!htrackerProc.listAutoRefreshTimer) {
                htrackerProc.listAutoRefreshTimer = window.setTimeout(function() {
                    htrackerProc.listAutoRefreshTimer = null;
                    htrackerProc.ListRefresh();
                }, htrackerProc.listAutoRefreshTimeRange);
            }
        },
    });
}


htrackerProc.ListRefreshQuery = function() {
    $("#htracker-proc-list-status-msg").text(l4i.T("Search") + " ...");
    htrackerProc.ListRefresh();
}

htrackerProc.EntryView = function(pid) {

    htracker.ApiCmd("proc/entry?pid=" + pid, {
        callback: function(err, data) {

            if (err) {
                return l4iAlert.Open("error",
                    l4i.T("Failed to get %s", l4i.T("Process")));
            }

            if (data.error) {
                return l4iAlert.Open("error", data.error.message);
            }

            htrackerProc.entryActive = data;

            l4iModal.Open({
                id: "htracker-proc-projnew",
                title: l4i.T("Process Overview"),
                data: data,
                tplid: "htracker-proc-entry-tpl",
                width: 900,
                height: 500,
                buttons: [{
                    title: l4i.T("Trace by %s", l4i.T("Process ID")),
                    onclick: "htrackerProc.TraceByPid()",
                    style: "btn-primary",
                }, {
                    title: l4i.T("Trace by %s", l4i.T("Process Name")),
                    onclick: "htrackerProc.TraceByName()",
                    style: "btn-primary",
                // }, {
                //     title: "Trace by Command",
                //     onclick: "htrackerProc.TraceByCommand()",
                //     style: "btn-primary",
                }, {
                    title: l4i.T("Cancel"),
                    onclick: "l4iModal.Close()",
                }],
            });
        },
    });
}

htrackerProc.TraceByPid = function() {
    if (!htrackerProc.entryActive) {
        return;
    }
    htrackerProj.NewEntry({
        modal_id: "htracker-proc-projnew",
        filter: {
            proc_id: htrackerProc.entryActive.pid,
        },
    });
}

htrackerProc.TraceByName = function() {
    if (!htrackerProc.entryActive) {
        return;
    }
    htrackerProj.NewEntry({
        modal_id: "htracker-proc-projnew",
        filter: {
            proc_name: htrackerProc.entryActive.name,
        },
    });
}

htrackerProc.TraceByCommand = function() {
    if (!htrackerProc.entryActive) {
        return;
    }
    htrackerProj.NewEntry({
        modal_id: "htracker-proc-projnew",
        filter: {
            proc_command: htrackerProc.entryActive.command,
        },
    });
}

