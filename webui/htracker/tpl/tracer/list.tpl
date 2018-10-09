<div>
  <div class="htracker-div-container alert less-hide" id="htracker-tracer-list-alert"></div>
  <div class="htracker-div-light">
    <table class="table table-hover valign-middle">
      <thead>
      <tr>
        <th>Name</th>
        <th>Filter</th>
        <th>Created</th>
        <th>Hit Processes</th>
      </tr>
      </thead>
      <tbody id="htracker-tracer-list"></tbody>
    </table>
  </div>
</div>

<script type="text/html" id="htracker-tracer-list-menus">
<li>
  <button class="btn btn-dark btn-sm" disabled>Actives</button>
</li>
</script>

<script type="text/html" id="htracker-tracer-list-optools">
<div class="input-groupi">
  <form class="form-inline input-group" onsubmit="htrackerTracer.ListRefreshQuery(); return false;">
     <input class="form-control" type="search" placeholder="Search" id="htracker-tracer-list-query">
    <div class="input-group-append">
      <button class="btn btn-outline-secondary" type="button">Search</button>
    </div>
  </form>
</div>
</script>

<script type="text/html" id="htracker-tracer-list-tpl">
{[~it.items :v]}
<tr id="tracer-{[=v.id]}"
  class="htracker-div-hover"
  onclick="htrackerTracer.ProcList('{[=v.id]}')">
  <td>{[=v.name]}</td>
  <td>{[=v._filter_title]}</td>
  <td>{[=l4i.UnixTimeFormat(v.created, "Y-m-d H:i")]}</td>
  <td>
    <button class="btn btn-outline-primary btn-sm" onclick="htrackerTracer.ProcList('{[=v.id]}')" style="width:50px">{[=v.proc_num]}</button>
  </td>
</tr>
{[~]}
</script>

