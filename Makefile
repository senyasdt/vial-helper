SHELL := /bin/sh

.PHONY: build win windows linux macos clean help

GO ?= go
APP := vial-helperd
CMD := ./cmd/vial-helperd
DIST := dist

PLATFORM_GOAL := $(strip $(firstword $(filter win windows linux macos,$(MAKECMDGOALS))))

ifeq ($(PLATFORM_GOAL),win)
PLATFORM := win
endif
ifeq ($(PLATFORM_GOAL),windows)
PLATFORM := win
endif
ifeq ($(PLATFORM_GOAL),linux)
PLATFORM := linux
endif
ifeq ($(PLATFORM_GOAL),macos)
PLATFORM := macos
endif

ifeq ($(PLATFORM),win)
OUT_DIR := $(DIST)/local-windows
	BIN_NAME := $(APP).exe
	INSTALLER_FILES := scripts/install/install-windows.bat scripts/install/uninstall-windows.bat
	GOOS := windows
	GOARCH := amd64
endif

ifeq ($(PLATFORM),linux)
OUT_DIR := $(DIST)/local-linux
	BIN_NAME := $(APP)
	INSTALLER_FILES := scripts/install/install-linux.sh scripts/install/uninstall-linux.sh
	GOOS := linux
	GOARCH := amd64
endif

ifeq ($(PLATFORM),macos)
OUT_DIR := $(DIST)/local-macos
	BIN_NAME := $(APP)
	INSTALLER_FILES := scripts/install/install-macos.sh scripts/install/uninstall-macos.sh
	GOOS := darwin
	GOARCH := arm64
endif

build:
ifndef PLATFORM
	@echo "usage: make build [win|linux|macos]"
	@echo "examples:"
	@echo "  make build win"
	@echo "  make build linux"
	@echo "  make build macos"
	@echo "  make win"
	@exit 2
endif
	@mkdir -p "$(OUT_DIR)"
ifeq ($(PLATFORM),win)
	@CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o "$(OUT_DIR)/$(BIN_NAME)" $(CMD)
else
	@CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o "$(OUT_DIR)/$(BIN_NAME)" $(CMD)
	@chmod +x "$(OUT_DIR)/$(BIN_NAME)"
endif
	@cp $(INSTALLER_FILES) "$(OUT_DIR)/"
ifeq ($(PLATFORM),linux)
	@chmod +x "$(OUT_DIR)"/install-linux.sh "$(OUT_DIR)"/uninstall-linux.sh
endif
ifeq ($(PLATFORM),macos)
	@chmod +x "$(OUT_DIR)"/install-macos.sh "$(OUT_DIR)"/uninstall-macos.sh
endif
	@echo "built $(PLATFORM) package in $(OUT_DIR)"

win: build

windows: build

linux: build

macos: build

clean:
	@rm -rf "$(DIST)/local-linux" "$(DIST)/local-macos" "$(DIST)/local-windows"

help:
	@echo "usage:"
	@echo "  make build win"
	@echo "  make build linux"
	@echo "  make build macos"
	@echo "  make win"
	@echo "  make linux"
	@echo "  make macos"

%:
	@:
