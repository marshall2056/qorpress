.SILENT :

export GO111MODULE=on

# Base package
BASE_PACKAGE=github.com/x0rzkov

# App name
APPNAME=qorpress

# Go configuration
GOOS?=$(shell go env GOHOSTOS)
GOARCH?=$(shell go env GOHOSTARCH)

# Add exe extension if windows target
is_windows:=$(filter windows,$(GOOS))
EXT:=$(if $(is_windows),".exe","")
LDLAGS_LAUNCHER:=$(if $(is_windows),-ldflags "-H=windowsgui",)

# Archive name
ARCHIVE=$(APPNAME)-$(GOOS)-$(GOARCH).tgz

# Main executable name
MAIN_EXE=$(APPNAME)$(EXT)

# CLI executable name
CLI_EXE=$(APPNAME)-ctl$(EXT)

# Launcher filename
LAUNCHER_EXE=$(APPNAME)-launcher$(EXT)

# Plugin name
PLUGIN?=twitter

# Plugin filename
PLUGIN_SO=$(APPNAME)-$(PLUGIN).so

# Extract version infos
PKG_VERSION:=github.com/x0rzkov/$(APPNAME)/v1/pkg/version
VERSION:=`git describe --tags --always`
GIT_COMMIT:=`git rev-list -1 HEAD --abbrev-commit`
BUILT:=`date`
define LDFLAGS
-X '$(PKG_VERSION).Version=$(VERSION)' \
-X '$(PKG_VERSION).GitCommit=$(GIT_COMMIT)' \
-X '$(PKG_VERSION).Built=$(BUILT)'
endef

dep:
	@go mod vendor
.PHONY: dep

run:
	@go run main.go
.PHONY: run

run_bindatafs:
	@go run config/compile/compile.go
	@go run -tags bindatafs *.go
.PHONY: run_bindatafs

build: dep
	@go build main.go
.PHONY: build

build_bindatafs: dep
	@go run config/compile/compile.go
	@go build -tags bindatafs *.go
.PHONY: build_bindatafs

## Bulid plugin (defined by PLUGIN variable)
plugin:
	-mkdir -p release
	echo ">>> Building: $(PLUGIN_SO) $(VERSION) for $(GOOS)-$(GOARCH) ..."
	cd plugins/$(PLUGIN) && GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o ../../release/$(PLUGIN_SO)
.PHONY: plugin

## Build all plugins
plugins:
	# GOARCH=amd64 PLUGIN=qp-flickr make plugin
	# GOARCH=amd64 PLUGIN=qp-oniontree make plugin
	# GOARCH=amd64 PLUGIN=qp-twitter make plugin
.PHONY: plugins

test:
	@go test -race $(shell go list ./... | grep -v vendor)

fmt:
	@go fmt $(shell go list ./... | grep -v vendor | grep -v testdata)

## help			:	Print commands help.
.PHONY: help
help : Makefile
	@sed -n 's/^##//p' $<

# https://stackoverflow.com/a/6273809/1826109
%:
	@: