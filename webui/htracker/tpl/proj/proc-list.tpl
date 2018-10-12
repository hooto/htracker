
<div class="htracker-div-container alert less-hide" id="htracker-proj-proclist-alert"></div>

<div class="htracker-div-light" id="htracker-proj-proclist"></div>

<script type="text/html" id="htracker-proj-proclist-menus">
<li>
  <button type="button" class="btn btn-primary btn-sm" onclick="htrackerProj.Index()">
    <span class="icon16 icono-caretLeftCircle"></span>
    <span>Back to Project List</span>
  </button>
</li>
</script>


<script type="text/html" id="htracker-proj-proclist-optools">
<li>
  <button class="btn btn-outline-danger btn-sm" onclick="htrackerProj.EntryDel()">
    <span class="icon16 icono-cross"></span>
    Remove this Project
  </button>
</li>
</script>

<script type="text/html" id="htracker-proj-proclist-tpl">
<table class="table table-hover valign-middle">
<thead>
  <tr>
    <th>ID</th>
    <th width="30%">Command</th>
    <th>Created</th>
    {[? it._hit]}<th>Updated</th>{[?]}
    {[? it._exit]}<th>Exited</th>{[?]}
    <th width="360px"></th>
  </tr>
</thead>
<tbody>

{[~it.items :v]}
<tr id="proj-{[=v.pid]}-{[=v.created]}">
  <td>{[=v.pid]}</td>
  <td>
    {[if (v.cmd.length > 80) {]}
      {[=v.cmd.substr(0, 70)]}...
    {[} else {]}
      {[=v.cmd]}
    {[}]}
  </td>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i")]}</td>
  {[? it._hit]}<td>{[=l4i.UnixTimeFormat(v.updated, "Y-m-d H:i")]}</td>{[?]}
  {[? it._exit]}<td>{[=l4i.UnixTimeFormat(v.exited, "Y-m-d H:i")]}</td>{[?]}
  <td align="right">
    <button class="btn btn-outline-primary btn-sm" onclick="htrackerProj.ProcDyTraceList('{[=v.proj_id]}', {[=v.pid]}, {[=v.created]})">
      <i class="icon16 icono-barChart"></i>
      Dynamic Trace
    </button>
    <button class="btn btn-outline-primary btn-sm" onclick="htrackerProj.ProcStats('{[=v.proj_id]}', {[=v.pid]}, {[=v.created]})">
      <i class="icon16 icono-areaChart"></i>
      Resource Usage
    </button>
  </td>
</tr>
{[~]}
</table>

</script>

