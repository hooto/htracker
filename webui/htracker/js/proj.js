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


var htrackerProj = {
    entryActive: null,
    entryActiveId: null,
    proc_stats_active_past: 3600,
    hchart_def: {
        "type": "line",
        "options": {
            "height": "200px",
            "title": "",
        },
        "data": {
            "labels": [],
            "datasets": [],
        },
    },
    listMenus: [{
        name: "Current Active Projects",
        uri: "proj/list/active",
    }, {
        name: "History Projects",
        uri: "proj/list/history",
    }],
    listMenuActive: null,
    listLimit: 50,
    listOffset: null,
    listAutoRefreshTimer: null,
    procListMenus: [{
        name: "Current Running Processes",
        uri: "proj/proc/hit",
    }, {
        name: "Exited Processes",
        uri: "proj/proc/exit",
    }],
    procListMenuActive: null,
    procTraceListOffset: 0,
    procTraceListLimit: 50,
}

htrackerProj.Index = function() {

    htracker.KeyUpEscHook = null;
    htracker.ModuleNavbarLeftClean();

    if (!htrackerProj.listMenuActive) {
        htrackerProj.listMenuActive = "proj/list/active";
    }

    htracker.Loader("#htracker-module-content", "proj/list", {
        callback: function() {
            htracker.ModuleNavbarMenu("projmenus", htrackerProj.listMenus);
            l4i.UrlEventRegister("proj/list/active", htrackerProj.ListRefreshActive, "htracker-module-navbar-menus");
            l4i.UrlEventRegister("proj/list/history", htrackerProj.ListRefreshHistory, "htracker-module-navbar-menus");
            l4i.UrlEventHandler(htrackerProj.listMenuActive, true);
        },
    });
}

htrackerProj.ListRefreshActive = function() {
    htrackerProj.ListRefresh("proj/list/active");
}

htrackerProj.ListRefreshHistory = function() {
    htrackerProj.ListRefresh("proj/list/history");
}

htrackerProj.ListRefresh = function(list_active, options) {
    options = options || {};

    var elem = document.getElementById("htracker-projlist-table");
    if (!elem) {
        return;
    }

    var url = "limit=" + htrackerProj.listLimit;


    var elemq = document.getElementById("htracker-projlist-query");
    if (elemq && elemq.value.length > 0) {
        url += ("&q=" + elemq.value);
    }
    var alert_id = "#htracker-projlist-alert";

    if (!list_active && htrackerProj.listMenuActive) {
        list_active = htrackerProj.listMenuActive;
    }
    if (list_active != "proj/list/history") {
        list_active = "proj/list/active";
        options.offset = null;
    } else {
        url += "&filter_closed=true";
    }

    if (options.offset) {
        url += "&offset=" + options.offset;
    }

    htrackerProj.listMenuActive = list_active;
    htrackerProj.procListMenuActive = null;

    // htracker.ModuleNavbarMenuRefresh("htracker-projlist-menus");
    // htracker.ModuleNavbarMenu("projlist", htrackerProj.ListMenus, list_active);
    htracker.ModuleNavbarLeftClean();
    htracker.OpToolsRefresh("#htracker-projlist-optools");

    htracker.ApiCmd("proj/list?" + url, {
        callback: function(err, data) {

            $("#htracker-projlist-more").css({
                "display": "none"
            });
            htrackerProj.listOffset = null;

            if (err) {
                $("#htracker-projlist").empty();
                return l4i.InnerAlert(alert_id, "error", err);
            }
            if (data.error) {
                $("#htracker-projlist").empty();
                return l4i.InnerAlert(alert_id, "error", data.error.message);
            }
            if (!data.items || data.items.length < 1) {
                $("#htracker-projlist").empty();
                return l4i.InnerAlert(alert_id, "warn", "No Project Found");
            }

            var waiting = false;

            for (var i in data.items) {
                var filter_title = [];
                if (data.items[i].filter.proc_id > 0) {
                    filter_title.push("ProcID: " + data.items[i].filter.proc_id);
                } else if (data.items[i].filter.proc_name) {
                    filter_title.push("ProcName: " + data.items[i].filter.proc_name);
                } else if (data.items[i].filter.proc_cmd) {
                    filter_title.push("ProcCmd: " + data.items[i].filter.proc_cmd);
                }
                if (!data.items[i].proc_num) {
                    data.items[i].proc_num = 0;
                    waiting = true;
                }
                if (!data.items[i].closed) {
                    data.items[i].closed = 0;
                }
                data.items[i]._filter_title = filter_title.join(", ");
            }
            if (list_active == "proj/list/history") {
                data._history = true;
            }

            var elemt = document.getElementById("htracker-projlist");
            if (!elemt) {
                l4iTemplate.Render({
                    dstid: "htracker-projlist-table",
                    tplid: "htracker-projlist-table-tpl",
                    data: {
                        _history: data._history,
                    },
                });
            }

            var append = false;
            if (options.offset) {
                append = true;
            }
            l4iTemplate.Render({
                dstid: "htracker-projlist",
                tplid: "htracker-projlist-tpl",
                data: data,
                append: append,
            });

            l4i.InnerAlert(alert_id);
            htrackerProj.entryActiveId = null;

            if (list_active == "proj/list/history" && data.items.length >= htrackerProj.listLimit) {
                $("#htracker-projlist-more").css({
                    "display": "block"
                });
                htrackerProj.listOffset = data.items[data.items.length - 1].id;
            }

            if (waiting && list_active == "proj/list/active" && !htrackerProj.listAutoRefreshTimer) {
                htrackerProj.listAutoRefreshTimer = window.setTimeout(function() {
                    htrackerProj.listAutoRefreshTimer = null;
                    htrackerProj.ListRefresh();
                }, 10000);
            }
        },
    });
}


htrackerProj.ListRefreshQuery = function() {
    htrackerProj.ListRefresh();
}

htrackerProj.ListMore = function() {
    htrackerProj.ListRefresh(null, {
        offset: htrackerProj.listOffset,
    });
}


htrackerProj.EntryView = function(id) {

    htracker.ApiCmd("proj/entry?id=" + id, {
        callback: function(err, data) {

            if (err) {
                return l4iAlert.Open("error", "Failed to get proj (#" + pid + ")");
            }

            if (data.error) {
                return l4iAlert.Open("error", data.error.message);
            }

            htrackerProj.entryActive = data;

            l4iModal.Open({
                id: "htracker-proj-new",
                title: "Project Overview",
                data: data,
                tplid: "htracker-proj-entry-tpl",
                width: 900,
                height: 500,
                buttons: [{
                    title: "Trace by Name",
                    onclick: "htrackerProj.TraceByName()",
                    style: "btn-primary",
                }, {
                    title: "Close",
                    onclick: "l4iModal.Close()",
                }],
            });
        },
    });
}

htrackerProj.newEntryOptions = null;

htrackerProj.NewEntry = function(options) {
    options = options || {};
    if (!options.modal_id) {
        options.modal_id = "htracker-projnew";
    }
    var fn = null;
    if (!options.filter) {
        fn = htrackerProj.NewEntrySelector;
    } else {
        if (options.filter.proc_id && options.filter.proc_id > 0) {
            options.modal_id = "htracker-projnew-pid";
            fn = htrackerProj.NewEntryProcId;
        // htrackerProj.NewEntryProcId(options);
        } else if (options.filter.proc_name) {
            options.modal_id = "htracker-projnew-pname";
            fn = htrackerProj.NewEntryProcName;
        // htrackerProj.NewEntryProcName(options);
        }
    }
    htrackerProj.newEntryOptions = options;
    if (fn) {
        fn(htrackerProj.newEntryOptions);
    }
}

htrackerProj.NewEntrySelector = function(options) {

    l4iModal.Open({
        id: options.modal_id,
        title: "New Project",
        tpluri: htracker.TplPath("proj/entry-new-selector"),
        width: 900,
        height: 450,
        buttons: [{
            title: "Close",
            onclick: "l4iModal.Close()",
        }],
    });
}

htrackerProj.NewEntryProcId = function(options) {

    options = options || {};
    options.modal_id = options.modal_id || "htracker-projnew-pid";
    options.filter = options.filter || {};
    options.filter.proc_id = options.filter.proc_id || "";

    l4iModal.Open({
        id: options.modal_id,
        title: "Filter by Process ID",
        data: {
            name: "",
            filter: options.filter,
        },
        tpluri: htracker.TplPath("proj/entry-new-pid"),
        width: 900,
        height: 450,
        backEnable: true,
        buttons: [{
            title: "Next",
            onclick: "htrackerProj.NewEntryCommit()",
            style: "btn-primary",
        }],
    });
}

htrackerProj.NewEntryProcName = function(options) {

    options = options || {};
    if (!options.modal_id) {
        options.modal_id = "htracker-projnew-pname";
    }
    options.filter = options.filter || {};
    options.filter.proc_name = options.filter.proc_name || "";

    l4iModal.Open({
        id: options.modal_id,
        title: "Filter by Process Name",
        data: {
            name: "",
            filter: options.filter,
        },
        tpluri: htracker.TplPath("proj/entry-new-pname"),
        width: 900,
        height: 450,
        backEnable: true,
        buttons: [{
            title: "Next",
            onclick: "htrackerProj.NewEntryCommit()",
            style: "btn-primary",
        }],
    });
}

htrackerProj.NewEntryProcCommand = function(options) {

    options = options || {};
    if (!options.modal_id) {
        options.modal_id = "htracker-projnew-pcmd";
    }
    options.filter = options.filter || {};
    options.filter.proc_cmd = options.filter.proc_cmd || "";

    l4iModal.Open({
        id: options.modal_id,
        title: "Filter by Process Command line content",
        data: {
            name: "",
            filter: options.filter,
        },
        tpluri: htracker.TplPath("proj/entry-new-pcmd"),
        width: 900,
        height: 450,
        backEnable: true,
        buttons: [{
            title: "Next",
            onclick: "htrackerProj.NewEntryCommit()",
            style: "btn-primary",
        }],
    });
}


htrackerProj.NewEntryCommit = function() {
    var alert_id = "#htracker-projset-alert";
    var req = {
        filter: {},
    };
    try {
        var elem = $("#htracker_projset_proc_id");
        if (elem) {
            req.filter.proc_id = parseInt(elem.val());
        }
        elem = $("#htracker_projset_proc_name");
        if (elem) {
            req.filter.proc_name = elem.val();
        }
        elem = $("#htracker_projset_proc_cmd");
        if (elem) {
            req.filter.proc_cmd = elem.val();
        }

        elem = $("#htracker_projset_name");
        if (elem) {
            req.name = elem.val();
        }
    } catch (err) {
        return l4i.InnerAlert(alert_id, 'error', err);
    }

    htracker.ApiCmd("proj/set", {
        method: "POST",
        data: JSON.stringify(req),
        callback: function(err, rsj) {

            if (err) {
                return l4i.InnerAlert(alert_id, 'error', err);
            }

            if (!rsj || rsj.kind != "ProjEntry") {
                var msg = "Bad Request";
                if (rsj.error) {
                    msg = rsj.error.message;
                }
                return l4i.InnerAlert(alert_id, 'error', msg);
            }

            l4i.InnerAlert(alert_id, 'ok', "Successful operation");

            window.setTimeout(function() {
                htrackerProj.ListRefresh("proj/list/active");
                l4iModal.Close();
            }, 1000);
        }
    })
}

htrackerProj.EntryDel = function(id, is_confirm) {

    if (!id) {
        if (!htrackerProj.entryActiveId) {
            return;
        }
        id = htrackerProj.entryActiveId;
    }

    if (!is_confirm) {
        l4iModal.Open({
            title: "Remove this Project",
            tplsrc: '<div id="hpm-node-del" class="alert alert-danger">Are you sure to delete this Project?</div>',
            width: 600,
            height: 200,
            buttons: [{
                title: "Confirm and Remove",
                onclick: "htrackerProj.EntryDel(\"" + id + "\", true)",
                style: "btn-danger",
            }, {
                title: "Cancel",
                onclick: "l4iModal.Close()",
                style: "btn-primary",
            }],
        });
        return;
    }

    var alert_id = "#hpm-node-del";
    l4i.InnerAlert(alert_id, 'warn', "pending ...");

    htracker.ApiCmd("proj/del?id=" + id, {
        callback: function(err, rsj) {

            if (err) {
                return l4i.InnerAlert(alert_id, 'error', err);
            }

            if (!rsj || rsj.kind != "ProjEntry") {
                var msg = "Bad Request";
                if (rsj.error) {
                    msg = rsj.error.message;
                }
                return l4i.InnerAlert(alert_id, 'error', msg);
            }

            // $("#proj-" + id).remove();

            l4i.InnerAlert(alert_id, 'ok', "Successful operation");
            window.setTimeout(function() {
                htrackerProj.Index();
                l4iModal.Close();
            }, 1000);
        }
    });
}


htrackerProj.ProcIndex = function(proj_id) {

    if (!proj_id) {
        proj_id = htrackerProj.entryActiveId;
        if (!proj_id) {
            proj_id = l4iSession.Get("htproj_active_id");
        }
    } else {
        htrackerProj.entryActiveId = proj_id;
    }

    if (!proj_id) {
        console.log("ProjId Not Setup");
        return;
    }
    l4iSession.Set("htproj_active_id", proj_id);

    if (!htrackerProj.procListMenuActive) {
        htrackerProj.procListMenuActive = "proj/proc/hit";
    }

    htracker.Loader("#htracker-module-content", "proj/proc-list", {
        callback: function() {
            htracker.ModuleNavbarLeftClean();

            if (htrackerProj.listMenuActive == "proj/list/history") {
                htracker.ModuleNavbarMenuClean();
                return htrackerProj.ProcListExit();
            }

            htracker.ModuleNavbarMenu("projProcMenus", htrackerProj.procListMenus);
            l4i.UrlEventRegister("proj/proc/hit", htrackerProj.ProcListHit, "htracker-module-navbar-menus");
            l4i.UrlEventRegister("proj/proc/exit", htrackerProj.ProcListExit, "htracker-module-navbar-menus");
            l4i.UrlEventHandler(htrackerProj.procListMenuActive, true);
        },
    });
}

htrackerProj.ProcListHit = function() {
    htrackerProj.ProcList(null, "proj/proc/hit");
}

htrackerProj.ProcListExit = function() {
    htrackerProj.ProcList(null, "proj/proc/exit");
}


htrackerProj.ProcList = function(proj_id, list_active) {

    if (!proj_id) {
        proj_id = htrackerProj.entryActiveId;
        if (!proj_id) {
            proj_id = l4iSession.Get("htproj_active_id");
        }
    } else {
        htrackerProj.entryActiveId = proj_id;
    }

    if (!proj_id) {
        console.log("ProjId Not Setup");
        return;
    }
    var alert_id = "#htracker-proj-proclist-alert";
    var url = "proj_id=" + proj_id;

    if (htrackerProj.listMenuActive == "proj/list/history") {
        list_active = "proj/proc/exit";
    }

    if (!list_active && htrackerProj.procListMenuActive) {
        list_active = htrackerProj.procListMenuActive;
    }
    if (list_active != "proj/proc/exit") {
        list_active = "proj/proc/hit";
    } else {
        url += "&filter_exit=true";
    }

    htrackerProj.procListMenuActive = list_active;
    l4i.InnerAlert(alert_id);

    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("data", function(data) {

            htracker.ModuleNavbarLeftRefresh("htracker-proj-proclist-menus");
            htracker.OpToolsRefresh("#htracker-proj-proclist-optools");
            if (list_active == "proj/proc/exit") {
                // htracker.OpToolsClean();
            } else {
            }

            if (data.error) {
                $("#htracker-proj-proclist").empty();
                return l4i.InnerAlert(alert_id, 'error', data.error.message);
            }
            if (!data.items) {
                $("#htracker-proj-proclist").empty();
                return l4i.InnerAlert(alert_id, 'error', "No Process Found");
            }

            for (var i in data.items) {
                if (!data.items[i].cmd) {
                    data.items[i].cmd = "";
                }
                if (!data.items[i].exited) {
                    data.items[i].exited = 0;
                }
            }

            if (list_active == "proj/proc/exit") {
                data._exit = true;
            } else {
                data._hit = true;
            }


            l4iTemplate.Render({
                dstid: "htracker-proj-proclist",
                tplid: "htracker-proj-proclist-tpl",
                data: data,
            });
        });

        ep.fail(function(err) {
            alert("NetWork error, Please try again later");
        });

        // htracker.TplCmd("proj/proc-list", {
        //     callback: ep.done("tpl"),
        // });


        htracker.ApiCmd("proj/proc-list?" + url, {
            callback: ep.done("data"),
        });
    });
}



htrackerProj.procStatsActiveProcId = null;
htrackerProj.procStatsActiveProcTime = null;
htrackerProj.ProcStats = function(proj_id, proc_id, proc_time) {
    htrackerProj.procStatsActiveProcId = proc_id;
    htrackerProj.procStatsActiveProcTime = proc_time;
    htrackerProj.NodeStats(null);
}

htrackerProj.NodeStatsButton = function(obj) {
    $("#htracker-module-navbar-optools").find(".hover").removeClass("hover");
    obj.setAttribute("class", 'hover');
    htrackerProj.NodeStats(parseInt(obj.getAttribute('value')));
}

htrackerProj.nodeStatsFeedMaxValue = function(feed, names) {
    var max = 0;
    var arr = names.split(",");
    for (var i in feed.items) {
        if (arr.indexOf(feed.items[i].name) < 0) {
            continue;
        }
        for (var j in feed.items[i].items) {
            if (feed.items[i].items[j].value > max) {
                max = feed.items[i].items[j].value;
            }
        }
    }
    return max;
}

htrackerProj.NodeStats = function(time_past) {

    if (time_past) {
        htrackerProj.proc_stats_active_past = parseInt(time_past);
        if (!htrackerProj.proc_stats_active_past) {
            htrackerProj.proc_stats_active_past = 86400;
        }
    }
    if (htrackerProj.proc_stats_active_past < 600) {
        htrackerProj.proc_stats_active_past = 600;
    }
    if (htrackerProj.proc_stats_active_past > (30 * 86400)) {
        htrackerProj.proc_stats_active_past = 30 * 86400;
    }
    htracker.ModuleNavbarLeftClean();

    var stats_url = "proc_id=" + htrackerProj.procStatsActiveProcId;
    stats_url += "&proc_time=" + htrackerProj.procStatsActiveProcTime;

    var stats_query = {
        tc: 180,
        tp: htrackerProj.proc_stats_active_past,
        is: [
            {
                n: "cpu/p"
            },
            /*
                {
                    n: "cpu/sys",
                    d: true
                },
                {
                    n: "cpu/user",
                    d: true
                },
            */
            {
                n: "mem/rss"
            },
            /*
                {
                    n: "mem/vms"
                },
            */
            {
                n: "mem/data"
            },
            {
                n: "net/c"
            },
            {
                n: "net/rc",
                d: true
            },
            {
                n: "net/wc",
                d: true
            },
            {
                n: "net/rb",
                d: true
            },
            {
                n: "net/wb",
                d: true
            },
            {
                n: "io/rc",
                d: true
            },
            {
                n: "io/wc",
                d: true
            },
            {
                n: "io/rb",
                d: true
            },
            {
                n: "io/wb",
                d: true
            },
            {
                n: "io/fd"
            },
            {
                n: "io/td"
            }
        ],
    };

    var wlimit = 900;
    var tfmt = "";
    var ww = $(window).width();
    var hh = $(window).height();
    if (ww > wlimit) {
        ww = wlimit;
    }
    if (hh < 800) {
        htrackerProj.hchart_def.options.height = "150px";
    } else {
        htrackerProj.hchart_def.options.height = "200px";
    }
    if (stats_query.tp > (10 * 86400)) {
        stats_query.tc = 6 * 3600;
        tfmt = "m-d H";
    } else if (stats_query.tp > (3 * 86400)) {
        stats_query.tc = 3 * 3600;
        tfmt = "m-d H";
    } else if (stats_query.tp > 86400) {
        stats_query.tc = 3600;
        tfmt = "m-d H";
    } else if (stats_query.tp >= (3 * 3600)) {
        stats_query.tc = 1800;
        tfmt = "H:i";
    } else if (stats_query.tp >= (3 * 600)) {
        stats_query.tc = 120;
        tfmt = "H:i";
    } else {
        stats_query.tc = 60;
        tfmt = "i:s";
    }

    stats_url += "&qry=" + btoa(JSON.stringify(stats_query));
    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("tpl", "stats", function(tpl, stats) {

            if (tpl) {
                $("#htracker-module-content").html(tpl);
                $(".htracker-proj-procstats-item").css({
                    "flex-basis": ww + "px"
                });
                htracker.ModuleNavbarMenuRefresh("htracker-proj-procstats-menus");
                htracker.OpToolsRefresh("#htracker-proj-node-optools-stats");
            }

            var max = 0;
            var tc_title = stats.cycle + " seconds";
            if (stats.cycle >= 86400 && stats.cycle % 86400 == 0) {
                tc_title = (stats.cycle / 86400) + " Day";
                if (stats.cycle > 86400) {
                    tc_title += "s";
                }
            } else if (stats.cycle >= 3600 && stats.cycle % 3600 == 0) {
                tc_title = (stats.cycle / 3600) + " Hour";
                if (stats.cycle > 3600) {
                    tc_title += "s";
                }
            } else if (stats.cycle >= 60 && stats.cycle % 60 == 0) {
                tc_title = (stats.cycle / 60) + " Minute";
                if (stats.cycle > 60) {
                    tc_title += "s";
                }
            }


            //
            var stats_cpu = l4i.Clone(htrackerProj.hchart_def);
            stats_cpu.options.title = l4i.T("CPU Usage %");

            //
            var stats_mem = l4i.Clone(htrackerProj.hchart_def);
            stats_mem.options.title = l4i.T("Memory Usage (MB)");
            stats_mem._fix = 1024 * 1024;

            //
            var stats_netcc = l4i.Clone(htrackerProj.hchart_def);
            stats_netcc.options.title = l4i.T("Network Connections");

            //
            var stats_netc = l4i.Clone(htrackerProj.hchart_def);
            stats_netc.options.title = l4i.T("Network Packets / %s", tc_title);

            //
            var stats_netb = l4i.Clone(htrackerProj.hchart_def);
            max = htrackerProj.nodeStatsFeedMaxValue(stats, "net/rb,net/wb");
            if (max > (1024 * 1024)) {
                stats_netb.options.title = l4i.T("Network Bytes (MB / %s)", tc_title);
                stats_netb._fix = 1024 * 1024;
            } else if (max > 1024) {
                stats_netb.options.title = l4i.T("Network Bytes (KB / %s)", tc_title);
                stats_netb._fix = 1024;
            } else {
                stats_netb.options.title = l4i.T("Network Bytes (Bytes / %s)", tc_title);
            }

            //
            var stats_ioc = l4i.Clone(htrackerProj.hchart_def);
            stats_ioc.options.title = l4i.T("IO Count / %s", tc_title);

            //
            var stats_iob = l4i.Clone(htrackerProj.hchart_def);
            max = htrackerProj.nodeStatsFeedMaxValue(stats, "io/rb,io/wb");
            if (max > (1024 * 1024)) {
                stats_iob.options.title = l4i.T("IO Bytes (MB / %s)", tc_title);
                stats_iob._fix = 1024 * 1024;
            } else if (max > 1024) {
                stats_iob.options.title = l4i.T("IO Bytes (KB / %s)", tc_title);
                stats_iob._fix = 1024;
            } else {
                stats_iob.options.title = l4i.T("IO Bytes (Bytes / %s)", tc_title);
            }

            //
            var stats_iofd = l4i.Clone(htrackerProj.hchart_def);
            stats_iofd.options.title = l4i.T("Number of File Descriptors");


            //
            var stats_iotd = l4i.Clone(htrackerProj.hchart_def);
            stats_iotd.options.title = l4i.T("Number of Threads");



            //
            for (var i in stats.items) {

                var v = stats.items[i];
                var dataset = {
                    data: []
                };
                var labels = [];
                var fix = 1;
                switch (v.name) {
                    case "cpu/p":
                        fix = 10000;
                        break;

                    case "mem/rss":
                    case "mem/vms":
                    case "mem/data":
                        if (stats_mem._fix && stats_mem._fix > 1) {
                            fix = stats_mem._fix;
                        }
                        break;


                    case "net/rb":
                    case "net/wb":
                        if (stats_netb._fix && stats_netb._fix > 1) {
                            fix = stats_netb._fix;
                        }
                        break;

                    case "io/rb":
                    case "io/wb":
                        if (stats_iob._fix && stats_iob._fix > 1) {
                            fix = stats_iob._fix;
                        }
                        break;
                }

                for (var j in v.items) {

                    var v2 = v.items[j];

                    var t = new Date(v2.time * 1000);
                    labels.push(t.l4iTimeFormat(tfmt));

                    if (!v2.value) {
                        v2.value = 0;
                    }

                    if (fix > 1) {
                        v2.value = (v2.value / fix).toFixed(2);
                    } else {
                        v2.value = parseInt(v2.value);
                    }

                    dataset.data.push(v2.value);
                }

                switch (v.name) {
                    case "cpu/p":
                        stats_cpu.data.labels = labels;
                        dataset.label = "CPU Percent";
                        stats_cpu.data.datasets.push(dataset);
                        break;

                    case "mem/rss":
                        stats_mem.data.labels = labels;
                        dataset.label = "RSS";
                        stats_mem.data.datasets.push(dataset);
                        break

                    case "mem/vms":
                        stats_mem.data.labels = labels;
                        dataset.label = "VMS";
                        stats_mem.data.datasets.push(dataset);
                        break

                    case "mem/data":
                        stats_mem.data.labels = labels;
                        dataset.label = "Data";
                        stats_mem.data.datasets.push(dataset);
                        break

                    case "net/c":
                        stats_netcc.data.labels = labels;
                        dataset.label = "Connections";
                        stats_netcc.data.datasets.push(dataset);
                        break

                    case "net/rc":
                        stats_netc.data.labels = labels;
                        dataset.label = "Read Packets";
                        stats_netc.data.datasets.push(dataset);
                        break

                    case "net/rb":
                        stats_netb.data.labels = labels;
                        dataset.label = "Read Bytes";
                        stats_netb.data.datasets.push(dataset);
                        break

                    case "net/wc":
                        stats_netc.data.labels = labels;
                        dataset.label = "Sent Packets";
                        stats_netc.data.datasets.push(dataset);
                        break

                    case "net/wb":
                        stats_netb.data.labels = labels;
                        dataset.label = "Sent Bytes";
                        stats_netb.data.datasets.push(dataset);
                        break

                    case "io/rc":
                        stats_ioc.data.labels = labels;
                        dataset.label = "Read Count";
                        stats_ioc.data.datasets.push(dataset);
                        break

                    case "io/rb":
                        stats_iob.data.labels = labels;
                        dataset.label = "Read Bytes";
                        stats_iob.data.datasets.push(dataset);
                        break

                    case "io/wc":
                        stats_ioc.data.labels = labels;
                        dataset.label = "Write Count";
                        stats_ioc.data.datasets.push(dataset);
                        break

                    case "io/wb":
                        stats_iob.data.labels = labels;
                        dataset.label = "Write Bytes";
                        stats_iob.data.datasets.push(dataset);
                        break

                    case "io/fd":
                        stats_iofd.data.labels = labels;
                        dataset.label = "FDs";
                        stats_iofd.data.datasets.push(dataset);
                        break

                    case "io/td":
                        stats_iotd.data.labels = labels;
                        dataset.label = "Threads";
                        stats_iotd.data.datasets.push(dataset);
                        break
                }
            }

            hooto_chart.RenderElement(stats_cpu, "htracker-proj-node-stats-cpu");
            hooto_chart.RenderElement(stats_mem, "htracker-proj-node-stats-mem");
            hooto_chart.RenderElement(stats_netcc, "htracker-proj-node-stats-netcc");
            hooto_chart.RenderElement(stats_netc, "htracker-proj-node-stats-netc");
            hooto_chart.RenderElement(stats_netb, "htracker-proj-node-stats-netb");
            hooto_chart.RenderElement(stats_ioc, "htracker-proj-node-stats-ioc");
            hooto_chart.RenderElement(stats_iob, "htracker-proj-node-stats-iob");
            hooto_chart.RenderElement(stats_iofd, "htracker-proj-node-stats-iofd");
            hooto_chart.RenderElement(stats_iotd, "htracker-proj-node-stats-iotd");
        });

        ep.fail(function(err) {
            alert("Network Connection Error, Please try again later (EC:htracker-proj-node)");
        });

        htracker.ApiCmd("proj/proc-stats?" + stats_url, {
            callback: ep.done("stats"),
        });

        htracker.TplCmd("proj/proc-stats", {
            callback: ep.done("tpl"),
        });
    });
}


htrackerProj.procDyTraceListProcId = 0;
htrackerProj.procDyTraceListProcTime = 0;

htrackerProj.ProcDyTraceList = function(proj_id, pid, pcreated, options) {

    if (!proj_id) {
        proj_id = htrackerProj.entryActiveId;
        if (!proj_id) {
            proj_id = l4iSession.Get("htproj_active_id");
            if (!proj_id) {
                return;
            }
        }
    }

    htrackerProj.procDyTraceListProcId = pid;
    htrackerProj.procDyTraceListProcTime = pcreated;

    options = options || {};

    htracker.ModuleNavbarLeftClean();

    var alert_id = "#htracker-proj-ptrace-list-alert";
    var url = "proj_id=" + proj_id + "&proc_id=" + pid;
    url += "&proc_time=" + pcreated + "&limit=" + htrackerProj.procTraceListLimit;
    if (options.offset) {
        url += "&offset=" + options.offset;
    }

    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("tpl", "data", function(tpl, data) {

            if (tpl) {
                $("#htracker-module-content").html(tpl);
                htracker.ModuleNavbarMenuRefresh("htracker-proj-ptrace-list-menus");
                htracker.OpToolsClean("#htracker-proj-proclist-optools");
            }

            if (data.error) {
                return l4i.InnerAlert(alert_id, 'error', data.error.message);
            }
            if (!data.items) {
                if (!options.offset) {
                    return l4i.InnerAlert(alert_id, 'warn', "No Process/Trace Found");
                }
                data.items = [];
            }
            if (options.offset && data.items.length < 1) {
                return $("#htracker-proj-ptrace-list-more").css({
                    "display": "none"
                });
            }
            for (var i in data.items) {
                if (!data.items[i].perf_size) {
                    data.items[i].perf_size = 0;
                }
            }

            var append = false;
            if (options.offset) {
                append = true;
            }

            l4iTemplate.Render({
                dstid: "htracker-proj-ptrace-list",
                tplid: "htracker-proj-ptrace-list-tpl",
                data: data,
                append: append,
                callback: function() {
                    if (data.items.length >= htrackerProj.procTraceListLimit) {
                        $("#htracker-proj-ptrace-list-more").css({
                            "display": "block"
                        });
                    } else {
                        $("#htracker-proj-ptrace-list-more").css({
                            "display": "none"
                        });
                    }
                    if (data.items.length > 0) {
                        htrackerProj.procTraceListOffset = data.items[data.items.length - 1].created;
                    }
                },
            });
        });

        ep.fail(function(err) {
            alert("NetWork error, Please try again later");
        });

        if (options.offset) {
            ep.emit("tpl", null);
        } else {
            htracker.TplCmd("proj/ptrace-list", {
                callback: ep.done("tpl"),
            });
        }

        htracker.ApiCmd("proj/proc-trace-list?" + url, {
            callback: ep.done("data"),
        });
    });
}

htrackerProj.ProcDyTraceListMore = function() {
    htrackerProj.ProcDyTraceList(
        htrackerProj.entryActiveId,
        htrackerProj.procDyTraceListProcId,
        htrackerProj.procDyTraceListProcTime, {
            offset: htrackerProj.procTraceListOffset,
        });
}

htrackerProj.flamegraphRender = null;
htrackerProj.flamegraphRenderType = "svg";

htrackerProj.ProcDyTraceView = function(pid, pcreated, created) {

    if (!htrackerProj.entryActiveId) {
        htrackerProj.entryActiveId = l4iSession.Get("htproj_active_id");
    }

    htrackerProj.flamegraphRender = null;
    var url = "proj_id=" + htrackerProj.entryActiveId;
    url += "&created=" + created;
    url += "&proc_id=" + pid;
    url += "&proc_time=" + pcreated;

    var api_svg = htracker.api + "/proj/proc-trace-graph/?" + url;


    var buttons = [{
        title: "Open in new tab",
        href: api_svg,
    }];

    /*
    if (htrackerProj.flamegraphRenderType == "js") {
        buttons.push({
            title: "Reset Zoom",
            onclick: "htrackerProj.ProcDyTraceViewReset()",
        });
    }
    */
    buttons.push({
        title: "Close",
        onclick: "l4iModal.Close()",
    });



    l4iModal.Open({
        title: "On-CPU Flame Graph",
        tplsrc: "<div id='htracker-proj-flamegraph-body'></div>",
        width: "max",
        height: "max",
        buttons: buttons,
        callback: function() {

            $("#htracker-proj-flamegraph-body").html("loading ...");

            if (htrackerProj.flamegraphRenderType == "svg") {

                var api_url = htracker.api + "/proj/proc-trace-graph/?" + url;

                var obj = l4i.T('<object data="%s" type="image/svg+xml" width="%d" height="%d"></object>',
                    api_svg, l4iModal.CurOptions.inlet_width, l4iModal.CurOptions.inlet_height - 5);
                    // console.log("w " + l4iModal.CurOptions.inlet_width);
                    // console.log("h " + l4iModal.CurOptions.inlet_height);

                $("#htracker-proj-flamegraph-body").html(obj);
            }

        /*
            if (htrackerProj.flamegraphRenderType == "js") {


                htracker.ApiCmd("proj/proc-trace-graph-burn?" + url, {
                    callback: function(err, data) {
                        if (err || data.error) {
                            return; // TODO
                        }

                        htrackerProj.flamegraphRender = d3.flamegraph();
                        if (l4iModal.CurOptions.inlet_width) {
                            htrackerProj.flamegraphRender.width(l4iModal.CurOptions.inlet_width);
                            htrackerProj.flamegraphRender.height(l4iModal.CurOptions.inlet_height - 5);
                        }
                        htrackerProj.flamegraphRender.transitionDuration(300);
                        htrackerProj.flamegraphRender.cellHeight(16);
                        // htrackerProj.flamegraphRender.minFrameSize(2);
                        $("#htracker-proj-flamegraph-body").html("");
                        d3.select("#htracker-proj-flamegraph-body")
                            .datum(data.graph_burn)
                            .call(htrackerProj.flamegraphRender);
                    },
                });

            }
        */
        },
    });
}




htrackerProj.ProcDyTraceViewReset = function() {
    if (htrackerProj.flamegraphRender) {
        htrackerProj.flamegraphRender.resetZoom();
    }
}
