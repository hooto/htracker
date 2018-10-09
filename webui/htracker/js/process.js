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


var htrackerProcess = {
    listLastUpdated: 0,
    entryActive: null,
}

htrackerProcess.Index = function() {
    htracker.Loader("#htracker-module-content", "process/list", {
        callback: function() {
            htrackerProcess.ListRefresh();
        },
    });
}

htrackerProcess.ListRefresh = function() {

    var elem = document.getElementById("htracker-process-list");
    if (!elem) {
        return;
    }
    var url = "limit=20";
    var elemq = document.getElementById("htracker-process-list-query");
    if (elemq && elemq.value.length > 0) {
        url += ("&q=" + elemq.value);
    }

    htracker.ModuleNavbarMenuRefresh("htracker-process-list-menus");
    htracker.OpToolsRefresh("#htracker-process-list-optools");


    htracker.ApiCmd("process/list?" + url, {
        callback: function(err, data) {

            if (!err && !data.error && data.updated > htrackerProcess.listLastUpdated) {
                l4iTemplate.Render({
                    dstid: "htracker-process-list",
                    tplid: "htracker-process-list-tpl",
                    data: data,
                });
                htrackerProcess.listLastUpdated = data.updated;
            }

            var msg = "Updated at " + l4i.UnixTimeFormat(data.updated, "Y-m-d H:i:s");
            $("#htracker-process-list-status-msg").text(msg);

            window.setTimeout(function() {
                htrackerProcess.ListRefresh();
            }, 5000);
        },
    });
}


htrackerProcess.ListRefreshQuery = function() {
    $("#htracker-process-list-status-msg").text("search ...");
    htrackerProcess.ListRefresh();
}

htrackerProcess.EntryView = function(pid) {

    htracker.ApiCmd("process/entry?pid=" + pid, {
        callback: function(err, data) {

            if (err) {
                return l4iAlert.Open("error", "Failed to get process (#" + pid + ")");
            }

            if (data.error) {
                return l4iAlert.Open("error", data.error.message);
            }

            htrackerProcess.entryActive = data;

            l4iModal.Open({
                id: "htracker-process-trace-new",
                title: "Process Overview",
                data: data,
                tplid: "htracker-process-entry-tpl",
                width: 900,
                height: 500,
                buttons: [{
                    title: "Trace by PID",
                    onclick: "htrackerProcess.TraceByPid()",
                    style: "btn-primary",
                }, {
                    title: "Trace by Name",
                    onclick: "htrackerProcess.TraceByName()",
                    style: "btn-primary",
                // }, {
                //     title: "Trace by Command",
                //     onclick: "htrackerProcess.TraceByCommand()",
                //     style: "btn-primary",
                }, {
                    title: "Cancel",
                    onclick: "l4iModal.Close()",
                }],
            });
        },
    });
}

htrackerProcess.TraceByPid = function() {
    if (!htrackerProcess.entryActive) {
        return;
    }
    htrackerTracer.NewEntry({
        modal_id: "htracker-process-trace-new",
        filter: {
            proc_id: htrackerProcess.entryActive.pid,
        },
    });
}

htrackerProcess.TraceByName = function() {
    if (!htrackerProcess.entryActive) {
        return;
    }
    htrackerTracer.NewEntry({
        modal_id: "htracker-process-trace-new",
        filter: {
            proc_name: htrackerProcess.entryActive.name,
        },
    });
}

htrackerProcess.TraceByCommand = function() {
    if (!htrackerProcess.entryActive) {
        return;
    }
    htrackerTracer.NewEntry({
        modal_id: "htracker-process-trace-new",
        filter: {
            proc_command: htrackerProcess.entryActive.command,
        },
    });
}

