
<div class="htracker-div-container alert less-hide" id="htracker-tracer-proc-list-alert"></div>

<div class="htracker-div-light">
    <table class="table table-hover valign-middle">
      <thead>
      <tr>
        <th>ID</th>
        <th width="30%">Command</th>
        <th>Status</th>
        <th>Created</th>
        <th>Updated</th>
        <th width="360px"></th>
      </tr>
      </thead>
      <tbody id="htracker-tracer-proc-list"></tbody>
    </table>
</div>

<script type="text/html" id="htracker-tracer-proc-list-menus">
<li>
  <button type="button" class="btn btn-primary btn-sm" onclick="htrackerTracer.Index()">
    <span class="icon16 icono-caretLeftCircle"></span>
    <span>Back to Project List</span>
  </button>
</li>
</script>


<script type="text/html" id="htracker-tracer-proc-list-optools">
<li>
  <button class="btn btn-outline-danger btn-sm" onclick="htrackerTracer.EntryDel()">
    <span class="icon16 icono-cross"></span>
    Remove this Project
  </button>
</li>
</script>

<script type="text/html" id="htracker-tracer-proc-list-tpl">
{[~it.items :v]}
<tr id="tracer-{[=v.pid]}-{[=v.created]}">
  <td>{[=v.pid]}</td>
  <td>
    {[if (v.cmd.length > 80) {]}
      {[=v.cmd.substr(0, 70)]}...
    {[} else {]}
      {[=v.cmd]}
    {[}]}
  </td>
  <td></td>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i")]}</td>
  <td>{[=l4i.UnixTimeFormat(v.updated, "Y-m-d H:i")]}</td>
  <td align="right">
    <button class="btn btn-outline-primary btn-sm" onclick="htrackerTracer.ProcDyTraceList('{[=v.tid]}', {[=v.pid]}, {[=v.created]})">
      <i class="icon16 icono-barChart"></i>
	  Dynamic Trace
	</button>
    <button class="btn btn-outline-primary btn-sm" onclick="htrackerTracer.ProcStats('{[=v.tid]}', {[=v.pid]}, {[=v.created]})">
      <i class="icon16 icono-areaChart"></i>
	  Resource Usage
    </button>
  </td>
</tr>
{[~]}
</script>

