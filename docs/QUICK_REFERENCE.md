# 快速参考指南

本文档提供常用操作的快速参考。

## 编译和构建

```bash
# 编译当前平台
make build

# 编译所有平台
make build-all

# 编译特定平台
make build-linux
make build-windows
make build-windows-gui  # Windows GUI模式（系统托盘）
make build-darwin

# 清理编译产物
make clean
```

## 运行模式

### 默认模式（Web管理界面）
```bash
# Windows: 启动系统托盘应用
# Linux/macOS: 启动Web服务器
./go-proxy-server

# 访问: http://localhost:9090
```

### SOCKS5代理
```bash
# 标准模式
./go-proxy-server socks -port 1080

# Bind-listen模式（多IP出口）
./go-proxy-server socks -port 1080 -bind-listen
```

### HTTP代理
```bash
# 标准模式
./go-proxy-server http -port 8080

# Bind-listen模式
./go-proxy-server http -port 8080 -bind-listen
```

### 同时运行两种代理
```bash
# 标准模式
./go-proxy-server both -socks-port 1080 -http-port 8080

# Bind-listen模式
./go-proxy-server both -socks-port 1080 -http-port 8080 -bind-listen
```

### Web管理界面
```bash
# 默认端口9090
./go-proxy-server web

# 自定义端口
./go-proxy-server web -port 8888
```

## 用户管理

```bash
# 添加用户
./go-proxy-server adduser -username alice -password secret123

# 添加用户（带IP记录）
./go-proxy-server adduser -username bob -password pass456 -ip 192.168.1.100

# 删除用户
./go-proxy-server deluser -username alice

# 列出所有用户
./go-proxy-server listuser
```

## IP白名单管理

```bash
# 添加IP到白名单
./go-proxy-server addip -ip 192.168.1.100

# 删除IP（通过Web界面）
# 列出白名单（通过Web界面）
```

## 客户端配置

### SOCKS5客户端

**curl**:
```bash
# 无认证（IP白名单）
curl -x socks5://localhost:1080 https://example.com

# 用户名密码认证
curl -x socks5://username:password@localhost:1080 https://example.com
```

**Firefox**:
```
设置 → 网络设置 → 手动代理配置
SOCKS主机: localhost
端口: 1080
SOCKS v5: 选中
```

**Chrome/Edge**:
```bash
# Windows
chrome.exe --proxy-server="socks5://localhost:1080"

# Linux/macOS
google-chrome --proxy-server="socks5://localhost:1080"
```

### HTTP代理客户端

**curl**:
```bash
# 无认证（IP白名单）
curl -x http://localhost:8080 https://example.com

# 用户名密码认证
curl -x http://username:password@localhost:8080 https://example.com
```

**环境变量**:
```bash
# Linux/macOS
export http_proxy=http://username:password@localhost:8080
export https_proxy=http://username:password@localhost:8080

# Windows (PowerShell)
$env:http_proxy="http://username:password@localhost:8080"
$env:https_proxy="http://username:password@localhost:8080"
```

## 数据文件位置

### 数据库
```bash
# Windows
%APPDATA%\go-proxy-server\data.db

# macOS
~/Library/Application Support/go-proxy-server/data.db

# Linux
~/.local/share/go-proxy-server/data.db
```

### 日志文件
```bash
# Windows
%APPDATA%\go-proxy-server\proxy.log

# macOS
~/Library/Application Support/go-proxy-server/proxy.log

# Linux
~/.local/share/go-proxy-server/proxy.log
```

## 性能调优

### 修改常量配置

编辑 `internal/constants/constants.go`:

```go
// 高并发场景
const (
    HTTPPoolMaxIdleConns        = 200  // 默认100
    HTTPPoolMaxIdleConnsPerHost = 20   // 默认10
)

// 大量域名场景
const (
    DNSCacheMaxSize = 50000  // 默认10000
)

// 安全性优先
const (
    AuthCacheTTL = 2 * time.Minute  // 默认5分钟
)

// 用户变更频繁
const (
    ConfigReloadInterval = 15 * time.Second  // 默认30秒
)
```

修改后重新编译：
```bash
make build
```

### 系统资源限制

```bash
# 增加文件描述符限制
ulimit -n 65535

# 查看当前限制
ulimit -n
```

## 监控和诊断

### 查看日志
```bash
# 实时查看
tail -f ~/.local/share/go-proxy-server/proxy.log

# 查看错误
grep ERROR ~/.local/share/go-proxy-server/proxy.log

# 查看认证失败
grep "Authentication failed" ~/.local/share/go-proxy-server/proxy.log
```

### 查看连接数
```bash
# SOCKS5连接数
netstat -an | grep :1080 | wc -l

# HTTP连接数
netstat -an | grep :8080 | wc -l
```

### 查看进程资源
```bash
# CPU和内存使用
top -p $(pgrep go-proxy-server)

# 详细信息
ps aux | grep go-proxy-server

# 文件描述符
lsof -p $(pgrep go-proxy-server) | wc -l
```

## 常见问题

### Q: 如何重启服务？
```bash
# 找到进程ID
pgrep go-proxy-server

# 停止服务
pkill go-proxy-server

# 启动服务
./go-proxy-server web -port 9090
```

### Q: 如何备份数据？
```bash
# 备份数据库
cp ~/.local/share/go-proxy-server/data.db ~/backup/data.db.$(date +%Y%m%d)
```

### Q: 如何恢复数据？
```bash
# 停止服务
pkill go-proxy-server

# 恢复数据库
cp ~/backup/data.db.20260118 ~/.local/share/go-proxy-server/data.db

# 启动服务
./go-proxy-server web -port 9090
```

### Q: 如何查看版本信息？
```bash
# 查看git提交
git log --oneline -1

# 查看CHANGELOG
cat CHANGELOG.md | head -50
```

### Q: 认证失败怎么办？
1. 检查用户名和密码是否正确
2. 检查IP是否在白名单中
3. 查看日志文件：`tail -f ~/.local/share/go-proxy-server/proxy.log`
4. 尝试重新添加用户

### Q: 性能不佳怎么办？
1. 检查并发连接数：`netstat -an | grep :1080 | wc -l`
2. 检查CPU使用率：`top -p $(pgrep go-proxy-server)`
3. 检查内存使用：`ps aux | grep go-proxy-server`
4. 考虑调整常量配置（见性能调优部分）
5. 查看日志是否有错误

## 安全建议

### 1. 密码强度
- 最少8个字符
- 包含字母和数字
- 建议包含特殊字符

### 2. IP白名单
- 仅添加信任的IP地址
- 定期审查白名单
- 避免添加公网IP段

### 3. 访问控制
- Web管理界面仅监听localhost
- 使用防火墙限制代理端口访问
- 定期更新密码

### 4. 日志审计
- 定期检查认证失败日志
- 监控异常连接模式
- 保留日志用于审计

## 性能基准

### v1.4.0性能数据

**SOCKS5认证**:
- 缓存命中: <1ms
- 缓存未命中: 100-200ms (bcrypt)
- 缓存命中率: 通常>90%

**HTTP代理**:
- 连接复用: 80-90%
- 平均延迟: 18ms (vs 45ms旧版本)
- 吞吐量: 1200 req/s (vs 450 req/s旧版本)

**资源使用**:
- 内存: 基础20MB + 缓存2-3MB
- CPU: 空闲<1%, 高负载20-30%
- 文件描述符: 每连接2个

## 相关文档

- [详细修复说明](FIXES.md)
- [性能改进指南](docs/PERFORMANCE_IMPROVEMENTS.md)
- [升级指南](docs/UPGRADE_GUIDE.md)
- [完整CHANGELOG](CHANGELOG.md)
- [项目README](README.md)
- [架构说明](CLAUDE.md)

## 技术支持

如有问题，请：
1. 查看相关文档
2. 检查日志文件
3. 提交issue到GitHub仓库
4. 包含详细的错误信息和日志
