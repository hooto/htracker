<div>
  <div class="htracker-div-light">
    <table class="table table-hover valign-middle">
      <thead>
      <tr>
        <th>Name</th>
        <th>User</th>
        <th>CPU %</th>
        <th>Memory</th>
        <th>Command</th>
        <th width="30"></th>
      </tr>
      </thead>
      <tbody id="htracker-process-list"></tbody>
    </table>
  </div>
</div>

<script type="text/html" id="htracker-process-list-menus">
<li>
  <div id="htracker-process-list-status-msg" class="badge badge-success" style="padding: 5px 10px;font-size:14px">loading</div>
</li>
</script>

<script type="text/html" id="htracker-process-list-optools">
<form class="input-group mb-3" onsubmit="htrackerProcess.ListRefreshQuery(); return false;">
  <input class="form-control" type="text" id="htracker-process-list-query">
  <div class="input-group-append">
    <button class="btn btn-outline-secondary" type="button">Search</button>
  </div>
</form>
</script>


<script type="text/html" id="htracker-process-list-tpl">
{[~it.items :v]}
<tr class="htracker-div-hover" onclick="htrackerProcess.EntryView('{[=v.pid]}')">
  <td>{[=v.name]}</td>
  <td>{[=v.user]}</td>
  <td>{[=v.cpu_p]}</td>
  <td>{[=htracker.UtilResSizeFormat(v.mem_rss)]}</td>
  <td>
    {[if (v.cmd.length > 60) {]}
      {[=v.cmd.substr(0, 50)]}...
    {[} else {]}
      {[=v.cmd]}
    {[}]}
  </td>
  <td align="right">
    <i class="icono-caretRight" style="zoom: 85%; margin-right:0;"></i>
  </td>
</tr>
{[~]}
</script>

<script type="text/html" id="htracker-process-entry-tpl">
<table class="htracker-table">
<tr>
  <td class="htracker-table-item-name" width="200px">Name</td>
  <td>{[=it.name]}</td>
</tr>
<tr>
  <td class="htracker-table-item-name">User</td>
  <td>{[=it.user]}</td>
</tr>
<tr>
  <td class="htracker-table-item-name">CPU Percent</td>
  <td>{[=it.cpu_p]}</td>
</tr>
<tr>
  <td class="htracker-table-item-name">Memory RSS</td>
  <td>{[=htracker.UtilResSizeFormat(it.mem_rss)]}</td>
</tr>
<tr>
  <td class="htracker-table-item-name" valign="top">Command</td>
  <td><p>{[=it.cmd]}</p></td>
</tr>
</table>
</script>
