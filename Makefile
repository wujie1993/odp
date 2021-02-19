.PHONY: build clean tool lint help

module = github.com/wujie1993/waves
version = v1.2.0
commit = $(shell git log -n 1 --pretty=format:"%H")
build = $(shell git log -n 1 --format="%ad")
author = $(shell git log -n 1 --format="%ae")
go_version = $(shell go version)
build_dir = ./build

all: build

prebuild:
	mkdir -p $(build_dir)/outputs

# 编译客户端二进制程序
ctl: prebuild gen
	go build -v -o $(build_dir)/outputs/wavectl -ldflags " \
		-X '$(module)/pkg/version.Version=$(version)' \
		-X '$(module)/pkg/version.GoVersion=$(go_version)' \
		-X '$(module)/pkg/version.Commit=$(commit)' \
		-X '$(module)/pkg/version.Build=$(build)' \
		-X '$(module)/pkg/version.Author=$(author)' \
	" cmd/wavectl/wavectl.go

# 编译服务端二进制程序
build: prebuild gen doc
	go build -v -o $(build_dir)/outputs/waves -ldflags " \
		-X '$(module)/pkg/version.Version=$(version)' \
		-X '$(module)/pkg/version.GoVersion=$(go_version)' \
		-X '$(module)/pkg/version.Commit=$(commit)' \
		-X '$(module)/pkg/version.Build=$(build)' \
		-X '$(module)/pkg/version.Author=$(author)' \
	"
	cp -rf conf $(build_dir)/outputs/

# 生成swagger api文档
doc:
	swag init --propertyStrategy pascalcase

# 生成orm库代码
gen-orm:
	rm -f ./pkg/orm/runtime/zz_generated*
	rm -f ./pkg/orm/v1/zz_generated*
	rm -f ./pkg/orm/v2/zz_generated*

	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/runtime
	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/v1
	@go run cmd/codegen/codegen.go orm --pkg-path=./pkg/orm/v2

# 生成客户端库代码
gen-client:
	rm -f pkg/client/v1/zz_generated*
	rm -f pkg/client/v2/zz_generated*

	@go run cmd/codegen/codegen.go client -i pkg/orm/v1/ -o pkg/client/v1/
	@go run cmd/codegen/codegen.go client -i pkg/orm/v2/ -o pkg/client/v2/

# 生成项目代码！！！注意！！！在提交代码前先执行代码生成
gen: gen-orm gen-client
	gofmt -w -s ./

# 运行服务端
run: gen doc
	go run main.go

# 打包项目构建后的产物，包括客户端与服务端以及配置文件
release: build ctl
	mkdir -p $(build_dir)/releases
	cd $(build_dir)/outputs && tar -cvf ../releases/waves.tar ./

test: 
	go test ./... -cover -ldflags " \
		-X '$(module)/tests.EtcdEndpoint=localhost:2379' \
	"

image:
	docker build . -t waves:latest	

tool:
	go vet ./...; true
	gofmt -w .

lint:
	golint ./...

# 清理项目中的临时文件
clean:
	rm -rf $(build_dir)
	go clean -i .

help:
	@echo "make: compile packages and dependencies"
	@echo "make pack: build and pack binary and configuration file"
	@echo "make test: run unit tests. please make sure you have already setup etcd server and listen on 'localhost:2379'."
	@echo "make tool: run specified go tool"
	@echo "make lint: golint ./..."
	@echo "make clean: remove object files and cached files"
	@echo "make doc: generate swagger api documents. please make sure you have swag(https://github.com/swaggo/swag) installed"
