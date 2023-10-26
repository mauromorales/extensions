UBINDIR ?= /usr/bin
DESTDIR ?=
EXTNAME := $(shell basename $(shell pwd))

# go tool nm ./luet | grep Commit
# override LDFLAGS += -X "github.com/Luet-lab/extensions/extensions/${EXTNAME}/pkg/devkit.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
# override LDFLAGS += -X "github.com/Luet-lab/extensions/extensions/${EXTNAME}/pkg/devkit.BuildCommit=$(shell git rev-parse HEAD)"

all: build install

build:
	# CGO_ENABLED=0 go build -o luet-$(EXTNAME) -ldflags '$(LDFLAGS)' luet-${EXTNAME}.go

install: build
	install -d $(DESTDIR)/$(UBINDIR)
	install -m 0755 luet-$(EXTNAME) $(DESTDIR)/$(UBINDIR)/

test:
	BASE_DIRECTORY="test_packages" ./validate_yaml_files.sh
	git checkout test_packages
