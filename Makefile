PREFIX=/opt/hooto/tracker
CC=go
CARGS=build
CFLAGS=""

BUILDCOLOR="\033[34;1m"
BINCOLOR="\033[37;1m"
ENDCOLOR="\033[0m"

ifndef V
	QUIET_BUILD = @printf '%b %b\n' $(BUILDCOLOR)BUILD$(ENDCOLOR) $(BINCOLOR)$@$(ENDCOLOR) 1>&2;
	QUIET_INSTALL = @printf '%b %b\n' $(BUILDCOLOR)INSTALL$(ENDCOLOR) $(BINCOLOR)$@$(ENDCOLOR) 1>&2;
endif


all: hooto-tracker
	@echo ""
	@echo "build complete"
	@echo ""

install: install_init install_bin install_static install_systemd
	@echo ""
	@echo "install complete"
	@echo ""

hooto-tracker:
	$(QUIET_BUILD)$(CC) $(CARGS) -o ./bin/hooto-tracker ./main.go$(CCLINK)

burn:
	$(QUIET_BUILD)$(CC) $(CARGS) -o ./bin/burn vendor/github.com/spiermar/burn/main.go$(CCLINK)

install_init:
	# mkdir -p $(PREFIX)/{etc,bin,var/tracker_db,var/log,var/tmp,webui,misc,deps/FlameGraph}
	mkdir -p $(PREFIX)/etc
	mkdir -p $(PREFIX)/bin
	mkdir -p $(PREFIX)/var/tracker_db
	mkdir -p $(PREFIX)/var/log
	mkdir -p $(PREFIX)/var/tmp
	mkdir -p $(PREFIX)/webui
	mkdir -p $(PREFIX)/misc
	mkdir -p $(PREFIX)/deps/FlameGraph

install_bin:
	$(QUIET_INSTALL)
	install bin/hooto-tracker $(PREFIX)/bin/hooto-tracker$(CCLINK)
	# install bin/burn $(PREFIX)/bin/burn

install_static:
	$(QUIET_INSTALL)
	rsync -av --include="*/" --include="*.js" --include="*.css" --include="*.tpl" --exclude="*" ./webui/ $(PREFIX)/webui/
	sed -i 's/debug:\ true/debug:\ false/g' $(PREFIX)/webui/htracker/js/main.js
	rsync -av misc/* $(PREFIX)/misc/
	rsync -av deps/FlameGraph/*.pl $(PREFIX)/deps/FlameGraph/

install_systemd:
	$(QUIET_INSTALL)
	install misc/systemd/systemd.service /usr/lib/systemd/system/hooto-tracker.service
	systemctl daemon-reload

clean:
	rm -f ./bin/hooto-tracker
	# rm -f ./bin/burn

