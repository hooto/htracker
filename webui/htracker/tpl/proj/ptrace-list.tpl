
<div class="htracker-div-container alert less-hide" id="htracker-proj-ptrace-list-alert"></div>

<div class="htracker-div-light" id="htracker-proj-ptrace-list-box"></div>

<script type="text/html" id="htracker-proj-ptrace-list-box-tpl">
<table class="table table-hover valign-middle">
<thead>
  <tr>
    <th>{[=l4i.T("Start Time")]}</th>
    <th>{[=l4i.T("End Time")]}</th>
    <th>{[=l4i.T("Log Size")]}</th>
    <th style="text-align:right">{[=l4i.T("Flame Graph")]}</th>
  </tr>
</thead>
<tbody id="htracker-proj-ptrace-list"></tbody>
</table>

<div id="htracker-proj-ptrace-list-more" style="display: none; padding: 0 0 10px 10px">
  <button class="btn btn-primary btn-sm"
    onclick="htrackerProj.ProcDyTraceListMore()">
    {[=l4i.T("More items")]} ...
  </button>
</div>
</script>

<script type="text/html" id="htracker-proj-ptrace-list-menus">
<li>
  <button type="button" class="btn btn-outline-primary btn-sm" onclick="htrackerProj.ProcIndex()">
    <span class="icon16 icono-caretLeftCircle"></span>
    {[=l4i.T("Back to Process List")]}
  </button>
</li>
</script>

<script type="text/html" id="htracker-proj-ptrace-list-optools">
<li>
  <div id="htracker-proj-ptrace-list-status-msg" class="item-status-msg badge badge-light" style="display:none"></div>
</li>
<li>
  <button class="btn btn-outline-primary btn-sm" onclick="htrackerProj.ProcDyTraceNew()">
    <span class="icon16 icono-plus"></span>
    {[=l4i.T("New Trace Now")]}
  </button>
</li>
</script>


<script type="text/html" id="htracker-proj-ptrace-list-tpl">
{[~it.items :v]}
<tr id="{[=v._id]}">
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i:s")]}</td>
  <td id="{[=v._id]}-updated" value="{[=v.updated]}">
    {[if (v.updated > 100000000) {]}
      {[=l4i.UnixTimeFormat(v.updated, "Y-m-d H:i:s")]}
    {[}]}
  </td>
  <td id="{[=v._id]}-ps" value="{[=v.perf_size]}">
    {[if (v.perf_size > 0) {]}
      {[=htracker.UtilResSizeFormat(v.perf_size)]}
    {[} else if (v.updated < 100000000) {]}
      {[=l4i.T("tracing")]}
    {[} else if (v.updated > 100000000 && v.perf_size < 1) {]}
      {[=l4i.T("proc-trace-none-msg")]}
	{[}]}
  </td>
  <td align="right">
    {[if (v.updated > 100000000 && v.perf_size > 0) {]}
    <button class="btn btn-primary btn-sm"
	  onclick="htrackerProj.ProcDyTraceView({[=v.pid]}, {[=v.pcreated]}, {[=v.created]})">
	  On-CPU
	</button>
	{[}]}
  </td>
</tr>
{[~]}
</script>

