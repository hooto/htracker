
<div>
  <div class="htracker-div-light" id="htracker-sys-host-info"></div>
</div>

<div class="htracker-proj-procstats-box">
  <div id="htracker-sys-node-stats-cpu" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-sys-node-stats-ram" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-sys-node-stats-net" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-sys-node-stats-fss" class="htracker-proj-procstats-item htracker-div-light"></div>
  <div id="htracker-sys-node-stats-fsn" class="htracker-proj-procstats-item htracker-div-light"></div>
</div>


<script type="text/html" id="htracker-sys-host-info-tpl">
<table width="100%">
<tr>
{[~it.items :v]}
  <td>
    <strong>{[=v.name]}</strong>
    <p>{[=v.value]}</p>
  </td>
{[~]}
</tr>
</table>
</script>

<script type="text/html" id="htracker-sys-optools-stats">
<li class="item">{[=l4i.T("the Last")]}</li>
<li>
  <a href="#" value="3600" onclick="htrackerSys.NodeStatsButton(this)" class="l4i-nav-item hover">
    1 {[=l4i.T("Hour")]}
  </a>
</li>
<li>
  <a href="#" value="86400" onclick="htrackerSys.NodeStatsButton(this)" class="l4i-nav-item">
    24 {[=l4i.T("Hours")]}
  </a>
</li>
<li>
  <a href="#" value="259200" onclick="htrackerSys.NodeStatsButton(this)" class="l4i-nav-item">
    3 {[=l4i.T("Days")]}
  </a>
</li>
<li>
  <a href="#" value="864000" onclick="htrackerSys.NodeStatsButton(this)" class="l4i-nav-item">
    10 {[=l4i.T("Days")]}
  </a>
</li>
<li>
  <a href="#" value="2592000" onclick="htrackerSys.NodeStatsButton(this)" class="l4i-nav-item">
    30 {[=l4i.T("Days")]}
  </a>
</li>
</script>


