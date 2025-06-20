VERSION := $(shell git describe --tags)
COMMIT  := $(shell git log -1 --format='%H')

all: build

LD_FLAGS = -X svm-lsd-relay/cmd.Version=$(VERSION) \
	-X svm-lsd-relay/cmd.Commit=$(COMMIT) \

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

get:
	@echo "  >  \033[32mDownloading & Installing all the modules...\033[0m "
	go mod tidy && go mod download

build:
	@echo " > \033[32mBuilding sonic-lsd-relay...\033[0m "
	go build -mod readonly $(BUILD_FLAGS) -o build/sonic-lsd-relay main.go

install: 
	@echo " > \033[32mInstalling sonic-lsd-relay...\033[0m "
	go install -mod readonly $(BUILD_FLAGS) ./...


build-linux:
	@GOOS=linux GOARCH=amd64 go build --mod readonly $(BUILD_FLAGS) -o ./build/sonic-lsd-relay main.go


clean:
	@echo " > \033[32mCleanning build files ...\033[0m "
	rm -rf build
fmt :
	@echo " > \033[32mFormatting go files ...\033[0m "
	go fmt ./...


# repo: github.com:gagliardetto/anchor-go.git
anchor:
	@echo "  >  \033[32mGenerating anchor bindings...\033[0m "
	anchor-go -src ./pkg/lsd_program/lsd_program.json -dst ./pkg/lsd_program/

get-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s latest

lint:
	golangci-lint run ./... --skip-files ".+_test.go"

.PHONY: all lint test race msan tools clean build
