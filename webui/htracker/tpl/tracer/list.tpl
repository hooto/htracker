<div>
  <div class="htracker-div-container alert less-hide" id="htracker-tracer-list-alert"></div>
  <div class="htracker-div-light" id="htracker-tracer-list"></div>
</div>

<script type="text/html" id="htracker-tracer-list-menus">
<li>
  <button class="btn btn-sm btn-dark" id="htracker-tracer-list-menus-active">Current Active Projects</button>
</li>
<li>
  <button class="btn btn-sm" id="htracker-tracer-list-menus-history">History Projects</button>
</li>
</script>

<script type="text/html" id="htracker-tracer-list-optools">
<div>
  <!-- <form class="form-inline input-group" onsubmit="htrackerTracer.ListRefreshQuery(); return false;">
     <input class="form-control" type="search" placeholder="Search" id="htracker-tracer-list-query">
    <div class="input-group-append">
      <button class="btn btn-outline-secondary" type="button">Search</button>
    </div>
  </form> -->
</div>
<li>
  <button class="btn btn-outline-primary btn-sm" onclick="htrackerTracer.EntryNew()">
    <span class="icon16 icono-plus"></span>
    New Project
  </button>
</li>
</script>

<script type="text/html" id="htracker-tracer-list-tpl">
<table class="table table-hover valign-middle">
  <thead>
  <tr>
    <th>Name</th>
    <th>Filter</th>
    <th>Created</th>
    {[? it._history]}<th>Closed</th>{[?]}
    <th>Hit Processes</th>
    <th width="30"></th>
  </tr>
  </thead>
<tbody>
{[~it.items :v]}
<tr id="tracer-{[=v.id]}"
  {[if (v.proc_num > 0) {]}
  class="htracker-div-hover"
  onclick="htrackerTracer.ProcList('{[=v.id]}')"
  {[}]}>
  <td>{[=v.name]}</td>
  <td>{[=v._filter_title]}</td>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i")]}</td>
  {[? it._history]}<td>{[=l4i.UnixTimeFormat(v.closed, "Y-m-d H:i")]}</td>{[?]}
  <td>
    {[if (v.proc_num > 0) {]}
      {[=v.proc_num]}
    {[} else {]}
	  waiting
	{[}]}
  </td>
  <td align="right">
    {[if (v.proc_num > 0) {]}
    <i class="icono-caretRight" style="zoom: 85%; margin:0;"></i>
	{[}]}
  </td>
</tr>
{[~]}
</tbody>
</table>
</script>

