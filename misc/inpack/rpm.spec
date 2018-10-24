%define app_home    /opt/hooto/tracker

Name:    hooto-tracker
Version: __version__
Release: __release__%{?dist}
Vendor:  hooto.com
Summary: Productivity Tools for Enterprise
License: Apache 2
Group:   Applications

Source0:   %{name}-__version__.tar.gz
BuildRoot: %{_tmppath}/%{name}-%{version}-%{release}

Requires:       redhat-lsb-core
Requires(pre):  perf

%description
%prep
%setup -q -n %{name}-%{version}
%build

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{app_home}/
mkdir -p %{buildroot}/lib/systemd/system/

cp -rp * %{buildroot}%{app_home}/
install -m 600 misc/systemd/systemd.service %{buildroot}/lib/systemd/system/hooto-tracker.service

%clean
rm -rf %{buildroot}

%pre

%post
systemctl daemon-reload

%preun

%postun

%files
%defattr(-,root,root,-)
%dir %{app_home}
/lib/systemd/system/hooto-tracker.service
%{app_home}/

