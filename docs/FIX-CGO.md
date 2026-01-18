# 修复说明

## 问题
之前使用的 SQLite 驱动 (`mattn/go-sqlite3`) 需要 CGO 支持，在 Windows 交叉编译时会失败，错误信息：
```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work.
```

## 解决方案
已替换为纯 Go 实现的 SQLite 驱动：`github.com/glebarez/sqlite`

### 优点
- ✅ 无需 CGO，可以直接交叉编译
- ✅ 不需要安装 mingw-w64 或其他 C 编译器
- ✅ 支持所有 GORM 功能
- ✅ 性能与原生驱动相当

## 如何重新编译

### 1. 更新依赖
```bash
export GOPROXY=https://goproxy.cn,direct  # 国内用户推荐
go mod tidy
```

### 2. 编译 Windows 版本
```bash
# 使用 Makefile（推荐）
make build-windows         # 控制台版本（输出到 bin/go-proxy-server.exe）
make build-windows-gui     # 系统托盘版本（输出到 bin/go-proxy-server-gui.exe）

# 或手动编译
mkdir -p bin
GOOS=windows GOARCH=amd64 go build -o bin/go-proxy-server.exe ./cmd/server
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o bin/go-proxy-server-gui.exe ./cmd/server
```

## 测试
将 `bin/go-proxy-server.exe` 复制到 Windows 系统，双击运行：
1. 应该看到控制台窗口
2. 日志显示数据库打开成功
3. Web管理界面启动成功

或使用GUI版本 `bin/go-proxy-server-gui.exe`（系统托盘模式）

## 日志位置
```
%APPDATA%\go-proxy-server\app.log
```

示例：
```
C:\Users\xiaoming\AppData\Roaming\go-proxy-server\app.log
```

## 变更的文件
- `go.mod` - 更新依赖，使用 `github.com/glebarez/sqlite`
- `cmd/server/main.go` - 修改导入语句
- `internal/logger/logger.go` - 添加详细日志
- `Makefile` - 统一的构建流程

## 编译输出
所有二进制文件输出到 `bin/` 目录：
- `bin/go-proxy-server.exe` - Windows 控制台版本
- `bin/go-proxy-server-gui.exe` - Windows 系统托盘版本（隐藏控制台）
