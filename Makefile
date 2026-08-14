# NecoE (necore) 开发常用命令
# 用法：make fmt / make test / make build / make run / make all

# Windows 的应用程序控制策略可能拦截默认 Temp 目录下的测试二进制，
# 把 Go 的临时目录固定到仓库内可绕开（非 Windows 无影响）。
GOTMPDIR ?= $(CURDIR)/.gotmp
export GOTMPDIR

.PHONY: all fmt vet build test run tidy

all: fmt vet build test

# 语法格式化（gofmt 全仓库）
fmt:
	gofmt -w .

# 静态检查
vet:
	go vet ./...

# 编译（不含测试）
build:
	go build ./...

# 单元/集成测试（含 -race 竞态检测；TESTFLAGS 可追加参数，如 make test TESTFLAGS="-run TestXxx -v"）
test:
	go test -race ./... $(TESTFLAGS)

# 直接运行后端
run:
	go run .
