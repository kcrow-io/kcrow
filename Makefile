#!/usr/bin/make -f

# Copyright 2022 Authors of kcrow
# SPDX-License-Identifier: Apache-2.0


include Makefile.defs

all: build-bin

.PHONY: all build install

CONTROLLER_BIN_SUBDIRS := cmd/daemon

SUBDIRS := $(CONTROLLER_BIN_SUBDIRS)

build-bin:
	for i in $(SUBDIRS); do $(MAKE) $(SUBMAKEOPTS) -C $$i all; done

install-bin:
	$(QUIET)$(INSTALL) -m 0755 -d $(DESTDIR_BIN)
	for i in $(SUBDIRS); do $(MAKE) $(SUBMAKEOPTS) -C $$i install; done

install-bash-completion:
	$(QUIET)$(INSTALL) -m 0755 -d $(DESTDIR_BIN)
	for i in $(SUBDIRS); do $(MAKE) $(SUBMAKEOPTS) -C $$i install-bash-completion; done


# ============ build-load-image ============
.PHONY: image
image:
	@for NAME in $(IMAGES); do \
		docker build --build-arg RACE --build-arg GIT_COMMIT_VERSION=$(GIT_COMMIT_VERSION) \
				--build-arg GIT_COMMIT_TIME=$(GIT_COMMIT_TIME) \
				--build-arg VERSION=$(GIT_COMMIT_VERSION) \
				--file $(ROOT_DIR)/images/"$${NAME##*/}"/Dockerfile \
				--output type=docker --tag $${NAME}:$(IMAGE_TAG) . ; \
		echo "Image $${NAME##*/} build success" ; \
	done

.PHONY: one-image
one-image:
	@for NAME in $(IMAGES); do \
		buildah bud --platform linux/amd64 \
				--build-arg GIT_COMMIT_VERSION=$(GIT_COMMIT_VERSION) \
				--build-arg GIT_COMMIT_TIME=$(GIT_COMMIT_TIME) \
				--build-arg VERSION=$(GIT_COMMIT_VERSION) \
				--file $(ROOT_DIR)/Dockerfile \
				--tag $${NAME}:$(IMAGE_TAG) . ; \
		echo "Image $${NAME##*/} build success" ; \
	done


#============ lints ====================

.PHONY: lint-golang
lint-golang: golangci-lint
	@$(ECHO_CHECK) contrib/scripts/check-go-fmt.sh
	$(QUIET) contrib/scripts/check-go-fmt.sh
	@$(ECHO_CHECK) vetting all GOFILES...
	$(QUIET) $(GO_MOD_TIDY) 
	$(QUIET) $(GO_VET) \
    ./cmd/... \
    ./pkg/... \
    ./contrib/...
	@$(ECHO_CHECK) golangci-lint
	$(QUIET) golangci-lint run

.PHONY: lint-markdown-format
lint-markdown-format:
	@$(CONTAINER_ENGINE) container run --rm \
		--entrypoint sh -v $(ROOT_DIR):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest \
		-c '/usr/local/bin/markdownlint -c /workdir/.github/markdownlint.yaml -p /workdir/.github/markdownlintignore  /workdir/' ; \
		if (($$?==0)) ; then echo "congratulations ,all pass" ; else echo "error, pealse refer <https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md> " ; fi

.PHONY: fix-markdown-format
fix-markdown-format:
	@$(CONTAINER_ENGINE) container run --rm  \
		--entrypoint sh -v $(ROOT_DIR):/workdir ghcr.io/igorshubovych/markdownlint-cli:latest \
		-c '/usr/local/bin/markdownlint -f -c /workdir/.github/markdownlint.yaml -p /workdir/.github/markdownlintignore  /workdir/'

.PHONY: lint-yaml
lint-yaml:
	@$(CONTAINER_ENGINE) container run --rm \
		--entrypoint sh -v $(ROOT_DIR):/data cytopia/yamllint \
		-c '/usr/bin/yamllint -c /data/.github/yamllint-conf.yml /data' ; \
		if (($$?==0)) ; then echo "congratulations ,all pass" ; else echo "error, pealse refer <https://yamllint.readthedocs.io/en/stable/rules.html> " ; fi

.PHONY: lint-code-spell
lint-code-spell:
	$(QUIET) if ! which codespell &> /dev/null ; then \
  				echo "try to install codespell" ; \
  				if ! pip3 install codespell==2.2.1 ; then \
  					echo "error, miss tool codespell, install it: pip3 install codespell" ; \
  					exit 1 ; \
  				fi \
  			fi ;\
  			codespell --config .github/codespell-config

.PHONY: fix-code-spell
fix-code-spell:
	$(QUIET) if ! which codespell &> /dev/null ; then \
  				echo "try to install codespell" ; \
  				if ! pip3 install codespell ; then \
  					echo "error, miss tool codespell, install it: pip3 install codespell" ; \
  					exit 1 ;\
  				fi \
  			fi; \
  			codespell --config .github/codespell-config  --write-changes

#============ manifests ====================

.PHONY: manifests
manifests:
	@echo "Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects."
	@$(QUIET) tools/k8s-controller-gen/update-controller-gen.sh manifests

.PHONY: generate-k8s-api
generate-k8s-api:
	@echo "Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations."
	@$(QUIET) tools/k8s-controller-gen/update-controller-gen.sh deepcopy

.PHONY: manifests-verify
manifests-verify:
	@echo "Verify WebhookConfiguration, ClusterRole and CustomResourceDefinition objects."
	@$(QUIET) tools/k8s-controller-gen/update-controller-gen.sh verify

.PHONY: gofmt
gofmt: ## Run gofmt on Go source files in the repository.
	$(QUIET)for pkg in $(GOFILES); do $(GO) fmt $$pkg; done

.PHONY: dev-doctor
dev-doctor:
	@ $(GO) version 2>/dev/null || ( echo "go not found, see https://golang.org/doc/install" ; false )
	@ JUST_CLI_CHECK=true test/scripts/install-tools.sh
	@ echo "all tools ready"

#============ tools ====================

.PHONY: update-authors
update-authors: ## Update AUTHORS file for the repository.
	@echo "Updating AUTHORS file..."
	@echo "The following people, in alphabetical order, have either authored or signed" > AUTHORS
	@echo "off on commits in the Kcrow repository:" >> AUTHORS
	@echo "" >> AUTHORS
	@contrib/authorgen/authorgen.sh >> AUTHORS

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.59.1
.PHONY: golangci-lint
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: licenses-all
licenses-all: ## Generate file with all the License from dependencies.
	@$(GO) run ./contrib/licensegen > LICENSE.all || ( rm -f LICENSE.all ; false )

.PHONY: licenses-check
licenses-check:
	@$(ECHO_CHECK) tools/scripts/check-miss-license.sh
	$(QUIET) tools/scripts/check-miss-license.sh

.PHONY: update-go-version
update-go-version: ## Update Go version for all the components
	@echo "GO_MAJOR_AND_MINOR_VERSION=${GO_MAJOR_AND_MINOR_VERSION}"
	@echo "GO_IMAGE_VERSION=${GO_IMAGE_VERSION}"
	# ===== Update Go version for GitHub workflow
	$(QUIET) for fl in $(shell find .github/workflows -name "*.yaml" -print) ; do \
  			sed -i 's/go-version: .*/go-version: ${GO_IMAGE_VERSION}/g' $$fl ; \
  			done
	@echo "Updated go version in GitHub Actions to $(GO_IMAGE_VERSION)"
	# ======= Update Go version in main.go.
	$(QUIET) for fl in $(shell find .  -name main.go -not -path "./vendor/*" -print); do \
		sed -i \
			-e 's|^//go:build go.*|//go:build go${GO_MAJOR_AND_MINOR_VERSION}|g' \
			-e 's|^// +build go.*|// +build go${GO_MAJOR_AND_MINOR_VERSION}|g' \
			$$fl ; \
	done
	# ====== Update Go version in go.mod
	$(QUIET) sed -i -E 's/^go .*/go '$(GO_MAJOR_AND_MINOR_VERSION)'/g' go.mod
	@echo "Updated go version in go.mod to $(GO_VERSION)"
ifeq (${shell [ -d ./test ] && echo done},done)
	# ======= Update Go version in test scripts.
	@echo "Updated go version in test scripts to $(GO_VERSION)"
endif
	# ===== Update Go version in Dockerfiles.
	$(QUIET) $(MAKE) -C images update-golang-image
	@echo "Updated go version in image Dockerfiles to $(GO_IMAGE_VERSION)"

# ==========================


# should label for each test file
.PHONY: check_test_label
check_test_label:
	@ALL_TEST_FILE=` find  ./  -name "*_test.go" -not -path "./vendor/*" ` ; FAIL="false" ; \
		for ITEM in $$ALL_TEST_FILE ; do \
			[[ "$$ITEM" == *_suite_test.go ]] && continue  ; \
			! grep 'Label(' $${ITEM} &>/dev/null && FAIL="true" && echo "error, miss Label in $${ITEM}" ; \
		done ; \
		[ "$$FAIL" == "true" ] && echo "error, label check fail" && exit 1 ; \
		echo "each test.go is labeled right"


.PHONY: unitest-tests
unitest-tests: check_test_label
	@echo "run unitest-tests"
	$(QUIET) $(ROOT_DIR)/tools/scripts/ginkgo.sh \
		--cover --coverprofile=./coverage.out --covermode set \
		--json-report unitestreport.json \
		-randomize-suites -randomize-all --keep-going  --timeout=1h  -p \
		-vv  -r $(ROOT_DIR)/pkg $(ROOT_DIR)/cmd
	$(QUIET) go tool cover -html=./coverage.out -o coverage-all.html

.PHONY: preview_doc
preview_doc: PROJECT_DOC_DIR := ${ROOT_DIR}/docs
preview_doc:
	-docker stop doc_previewer &>/dev/null
	-docker rm doc_previewer &>/dev/null
	@echo "set up preview http server  "
	@echo "you can visit the website on browser with url 'http://127.0.0.1:8000' "
	[ -f "docs/mkdocs.yml" ] || { echo "error, miss docs/mkdocs.yml "; exit 1 ; }
	docker run --rm  -p 8000:8000 --name doc_previewer -v $(PROJECT_DOC_DIR):/host/docs \
        --entrypoint sh \
        --stop-timeout 3 \
        --stop-signal "SIGKILL" \
        squidfunk/mkdocs-material:8.5.11  -c "cd /host ; cp docs/mkdocs.yml ./ ;  mkdocs serve -a 0.0.0.0:8000"
	#sleep 10 ; if curl 127.0.0.1:8000 &>/dev/null  ; then echo "succeeded to set up preview server" ; else echo "error, failed to set up preview server" ; docker stop doc_previewer ; exit 1 ; fi


.PHONY: build_doc
build_doc: PROJECT_DOC_DIR := ${ROOT_DIR}/docs
build_doc: OUTPUT_TAR := site.tar.gz
build_doc:
	-docker stop doc_builder &>/dev/null
	-docker rm doc_builder &>/dev/null
	[ -f "docs/mkdocs.yml" ] || { echo "error, miss docs/mkdocs.yml "; exit 1 ; }
	-@ rm -f ./docs/$(OUTPUT_TAR)
	@echo "build doc html " ; \
		docker run --rm --name doc_builder  \
		-v ${PROJECT_DOC_DIR}:/host/docs \
                --entrypoint sh \
                squidfunk/mkdocs-material:8.5.11 -c "cd /host && cp ./docs/mkdocs.yml ./ && mkdocs build && cd site && tar -czvf site.tar.gz * && mv ${OUTPUT_TAR} ../docs/"
	@ [ -f "$(PROJECT_DOC_DIR)/$(OUTPUT_TAR)" ] || { echo "failed to build site to $(PROJECT_DOC_DIR)/$(OUTPUT_TAR) " ; exit 1 ; }
	@ echo "succeeded to build site to $(PROJECT_DOC_DIR)/$(OUTPUT_TAR) "


.PHONY: check_doc
check_doc: PROJECT_DOC_DIR := ${ROOT_DIR}/docs
check_doc: OUTPUT_TAR := site.tar.gz
check_doc:
	-docker stop doc_builder &>/dev/null
	-docker rm doc_builder &>/dev/null
	[ -f "docs/mkdocs.yml" ] || { echo "error, miss docs/mkdocs.yml "; exit 1 ; }
	-@ rm -f ./docs/$(OUTPUT_TAR)
	@echo "check doc" ; \
	MESSAGE=`docker run --rm --name doc_builder  \
		-v ${PROJECT_DOC_DIR}:/host/docs \
                --entrypoint sh \
                squidfunk/mkdocs-material:8.5.11 -c "cd /host && cp ./docs/mkdocs.yml ./ && mkdocs build 2>&1 && cd site && tar -czvf site.tar.gz * && mv ${OUTPUT_TAR} ../docs/" 2>&1` ; \
        if (( $$? !=0 )) ; then \
        	echo "!!! error, failed to build doc" ; \
        	exit 1 ; \
        fi ; \
        if grep -E "WARNING .* which is not found" <<< "$${MESSAGE}" ; then  \
        	echo "!!! error, some link is bad" ; \
        	exit 1 ; \
        fi
	@ [ -f "$(PROJECT_DOC_DIR)/$(OUTPUT_TAR)" ] || { echo "failed to build site to $(PROJECT_DOC_DIR)/$(OUTPUT_TAR) " ; exit 1 ; }
	-@ rm -f ./docs/$(OUTPUT_TAR)
	@ echo "all doc is ok "


# ==========================
.PHONY: e2e
e2e: vendor
	-$(QUIET) cd "./tests/e2e" && $(GO_TEST) -v ./...


.PHONY: clean_e2e
clean_e2e:
	-$(QUIET) make -C test clean
	-$(QUIET) rm -f e2ereport.json

.PHONY: clean
clean: clean_e2e
	-$(QUIET) for i in $(SUBDIRS); do $(MAKE) $(SUBMAKEOPTS) -C $$i clean; done
	-$(QUIET) rm -rf $(DESTDIR_BIN)
	-$(QUIET) rm -rf $(DESTDIR_BASH_COMPLETION)

.PHONY: codegen
codegen:
	@echo "Generate k8s SDK with code-generator."
	$(QUIET)  ./tools/scripts/update-codegen.sh

.PHONY: codegen-verify
codegen-verify:
	@echo "Verify generated k8s SDK."
	$(QUIET)  ./tools/scripts/verify-codegen.sh

.PHONY: lint_image_trivy
lint_image_trivy: IMAGE_NAME ?=
lint_image_trivy:
	@ [ -n "$(IMAGE_NAME)" ] || { echo "error, please input IMAGE_NAME" && exit 1 ; }
	@ docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
          -v $(ROOT_DIR)/.trivyignore:/.trivyignore \
          -v /tmp/trivy:/root/trivy.cache/  \
          aquasec/trivy:$(TRIVY_VERSION) image --exit-code 1  --severity $(LINT_TRIVY_SEVERITY_LEVEL)  $(IMAGE_NAME) ; \
      (($$?==0)) || { echo "error, failed to check dockerfile trivy", $(IMAGE_NAME)  && exit 1 ; } ; \
      echo "trivy check: $(IMAGE_NAME) pass"


.PHONY: lint_dockerfile_trivy
lint_dockerfile_trivy:
	@ docker run --rm \
          -v $(ROOT_DIR)/.trivyignore:/.trivyignore \
          -v /tmp/trivy:/root/trivy.cache/  \
          -v $(ROOT_DIR):/tmp/src  \
          aquasec/trivy:$(TRIVY_VERSION) config --exit-code 1  --severity $(LINT_TRIVY_SEVERITY_LEVEL) /tmp/src/images  ; \
      (($$?==0)) || { echo "error, failed to check dockerfile trivy" && exit 1 ; } ; \
      echo "dockerfile trivy check: pass"


.PHONY: lint_chart_trivy
lint_chart_trivy:
	@ docker run --rm \
          -v $(ROOT_DIR)/.trivyignore:/.trivyignore \
          -v /tmp/trivy:/root/trivy.cache/  \
          -v $(ROOT_DIR):/tmp/src  \
          aquasec/trivy:$(TRIVY_VERSION) config --exit-code 1  --severity $(LINT_TRIVY_SEVERITY_LEVEL) /tmp/src/charts  ; \
      (($$?==0)) || { echo "error, failed to check chart trivy" && exit 1 ; } ; \
      echo "chart trivy check: pass"

# ==========================
.PHONY: chart
chart:
	@cd charts ; make
