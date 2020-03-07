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

# To do
# git rev
docker-build:
	@docker build -t qorpress/qorpress .

docker-run:
	@docker run -ti -p 443:443 -p 80:80 -p 7000:7000 qorpress/qor-example

## Bulid plugin (defined by PLUGIN variable)
plugin:
	-mkdir -p release
	echo ">>> Building: $(PLUGIN_SO) $(VERSION) for $(GOOS)-$(GOARCH) ..."
	cd plugins/$(PLUGIN) && GOOS=$(GOOS) GOARCH=$(GOARCH) go build -buildmode=plugin -o ../../release/$(PLUGIN_SO)
.PHONY: plugin

## Build all plugins
plugins:
	GOARCH=amd64 PLUGIN=oniontree make plugin
	# GOARCH=amd64 PLUGIN=flickr make plugin
	# GOARCH=amd64 PLUGIN=twitter make plugin
.PHONY: plugins  