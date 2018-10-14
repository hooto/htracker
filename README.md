# About

hooto tracker is a web frontend visualization and analysis tool for APM (application performance management).

## Features

* Compatible with Linux systems
* Probe point based on process, no need to modify the analyzed application's code
* Continuous tracking and state saving
* Historical rollup of all data (default setup in 30 days)
* Resource Usage capture and charts (CPU, Memory, Network, IO, ...)
* Dynamic Trace capture and graphs (On-CPU, Off-CPU, Memory, ...)

# Install

todo ...

## CentOS 7.x

``` shell
sudo yum install -y perf
```
## Ubuntu 18.04

``` shell
sudo apt install linux-tools-commom linux-tools-generic
```

## Source

``` shell
# pull source code
git clone https://github.com/hooto/htracker.git
cd htracker
git submodule update

# build
make
sudo make install

# setup service, and start
systemctl enable hooto-tracker
systemctl start hooto-tracker
```

## Setup

# Getting Started

todo ...

# Dependent Projects or Documents

* Brendan Gregg's site for computer performance analysis and methodology [http://www.brendangregg.com/](http://www.brendangregg.com/)
* Stack trace visualizer [https://github.com/brendangregg/FlameGraph](https://github.com/brendangregg/FlameGraph)
* Cross-platform lib for process and system monitoring [https://github.com/shirou/gopsutil](https://github.com/shirou/gopsutil)
* D3.js JavaScript visualization lib [https://d3js.org/](https://d3js.org/)
* D3.js plugin that produces flame graphs [https://github.com/spiermar/d3-flame-graph](https://github.com/spiermar/d3-flame-graph)
* jQuery Web library [http://jquery.com/](http://jquery.com/)
* Bootstrap Web	UI library [http://getbootstrap.com/](http://getbootstrap.com/)

