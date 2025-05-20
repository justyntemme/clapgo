# ClapGo - Makefile for CLAP plugins in Go
# This Makefile centralizes all build and installation steps

# Configuration
PLUGIN_DIR := $(HOME)/.clap
CLAP_PLUGIN_DIR := clap-plugins
BUILD_DIR := build
INSTALL_DIR := $(PLUGIN_DIR)

# Platform detection
UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
    PLATFORM := macos
    SO_EXT := dylib
else ifeq ($(UNAME), Linux)
    PLATFORM := linux
    SO_EXT := so
else
    PLATFORM := windows
    SO_EXT := dll
endif

# Go configuration
GO := go
CGO_ENABLED := 1
GO_FLAGS := -buildmode=c-shared
GO_BUILD_FLAGS := -ldflags="-s -w"
ifeq ($(DEBUG), 1)
    GO_BUILD_FLAGS := -gcflags="all=-N -l"
endif

# C compilation
CC := gcc
LD := gcc
CFLAGS := -I./include/clap/include -fPIC -Wall -Wextra
LDFLAGS := -shared
ifeq ($(DEBUG), 1)
    CFLAGS += -g -O0 -DDEBUG
else
    CFLAGS += -O2 -DNDEBUG
endif

# Directories
C_SRC_DIR := src/c
GO_SRC_DIR := src/goclap
INTERNAL_DIR := internal
PKG_DIR := pkg
EXAMPLES_DIR := examples

# Find all example plugins
EXAMPLE_PLUGINS := $(wildcard $(EXAMPLES_DIR)/*)

# Main targets
.PHONY: all clean install uninstall build-go build-plugins examples test

all: build-go build-plugins

# Build Go bridge library
build-go:
	@echo "Building Go bridge library..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) \
		-o $(BUILD_DIR)/libgoclap.$(SO_EXT) \
		./cmd/goclap

# Build all plugins
build-plugins: build-go
	@echo "Building CLAP plugins..."
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			echo "  Building $$plugin_name..."; \
			$(MAKE) -C $$plugin || exit 1; \
		fi; \
	done

# Install plugins to plugin directory
install: all
	@echo "Installing plugins to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp -f $(BUILD_DIR)/libgoclap.$(SO_EXT) $(INSTALL_DIR)/
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			if [ -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" ]; then \
				echo "  Installing $$plugin_name.clap..."; \
				cp -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" $(INSTALL_DIR)/; \
			fi; \
		fi; \
	done
	@chmod 755 $(INSTALL_DIR)/*.clap $(INSTALL_DIR)/*.$(SO_EXT)
	@echo "Installation complete!"

# Remove installed plugins
uninstall:
	@echo "Uninstalling plugins from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/libgoclap.$(SO_EXT)
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			echo "  Removing $$plugin_name.clap..."; \
			rm -f "$(INSTALL_DIR)/$$plugin_name.clap"; \
		fi; \
	done
	@echo "Uninstallation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			echo "  Cleaning $$plugin..."; \
			$(MAKE) -C $$plugin clean || echo "  No makefile in $$plugin"; \
		fi; \
	done
	@echo "Clean complete!"

# Test plugins
test: all
	@echo "Testing plugins..."
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			if [ -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" ]; then \
				echo "  Testing $$plugin_name.clap..."; \
				./test_plugin.sh "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" || echo "  Test failed for $$plugin_name"; \
			fi; \
		fi; \
	done
	@echo "Testing complete!"

# Help
help:
	@echo "ClapGo Makefile Usage:"
	@echo "  make              - Build all plugins"
	@echo "  make build-go     - Build only the Go bridge library"
	@echo "  make build-plugins - Build all CLAP plugins"
	@echo "  make install      - Install plugins to $(INSTALL_DIR)"
	@echo "  make uninstall    - Remove installed plugins"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Test plugins"
	@echo "  make help         - Display this help"
	@echo ""
	@echo "Options:"
	@echo "  DEBUG=1           - Build with debug symbols and no optimization"
	@echo "  INSTALL_DIR=path  - Install to a custom directory"