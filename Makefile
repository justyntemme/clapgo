.PHONY: all clean build examples

# Default target
all: build examples

# Build the build tool
build:
	go build -o dist/clapgo-build ./cmd/build

# Clean build artifacts
clean:
	rm -rf dist/

# Build all example plugins
examples: build
	./dist/clapgo-build --name gain --output dist/
	
# Install plugins to the appropriate directory
install: examples
ifeq ($(OS),Windows_NT)
	mkdir -p "$(APPDATA)\CLAP"
	cp dist/*.clap "$(APPDATA)\CLAP"
else
	uname_s := $(shell uname -s)
	ifeq ($(uname_s),Darwin)
		mkdir -p ~/Library/Audio/Plug-Ins/CLAP
		cp -r dist/*.clap ~/Library/Audio/Plug-Ins/CLAP/
	else
		mkdir -p ~/.clap
		cp dist/*.clap ~/.clap/
	endif
endif