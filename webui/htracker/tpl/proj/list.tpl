<div>
  <div class="htracker-div-container alert less-hide" id="htracker-projlist-alert"></div>
  <div class="htracker-div-light" id="htracker-projlist"></div>
</div>

<script type="text/html" id="htracker-projlist-menus">
<li>
  <button class="btn btn-sm btn-dark" id="htracker-projlist-menus-active">Current Active Projects</button>
</li>
<li>
  <button class="btn btn-sm" id="htracker-projlist-menus-history">History Projects</button>
</li>
</script>

<script type="text/html" id="htracker-projlist-optools">
<li>
  <button class="btn btn-outline-primary btn-sm" onclick="htrackerProj.NewEntry()">
    <span class="icon16 icono-plus"></span>
    New Project
  </button>
</li>
</script>

<script type="text/html" id="htracker-projlist-tpl">
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
<tr id="proj-{[=v.id]}"
  class="htracker-div-hover"
  onclick="htrackerProj.ProcIndex('{[=v.id]}')">
  <td>{[=v.name]}</td>
  <td>{[=v._filter_title]}</td>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i")]}</td>
  {[? it._history]}<td>{[=l4i.UnixTimeFormat(v.closed, "Y-m-d H:i")]}</td>{[?]}
  <td>
    {[if (v.proc_num > 0 || it._history) {]}
      {[=v.proc_num]}
    {[} else {]}
      waiting
    {[}]}
  </td>
  <td align="right">
    <i class="icono-caretRight" style="zoom: 85%; margin:0;"></i>
  </td>
</tr>
{[~]}
</tbody>
</table>
</script>

