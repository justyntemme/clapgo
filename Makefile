# ClapGo Makefile

# Detect OS
UNAME := $(shell uname -s)
ifeq ($(UNAME),Darwin)
    LIBEXT = .dylib
    CLAP_EXT = .clap
else ifeq ($(OS),Windows_NT)
    LIBEXT = .dll
    CLAP_EXT = .clap
else
    LIBEXT = .so
    CLAP_EXT = .clap
endif

# Directories
DIST_DIR = dist
EXAMPLES_DIR = examples
INCLUDE_DIR = include
SRC_DIR = src

# Compiler flags
GO = go
GCC = gcc
GCCFLAGS = -shared -fPIC -I$(INCLUDE_DIR)/clap/include
GO_LDFLAGS = -ldflags="-s -w"
GO_GCFLAGS = 

# Debug mode
ifdef DEBUG
    GCCFLAGS = -g -shared -fPIC -I$(INCLUDE_DIR)/clap/include
    GO_LDFLAGS = 
    GO_GCFLAGS = -gcflags="all=-N -l"
else
    GCCFLAGS += -O3
endif

# Find all plugins
PLUGINS = $(notdir $(wildcard $(EXAMPLES_DIR)/*))

# Default target
all: examples

# Create distribution directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Clean build artifacts
clean:
	rm -rf $(DIST_DIR)

# Help target
help:
	@echo "ClapGo - CLAP Plugin Framework for Go"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build all plugins"
	@echo "  make clean        Remove build artifacts"
	@echo "  make gain         Build the gain example plugin"
	@echo "  make install      Install plugins to system directory"
	@echo "  make new NAME=x   Create a new plugin named x"
	@echo ""
	@echo "Available plugins: $(PLUGINS)"

# Build a specific plugin
define PLUGIN_template
$(1): $(DIST_DIR)
	@echo "Building $(1) plugin..."
	cd $(CURDIR) && $(GO) build $(GO_GCFLAGS) $(GO_LDFLAGS) -buildmode=c-shared -o $(DIST_DIR)/lib$(1).so ./$(EXAMPLES_DIR)/$(1)
	$(GCC) $(GCCFLAGS) -o $(DIST_DIR)/$(1)$(LIBEXT) $(SRC_DIR)/c/plugin.c -L$(DIST_DIR) -l$(1)
	cp $(DIST_DIR)/$(1)$(LIBEXT) $(DIST_DIR)/$(1)$(CLAP_EXT)
	@echo "Successfully built $(1) plugin"
endef

# Create build targets for all plugins
$(foreach plugin,$(PLUGINS),$(eval $(call PLUGIN_template,$(plugin))))

# Build all example plugins
examples: $(PLUGINS)

# Install plugins
install: examples
ifeq ($(OS),Windows_NT)
	mkdir -p "$(APPDATA)\CLAP"
	cp $(DIST_DIR)/*.clap "$(APPDATA)\CLAP"
else ifeq ($(UNAME),Darwin)
	mkdir -p ~/Library/Audio/Plug-Ins/CLAP
	cp -r $(DIST_DIR)/*.clap ~/Library/Audio/Plug-Ins/CLAP/
else
	mkdir -p ~/.clap
	cp $(DIST_DIR)/*.clap ~/.clap/
endif

# Create a new plugin from template
new:
ifndef NAME
	$(error Please specify a plugin name with NAME=pluginname)
endif
	@echo "Creating new plugin: $(NAME)"
	mkdir -p $(EXAMPLES_DIR)/$(NAME)
	cp $(EXAMPLES_DIR)/gain/main.go $(EXAMPLES_DIR)/$(NAME)/main.go
	sed -i 's/GainPlugin/$(NAME)Plugin/g' $(EXAMPLES_DIR)/$(NAME)/main.go
	sed -i 's/Simple Gain/Simple $(NAME)/g' $(EXAMPLES_DIR)/$(NAME)/main.go
	sed -i 's/com.clapgo.gain/com.clapgo.$(NAME)/g' $(EXAMPLES_DIR)/$(NAME)/main.go
	sed -i 's/gainPlugin/$(NAME)Plugin/g' $(EXAMPLES_DIR)/$(NAME)/main.go
	sed -i 's/GainGetPluginCount/$(NAME)GetPluginCount/g' $(EXAMPLES_DIR)/$(NAME)/main.go
	@echo "Example plugin created at $(EXAMPLES_DIR)/$(NAME)/main.go"
	@echo "Build it with: make $(NAME)"

# Declare phony targets
.PHONY: all clean help examples install new $(PLUGINS)