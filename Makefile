.PHONY: build clean tool lint help

module = github.com/wujie1993/waves
version = v1.2.0
commit = $(shell git log -n 1 --pretty=format:"%H")
build = $(shell git log -n 1 --format="%ad")
author = $(shell git log -n 1 --format="%ae")
go_version = $(shell go version)

all: build

build: gen doc
	mkdir -p build
	@go build -v -o build/waves -ldflags " \
		-X '$(module)/pkg/version.Version=$(version)' \
		-X '$(module)/pkg/version.GoVersion=$(go_version)' \
		-X '$(module)/pkg/version.Commit=$(commit)' \
		-X '$(module)/pkg/version.Build=$(build)' \
		-X '$(module)/pkg/version.Author=$(author)' \
	"

doc:
	swag init --propertyStrategy pascalcase

gen:
	rm -f ./pkg/orm/runtime/zz_generated*
	rm -f ./pkg/orm/v1/zz_generated*
	rm -f ./pkg/orm/v2/zz_generated*

	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/runtime
	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/v1
	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/v2

	gofmt -w -s ./pkg/orm/runtime
	gofmt -w -s ./pkg/orm/v1
	gofmt -w -s ./pkg/orm/v2

run: gen doc
	sh scripts/asset/bindata.sh
	go run main.go

pack: build
	cp -r conf build/
	cd build/ ; tar -cvf waves.tar .

test: 
	go test ./pkg/db/*_test.go -cover -coverpkg "$(module)/pkg/db"
	go test ./pkg/orm/v1/*_test.go -cover -coverpkg "$(module)/pkg/orm/v1"

tool:
	go vet ./...; true
	gofmt -w .

lint:
	golint ./...

clean:
	rm -rf ./build
	go clean -i .

help:
	@echo "make: compile packages and dependencies"
	@echo "make pack: build and pack binary and configuration file"
	@echo "make test: run unit tests. please make sure you have already setup etcd server and listen on 'localhost:2379'."
	@echo "make tool: run specified go tool"
	@echo "make lint: golint ./..."
	@echo "make clean: remove object files and cached files"
	@echo "make doc: generate swagger api documents. please make sure you have swag(https://github.com/swaggo/swag) installed"
