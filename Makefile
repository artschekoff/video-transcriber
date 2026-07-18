# Adapted from the Joplin "jira-cli — Makefile release template" note, extended
# for this project's native deps: it bundles a whisper model and links against
# whisper.cpp (CGo). `build` first compiles the whisper.cpp static lib from the
# Go module source, then builds the app with the right CGo include/lib paths.
# ponytail: macOS/Metal only. For Linux, drop the Metal frameworks from EXT_LDFLAGS
#           and adjust LIB_PATHS (add fyne-cross if multi-OS releases are needed).
.PHONY: help build run test clean install download-model whisper-lib bundle install-app release

APP_NAME     := VideoTranscriber
DIST_DIR     := dist
BIN          := $(DIST_DIR)/$(APP_NAME)
MODEL_DIR    := models
MODEL_FILE   := $(MODEL_DIR)/ggml-base.bin
MODEL_URL    := https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
INSTALL_DIR  := /usr/local/bin
APP_BUNDLE   := $(DIST_DIR)/$(APP_NAME).app
APPLICATIONS := /Applications
BUNDLE_TAG      = $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
BUNDLE_VERSION ?= $(or $(BUNDLE_TAG),1.0.0)

# whisper.cpp native build (static libs). The Go binding ships no C sources, so
# we fetch the pinned whisper.cpp module, copy it out of the read-only module
# cache (its CMake writes into its own source tree), and cmake-build the lib.
WHISPER_VER   := v1.9.1
WHISPER_CACHE  = $(shell go env GOMODCACHE)/github.com/ggerganov/whisper.cpp@$(WHISPER_VER)
WHISPER_SRC   := build/whisper-src
WHISPER_BUILD := build/whisper
WHISPER_LIB   := $(WHISPER_BUILD)/src/libwhisper.a
INCLUDE_PATHS := $(WHISPER_SRC)/include:$(WHISPER_SRC)/ggml/include
LIB_PATHS     := $(WHISPER_BUILD)/src:$(WHISPER_BUILD)/ggml/src:$(WHISPER_BUILD)/ggml/src/ggml-blas:$(WHISPER_BUILD)/ggml/src/ggml-metal
EXT_LDFLAGS   := -framework Foundation -framework Metal -framework MetalKit -lggml-metal -lggml-blas -lggml-cpu
GO_BUILD      := C_INCLUDE_PATH="$(INCLUDE_PATHS)" LIBRARY_PATH="$(LIB_PATHS)" CGO_ENABLED=1 \
                 go build -ldflags "-extldflags '$(EXT_LDFLAGS)'"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-16s %s\n", $$1, $$2}'

whisper-lib: $(WHISPER_LIB) ## Build the whisper.cpp static library (needs cmake)
$(WHISPER_LIB):
	go mod download github.com/ggerganov/whisper.cpp@$(WHISPER_VER)
	rm -rf $(WHISPER_SRC) && mkdir -p $(WHISPER_SRC)
	cp -R "$(WHISPER_CACHE)/." $(WHISPER_SRC)/ && chmod -R u+w $(WHISPER_SRC)
	cmake -S $(WHISPER_SRC) -B $(WHISPER_BUILD) -DCMAKE_BUILD_TYPE=Release \
		-DBUILD_SHARED_LIBS=OFF -DWHISPER_BUILD_TESTS=OFF -DWHISPER_BUILD_EXAMPLES=OFF
	cmake --build $(WHISPER_BUILD) --target whisper -j4

build: $(MODEL_FILE) $(WHISPER_LIB) ## Build the app into dist/
	@mkdir -p $(DIST_DIR)
	$(GO_BUILD) -o $(BIN) .

download-model: $(MODEL_FILE) ## Download the whisper base model (~142MB)

$(MODEL_FILE):
	@mkdir -p $(MODEL_DIR)
	@echo "Downloading whisper base model (~142MB)..."
	curl -L --progress-bar -o $(MODEL_FILE) "$(MODEL_URL)"

run: build ## Build and run the app
	./$(BIN)

test: $(WHISPER_LIB) ## Run all tests
	C_INCLUDE_PATH="$(INCLUDE_PATHS)" LIBRARY_PATH="$(LIB_PATHS)" CGO_ENABLED=1 \
		go test -ldflags "-extldflags '$(EXT_LDFLAGS)'" ./...

install: build ## Build and install the raw binary onto PATH (CLI use)
	sudo install -m 0755 $(BIN) $(INSTALL_DIR)/$(APP_NAME)
	@echo "Installed $(APP_NAME) to $(INSTALL_DIR)"

bundle: build ## Package into dist/VideoTranscriber.app (macOS)
	rm -rf $(APP_BUNDLE)
	mkdir -p $(APP_BUNDLE)/Contents/MacOS
	cp $(BIN) $(APP_BUNDLE)/Contents/MacOS/$(APP_NAME)
	sed 's/__VERSION__/$(BUNDLE_VERSION)/g' packaging/darwin/Info.plist > $(APP_BUNDLE)/Contents/Info.plist
	@echo "Built $(APP_BUNDLE)"

install-app: bundle ## Install the .app into /Applications
	rm -rf $(APPLICATIONS)/$(APP_NAME).app
	cp -R $(APP_BUNDLE) $(APPLICATIONS)/
	@echo "Installed $(APP_NAME).app to $(APPLICATIONS)"

clean: ## Remove build artifacts, native libs, and the downloaded model
	rm -rf $(DIST_DIR) $(MODEL_DIR) build

release: ## Tag + publish a GitHub release (interactive semver bump)
	@set -e; \
	git diff --quiet && git diff --cached --quiet || { echo "Working tree is dirty; commit or stash first."; exit 1; }; \
	git rev-parse --abbrev-ref @{u} >/dev/null 2>&1 || { echo "No upstream tracking branch; push first."; exit 1; }; \
	test -z "$$(git log @{u}.. --oneline)" || { echo "You have unpushed commits; push first."; exit 1; }; \
	$(MAKE) test; \
	current=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	echo "Current version: $$current"; \
	printf "Bump type? [major/minor/patch] (default: patch): "; \
	read bump; \
	bump=$${bump:-patch}; \
	ver=$${current#v}; \
	major=$$(echo $$ver | cut -d. -f1); \
	minor=$$(echo $$ver | cut -d. -f2); \
	patch=$$(echo $$ver | cut -d. -f3); \
	case $$bump in \
		major) major=$$((major+1)); minor=0; patch=0 ;; \
		minor) minor=$$((minor+1)); patch=0 ;; \
		*)     patch=$$((patch+1)) ;; \
	esac; \
	new="v$${major}.$${minor}.$${patch}"; \
	printf "Tag and release %s? [y/N]: " "$$new"; \
	read confirm; \
	case $$confirm in [yY]*) ;; *) echo "Aborted."; exit 1 ;; esac; \
	echo "Building and releasing $$new..."; \
	$(MAKE) build; \
	git tag $$new; \
	git push origin $$new; \
	gh release create $$new "$(BIN)" --generate-notes --title "Release $$new"
