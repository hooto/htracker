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


var htrackerSys = {
    listLimit: 30,
    listLastUpdated: 0,
    listAutoRefreshTimer: null,
    listAutoRefreshTimeRange: 10000,
    entryActive: null,
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

htrackerSys.Index = function() {

    htracker.KeyUpEscHook = null;
    htracker.ModuleNavbarOn();
    htracker.ModuleNavbarLeftClean();
    htracker.ModuleNavbarMenuClean();

    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("tpl", "node", function(tpl, node) {

            $("#htracker-module-content").html(tpl);
            htracker.OpToolsRefresh("#htracker-sys-optools-stats");

            var items = [];

            if (node.spec.platform.kernel) {
                node.spec.platform.kernel = node.spec.platform.kernel.replace(/.x86_64$/g, '');
                items.push({
                    "name": "Kernel",
                    "value": node.spec.platform.kernel
                });
            }

            if (node.spec.capacity) {
                items.push({
                    "name": "CPU/Memory",
                    "value": node.spec.capacity.cpu + " / " + htracker.UtilResSizeFormat(node.spec.capacity.mem, 0),
                });
            }

            if (node.status.uptime) {
                var tu = parseInt((new Date()) / 1000) - node.status.uptime;
                items.push({
                    "name": l4i.T("Uptime"),
                    "value": htracker.UtilTimeUptime(tu),
                });
            }


            if (!node.status.volumes) {
                node.status.volumes = [];
            }
            for (var i in node.status.volumes) {
                if (!node.status.volumes[i].total) {
                    continue;
                }
                if (!node.status.volumes[i].used) {
                    node.status.volumes[i].used = 1;
                }
                items.push({
                    "name": l4i.T("Volume ") + node.status.volumes[i].name,
                    "value": htracker.UtilResSizeFormat(node.status.volumes[i].used, 0) + " / " +
                        htracker.UtilResSizeFormat(node.status.volumes[i].total, 0),
                });
            }

            if (!node.spec.exp_docker_version) {
                node.spec.exp_docker_version = "disable";
            }
            if (!node.spec.exp_pouch_version) {
                node.spec.exp_pouch_version = "disable";
            }

            l4iTemplate.Render({
                dstid: "htracker-sys-host-info",
                tplid: "htracker-sys-host-info-tpl",
                data: {
                    items: items,
                },
            });

            htrackerSys.NodeStats(3600);
        });

        ep.fail(function(err) {
            alert(l4i.T("Network error, Please try again later"));
        });

        htracker.ApiCmd("sys/item", {
            callback: ep.done("node"),
        });

        htracker.TplCmd("sys/index", {
            callback: ep.done("tpl"),
        });
    });
}


htrackerSys.NodeStatsButton = function(obj) {
    $("#htracker-module-navbar-optools").find(".hover").removeClass("hover");
    obj.setAttribute("class", 'hover');
    htrackerSys.NodeStats(parseInt(obj.getAttribute('value')));
}

htrackerSys.nodeStatsFeedMaxValue = function(feed, names) {
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

htrackerSys.NodeStats = function(time_past) {

    if (time_past) {
        htrackerSys.node_active_past = parseInt(time_past);
        if (!htrackerSys.node_active_past) {
            htrackerSys.node_active_past = 86400;
        }
    }
    if (htrackerSys.node_active_past < 600) {
        htrackerSys.node_active_past = 600;
    }
    if (htrackerSys.node_active_past > (30 * 86400)) {
        htrackerSys.node_active_past = 30 * 86400;
    }


    var stats_query = {
        tc: 180,
        tp: htrackerSys.node_active_past,
        is: [
            /**
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
                n: "cpu/p"
            },
            {
                n: "ram/us"
            },
            {
                n: "ram/cc"
            },
            {
                n: "net/rs",
                d: true
            },
            {
                n: "net/ws",
                d: true
            },
            {
                n: "fs/sp/rs",
                d: true
            },
            {
                n: "fs/sp/rn",
                d: true
            },
            {
                n: "fs/sp/ws",
                d: true
            },
            {
                n: "fs/sp/wn",
                d: true
            },
        ],
    };

    var wlimit = 700;
    var tfmt = "";
    var ww = $(window).width();
    var hh = $(window).height();
    if (ww > wlimit) {
        ww = wlimit;
    }
    if (hh < 800) {
        htrackerSys.hchart_def.options.height = "180px";
    } else {
        htrackerSys.hchart_def.options.height = "220px";
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

    var stats_url = "qry=" + btoa(JSON.stringify(stats_query));
    seajs.use(["ep"], function(EventProxy) {

        var ep = EventProxy.create("stats", function(stats) {

            $(".incp-podentry-stats-item").css({
                "flex-basis": ww + "px"
            });
            // inCp.OpToolsRefresh("#htracker-sys-node-optools-stats");

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
            var stats_cpu = l4i.Clone(htrackerSys.hchart_def);
            stats_cpu.options.title = l4i.T("CPU Usage %");

            //
            var stats_ram = l4i.Clone(htrackerSys.hchart_def);
            stats_ram.options.title = l4i.T("Memory Usage (MB)");
            stats_ram._fix = 1024 * 1024;

            //
            var stats_net = l4i.Clone(htrackerSys.hchart_def);
            max = htrackerSys.nodeStatsFeedMaxValue(stats, "net/rs,net/ws");
            if (max > (1024 * 1024)) {
                stats_net.options.title = l4i.T("Network Bytes (MB / %s)", tc_title);
                stats_net._fix = 1024 * 1024;
            } else if (max > 1024) {
                stats_net.options.title = l4i.T("Network Bytes (KB / %s)", tc_title);
                stats_net._fix = 1024;
            } else {
                stats_net.options.title = l4i.T("Network Bytes (Bytes / %s)", tc_title);
            }

            //
            var stats_fsn = l4i.Clone(htrackerSys.hchart_def);
            stats_fsn.options.title = l4i.T("Storage IO / %s", tc_title);

            //
            var stats_fss = l4i.Clone(htrackerSys.hchart_def);
            max = htrackerSys.nodeStatsFeedMaxValue(stats, "fs/sp/rs,fs/sp/ws");
            if (max > (1024 * 1024)) {
                stats_fss.options.title = l4i.T("Storage IO Bytes (MB / %s)", tc_title);
                stats_fss._fix = 1024 * 1024;
            } else if (max > 1024) {
                stats_fss.options.title = l4i.T("Storage IO Bytes (KB / %s)", tc_title);
                stats_fss._fix = 1024;
            } else {
                stats_fss.options.title = l4i.T("Storage IO Bytes (Bytes / %s)", tc_title);
            }


            for (var i in stats.items) {

                var v = stats.items[i];
                var dataset = {
                    data: []
                };
                var labels = [];
                var fix = 1;
                switch (v.name) {
                    /**
                                case "cpu/sys":
                                case "cpu/user":
                    */
                    case "cpu/p":
                        fix = 100;
                        break;

                    case "ram/us":
                    case "ram/cc":
                        if (stats_ram._fix && stats_ram._fix > 1) {
                            fix = stats_ram._fix;
                        }
                        break;


                    case "net/rs":
                    case "net/ws":
                        if (stats_net._fix && stats_net._fix > 1) {
                            fix = stats_net._fix;
                        }
                        break;

                    case "fs/sp/rs":
                    case "fs/sp/ws":
                        if (stats_fss._fix && stats_fss._fix > 1) {
                            fix = stats_fss._fix;
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
                    /**
                                case "cpu/sys":
                                    stats_cpu.data.labels = labels;
                                    dataset.label = "System";
                                    stats_cpu.data.datasets.push(dataset);
                                    break;

                                case "cpu/user":
                                    stats_cpu.data.labels = labels;
                                    dataset.label = "User";
                                    stats_cpu.data.datasets.push(dataset);
                                    break;
                    */
                    case "cpu/p":
                        stats_cpu.data.labels = labels;
                        dataset.label = "CPU Percent";
                        stats_cpu.data.datasets.push(dataset);
                        break;

                    case "ram/us":
                        stats_ram.data.labels = labels;
                        dataset.label = "Usage";
                        stats_ram.data.datasets.push(dataset);
                        break

                    case "ram/cc":
                        stats_ram.data.labels = labels;
                        dataset.label = "Cache";
                        stats_ram.data.datasets.push(dataset);
                        break

                    case "net/rs":
                        stats_net.data.labels = labels;
                        dataset.label = "Read";
                        stats_net.data.datasets.push(dataset);
                        break

                    case "net/ws":
                        stats_net.data.labels = labels;
                        dataset.label = "Send";
                        stats_net.data.datasets.push(dataset);
                        break

                    case "fs/sp/rs":
                        stats_fss.data.labels = labels;
                        dataset.label = "Read";
                        stats_fss.data.datasets.push(dataset);
                        break

                    case "fs/sp/ws":
                        stats_fss.data.labels = labels;
                        dataset.label = "Write";
                        stats_fss.data.datasets.push(dataset);
                        break

                    case "fs/sp/rn":
                        stats_fsn.data.labels = labels;
                        dataset.label = "Read";
                        stats_fsn.data.datasets.push(dataset);
                        break

                    case "fs/sp/wn":
                        stats_fsn.data.labels = labels;
                        dataset.label = "Write";
                        stats_fsn.data.datasets.push(dataset);
                        break
                }
            }

            hooto_chart.RenderElement(stats_cpu, "htracker-sys-node-stats-cpu");
            hooto_chart.RenderElement(stats_ram, "htracker-sys-node-stats-ram");
            hooto_chart.RenderElement(stats_net, "htracker-sys-node-stats-net");
            hooto_chart.RenderElement(stats_fss, "htracker-sys-node-stats-fss");
            hooto_chart.RenderElement(stats_fsn, "htracker-sys-node-stats-fsn");
        });

        ep.fail(function(err) {
            alert(l4i.T("Network error, Please try again later"));
        });

        htracker.ApiCmd("sys/stats-feed?" + stats_url, {
            callback: ep.done("stats"),
        });
    });
}

