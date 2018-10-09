var htrackerTracer = {
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
}

htrackerTracer.Index = function() {
    htracker.Loader("#htracker-module-content", "tracer/list", {
        callback: function() {
            htrackerTracer.ListRefresh();
        },
    });
}

htrackerTracer.ListRefresh = function() {

    htrackerTracer.entryActiveId = null;

    var elem = document.getElementById("htracker-tracer-list");
    if (!elem) {
        return;
    }
    var url = "limit=20";
    var elemq = document.getElementById("htracker-tracer-list-query");
    if (elemq && elemq.value.length > 0) {
        url += ("&q=" + elemq.value);
    }
    var alert_id = "#htracker-tracer-list-alert";

    htracker.ModuleNavbarMenuRefresh("htracker-tracer-list-menus");
    htracker.OpToolsRefresh("#htracker-tracer-list-optools");

    htracker.ApiCmd("tracer/list?" + url, {
        callback: function(err, data) {

            if (err) {
                return l4i.InnerAlert(alert_id, "error", err);
            }
            if (data.error) {
                return l4i.InnerAlert(alert_id, "error", data.error.message);
            }
            if (!data.items || data.items.length < 1) {
                return l4i.InnerAlert(alert_id, "warn", "No Tracer Found");
            }
            for (var i in data.items) {
                var filter_title = [];
                if (data.items[i].filter.proc_id > 0) {
                    filter_title.push("PID: " + data.items[i].filter.proc_id);
                } else if (data.items[i].filter.proc_name) {
                    filter_title.push("PNAME: " + data.items[i].filter.proc_name);
                }
                if (!data.items[i].proc_num) {
                    data.items[i].proc_num = 0;
                }
                data.items[i]._filter_title = filter_title.join(", ");
            }

            l4iTemplate.Render({
                dstid: "htracker-tracer-list",
                tplid: "htracker-tracer-list-tpl",
                data: data,
            });
        },
    });
}


htrackerTracer.ListRefreshQuery = function() {
    htrackerTracer.ListRefresh();
}

htrackerTracer.EntryView = function(id) {

    htracker.ApiCmd("tracer/entry?id=" + id, {
        callback: function(err, data) {

            if (err) {
                return l4iAlert.Open("error", "Failed to get tracer (#" + pid + ")");
            }

            if (data.error) {
                return l4iAlert.Open("error", data.error.message);
            }

            htrackerTracer.entryActive = data;

            l4iModal.Open({
                id: "htracker-tracer-new",
                title: "Tracer Overview",
                data: data,
                tplid: "htracker-tracer-entry-tpl",
                width: 900,
                height: 500,
                buttons: [{
                    title: "Trace by Name",
                    onclick: "htrackerTracer.TraceByName()",
                    style: "btn-primary",
                }, {
                    title: "Close",
                    onclick: "l4iModal.Close()",
                }],
            });
        },
    });
}

htrackerTracer.newEntryOptions = null;

htrackerTracer.NewEntry = function(options) {
    options = options || {};
    if (!options.modal_id) {
        options.modal_id = "htracker-trace-new";
    }
    if (!options.filter) {
    } else {
        if (options.filter.proc_id && options.filter.proc_id > 0) {
            options.modal_id = "htracker-trace-new-pid";
            htrackerTracer.NewEntryProcessId(options);
        } else if (options.filter.proc_name) {
            options.modal_id = "htracker-trace-new-pname";
            htrackerTracer.NewEntryProcessName(options);
        }
    }
    htrackerTracer.newEntryOptions = options;
}

htrackerTracer.NewEntryProcessId = function(options) {

    l4iModal.Open({
        id: options.modal_id,
        title: "Tracer Settings",
        data: {
            filter: options.filter,
        },
        tpluri: htracker.TplPath("tracer/entry-new-pid"),
        width: 900,
        height: 450,
        backEnable: true,
        buttons: [{
            title: "Next",
            onclick: "htrackerTracer.NewEntryCommit()",
            style: "btn-primary",
        }],
    });
}

htrackerTracer.NewEntryProcessName = function(options) {

    l4iModal.Open({
        id: options.modal_id,
        title: "Tracer Settings",
        data: {
            filter: options.filter,
        },
        tpluri: htracker.TplPath("tracer/entry-new-pname"),
        width: 900,
        height: 450,
        backEnable: true,
        buttons: [{
            title: "Next",
            onclick: "htrackerTracer.NewEntryCommit()",
            style: "btn-primary",
        }],
    });
}


htrackerTracer.NewEntryCommit = function() {
    var alert_id = "#htracker-tracerset-alert";
    var req = {
        filter: {},
    };
    try {
        if (htrackerTracer.newEntryOptions.filter.proc_id) {
            req.filter.proc_id = parseInt($("#htracker_tracerset_proc_id").val());
        } else if (htrackerTracer.newEntryOptions.filter.proc_name) {
            req.filter.proc_name = $("#htracker_tracerset_proc_name").val();
        }
    } catch (err) {
        return l4i.InnerAlert(alert_id, 'error', err);
    }

    htracker.ApiCmd("tracer/set", {
        method: "POST",
        data: JSON.stringify(req),
        callback: function(err, rsj) {

            if (err) {
                return l4i.InnerAlert(alert_id, 'error', err);
            }

            if (!rsj || rsj.kind != "TracerEntry") {
                var msg = "Bad Request";
                if (rsj.error) {
                    msg = rsj.error.message;
                }
                return l4i.InnerAlert(alert_id, 'error', msg);
            }

            l4i.InnerAlert(alert_id, 'ok', "Successful operation");

            window.setTimeout(function() {
                l4iModal.Close();
            }, 600);
        }
    })
}

htrackerTracer.EntryDel = function(id, is_confirm) {

    if (!id) {
        if (!htrackerTracer.entryActiveId) {
            return;
        }
        id = htrackerTracer.entryActiveId;
    }

    if (!is_confirm) {
        l4iModal.Open({
            title: "Delete this Tracer",
            tplsrc: '<div id="hpm-node-del" class="alert alert-danger">Are you sure to delete this Tracer?</div>',
            width: 600,
            height: 200,
            buttons: [{
                title: "Confirm and Delete",
                onclick: "htrackerTracer.EntryDel(\"" + id + "\", true)",
                style: "btn-danger",
            }, {
                title: "Cancel",
                onclick: "l4iModal.Close()",
                style: "btn-primary",
            }],
        });
        return;
    }

    var alert_id = "#htracker-tracer-list-alert";

    htracker.ApiCmd("tracer/del?id=" + id, {
        callback: function(err, rsj) {

            if (err) {
                return l4i.InnerAlert(alert_id, 'error', err);
            }

            if (!rsj || rsj.kind != "TracerEntry") {
                var msg = "Bad Request";
                if (rsj.error) {
                    msg = rsj.error.message;
                }
                return l4i.InnerAlert(alert_id, 'error', msg);
            }

            $("#tracer-" + id).remove();
            l4iModal.Close();

            l4i.InnerAlert(alert_id, 'ok', "Successful operation");
            window.setTimeout(function() {
                l4i.InnerAlert(alert_id, '');
            }, 2000);

        }
    });
}


htrackerTracer.ProcList = function(tid) {
    if (!tid) {
        tid = htrackerTracer.entryActiveId;
    } else {
        htrackerTracer.entryActiveId = tid;
    }

    if (!tid) {
        return;
    }

    var alert_id = "#htracker-tracer-proc-list-alert";
    var url = "tracer_id=" + tid + "&limit=20";

    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("tpl", "data", function(tpl, data) {

            if (tpl) {
                $("#htracker-module-content").html(tpl);
                htracker.ModuleNavbarMenuRefresh("htracker-tracer-proc-list-menus");
                htracker.OpToolsRefresh("#htracker-tracer-proc-list-optools");

            }

            if (data.error) {
                return l4i.InnerAlert(alert_id, 'error', data.error.message);
            }
            if (!data.items) {
                return l4i.InnerAlert(alert_id, 'error', "No Process Found");
            }
            for (var i in data.items) {
                if (!data.items[i].cmd) {
                    data.items[i].cmd = "";
                }
            }

            l4iTemplate.Render({
                dstid: "htracker-tracer-proc-list",
                tplid: "htracker-tracer-proc-list-tpl",
                data: data,
            });
        });

        ep.fail(function(err) {
            alert("NetWork error, Please try again later");
        });

        htracker.TplCmd("tracer/proc-list", {
            callback: ep.done("tpl"),
        });


        htracker.ApiCmd("tracer/proc-list?" + url, {
            callback: ep.done("data"),
        });

    });
}



htrackerTracer.procStatsActiveProcId = null;
htrackerTracer.procStatsActiveProcTime = null;
htrackerTracer.ProcStats = function(tracer_id, proc_id, proc_time) {
    htrackerTracer.procStatsActiveProcId = proc_id;
    htrackerTracer.procStatsActiveProcTime = proc_time;
    htrackerTracer.NodeStats(null);
}

htrackerTracer.NodeStatsButton = function(obj) {
    $("#htracker-module-navbar-optools").find(".hover").removeClass("hover");
    obj.setAttribute("class", 'hover');
    htrackerTracer.NodeStats(parseInt(obj.getAttribute('value')));
}

htrackerTracer.nodeStatsFeedMaxValue = function(feed, names) {
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

htrackerTracer.NodeStats = function(time_past) {

    if (time_past) {
        htrackerTracer.proc_stats_active_past = parseInt(time_past);
        if (!htrackerTracer.proc_stats_active_past) {
            htrackerTracer.proc_stats_active_past = 86400;
        }
    }
    if (htrackerTracer.proc_stats_active_past < 600) {
        htrackerTracer.proc_stats_active_past = 600;
    }
    if (htrackerTracer.proc_stats_active_past > (30 * 86400)) {
        htrackerTracer.proc_stats_active_past = 30 * 86400;
    }

    var stats_url = "proc_id=" + htrackerTracer.procStatsActiveProcId;
    stats_url += "&proc_time=" + htrackerTracer.procStatsActiveProcTime;

    var stats_query = {
        tc: 180,
        tp: htrackerTracer.proc_stats_active_past,
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
        htrackerTracer.hchart_def.options.height = "150px";
    } else {
        htrackerTracer.hchart_def.options.height = "200px";
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
                $(".htracker-tracer-proc-stats-item").css({
                    "flex-basis": ww + "px"
                });
                htracker.ModuleNavbarMenuRefresh("htracker-tracer-proc-stats-menus");
                htracker.OpToolsRefresh("#htracker-tracer-node-optools-stats");
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
            var stats_cpu = l4i.Clone(htrackerTracer.hchart_def);
            stats_cpu.options.title = l4i.T("CPU Usage %");

            //
            var stats_mem = l4i.Clone(htrackerTracer.hchart_def);
            stats_mem.options.title = l4i.T("Memory Usage (MB)");
            stats_mem._fix = 1024 * 1024;

            //
            var stats_netcc = l4i.Clone(htrackerTracer.hchart_def);
            stats_netcc.options.title = l4i.T("Network Connections");

            //
            var stats_netc = l4i.Clone(htrackerTracer.hchart_def);
            stats_netc.options.title = l4i.T("Network Packets / %s", tc_title);

            //
            var stats_netb = l4i.Clone(htrackerTracer.hchart_def);
            max = htrackerTracer.nodeStatsFeedMaxValue(stats, "net/rb,net/wb");
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
            var stats_ioc = l4i.Clone(htrackerTracer.hchart_def);
            stats_ioc.options.title = l4i.T("IO Count / %s", tc_title);

            //
            var stats_iob = l4i.Clone(htrackerTracer.hchart_def);
            max = htrackerTracer.nodeStatsFeedMaxValue(stats, "io/rb,io/wb");
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
            var stats_iofd = l4i.Clone(htrackerTracer.hchart_def);
            stats_iofd.options.title = l4i.T("Number of File Descriptors");


            //
            var stats_iotd = l4i.Clone(htrackerTracer.hchart_def);
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

            hooto_chart.RenderElement(stats_cpu, "htracker-tracer-node-stats-cpu");
            hooto_chart.RenderElement(stats_mem, "htracker-tracer-node-stats-mem");
            hooto_chart.RenderElement(stats_netcc, "htracker-tracer-node-stats-netcc");
            hooto_chart.RenderElement(stats_netc, "htracker-tracer-node-stats-netc");
            hooto_chart.RenderElement(stats_netb, "htracker-tracer-node-stats-netb");
            hooto_chart.RenderElement(stats_ioc, "htracker-tracer-node-stats-ioc");
            hooto_chart.RenderElement(stats_iob, "htracker-tracer-node-stats-iob");
            hooto_chart.RenderElement(stats_iofd, "htracker-tracer-node-stats-iofd");
            hooto_chart.RenderElement(stats_iotd, "htracker-tracer-node-stats-iotd");
        });

        ep.fail(function(err) {
            alert("Network Connection Error, Please try again later (EC:htracker-tracer-node)");
        });

        htracker.ApiCmd("tracer/proc-stats?" + stats_url, {
            callback: ep.done("stats"),
        });

        htracker.TplCmd("tracer/proc-stats", {
            callback: ep.done("tpl"),
        });
    });
}


htrackerTracer.ProcDyTraceList = function(tid, pid, pcreated) {

    if (!tid) {
        tid = htrackerTracer.entryActiveId;
        if (!tid) {
            return;
        }
    }

    var alert_id = "#htracker-tracer-ptrace-list-alert";
    var url = "tracer_id=" + tid + "&proc_id=" + pid + "&proc_time=" + pcreated + "&limit=20";

    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("tpl", "data", function(tpl, data) {

            if (tpl) {
                $("#htracker-module-content").html(tpl);
                htracker.ModuleNavbarMenuRefresh("htracker-tracer-ptrace-list-menus");
                htracker.OpToolsClean("#htracker-tracer-proc-list-optools");
            }

            if (data.error) {
                return l4i.InnerAlert(alert_id, 'error', data.error.message);
            }
            if (!data.items) {
                return l4i.InnerAlert(alert_id, 'error', "No Process/Trace Found");
            }

            l4iTemplate.Render({
                dstid: "htracker-tracer-ptrace-list",
                tplid: "htracker-tracer-ptrace-list-tpl",
                data: data,
            });
        });

        ep.fail(function(err) {
            alert("NetWork error, Please try again later");
        });

        htracker.TplCmd("tracer/ptrace-list", {
            callback: ep.done("tpl"),
        });


        htracker.ApiCmd("tracer/proc-trace-list?" + url, {
            callback: ep.done("data"),
        });

    });
}

htrackerTracer.flamegraphRender = null;

htrackerTracer.ProcDyTraceView = function(pid, pcreated, created) {

    htrackerTracer.flamegraphRender = null;
    var url = "tracer_id=" + htrackerTracer.entryActiveId;
    url += "&created=" + created;
    url += "&proc_id=" + pid;
    url += "&proc_time=" + pcreated;


    l4iModal.Open({
        title: "On-CPU Flame Graph",
        tplsrc: "<div id='htracker-tracer-flamegraph-body'></div>",
        width: "max",
        height: "max",
        buttons: [{
            //     title: "Reset Zoom",
            //     onclick: "htrackerTracer.ProcDyTraceViewReset()",
            // }, {
            title: "Close",
            onclick: "l4iModal.Close()",
        }],
        _callback: function() {

            var api_url = h3tracker.api + "/tracer/proc-trace-graph/?" + url;
            api_url += "&svg_w=" + l4iModal.CurOptions.inlet_width;
            api_url += "&svg_h=" + l4iModal.CurOptions.inlet_height;

            var obj = l4i.T('<object data="%s" type="image/svg+xml" width="%d" height="%d"></object>',
                api_url, l4iModal.CurOptions.inlet_width, l4iModal.CurOptions.inlet_height - 5);
            console.log("w " + l4iModal.CurOptions.inlet_width);
            console.log("h " + l4iModal.CurOptions.inlet_height);

            $("#htracker-tracer-flamegraph-body").html(obj);
        },
        callback: function() {

            htracker.ApiCmd("tracer/proc-trace-graph-burn?" + url, {
                callback: function(err, data) {

                    htrackerTracer.flamegraphRender = d3.flamegraph();
                    if (l4iModal.CurOptions.inlet_width) {
                        htrackerTracer.flamegraphRender.width(l4iModal.CurOptions.inlet_width);
                        htrackerTracer.flamegraphRender.height(l4iModal.CurOptions.inlet_height - 5);
                    }
                    htrackerTracer.flamegraphRender.transitionDuration(300);
                    htrackerTracer.flamegraphRender.cellHeight(16);
                    // htrackerTracer.flamegraphRender.minFrameSize(2);
                    d3.select("#htracker-tracer-flamegraph-body")
                        .datum(data.graph_burn)
                        .call(htrackerTracer.flamegraphRender);
                },
            });
        },
    });
}




htrackerTracer.ProcDyTraceViewReset = function() {
    if (htrackerTracer.flamegraphRender) {
        htrackerTracer.flamegraphRender.resetZoom();
    }
}
