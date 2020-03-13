#---* Makefile *---#
.SILENT :

export GO111MODULE=on

# Base package
BASE_PACKAGE=github.com/qorpress

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

# Plugin name
PLUGIN?=oniontree

# Plugin filename
PLUGIN_SO=$(APPNAME)-$(PLUGIN).so

# Extract version infos
VERSION:=`git describe --tags --always`
GIT_COMMIT:=`git rev-list -1 HEAD --abbrev-commit`
BUILT:=`date`

## docker-build			:	build qorpress inside a docker container.
docker-build:
	@docker build -t qorpress/qorpress .
.PHONY: docker-build

## docker-run			:	run qorpress from docker container.
docker-run:
	@docker run -ti -p 443:443 -p 80:80 -p 7000:7000 qorpress/qorpress
.PHONY: docker-run

## manticore-darwin		:	install manticore on mac osx.
manticore-darwin:
	@brew install manticoresearch
.PHONY: manticore-darwin

## manticore-start		:	start local manticore searchd with qorpress config.
manticore-start:
	@searchd --config ./.docker/manticore/manticore.conf
.PHONY: manticore-start

## manticore-stop			:	stop local manticore searchd.
manticore-stop:
	@searchd --stop
.PHONY: manticore-stop

## manticore-index			:	stop local manticore searchd.
manticore-index:
	@indexer --config ./.docker/manticore/manticore.conf
.PHONY: manticore-index

## plugin				:	Build plugin (defined by PLUGIN variable).
plugin:
	-mkdir -p release
	echo ">>> Building: $(PLUGIN_SO) $(VERSION) for $(GOOS)-$(GOARCH) ..."
	cd plugins/$(PLUGIN) && GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o ../../release/$(PLUGIN_SO)
.PHONY: plugin

## plugins			:	Build all qorpress plugins
plugins:
	GOARCH=amd64 PLUGIN=flickr make plugin
	GOARCH=amd64 PLUGIN=twitter make plugin
	GOARCH=amd64 PLUGIN=oniontree make plugin
.PHONY: plugins  

## help				:	Print commands help.
help : Makefile
	@sed -n 's/^##//p' $<
.PHONY: help

# https://stackoverflow.com/a/6273809/1826109
%:
	@:
