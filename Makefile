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
    CLAP_FORMAT := bundle
else ifeq ($(UNAME), Linux)
    PLATFORM := linux
    SO_EXT := so
    CLAP_FORMAT := so
else
    PLATFORM := windows
    SO_EXT := dll
    CLAP_FORMAT := dll
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
ifeq ($(PLATFORM), linux)
    LDFLAGS += -Wl,-rpath,'$$$$ORIGIN'
endif
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
CLAP_INCLUDE_DIR := ./include/clap/include

# Define plugin directories
EXAMPLE_PLUGINS := $(EXAMPLES_DIR)/gain $(EXAMPLES_DIR)/synth

# Main targets
.PHONY: all clean install uninstall build-go build-plugins examples test print-plugin-id

# Helper target to print plugin ID (deprecated, kept for backward compatibility)
print-plugin-id:
	@echo "Plugin IDs are now handled through exported Go functions in each plugin"

all: build-go build-plugins

# Build Go bridge library
build-go:
	@echo "Building Go bridge library..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) \
		-o $(BUILD_DIR)/libgoclap.$(SO_EXT) \
		./pkg/bridge

# Plugin build rules
# Common function to build a plugin
define build_plugin
# Create build directory
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR):
	@mkdir -p $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)

# Go library for the plugin
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/lib$(1).$(SO_EXT): $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Building Go plugin library for $(1)..."
	@cd $(EXAMPLES_DIR)/$(1) && \
	CGO_ENABLED=$(CGO_ENABLED) \
	$(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/lib$(1).$(SO_EXT) *.go

# C bridge objects
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o: $(C_SRC_DIR)/bridge.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C bridge for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/bridge.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o

$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o: $(C_SRC_DIR)/plugin.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C plugin for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/plugin.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o

# Final CLAP plugin
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/$(1).clap: $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/lib$(1).$(SO_EXT) $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o
	@echo "Linking $(1).clap..."
	$(LD) $(LDFLAGS) -o $$@ $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o -L$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR) -l$(1)

# Build target for each plugin
build-$(1): build-go $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/$(1).clap
	@echo "Built $(1) plugin (ID: $(PLUGIN_ID_$(1)))"

endef

# Apply the build_plugin function for each plugin
$(foreach plugin,gain synth,$(eval $(call build_plugin,$(plugin))))

# Build all plugins
build-plugins: build-gain build-synth
	@echo "All plugins built."

# Install plugins to plugin directory
install: all
	@echo "Installing plugins to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@if [ -f $(BUILD_DIR)/libgoclap.$(SO_EXT) ]; then \
		cp -f $(BUILD_DIR)/libgoclap.$(SO_EXT) $(INSTALL_DIR)/; \
		echo "  Installed libgoclap.$(SO_EXT)"; \
	else \
		echo "  Error: libgoclap.$(SO_EXT) not found"; \
	fi
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			if [ -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" ]; then \
				echo "  Installing $$plugin_name.clap..."; \
				cp -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" $(INSTALL_DIR)/; \
				# Copy the plugin's library to the same directory as the plugin for runtime loading\
				if [ -f "$$plugin/$(BUILD_DIR)/lib$$plugin_name.$(SO_EXT)" ]; then \
					echo "  Installing lib$$plugin_name.$(SO_EXT)..."; \
					cp -f "$$plugin/$(BUILD_DIR)/lib$$plugin_name.$(SO_EXT)" $(INSTALL_DIR)/; \
				fi; \
			else \
				echo "  Warning: $$plugin_name.clap not found, skipping"; \
			fi; \
		fi; \
	done
	@if ls $(INSTALL_DIR)/*.clap >/dev/null 2>&1 && ls $(INSTALL_DIR)/*.$(SO_EXT) >/dev/null 2>&1; then \
		chmod 755 $(INSTALL_DIR)/*.clap $(INSTALL_DIR)/*.$(SO_EXT); \
	fi
	@echo "Note: Complex plugins like gain-with-gui require additional build steps with CMake"
	@echo "      To build them, use CMake directly or extend this Makefile in the future."
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
			rm -rf "$$plugin/$(BUILD_DIR)"; \
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
				./scripts/test_plugin.sh "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" || echo "  Test failed for $$plugin_name"; \
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
	@echo "  make build-gain   - Build only the gain plugin"
	@echo "  make build-synth  - Build only the synth plugin"
	@echo "  make install      - Install plugins to $(INSTALL_DIR)"
	@echo "  make uninstall    - Remove installed plugins"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Test plugins"
	@echo "  make help         - Display this help"
	@echo ""
	@echo "Options:"
	@echo "  DEBUG=1           - Build with debug symbols and no optimization"
	@echo "  INSTALL_DIR=path  - Install to a custom directory"