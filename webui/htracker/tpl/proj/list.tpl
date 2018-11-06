<div>
  <div class="htracker-div-container alert less-hide" id="htracker-projlist-alert"></div>
  <div class="htracker-div-light">
    <div id="htracker-projlist-table"></div>
    <div id="htracker-projlist-more" style="display: none; padding: 0 0 10px 10px">
      <button class="btn btn-primary btn-sm"
        onclick="htrackerProj.ListMore()">
        More ...
      </button>
    </div>
  </div>
</div>

<script type="text/html" id="htracker-projlist-menus">
<li>
  <button class="btn btn-sm btn-dark" id="htracker-projlist-menus-active">{[=l4i.T("Current Active Projects")]}</button>
</li>
<li>
  <button class="btn btn-sm" id="htracker-projlist-menus-history">{[=l4i.T("History Projects")]}</button>
</li>
</script>

<script type="text/html" id="htracker-projlist-optools">
<li>
  <div id="htracker-proj-list-status-msg" class="item-status-msg badge badge-light" style="display:none"></div>
</li>
<li>
  <button class="btn btn-outline-primary btn-sm" onclick="htrackerProj.NewEntry()">
    <span class="icon16 icono-plus"></span>
    {[=l4i.T("New Project")]}
  </button>
</li>
</script>

<script type="text/html" id="htracker-projlist-table-tpl">
<table class="table table-hover valign-middle">
<thead>
  <tr>
    <th>{[=l4i.T("Project Name")]}</th>
    <th>{[=l4i.T("Filter")]}</th>
    <th>{[=l4i.T("Created")]}</th>
    {[? it._history]}<th>{[=l4i.T("Closed")]}</th>{[?]}
    <th>{[=l4i.T("Hit Processes")]}</th>
    <th width="30"></th>
  </tr>
</thead>
<tbody id="htracker-projlist"></tbody>
</table>
</script>

<script type="text/html" id="htracker-projlist-tpl">
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
      {[=l4i.T("waiting")]}
    {[}]}
  </td>
  <td align="right">
    <i class="icono-caretRight" style="zoom: 85%; margin:0;"></i>
  </td>
</tr>
{[~]}
</script>
