UBINDIR ?= /usr/bin
DESTDIR ?=

all: build install

build:
	@echo "Nothing to build"

install: build
	install -d $(DESTDIR)/$(UBINDIR)
	install -m 0755 luet-autobump-github $(DESTDIR)/$(UBINDIR)/

test:
	TREE_DIR=test_packages FAIL_ON_ERROR=true ./luet-autobump-github
	git checkout test_packages