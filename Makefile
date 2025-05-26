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

# Add json-c library
CFLAGS += $(shell pkg-config --cflags json-c)
LDFLAGS += $(shell pkg-config --libs json-c)

# Bridge source files
C_BRIDGE_SRCS := src/c/bridge.c src/c/plugin.c src/c/manifest.c src/c/preset_discovery.c

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
.PHONY: all clean clean-all install uninstall build-go build-plugins examples test print-plugin-id

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

# Go shared library for the plugin
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/lib$(1).so: $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Building Go shared library for $(1)..."
	@cd $(EXAMPLES_DIR)/$(1) && \
	CGO_ENABLED=$(CGO_ENABLED) \
	$(GO) build $(GO_FLAGS) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/lib$(1).so *.go
	@if [ -f "$(EXAMPLES_DIR)/$(1)/$(1).json" ]; then \
		echo "Copying manifest file for $(1)..."; \
		cp "$(EXAMPLES_DIR)/$(1)/$(1).json" "$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/"; \
	fi

# C bridge objects
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o: $(C_SRC_DIR)/bridge.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C bridge for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/bridge.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o

$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o: $(C_SRC_DIR)/plugin.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C plugin for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/plugin.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o

$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/manifest.o: $(C_SRC_DIR)/manifest.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C manifest for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/manifest.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/manifest.o

$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/preset_discovery.o: $(C_SRC_DIR)/preset_discovery.c | $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)
	@echo "Compiling C preset discovery for $(1)..."
	$(CC) $(CFLAGS) -I$(C_SRC_DIR) -c $(C_SRC_DIR)/preset_discovery.c -o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/preset_discovery.o

# Final CLAP plugin - linked with shared Go library
$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/$(1).clap: $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/lib$(1).so $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/manifest.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/preset_discovery.o
	@echo "Linking $(1).clap with shared library..."
	$(LD) $(LDFLAGS) -L$(EXAMPLES_DIR)/$(1)/$(BUILD_DIR) -o $$@ $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/bridge.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/plugin.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/manifest.o $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/preset_discovery.o -l$(1) $(shell pkg-config --libs json-c)

# Build target for each plugin
build-$(1): build-go $(EXAMPLES_DIR)/$(1)/$(BUILD_DIR)/$(1).clap
	@echo "Built $(1) plugin (ID: $(PLUGIN_ID_$(1)))"

endef

# Apply the build_plugin function for each plugin
$(foreach plugin,gain synth,$(eval $(call build_plugin,$(plugin))))

# Build all plugins
build-plugins: build-gain build-synth
	@echo "All plugins built."

# Install plugins to plugin directory with simplified structure
install: all
	@echo "Installing plugins to $(INSTALL_DIR)..."
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			plugin_dir="$(INSTALL_DIR)/$$plugin_name"; \
			echo "  Installing $$plugin_name to $$plugin_dir..."; \
			mkdir -p "$$plugin_dir"; \
			if [ -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" ]; then \
				cp -f "$$plugin/$(BUILD_DIR)/$$plugin_name.clap" "$$plugin_dir/"; \
				echo "    Copied $$plugin_name.clap"; \
				chmod 755 "$$plugin_dir"/$$plugin_name.clap; \
			fi; \
			if [ -f "$$plugin/$(BUILD_DIR)/lib$$plugin_name.so" ]; then \
				cp -f "$$plugin/$(BUILD_DIR)/lib$$plugin_name.so" "$$plugin_dir/"; \
				echo "    Copied lib$$plugin_name.so"; \
				chmod 755 "$$plugin_dir"/lib$$plugin_name.so; \
			fi; \
			if [ -f "$$plugin/$$plugin_name.json" ]; then \
				cp -f "$$plugin/$$plugin_name.json" "$$plugin_dir/$$plugin_name.json"; \
				echo "    Copied $$plugin_name.json manifest"; \
			fi; \
			if [ -d "$$plugin/presets/factory" ]; then \
				mkdir -p "$$plugin_dir/presets/factory"; \
				cp "$$plugin/presets/factory"/*.json "$$plugin_dir/presets/factory/" 2>/dev/null || true; \
				preset_count=$$(find "$$plugin_dir/presets/factory" -name "*.json" | wc -l); \
				if [ $$preset_count -gt 0 ]; then \
					echo "    Copied $$preset_count presets"; \
				fi; \
			fi; \
		else \
			echo "  Warning: $$plugin directory not found, skipping"; \
		fi; \
	done
	@echo "Installation complete!"
	@echo ""
	@echo "Directory structure:"
	@echo "  $(INSTALL_DIR)/gain/"
	@echo "    ├── gain.clap"
	@echo "    ├── libgain.so"
	@echo "    ├── gain.json"
	@echo "    └── presets/"
	@echo "        └── factory/"
	@echo "            ├── boost.json"
	@echo "            └── unity.json"
	@echo "  $(INSTALL_DIR)/synth/"
	@echo "    ├── synth.clap"
	@echo "    ├── libsynth.so"
	@echo "    ├── synth.json"
	@echo "    └── presets/"
	@echo "        └── factory/"
	@echo "            ├── lead.json"
	@echo "            └── pad.json"

# Remove installed plugins
uninstall:
	@echo "Uninstalling plugins from $(INSTALL_DIR)..."
	@for plugin in $(EXAMPLE_PLUGINS); do \
		if [ -d "$$plugin" ]; then \
			plugin_name=$$(basename $$plugin); \
			plugin_dir="$(INSTALL_DIR)/$$plugin_name"; \
			if [ -d "$$plugin_dir" ]; then \
				echo "  Removing $$plugin_name plugin directory..."; \
				rm -rf "$$plugin_dir"; \
			fi; \
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
	@echo "  Cleaning gain-with-gui CMake artifacts..."
	@if [ -d "$(EXAMPLES_DIR)/gain-with-gui/build" ]; then \
		rm -rf "$(EXAMPLES_DIR)/gain-with-gui/build"; \
	fi
	@echo "  Cleaning plugin installation directories..."
	@if [ -d "$(HOME)/.clap/gain" ]; then \
		rm -rf "$(HOME)/.clap/gain"; \
	fi
	@if [ -d "$(HOME)/.clap/synth" ]; then \
		rm -rf "$(HOME)/.clap/synth"; \
	fi
	@rmdir "$(HOME)/.clap" 2>/dev/null || true
	@echo "  Cleaning temporary Go build files..."
	@find . -name "*.h" -path "*/build/*" -delete 2>/dev/null || true
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "Clean complete!"

# Clean everything including installed files
clean-all: clean uninstall
	@echo "Deep cleaning..."
	@echo "  Removing CLAP plugin directory..."
	@if [ -d "$(HOME)/.clap" ]; then \
		rm -rf "$(HOME)/.clap"; \
		echo "  Removed $(HOME)/.clap"; \
	fi
	@echo "Deep clean complete!"

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
	@echo "  make clean-all    - Clean build artifacts AND installed files"
	@echo "  make test         - Test plugins"
	@echo "  make help         - Display this help"
	@echo ""
	@echo "Code Generation:"
	@echo "  make new-plugin                  - Interactive plugin creation wizard"
	@echo "  make generate-plugin NAME=<name> ID=<id> [TYPE=<type>] - Generate new plugin (one-time)"
	@echo "  make build-plugin NAME=<name>    - Build a generated plugin (preserves edits)"
	@echo "  make validate-plugin NAME=<name> - Validate a plugin with clap-validator"
	@echo ""
	@echo "Note: Code generation only happens once during plugin creation."
	@echo "      Subsequent builds preserve developer modifications."
	@echo ""
	@echo "Options:"
	@echo "  DEBUG=1           - Build with debug symbols and no optimization"
	@echo "  INSTALL_DIR=path  - Install to a custom directory"
	@echo "  TYPE=<type>       - Plugin type (audio-effect, instrument, note-effect, etc.)"

# Build clapgo-generate tool
build-generator:
	@echo "Building clapgo-generate tool..."
	@$(GO) build -o bin/clapgo-generate ./cmd/generate-manifest/

# Create a new plugin from template (interactive)
new-plugin: build-generator
	@./bin/clapgo-generate -interactive

# Create a new plugin from template (command-line)
generate-plugin: build-generator
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make generate-plugin NAME=my-plugin TYPE=audio-effect ID=com.example.myplugin"; \
		exit 1; \
	fi
	@if [ -z "$(ID)" ]; then \
		echo "Error: ID is required. Usage: make generate-plugin NAME=my-plugin TYPE=audio-effect ID=com.example.myplugin"; \
		exit 1; \
	fi
	@TYPE=$${TYPE:-audio-effect}; \
	mkdir -p plugins/$(NAME) && \
	./bin/clapgo-generate -type=$$TYPE -id=$(ID) -name="$(NAME)" -output-dir=plugins/$(NAME) -generate=true && \
	cd plugins/$(NAME) && go generate
	@echo "Plugin created at plugins/$(NAME)"
	@echo "Run 'make build-plugin NAME=$(NAME)' to build"

# Build a specific plugin
build-plugin:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make build-plugin NAME=my-plugin"; \
		exit 1; \
	fi
	@echo "Building $(NAME) plugin..."
	@cd plugins/$(NAME) && \
	go build -buildmode=c-shared -ldflags="-s -w" -o lib$(NAME).so .
	@echo "Linking with C bridge..."
	@gcc -shared -o plugins/$(NAME)/$(NAME).clap \
		-I./include/clap/include \
		-fPIC \
		plugins/$(NAME)/lib$(NAME).so \
		src/c/bridge.c src/c/manifest.c src/c/plugin.c \
		-lm -ldl -ljson-c
	@echo "Plugin built: plugins/$(NAME)/$(NAME).clap"

# Validate a plugin
validate-plugin:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make validate-plugin NAME=my-plugin"; \
		exit 1; \
	fi
	@if command -v clap-validator >/dev/null 2>&1; then \
		clap-validator validate plugins/$(NAME)/$(NAME).clap; \
	else \
		echo "Warning: clap-validator not found. Please install it to validate plugins."; \
	fi