PYTHON ?= python
GO ?= go

.PHONY: help poc build test acceptance fmt
help:
	@echo "poc        Run real GraphQL feasibility in disposable Docker containers"
	@echo "build      Build the provider binary"
	@echo "test       Run Go unit tests"
	@echo "acceptance Run Terraform CLI lifecycle against disposable GoAlert"
	@echo "fmt        Format Go and Terraform sources"
poc:
	$(PYTHON) scripts/fixture.py
build:
	$(GO) build -o bin/ ./...
test: test-harness
	$(GO) test -timeout=120s ./...
acceptance:
	$(PYTHON) scripts/acceptance.py
fmt:
	$(GO) fmt ./...
	terraform fmt -recursive examples

GORELEASER ?= goreleaser
.PHONY: vet check-format release-check release-snapshot release-smoke
vet:
	$(GO) vet ./...
check-format:
	$(PYTHON) scripts/check_format.py
release-check:
	$(GORELEASER) check
release-snapshot:
	$(GORELEASER) release --snapshot --clean --skip=publish,sign
release-smoke:
	$(PYTHON) scripts/release_smoke.py --goreleaser "$(GORELEASER)"

.PHONY: test-harness
test-harness:
	$(PYTHON) -m unittest discover -s scripts -p "test_*.py" -v
