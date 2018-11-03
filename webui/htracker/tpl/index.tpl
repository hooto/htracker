<nav class="navbar navbar-expand navbar-dark bg-dark">
  <a class="navbar-brand" href="{[=it.http_basepath]}/htracker/">
    <img src="{[=it.http_basepath]}/htracker/~/htracker/img/ht-topnav-light.png" width="30" height="30">
  </a>
  <a class="navbar-brand" href="{[=it.http_basepath]}/htracker/">Perf Tracker</a>
  <div id="htracker-nav" class="navbar-nav mr-auto" style="margin-left: 30px">
    <a class="nav-item nav-link l4i-nav-item" href="#proj/index">{[=l4i.T("Projects")]}</a>
    <a class="nav-item nav-link l4i-nav-item" href="#proc/index">{[=l4i.T("Processes")]}</a>
  </div>
  <div class="navbar-text">
    <a class="" href="#user/sign-out" onclick="htracker.UserSignOut()">{[=l4i.T("Sign Out")]}</a>
  </div>
</nav>

<div id="htracker-module-layout">
  <div id='htracker-module-navbar'>
    <ul id='htracker-module-navbar-menus' class='htracker-module-nav'></ul>
    <ul id='htracker-module-navbar-optools' class='htracker-module-nav htracker-nav-right'></ul>
  </div>
  <div id="htracker-module-content"></div>
</div>

<div class="htracker-body-footer">
  <span>© 2018 <a href="https://github.com/hooto/htracker" target="_blank">hooto tracker {[=it.version]}</a>
  &nbsp; is a web visualization and analysis tool for APM (application performance management)</span>
</div>

