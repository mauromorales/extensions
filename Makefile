BUNDLED_EXTENSIONS = autobump-github
UBINDIR ?= /usr/bin
DESTDIR = 

all: build

build:
	for d in $(BUNDLED_EXTENSIONS); do $(MAKE) -C extensions/$$d build; done

install: build
	for d in $(BUNDLED_EXTENSIONS); do $(MAKE) -C extensions/$$d install; done
