# add more targets (ci, docker...)
GO_VERSION			:= $(shell go version)
GO_GLIDE			:= $(shell which glide)

all: reset deps test

reset:
	@if [ -f $(CURDIR)/glide.yaml ]; then rm -f $(CURDIR)/glide.yaml ; fi
	@if [ -f $(CURDIR)/glide.lock ]; then rm -f $(CURDIR)/glide.lock ; fi

deps:
	@if [ ! -f $(GO_GLIDE) ]; then go get -v github.com/Masterminds/glide ; fi
	@if [ ! -f $(CURDIR)/glide.yaml ]; then glide create --non-interactive ; fi
	@if [ -f $(CURDIR)/glide.lock ]; then glide update --strip-vendor ; else glide install --strip-vendor ; fi

test:
	@if [ ! -f $(GO_GLIDE) ]; then go get -v github.com/Masterminds/glide ; fi
	@go test -v $(shell glide novendor)