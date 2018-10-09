
<div class="htracker-div-container alert less-hide" id="htracker-tracer-ptrace-list-alert"></div>

<div class="htracker-div-light">
    <table class="table table-hover valign-middle">
      <thead>
      <tr>
        <th>Start Time</th>
        <th>End Time</th>
        <th>Flame Graph</th>
      </tr>
      </thead>
      <tbody id="htracker-tracer-ptrace-list"></tbody>
    </table>
</div>

<script type="text/html" id="htracker-tracer-ptrace-list-menus">
<li>
  <button type="button" class="btn btn-primary btn-sm" onclick="htrackerTracer.ProcList()">Back to Process List</button>
</li>
</script>


<script type="text/html" id="htracker-tracer-ptrace-list-tpl">
{[~it.items :v]}
<tr>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i:s")]}</td>
  <td>{[=l4i.UnixTimeFormat(v.updated, "Y-m-d H:i:s")]}</td>
  <td>
    <button class="btn btn-primary btn-sm" onclick="htrackerTracer.ProcDyTraceView({[=v.pid]}, {[=v.pcreated]}, {[=v.created]})">On-CPU</button>
  </td>
</tr>
{[~]}
</script>

