LD_FLAGS=-ldflags " \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.idpbuilderVersion=$(shell git describe --always --tags --dirty --broken) \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.gitCommit=$(shell git rev-parse HEAD) \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
    "

# The name of the binary. Defaults to idpbuilder
OUT_FILE ?= idpbuilder

.PHONY: build
build: manifests generate fmt vet embedded-resources
	go build $(LD_FLAGS) -o $(OUT_FILE) main.go

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.1

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUSTOMIZE ?= $(LOCALBIN)/kustomize
HELM_TGZ ?= $(LOCALBIN)/helm.tar.gz
HELM ?= $(LOCALBIN)/helm

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.20.0

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
ifeq ($(RUN),)
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -p 1 --tags=integration ./... -coverprofile cover.out
else
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -p 1 --tags=integration ./... -coverprofile cover.out -run $(RUN)
endif

	

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=pkg/controllers/resources

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: ## Download kustomize if necessary
ifeq (,$(wildcard $(KUSTOMIZE)))
	cd $(LOCALBIN) && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
endif

helm_os := $(shell uname | tr '[:upper:]' '[:lower:]')
helm_version ?= 3.15.0
ifeq ($(shell uname -m), x86_64)
	helm_arch ?= amd64
endif
ifeq ($(shell uname -m), arm64)
	helm_arch ?= arm64
endif
ifeq ($(shell uname -m), aarch64)
	helm_arch ?= arm64
endif


.PHONY: helm
helm: ## Download helm if necessary
ifeq (,$(wildcard $(HELM)))
	curl https://get.helm.sh/helm-v$(helm_version)-$(helm_os)-$(helm_arch).tar.gz -o $(HELM_TGZ)
	tar xvzf $(HELM_TGZ) -C $(LOCALBIN) --strip-components 1 $(helm_os)-$(helm_arch)/helm
	chmod +x $(HELM)
endif

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: embedded-resources
embedded-resources: kustomize helm
	export PATH=$(LOCALBIN):$$PATH; ./hack/embedded-resources.sh;

.PHONY: e2e
e2e: build
	go test -v -p 1 -timeout 15m --tags=e2e ./tests/e2e/...

# ─── Upstream Management ────────────────────────────────────────────
# Targets for tracking and cherry-picking changes from cnoe-io/idpbuilder.
# The 'upstream' remote should point to git@github.com:cnoe-io/idpbuilder.git
#
# Workflow:
#   make fetch-upstream          # fetch latest upstream changes
#   make log-upstream            # see what's new since we diverged
#   make diff-upstream           # diff our main vs upstream/main
#   make cherry-pick COMMIT=abc  # cherry-pick a specific upstream commit

UPSTREAM_REMOTE ?= upstream
UPSTREAM_BRANCH ?= main

.PHONY: fetch-upstream
fetch-upstream: ## Fetch latest changes and tags from cnoe-io/idpbuilder
	@echo "Fetching from $(UPSTREAM_REMOTE)..."
	@git fetch $(UPSTREAM_REMOTE) --tags
	@echo "Latest upstream commit:"
	@git log -1 --oneline $(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH)

.PHONY: log-upstream
log-upstream: ## Show upstream commits not yet in our main branch
	@echo "Commits in $(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH) not in HEAD:"
	@git log --oneline HEAD..$(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH)

.PHONY: log-upstream-detail
log-upstream-detail: ## Show upstream commits with diffs (detailed)
	@git log -p HEAD..$(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH)

.PHONY: diff-upstream
diff-upstream: ## Show diff between our main and upstream/main
	@git diff HEAD...$(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH) --stat

.PHONY: cherry-pick
cherry-pick: ## Cherry-pick a specific upstream commit (usage: make cherry-pick COMMIT=<sha>)
ifndef COMMIT
	$(error COMMIT is required. Usage: make cherry-pick COMMIT=<sha>)
endif
	@echo "Cherry-picking $(COMMIT)..."
	@git cherry-pick $(COMMIT)

.PHONY: cherry-pick-range
cherry-pick-range: ## Cherry-pick a range of upstream commits (usage: make cherry-pick-range FROM=<sha> TO=<sha>)
ifndef FROM
	$(error FROM is required. Usage: make cherry-pick-range FROM=<sha> TO=<sha>)
endif
ifndef TO
	$(error TO is required. Usage: make cherry-pick-range FROM=<sha> TO=<sha>)
endif
	@echo "Cherry-picking range $(FROM)..$(TO)..."
	@git cherry-pick $(FROM)..$(TO)

.PHONY: upstream-status
upstream-status: ## Show how far ahead/behind we are from upstream
	@echo "Our main vs $(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH):"
	@printf "  Ahead:  %s\n" "$$(git rev-list --count $(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH)..HEAD)"
	@printf "  Behind: %s\n" "$$(git rev-list --count HEAD..$(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH))"
	@printf "  Last upstream fetch: %s\n" "$$(git log -1 --format='%ci' $(UPSTREAM_REMOTE)/$(UPSTREAM_BRANCH))"

# ─── Documentation ─────────────────────────────────────────────────
ORG_DIR := $(shell dirname $(CURDIR))

.PHONY: sync-docs
sync-docs: ## Sync documentation to Obsidian vault (via org-level docs repo)
	@if [ -x "$(ORG_DIR)/docs/scripts/sync-docs.sh" ]; then \
		bash "$(ORG_DIR)/docs/scripts/sync-docs.sh"; \
	else \
		printf "\033[0;31mx\033[0m docs repo not found at $(ORG_DIR)/docs/\n"; \
		printf "  Clone it: gh repo clone idpbuilder/docs $(ORG_DIR)/docs\n"; \
		exit 1; \
	fi

.PHONY: hooks
hooks: ## Git hooks are Nix-managed (see ADR-005)
	@printf "\033[0;36m[info]\033[0m Hooks are globally managed via Nix (cmdr git module)\n"
	@printf "  Deploy: unimart deli switch\n"
	@printf "  Extend: add .githooks/<hook-name> to this repo\n"
