
<div class="htracker-div-container alert less-hide" id="htracker-proj-ptrace-list-alert"></div>

<div class="htracker-div-light">
    <table class="table table-hover valign-middle">
      <thead>
      <tr>
        <th>Start Time</th>
        <th>End Time</th>
        <th>Log Size</th>
        <th style="text-align:right">Flame Graph</th>
      </tr>
      </thead>
      <tbody id="htracker-proj-ptrace-list"></tbody>
    </table>

  <div id="htracker-proj-ptrace-list-more" style="display: none; padding: 0 0 10px 10px">
     <button class="btn btn-primary btn-sm"
	   onclick="htrackerProj.ProcDyTraceListMore()">
       More items ...
	 </button>
  </div>
</div>

<script type="text/html" id="htracker-proj-ptrace-list-menus">
<li>
  <button type="button" class="btn btn-primary btn-sm" onclick="htrackerProj.ProcIndex()">
    <span class="icon16 icono-caretLeftCircle"></span>
    Back to Hit Process List</button>
</li>
</script>


<script type="text/html" id="htracker-proj-ptrace-list-tpl">
{[~it.items :v]}
<tr>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i:s")]}</td>
  <td>{[=l4i.UnixTimeFormat(v.updated, "Y-m-d H:i:s")]}</td>
  <td>{[=htracker.UtilResSizeFormat(v.perf_size)]}</td>
  <td align="right">
    <button class="btn btn-primary btn-sm"
	  onclick="htrackerProj.ProcDyTraceView({[=v.pid]}, {[=v.pcreated]}, {[=v.created]})">On-CPU</button>
  </td>
</tr>
{[~]}
</script>

