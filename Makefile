# pinned versions
FABRIC_VERSION ?= 2.5.0
FABRIC_TWO_DIGIT_VERSION = $(shell echo $(FABRIC_VERSION) | cut -d '.' -f 1,2)

# need to install fabric binaries outside of fsc tree for now (due to chaincode packaging issues)
FABRIC_BINARY_BASE=$(PWD)/../fabric
FAB_BINS ?= $(FABRIC_BINARY_BASE)/bin

TOP = .

all: install-tools checks unit-tests

VERSION := $(shell git describe --tags --dirty --always)
COMMIT := $(shell git rev-parse --short HEAD)

GO_LD_FLAGS += -X github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/internal.Version=$(VERSION)
GO_LD_FLAGS += -X github.com/hyperledger/fabric-x-endorser/cmd/fxconfig/internal.Commit=$(COMMIT)
GO_FLAGS = -ldflags "$(GO_LD_FLAGS)"

.PHONY: fxconfig
fxconfig:
	@env GOBIN=$(FAB_BINS) go install $(GO_FLAGS) ./cmd/fxconfig

.PHONY: install-tools
install-tools:
# Thanks for great inspiration https://marcofranssen.nl/manage-go-tools-via-go-modules
	@echo Installing tools from tools/tools.go
	@cd tools; cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: download-fabric
download-fabric:
	./scripts/download_fabric.sh $(FABRIC_BINARY_BASE) $(FABRIC_VERSION)

# include the checks target
include $(TOP)/checks.mk

.PHONY: unit-tests
unit-tests:
	@export FAB_BINS=$(FAB_BINS); go test -cover $(shell go list ./... | grep -v '/integration/')

.PHONY: protos
protos:
	@./scripts/compile_proto.sh

.PHONY: tidy
tidy:
	@./scripts/gomate.sh tidy

.PHONY: fakes
fakes:
	counterfeiter platform/fabricx/core/fabricx/committer/api/protoqueryservice QueryServiceClient

.PHONY: formartimports
formartimports:
	goimports -w $(find . -type f -name '*.go')

