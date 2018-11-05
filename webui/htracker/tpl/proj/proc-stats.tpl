<div class="htracker-proj-procstats-box">
  <div id="htracker-proj-node-stats-cpu" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-mem" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-netcc" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-netc" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-netb" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-ioc" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-iob" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-iofd" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-proj-node-stats-iotd" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div class="htracker-proj-procstats-item"></div>
  <div class="htracker-proj-procstats-item"></div>
  <div class="htracker-proj-procstats-item"></div>
</div>

<script type="text/html" id="htracker-proj-procstats-menus">
<li>
  <button type="button" class="btn btn-outline-primary btn-sm" onclick="htrackerProj.ProcIndex()">
    <span class="icon16 icono-caretLeftCircle"></span>
    {[=l4i.T("Back to Process List")]}
  </button>
</li>
</script>

<script type="text/html" id="htracker-proj-node-optools-stats">
<li class="item">{[=l4i.T("the Last")]}</li>
<li>
  <a href="#" value="3600" onclick="htrackerProj.NodeStatsButton(this)" class="l4i-nav-item hover">
    1 {[=l4i.T("Hour")]}
  </a>
</li>
<li>
  <a href="#" value="86400" onclick="htrackerProj.NodeStatsButton(this)" class="l4i-nav-item">
    24 {[=l4i.T("Hours")]}
  </a>
</li>
<li>
  <a href="#" value="259200" onclick="htrackerProj.NodeStatsButton(this)" class="l4i-nav-item">
    3 {[=l4i.T("Days")]}
  </a>
</li>
<li>
  <a href="#" value="864000" onclick="htrackerProj.NodeStatsButton(this)" class="l4i-nav-item">
    10 {[=l4i.T("Days")]}
  </a>
</li>
<li>
  <a href="#" value="2592000" onclick="htrackerProj.NodeStatsButton(this)" class="l4i-nav-item">
    30 {[=l4i.T("Days")]}
  </a>
</li>
</script>


