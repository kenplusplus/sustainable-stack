#Use Go command for build
GO ?= go

#directory of the project.
SOURCE := cmd
BUILD_DIRS := ${SOURCE}/amx ${SOURCE}/qat
BINARY_NAMES := amx_collector qat_collector
GOFILES := pkg cmd 

all: gofmt build

# format go code
gofmt: 
	@echo "Formatting go code..."
	@gofmt -s -w ${GOFILES}

# check format for go files
check_gofmt: 
	@echo "Checking go format..."
	@unformatted=$$(gofmt -e -d -s -l ${GOFILES}); \
	if [ -n "$$unformatted" ]; then \
		echo >&2 "The following Go files must be formatted with gofmt:"; \
		echo "$$unformatted" | awk '{print "  " $$0}'; \
		exit 1; \
	fi

# build all binaries from cmd
build: check_gofmt 
	@echo "Building..."
	@for i in $$(seq 1 $$(echo $$(($(words $(BINARY_NAMES)))))); do \
		BINARY_NAME=$$(echo $(BINARY_NAMES) | cut -d ' ' -f $$i); \
		BUILD_DIR=$$(echo $(BUILD_DIRS) | cut -d ' ' -f $$i); \
		cd $$BUILD_DIR && ${GO} build -o $$BINARY_NAME ; \
		cd $(CURDIR) ;\
	done

# build binary for AMX power model train data collector
amx:check_gofmt 
	@echo "Building amx_collector..."
	@cd ${SOURCE}/amx &&\
		${GO} build -o amx_collector 

# build binary for QAT power model train data collector 
qat:check_gofmt 
	@echo "Building qat_collector..."
	@cd ${SOURCE}/qat &&\
		${GO} build -o qat_collector  

# ensure all the go.mod/go.sum files are up-to-date including vendor/ directory 
vendor: 
	@$(GO) mod tidy
	@$(GO) mod vendor
	@$(GO) mod verify

# clean up binaries
clean:
	@echo "Cleaning..."
	@for i in $$(seq 1 $$(echo $$(($(words $(BINARY_NAMES)))))); do \
		BINARY_NAME=$$(echo $(BINARY_NAMES) | cut -d ' ' -f $$i); \
		BUILD_DIR=$$(echo $(BUILD_DIRS) | cut -d ' ' -f $$i); \
		cd $$BUILD_DIR && rm -f $$BINARY_NAME ; \
		cd $(CURDIR) ;\
	done

help: 
	@echo "Usage: make <target>"
	@echo
	@echo " * 'build'  - Build all binaries for accelerator power model train data collectors from cmd."
	@echo " * 'amx'    - Build binary for AMX power model train data collector."
	@echo " * 'qat'    - Build binary for QAT power model train data collector."
	@echo " * 'vendor' - Ensure all the go.mod/go.sum files are up-to-date including vendor/ directory."
	@echo " * 'clean'  - Clean all binaries."


.PHONY: \
	all \
	gofmt \
	check_format \
	build \
	amx \
	qat \
	vendor \
	clean \
	help