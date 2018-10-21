<nav class="navbar navbar-expand-lg navbar-dark bg-dark">
  <a class="navbar-brand" href="/htracker/">
    <img src="/htracker/~/htracker/img/ht-topnav-light.png" width="30" height="30">
  </a>
  <a class="navbar-brand" href="/htracker/">Perf Tracker</a>
  <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
    <span class="navbar-toggler-icon"></span>
  </button>
  <div class="collapse navbar-collapse" id="navbarNav">
    <ul id="htracker-nav" class="navbar-nav mr-auto" style="margin-left: 30px">
      <li class="nav-item l4i-nav-item">
        <a class="nav-link" href="#proj/index">Projects</a>
      </li>
      <li class="nav-item l4i-nav-item">
        <a class="nav-link" href="#proc/index">Processes</a>
      </li>
    </ul>
    <span class="navbar-text">
        <a class="" href="#user/sign-out" onclick="htracker.UserSignOut()">Sign Out</a>
    </span>
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

