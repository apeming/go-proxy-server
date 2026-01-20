# Go Proxy Server

一个功能完善的代理服务器，支持 SOCKS5 和 HTTP 协议、用户名/密码认证、IP 白名单管理和数据库存储。

## 功能特性

### 核心功能
- 标准 SOCKS5 协议实现（支持 IPv4、IPv6、域名）
- HTTP/HTTPS 代理实现（支持 CONNECT 隧道和 Keep-Alive）
- 用户名/密码认证（SHA-256 加盐哈希）
- IP 白名单访问控制
- SQLite 数据库存储（纯 Go 实现，无需 CGO）
- 支持 bind-listen 模式（多出口 IP 路由）
- 支持同时运行 SOCKS5 和 HTTP 代理
- Web 管理界面（仅监听 localhost）
- Windows 系统托盘应用
- 命令行管理工具
- 自动重载配置（每 10 秒）

### 安全特性（v1.3.0+）
- ✅ **SSRF 防护**：自动阻止访问私有 IP 地址（RFC 1918、RFC 3927、RFC 4193）
- ✅ **时序攻击防护**：防止通过响应时间枚举用户名
- ✅ **IPv6 支持**：完整支持 IPv6 地址（SOCKS5 地址类型 0x04）
- ✅ **协议验证增强**：严格的 SOCKS5 协议合规性检查
- ✅ **并发安全**：线程安全的内存缓存和数据库操作
- ✅ **精确错误码**：SOCKS5 返回符合 RFC 1928 的错误码（0x02-0x07）

## 系统要求

- Go 1.19 或更高版本

## 项目结构

```
.
├── assets/                # 资源文件（图标等）
├── bin/                   # 构建输出目录
├── cmd/
│   └── server/           # 主程序入口
│       └── main.go
├── internal/             # 内部包（不对外暴露）
│   ├── auth/            # 认证和授权
│   ├── autostart/       # 自动启动管理（Windows）
│   ├── cache/           # 通用缓存基础设施
│   ├── config/          # 配置管理
│   ├── constants/       # 集中配置常量
│   ├── logger/          # 日志管理
│   ├── models/          # 数据模型（User、Whitelist、ProxyConfig）
│   ├── proxy/           # 代理实现（SOCKS5 和 HTTP）
│   │   ├── socks5.go   # SOCKS5 协议实现
│   │   ├── http.go     # HTTP/HTTPS 代理实现
│   │   ├── limiter.go  # 连接速率限制
│   │   └── copy.go     # 数据中继工具
│   ├── security/        # SSRF 和安全防护
│   ├── singleinstance/  # Windows 单实例检查
│   ├── tray/            # 系统托盘（Windows）
│   └── web/             # Web 管理服务器
│       ├── handlers.go  # HTTP API 处理器
│       ├── manager.go   # 代理服务器生命周期管理
│       ├── static.go    # 静态文件服务
│       └── dist/        # 前端构建产物（来自 web-ui/）
├── web-ui/               # 前端源代码（React + Vite + Ant Design）
│   ├── src/             # React 组件和应用逻辑
│   │   ├── api/        # API 客户端函数
│   │   ├── components/ # React 组件
│   │   ├── types/      # TypeScript 类型定义
│   │   └── utils/      # 工具函数
│   ├── public/          # 静态资源
│   ├── dist/            # 构建输出（复制到 internal/web/dist/）
│   ├── package.json     # Node.js 依赖
│   └── vite.config.ts   # Vite 构建配置
├── scripts/              # 构建和运行脚本
├── docs/                 # 文档
│   └── archive/         # 归档文档
├── Makefile             # 构建配置
├── go.mod               # Go 模块定义
├── CHANGELOG.md         # 详细更新日志
├── CLAUDE.md            # Claude Code 项目指南
└── README.md            # 项目说明

```

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd socks5-proxy
go mod download

# 使用 Makefile 编译（推荐）
make build                  # 编译当前平台 -> bin/go-proxy-server
make build-linux           # 编译 Linux 版本 -> bin/go-proxy-server-linux-amd64
make build-windows         # 编译 Windows 版本（控制台模式）-> bin/go-proxy-server.exe
make build-windows-gui     # 编译 Windows 版本（GUI/托盘模式）-> bin/go-proxy-server-gui.exe
make build-darwin          # 编译 macOS 版本 -> bin/go-proxy-server-darwin-amd64
make build-all             # 编译所有平台

# 或直接使用 go build
mkdir -p bin && go build -o bin/go-proxy-server ./cmd/server
```

**注意**: 所有二进制文件都输出到 `bin/` 目录，避免与 `cmd/server/` 源代码目录混淆。

## 配置

程序首次运行时会自动创建数据目录。

### 数据目录位置

数据文件（数据库和日志）自动存储在用户数据目录：

- **Windows**: `%APPDATA%\go-proxy-server\` (例如: `C:\Users\用户名\AppData\Roaming\go-proxy-server\`)
- **macOS**: `~/Library/Application Support/go-proxy-server/`
- **Linux/Unix**: `~/.local/share/go-proxy-server/`
- **支持 XDG 规范**: `$XDG_DATA_HOME/go-proxy-server/`

### 数据文件

程序会在数据目录中自动创建以下文件：
- `data.db`: SQLite 数据库，存储用户认证信息、IP 白名单和代理配置
- `app.log`: 日志文件（Windows GUI 模式）

**注意**: 所有数据（用户、密码、白名单、代理配置）均存储在 SQLite 数据库中，便于管理和备份。

## 使用方法

### 1. 启动代理服务器

#### 启动 SOCKS5 代理服务器

```bash
./bin/go-proxy-server socks -port <端口号> [-bind-listen]
```

参数说明：
- `-port`: 监听端口号（默认：1080）
- `-bind-listen`: 多出口 IP 模式。启用后，服务器使用客户端连接的本地 IP 作为出口 IP 连接目标服务器。适用于服务器有多个 IP 地址（如 IPa、IPb、IPc）的场景，不同客户端连接到不同的 IP，流量会从对应的 IP 出口

示例：
```bash
# 在 1080 端口启动代理服务器（普通模式）
./bin/go-proxy-server socks -port 1080

# 多出口 IP 模式：服务器绑定 0.0.0.0，有 3 个公网 IP
# 客户端连接 IPa:8888，流量从 IPa 出口
# 客户端连接 IPb:8888，流量从 IPb 出口
./bin/go-proxy-server socks -port 8888 -bind-listen
```

#### 启动 HTTP 代理服务器

```bash
./bin/go-proxy-server http -port <端口号> [-bind-listen]
```

参数说明：
- `-port`: 监听端口号（默认：8080）
- `-bind-listen`: 多出口 IP 模式（同上）

示例：
```bash
# 在 8080 端口启动 HTTP 代理服务器
./bin/go-proxy-server http -port 8080

# 多出口 IP 模式
./bin/go-proxy-server http -port 8080 -bind-listen
```

#### 同时启动 SOCKS5 和 HTTP 代理服务器

```bash
./bin/go-proxy-server both -socks-port <SOCKS5端口> -http-port <HTTP端口> [-bind-listen]
```

参数说明：
- `-socks-port`: SOCKS5 监听端口号（默认：1080）
- `-http-port`: HTTP 监听端口号（默认：8080）
- `-bind-listen`: 多出口 IP 模式（同时应用于两个代理）

示例：
```bash
# 同时启动 SOCKS5（1080）和 HTTP（8080）代理
./bin/go-proxy-server both -socks-port 1080 -http-port 8080

# 多出口 IP 模式
./bin/go-proxy-server both -socks-port 1080 -http-port 8080 -bind-listen
```

### 2. 用户管理

#### 添加用户

```bash
./bin/go-proxy-server adduser -username <用户名> -password <密码> [-ip <IP地址>]
```

参数说明：
- `-username`: 用户名（必需）
- `-password`: 密码（必需）
- `-ip`: 可选，指定用户的连接 IP

示例：
```bash
# 添加普通用户
./bin/go-proxy-server adduser -username alice -password secret123

# 添加指定 IP 的用户
./bin/go-proxy-server adduser -username bob -password pass456 -ip 192.168.1.100
```

#### 删除用户

```bash
./bin/go-proxy-server deluser -username <用户名>
```

参数说明：
- `-username`: 用户名（必需）
- `-ip`: 已废弃，为向后兼容保留但会被忽略

示例：
```bash
# 删除用户
./bin/go-proxy-server deluser -username alice

# 删除用户（旧版本带 -ip 参数仍可使用但会被忽略）
./bin/go-proxy-server deluser -username bob -ip 192.168.1.100
```

**注意**：从 v1.3.0 开始，用户名是全局唯一的，删除用户只需要提供用户名即可。

#### 列出所有用户

```bash
./bin/go-proxy-server listuser
```

### 3. IP 白名单管理

#### 添加 IP 到白名单

```bash
./bin/go-proxy-server addip -ip <IP地址>
```

示例：
```bash
./bin/go-proxy-server addip -ip 192.168.1.100
./bin/go-proxy-server addip -ip 10.0.0.50
```

**安全说明**：
- ⚠️ **默认情况下，所有连接都需要认证**（包括本地连接）
- 如需本地免认证访问，请手动添加：`./bin/go-proxy-server addip -ip 127.0.0.1`
- 白名单中的 IP 可以无需用户名/密码直接访问代理

#### 列出白名单

通过 Web 管理界面查看：访问 `http://localhost:9090` → IP 白名单管理

**注意**：`listip` 命令行工具尚未实现，请使用 Web 界面或 API (`GET /api/whitelist`)

#### 删除白名单 IP

通过 Web 管理界面删除：访问 `http://localhost:9090` → IP 白名单管理

**注意**：`delip` 命令行工具尚未实现，请使用 Web 界面或 API (`DELETE /api/whitelist`)

### 4. Web 管理界面

#### 启动 Web 管理服务

**方式一：直接双击运行（Windows 推荐）**

直接双击 `bin/go-proxy-server-gui.exe`，程序会自动启动 Web 管理界面（默认端口 9090）。

**方式二：命令行启动**

```bash
./bin/go-proxy-server web [-port <端口号>]
```

参数说明：
- `-port`: Web 管理界面端口号（默认：9090）

示例：
```bash
# 使用默认端口 9090
./bin/go-proxy-server web

# 使用自定义端口
./bin/go-proxy-server web -port 8888
```

启动后，在浏览器中访问 `http://localhost:9090` 即可打开管理界面。

**安全说明**：
- ✅ Web 管理界面**仅监听 localhost（127.0.0.1）**，不对外网暴露
- ✅ 只能从本机访问，确保管理界面的安全性
- ✅ 如需远程管理，建议使用 SSH 隧道或 VPN 连接

**注意**：如果不带任何参数运行程序，会默认启动 Web 管理界面。这使得 Windows 用户可以直接双击运行。

#### Web 管理界面功能

Web 管理界面采用现代化技术栈构建（React + TypeScript + Vite + Ant Design），提供以下功能：

- **代理服务控制**
  - 动态启动/停止 SOCKS5 代理服务器
  - 动态启动/停止 HTTP 代理服务器
  - 配置端口号和 bind-listen 模式
  - 实时查看服务运行状态

- **用户管理**
  - 添加新用户（用户名、密码、可选 IP）
  - 删除用户
  - 查看所有用户列表
  - 显示用户创建时间

- **IP 白名单管理**
  - 添加 IP 到白名单
  - 删除白名单 IP
  - 查看白名单列表
  - 实时生效，无需重启服务

- **系统配置**
  - 连接速率限制配置
  - 超时参数配置
  - 实时应用配置更改

**技术特性**：
- 响应式设计，支持桌面和移动设备
- 实时状态更新（每 5 秒轮询）
- 友好的错误提示和成功通知
- 单页应用（SPA）架构，流畅的用户体验

#### Windows 便携式应用（系统托盘）

**功能特性**：
- ✅ 双击运行，自动最小化到系统托盘
- ✅ 带图标，显示在任务栏通知区域
- ✅ 右键菜单：打开管理界面、退出程序
- ✅ 首次启动仅运行管理服务，不启动任何代理
- ✅ 通过浏览器配置所有参数
- ✅ 后台运行，不显示命令行窗口

**编译 Windows 版本**：

```bash
# 使用 Makefile（推荐）
make build-windows-gui     # 生成 bin/go-proxy-server-gui.exe（隐藏控制台）
make build-windows         # 生成 bin/go-proxy-server.exe（显示控制台，用于调试）

# 手动编译（需要设置国内镜像）
# 设置 Go 代理（国内用户推荐）
export GOPROXY=https://goproxy.cn,direct

# 隐藏控制台窗口（推荐用于发布）
mkdir -p bin && GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o bin/go-proxy-server-gui.exe ./cmd/server

# 显示控制台窗口（推荐用于调试）
mkdir -p bin && GOOS=windows GOARCH=amd64 go build -o bin/go-proxy-server.exe ./cmd/server
```

**重要说明**：
- ✅ 使用纯 Go 实现的 SQLite（modernc.org/sqlite），无需 CGO
- ✅ 支持 Windows/Linux/macOS 交叉编译
- ✅ 不需要安装 mingw-w64 或其他 C 编译器

**避免 Windows 误报提示**：

Windows Defender可能会将未签名的代理程序误报为恶意软件（如 Trojan:Win32/Bearfoos.A!ml）。本项目已实施以下改进来降低误报率：

**✅ 改进1：安全的自动启动方式（v1.1+）**
- 不再修改注册表（Windows Defender最敏感的行为）
- 改为在启动文件夹创建快捷方式：`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`
- 用户可见且易于管理

**✅ 改进2：添加版本信息和清单**

```bash
# 首次构建前，安装资源编译工具（选择其一）
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest  # 推荐
# 或
go install github.com/tc-hib/go-winres@latest  # 替代方案

# 然后正常构建，资源文件会自动嵌入
make build-windows-gui
```

构建完成后，右键点击exe → 属性 → 详细信息，应该能看到完整的版本信息。

**✅ 改进3：使用纯Go实现替代VBScript（v1.2+，重要！）**
- 移除了临时VBScript文件创建和执行（这是触发 Trojan:Win32/Bearfoos.A!ml 的主要原因）
- 使用 `github.com/go-ole/go-ole` 库直接调用Windows COM接口
- 不再创建临时脚本文件，不调用外部命令
- **这是降低误报率的关键改进**

**如果仍被误报：**
- 向Microsoft提交误报申诉：https://www.microsoft.com/en-us/wdsi/filesubmission
- 或考虑代码签名（最彻底的解决方案）

详细说明请查看：
- [Windows构建指南](docs/WINDOWS_BUILD.md)
- [误报问题解决方案](docs/ANTIVIRUS_FIX.md)
- [更新日志](CHANGELOG.md)

**故障排查**：

如果双击程序后没有反应，请按以下步骤排查：

**步骤 1：运行调试版本**
- 直接运行 `bin/go-proxy-server.exe`（控制台模式，显示日志输出）

**步骤 2：查看日志文件**
- 手动查看：`%APPDATA%\go-proxy-server\app.log`
- 完整路径示例：`C:\Users\你的用户名\AppData\Roaming\go-proxy-server\app.log`

**步骤 3：检查端口占用**
- 打开命令提示符，运行：`netstat -ano | findstr :9090`
- 如果端口被占用，结束占用进程或使用 `bin/go-proxy-server.exe web -port 其他端口`

**步骤 4：检查防火墙**
- 确保防火墙允许程序访问网络
- Windows 防火墙可能会弹出允许访问的提示

**常见问题**：
- **看不到系统托盘图标**：检查任务栏右下角，可能需要点击 "^" 展开隐藏图标
- **端口 9090 被占用**：运行 `bin/go-proxy-server.exe web -port 8888` 使用其他端口
- **程序闪退**：运行 `bin/go-proxy-server.exe`（控制台模式）查看错误信息

**使用方法**：
1. 从 `bin/` 目录复制 `go-proxy-server-gui.exe` 到目标位置并双击运行
2. 程序会自动启动并最小化到系统托盘（右下角）
3. 托盘图标会显示绿色圆点
4. 右键点击托盘图标：
   - **打开管理界面**：在浏览器中打开 http://localhost:9090
   - **退出**：关闭程序

**自定义图标**（可选）：
如果想使用自定义图标，可以使用工具如 [Resource Hacker](http://www.angusj.com/resourcehacker/) 或 [rcedit](https://github.com/electron/rcedit) 修改 exe 文件的图标资源。

## 认证机制

代理服务器（SOCKS5 和 HTTP）采用严格的认证策略，支持两种认证方式：

### 默认安全策略

⚠️ **重要**：默认情况下，**所有连接都需要认证**，包括本地连接（127.0.0.1）。

- ❌ 未配置任何用户或白名单时，所有连接都会被拒绝
- ✅ 必须显式配置白名单或添加用户才能使用代理
- 💡 如需本地免认证访问，需手动添加：`./bin/go-proxy-server addip -ip 127.0.0.1`

### 1. IP 白名单认证（优先级最高）

如果客户端 IP 在白名单中，无需用户名/密码即可访问。

### 2. 用户名/密码认证

如果客户端 IP 不在白名单中，则要求提供有效的用户名和密码。

**SOCKS5 认证流程：**
1. 客户端连接代理服务器
2. 服务器首先检查客户端 IP 是否在白名单中
3. 如果在白名单中，直接允许访问（无需认证）
4. 如果不在白名单中：
   - 如果客户端支持用户名/密码认证，要求进行认证
   - 如果客户端不支持认证，拒绝连接
5. 认证成功后建立连接

**HTTP 认证流程：**
1. 客户端连接代理服务器
2. 服务器首先检查客户端 IP 是否在白名单中
3. 如果在白名单中，直接允许访问（无需认证）
4. 如果不在白名单中：
   - 检查 Proxy-Authorization header（HTTP Basic 认证）
   - 如果认证失败或缺失，返回 407 Proxy Authentication Required
5. 认证成功后建立连接

## 客户端配置示例

### SOCKS5 代理

#### curl

```bash
# 使用用户名密码
curl -x socks5://alice:secret123@127.0.0.1:1080 https://example.com

# IP 在白名单中（无需认证）
curl -x socks5://127.0.0.1:1080 https://example.com
```

#### SSH

```bash
ssh -o ProxyCommand="nc -X 5 -x 127.0.0.1:1080 %h %p" user@remote-host
```

#### 浏览器配置

在浏览器代理设置中：
- 类型：SOCKS5
- 地址：127.0.0.1
- 端口：1080
- 用户名：alice
- 密码：secret123

### HTTP 代理

#### curl

```bash
# 使用用户名密码（HTTP）
curl -x http://alice:secret123@127.0.0.1:8080 http://example.com

# 使用用户名密码（HTTPS）
curl -x http://alice:secret123@127.0.0.1:8080 https://example.com

# IP 在白名单中（无需认证）
curl -x http://127.0.0.1:8080 https://example.com
```

#### 浏览器配置

在浏览器代理设置中：
- 类型：HTTP
- 地址：127.0.0.1
- 端口：8080
- 用户名：alice
- 密码：secret123

#### wget

```bash
# 使用 HTTP 代理
export http_proxy="http://alice:secret123@127.0.0.1:8080"
export https_proxy="http://alice:secret123@127.0.0.1:8080"
wget https://example.com
```

## 工作原理

### SOCKS5 代理

1. **启动服务器**：监听指定端口，等待客户端连接
2. **认证协商**：
   - 优先检查客户端 IP 是否在白名单中
   - 白名单中的 IP 直接放行
   - 非白名单 IP 要求用户名/密码认证
3. **访问控制**：
   - 白名单 IP 无需认证即可访问
   - 验证用户名和密码（如果需要）
   - 使用时序攻击防护确保安全
4. **协议验证**（v1.3.0+）：
   - 验证 SOCKS5 版本（0x05）
   - 验证 CMD 字段（仅支持 CONNECT/0x01）
   - 验证认证子协议版本（0x01）
   - 支持地址类型：IPv4（0x01）、域名（0x03）、IPv6（0x04）
5. **SSRF 防护**（v1.3.0+）：
   - 检查目标地址是否为私有 IP
   - 对域名进行 DNS 解析并验证所有结果
   - 阻止访问内网地址，返回错误码 0x02
6. **建立连接**：连接到目标主机（30 秒超时）
7. **数据转发**：在客户端和目标主机之间双向转发数据，正确处理连接关闭（TCP half-close）
8. **配置重载**：每 30 秒自动重新加载用户数据库和 IP 白名单（使用读写锁保证并发安全）

### HTTP 代理

1. **启动服务器**：监听指定端口，等待客户端连接
2. **读取 HTTP 请求**：解析客户端的 HTTP 请求
3. **认证检查**：
   - 优先检查客户端 IP 是否在白名单中
   - 白名单中的 IP 直接放行
   - 非白名单 IP 检查 Proxy-Authorization header（HTTP Basic 认证）
   - 认证失败返回 407 Proxy Authentication Required
   - 使用时序攻击防护确保安全
4. **SSRF 防护**（v1.3.0+）：
   - 检查目标地址是否为私有 IP
   - 对域名进行 DNS 解析并验证所有结果
   - 阻止访问内网地址，返回 403 Forbidden
5. **请求处理**：
   - **CONNECT 方法**（HTTPS）：建立透明隧道，双向转发数据（30 秒连接超时）
   - **其他方法**（HTTP）：转发请求到目标服务器，返回响应（支持 Keep-Alive）
6. **配置重载**：每 30 秒自动重新加载用户数据库和 IP 白名单（使用读写锁保证并发安全）

## 数据库结构

SQLite 数据库包含以下表：

### users 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| ip | TEXT | 用户的连接 IP（仅用于审计和日志记录） |
| username | TEXT | 用户名（全局唯一） |
| password | BLOB | SHA-256 加盐哈希的密码（格式：`$sha256$<salt>$<hash>`）|
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### whitelist 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| ip | TEXT | 白名单 IP 地址（唯一） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### proxy_configs 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键 |
| type | TEXT | 代理类型（socks5 或 http） |
| port | INTEGER | 监听端口 |
| bind_listen | BOOLEAN | 是否启用 bind-listen 模式 |
| auto_start | BOOLEAN | 是否自动启动 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

注意：
- **v1.3.0 重要变更**：`username` 字段现在是全局唯一的（不再是 `ip` + `username` 组合唯一）
- `ip` 字段仅用于审计和日志记录，不影响用户身份验证
- 密码使用 SHA-256 + 随机盐哈希存储（格式：`$sha256$<salt>$<hash>`）
- 所有配置数据（用户、白名单、代理配置）都存储在数据库中，便于管理和备份

## 安全特性

### 内置安全保护（v1.3.0+）

本项目实现了多层安全防护机制：

#### 1. SSRF（服务器端请求伪造）防护

自动阻止对私有 IP 地址的访问，防止代理被用于攻击内网服务：

- **阻止的 IPv4 范围**：
  - `127.0.0.0/8` - 本地回环地址
  - `10.0.0.0/8` - 私有网络 A 类
  - `172.16.0.0/12` - 私有网络 B 类
  - `192.168.0.0/16` - 私有网络 C 类
  - `169.254.0.0/16` - 链路本地地址

- **阻止的 IPv6 范围**：
  - `::1` - 本地回环地址
  - `fc00::/7` - 唯一本地地址
  - `fe80::/10` - 链路本地地址

- **防护策略**：
  - 检查直接 IP 连接
  - 对域名进行 DNS 解析并检查所有解析结果
  - 失败时关闭连接（fail-closed）
  - SOCKS5 返回错误码 0x02（连接不允许）
  - HTTP 返回 403 Forbidden

**注意**：如果需要代理访问内网服务，必须将客户端 IP 添加到白名单。

#### 2. 时序攻击防护

防止通过响应时间差异枚举有效用户名：

- 对不存在的用户名也执行 SHA-256 比较（使用预计算的虚拟哈希）
- 确保用户名存在和不存在时的响应时间一致
- 统一返回"无效凭据"错误，不区分用户名或密码错误

#### 3. IPv6 支持

SOCKS5 协议完整支持 IPv6：

- 支持地址类型 0x04（16 字节 IPv6 地址）
- 支持 IPv4、IPv6 和域名三种地址类型
- 适配现代双栈网络环境

#### 4. 协议验证增强

严格的 SOCKS5 协议合规性检查：

- 验证 SOCKS5 版本（必须为 0x05）
- 验证 CMD 字段（仅支持 CONNECT/0x01）
- 验证认证子协议版本（必须为 0x01）
- 拒绝不支持的命令（BIND、UDP ASSOCIATE）
- 返回精确的错误码（0x02-0x07）

#### 5. 数据库并发安全

防止并发场景下的竞态条件：

- 使用数据库唯一约束而非"检查后插入"模式
- 原子性操作防止重复条目
- 线程安全的内存缓存（sync.RWMutex）

## 安全建议

1. **密码安全**
   - 使用强密码（至少 12 字符，包含大小写字母、数字和特殊字符）
   - 定期更新用户密码
   - 密码已使用 SHA-256 + 随机盐哈希存储

2. **访问控制**
   - 限制 IP 白名单范围，仅添加可信 IP
   - 定期审查白名单和用户列表
   - 默认拒绝所有连接，显式配置允许访问

3. **网络安全**
   - 在生产环境中使用防火墙限制访问
   - Web 管理界面仅监听 localhost（127.0.0.1）
   - 考虑使用 TLS/SSL 加密传输（可通过反向代理实现）

4. **SSRF 防护**
   - 内置 SSRF 防护默认启用
   - 如需访问内网服务，必须将客户端 IP 加入白名单
   - 评估安全风险后再允许内网访问

5. **监控和审计**
   - 定期检查日志文件（`%APPDATA%\go-proxy-server\app.log`）
   - 监控异常连接和认证失败
   - 记录所有用户操作和 IP 访问

6. **系统安全**
   - 保持操作系统和依赖库更新
   - 使用最小权限原则运行程序
   - 定期备份数据库文件

## 故障排除

### 连接被拒绝

检查：
- IP 是否在白名单中
- 用户名和密码是否正确
- 防火墙设置
- 服务器是否正常运行

### 认证失败

- 确认用户名和密码正确
- 检查数据库中是否存在该用户
- 使用 `listuser` 命令查看所有用户

### 无法连接到目标主机

- 检查网络连接
- 确认目标主机地址和端口正确
- 检查是否有防火墙阻止

## 开发依赖

```
- github.com/glebarez/sqlite - 纯 Go 实现的 SQLite 驱动（无需 CGO）
- github.com/getlantern/systray - 系统托盘图标（Windows）
- golang.org/x/crypto - 密码加密
- gorm.io/gorm - ORM 框架
- modernc.org/sqlite - SQLite 底层实现
```

**注意**：本项目使用纯 Go 实现的 SQLite，不需要 CGO，可以轻松进行跨平台编译。

## 许可证

请查看 LICENSE 文件获取许可证信息。

## 贡献

欢迎提交 Issue 和 Pull Request。

## 更新日志

### 最新版本：v1.3.0 (2026-01-17)

**重要安全更新和功能增强**：

- ✅ **P0 级别修复**：SOCKS5 协议验证、HTTP 连接管理、认证模块初始化
- ✅ **安全增强**：SSRF 防护、时序攻击防护、IPv6 支持
- ✅ **API 变更**：用户名全局唯一，删除用户不再需要 IP 参数
- ✅ **协议改进**：SOCKS5 精确错误码、认证版本验证
- ✅ **并发安全**：数据库竞态条件修复

详细更新内容请查看 [CHANGELOG.md](CHANGELOG.md)

### 历史版本

- **v1.2.1**: 连接超时修复、HTTP 请求体传输修复、凭据存储逻辑优化
- **v1.2.0**: 使用纯 Go 实现替代 VBScript，降低杀毒软件误报率
- **v1.1.0**: 改用启动文件夹快捷方式替代注册表自动启动
- **v1.0.0**: 初始版本，支持 SOCKS5/HTTP 代理、认证、Web 管理界面
