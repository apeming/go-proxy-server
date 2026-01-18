# React + Ant Design 前端迁移实施计划

## 一、项目概述

将当前内嵌在 Go 代码中的 HTML 字符串前端改造为使用 React + Ant Design 开发的现代化前端应用，最终编译打包成单个可执行文件。

## 二、技术栈选择

### 前端技术栈
- **框架**: React 18
- **UI 库**: Ant Design 5.x
- **语言**: TypeScript
- **构建工具**: Vite（快速、现代化）
- **HTTP 客户端**: axios
- **状态管理**: React Hooks（useState, useEffect）- 无需 Redux
- **路由**: 无需路由（单页应用）

### Go 集成方案
- **嵌入方式**: Go 1.16+ `embed` 包
- **静态文件服务**: `http.FileServer` + `http.FS`

## 三、前端项目结构设计

```
web-ui/                          # 新建前端项目目录
├── public/                      # 公共资源
│   └── favicon.ico
├── src/
│   ├── api/                     # API 调用封装
│   │   ├── index.ts            # axios 实例配置
│   │   ├── proxy.ts            # 代理相关 API
│   │   ├── user.ts             # 用户管理 API
│   │   ├── whitelist.ts        # 白名单 API
│   │   ├── system.ts           # 系统设置 API
│   │   └── timeout.ts          # 超时配置 API
│   ├── components/              # React 组件
│   │   ├── ProxyControl/       # 代理控制组件
│   │   │   ├── index.tsx
│   │   │   └── ProxyCard.tsx
│   │   ├── UserManagement/     # 用户管理组件
│   │   │   ├── index.tsx
│   │   │   ├── UserTable.tsx
│   │   │   └── AddUserForm.tsx
│   │   ├── WhitelistManagement/ # 白名单管理组件
│   │   │   ├── index.tsx
│   │   │   ├── WhitelistTable.tsx
│   │   │   └── AddIPForm.tsx
│   │   ├── SystemSettings/     # 系统设置组件
│   │   │   └── index.tsx
│   │   └── TimeoutConfig/      # 超时配置组件
│   │       └── index.tsx
│   ├── types/                   # TypeScript 类型定义
│   │   ├── proxy.ts
│   │   ├── user.ts
│   │   └── api.ts
│   ├── utils/                   # 工具函数
│   │   └── message.ts          # 消息提示封装
│   ├── App.tsx                  # 主应用组件
│   ├── App.css                  # 全局样式
│   ├── main.tsx                 # 应用入口
│   └── vite-env.d.ts           # Vite 类型声明
├── index.html                   # HTML 模板
├── package.json                 # 依赖配置
├── tsconfig.json               # TypeScript 配置
├── tsconfig.node.json          # Node TypeScript 配置
├── vite.config.ts              # Vite 配置
└── .gitignore                  # Git 忽略文件
```

## 四、Go 代码改造方案

### 4.1 新建文件

**internal/web/static.go** - 嵌入静态文件
```go
//go:build !dev
// +build !dev

package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetStaticFS returns the embedded static file system
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
```

**internal/web/static_dev.go** - 开发模式（可选）
```go
//go:build dev
// +build dev

package web

import (
	"io/fs"
	"os"
)

// GetStaticFS returns the local file system for development
func GetStaticFS() (fs.FS, error) {
	return os.DirFS("web-ui/dist"), nil
}
```

### 4.2 修改文件

**internal/web/server.go** - ���改静态文件服务
```go
// 修改 handleIndex 方法
func (wm *Manager) handleIndex(w http.ResponseWriter, r *http.Request) {
	// 如果请求的是 API 路径，返回 404
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	// 获取嵌入的静态文件系统
	staticFS, err := GetStaticFS()
	if err != nil {
		http.Error(w, "Failed to load static files", http.StatusInternalServerError)
		return
	}

	// 创建文件服务器
	fileServer := http.FileServer(http.FS(staticFS))

	// 如果请求的文件不存在，返回 index.html（支持 SPA 路由）
	if r.URL.Path != "/" {
		if _, err := staticFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			r.URL.Path = "/"
		}
	}

	fileServer.ServeHTTP(w, r)
}

// 修改 StartServer 方法中的路由注册
func (wm *Manager) StartServer() error {
	// API routes
	http.HandleFunc("/api/status", wm.handleStatus)
	http.HandleFunc("/api/users", wm.handleUsers)
	http.HandleFunc("/api/whitelist", wm.handleWhitelist)
	http.HandleFunc("/api/proxy/start", wm.handleProxyStart)
	http.HandleFunc("/api/proxy/stop", wm.handleProxyStop)
	http.HandleFunc("/api/proxy/config", wm.handleProxyConfig)
	http.HandleFunc("/api/system/settings", wm.handleSystemSettings)
	http.HandleFunc("/api/timeout", wm.handleTimeout)

	// Static files and SPA fallback (must be last)
	http.HandleFunc("/", wm.handleIndex)

	// ... rest of the code
}
```

### 4.3 删除文件

- **internal/web/html.go** - 删除整个文件（不再需要 HTML 字符串）

## 五、构建流程设计

### 5.1 Makefile 修改

```makefile
# 在现有 Makefile 中添加以下内容

# Frontend build directory
FRONTEND_DIR=web-ui
FRONTEND_DIST=$(FRONTEND_DIR)/dist

# Check if npm is installed
check-npm:
	@which npm > /dev/null || (echo "Error: npm is not installed. Please install Node.js and npm first." && exit 1)

# Install frontend dependencies
frontend-deps: check-npm
	@echo "Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && npm install

# Build frontend for production
frontend-build: check-npm
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm run build
	@echo "Frontend build complete: $(FRONTEND_DIST)"

# Clean frontend build
frontend-clean:
	@echo "Cleaning frontend build..."
	@rm -rf $(FRONTEND_DIST)
	@rm -rf $(FRONTEND_DIR)/node_modules

# Development: run frontend dev server
frontend-dev: check-npm
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && npm run dev

# 修改现有的 build 目标，添加 frontend-build 依赖
build: frontend-build
	@echo "Building for current platform..."
	@mkdir -p $(OUTPUT_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)"

# 修改其他 build 目标
build-linux: frontend-build
	@echo "Building for Linux..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64"

build-windows: frontend-build build-resources
	@echo "Building for Windows (console mode)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME).exe"

build-windows-gui: frontend-build build-resources
	@echo "Building for Windows (GUI mode - system tray)..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(WINDOWS_GUI_LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-gui.exe"

build-darwin: frontend-build
	@echo "Building for macOS..."
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64"

# 修改 clean 目标
clean: clean-resources frontend-clean
	@echo "Cleaning build artifacts..."
	rm -rf $(OUTPUT_DIR)
	@echo "Clean complete!"

# 添加新的帮助信息
help:
	@echo "Available targets:"
	@echo "  make build              - Build for current platform (includes frontend)"
	@echo "  make build-linux        - Build for Linux (includes frontend)"
	@echo "  make build-windows      - Build for Windows console mode (includes frontend)"
	@echo "  make build-windows-gui  - Build for Windows GUI/tray mode (includes frontend)"
	@echo "  make build-darwin       - Build for macOS (includes frontend)"
	@echo "  make build-all          - Build for all platforms (includes frontend)"
	@echo "  make frontend-build     - Build frontend only"
	@echo "  make frontend-dev       - Start frontend dev server"
	@echo "  make frontend-deps      - Install frontend dependencies"
	@echo "  make frontend-clean     - Clean frontend build and dependencies"
	@echo "  make clean              - Remove all build artifacts"
	@echo "  ... (other existing targets)"
```

### 5.2 前端构建配置

**web-ui/vite.config.ts**
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/',
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom'],
          'antd-vendor': ['antd'],
        },
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      },
    },
  },
})
```

## 六、开发模式和生产模式

### 6.1 开发模式

**前端开发**:
```bash
# 终端 1: 启动 Go 后端
make run

# 终端 2: 启动前端开发服务器
make frontend-dev
# 或
cd web-ui && npm run dev
```

前端开发服务器运行在 `http://localhost:3000`，通过 Vite 的 proxy 配置将 `/api` 请求代理到 `http://localhost:9090`。

### 6.2 生产模式

```bash
# 构建所有平台
make build-all

# 或单独构建
make build              # 当前平台
make build-windows-gui  # Windows GUI 版本
```

构建过程：
1. `make frontend-build` - 构建前端（生成 `web-ui/dist/`）
2. Go 编译时通过 `embed` 将 `dist/` 目录嵌入到二进制文件中
3. 生成单个可执行文件

## 七、详细实施步骤

### 步骤 1: 初始化前端项目

```bash
# 在项目根目录执行
npm create vite@latest web-ui -- --template react-ts
cd web-ui
npm install
npm install antd axios
npm install -D @types/node
```

### 步骤 2: 创建前端项目结构

创建以下目录和文件：
- `src/api/` - API 调用封装
- `src/components/` - React 组件
- `src/types/` - TypeScript 类型
- `src/utils/` - 工具函数

### 步骤 3: 实现 API 层

**src/api/index.ts** - axios 配置
```typescript
import axios from 'axios';
import { message } from 'antd';

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
});

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const errorMessage = error.response?.data?.error || error.message || '请求失败';
    message.error(errorMessage);
    return Promise.reject(error);
  }
);

export default api;
```

**src/api/proxy.ts** - 代理 API
```typescript
import api from './index';

export interface ProxyStatus {
  socks5: {
    running: boolean;
    port: number;
    bindListen: boolean;
    autoStart: boolean;
  };
  http: {
    running: boolean;
    port: number;
    bindListen: boolean;
    autoStart: boolean;
  };
}

export const getProxyStatus = () => api.get<ProxyStatus>('/status');

export const startProxy = (type: string, port: number, bindListen: boolean) =>
  api.post('/proxy/start', { type, port, bindListen });

export const stopProxy = (type: string) =>
  api.post('/proxy/stop', { type });

export const saveProxyConfig = (
  type: string,
  port: number,
  bindListen: boolean,
  autoStart: boolean
) => api.post('/proxy/config', { type, port, bindListen, autoStart });
```

类似地实现其他 API 文件（user.ts, whitelist.ts, system.ts, timeout.ts）。

### 步骤 4: 实现 React 组件

**src/App.tsx** - 主应用
```typescript
import React, { useEffect, useState } from 'react';
import { Layout, Tabs, message } from 'antd';
import ProxyControl from './components/ProxyControl';
import UserManagement from './components/UserManagement';
import WhitelistManagement from './components/WhitelistManagement';
import SystemSettings from './components/SystemSettings';
import TimeoutConfig from './components/TimeoutConfig';
import './App.css';

const { Header, Content } = Layout;

const App: React.FC = () => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ background: '#fff', textAlign: 'center' }}>
        <h1>Go Proxy Server - 管理后台</h1>
      </Header>
      <Content style={{ padding: '20px' }}>
        <Tabs
          defaultActiveKey="proxy"
          items={[
            {
              key: 'proxy',
              label: '代理控制',
              children: <ProxyControl />,
            },
            {
              key: 'users',
              label: '用户管理',
              children: <UserManagement />,
            },
            {
              key: 'whitelist',
              label: 'IP 白名单',
              children: <WhitelistManagement />,
            },
            {
              key: 'system',
              label: '系统设置',
              children: <SystemSettings />,
            },
            {
              key: 'timeout',
              label: '超时配置',
              children: <TimeoutConfig />,
            },
          ]}
        />
      </Content>
    </Layout>
  );
};

export default App;
```

实现各个组件（ProxyControl, UserManagement 等），使用 Ant Design 组件库。

### 步骤 5: 配置 Vite

创建 `web-ui/vite.config.ts`（见上文配置）。

### 步骤 6: 修改 Go 代码

1. 创建 `internal/web/static.go`（嵌入静态文件）
2. 修改 `internal/web/server.go`（静态文件服务）
3. 删除 `internal/web/html.go`

### 步骤 7: 修改 Makefile

添加前端构建相关的 target（见上文 Makefile 修改）。

### 步骤 8: 更新 .gitignore

```gitignore
# 在现有 .gitignore 中添加
web-ui/node_modules/
web-ui/dist/
web-ui/.vite/
```

### 步骤 9: 测试

**开发模式测试**:
```bash
# 终端 1
make run

# 终端 2
cd web-ui && npm run dev
```

访问 `http://localhost:3000` 测试前端功能。

**生产模式测试**:
```bash
make build
./bin/go-proxy-server
```

访问 `http://localhost:9090` 测试嵌入的前端。

### 步骤 10: 跨平台构建测试

```bash
make build-all
```

验证所有平台的二进制文件都包含了前端资源。

## 八、关键技术细节

### 8.1 Go Embed 工作原理

```go
//go:embed dist/*
var distFS embed.FS
```

- 在编译时将 `dist/` 目录下的所有文件嵌入到二进制文件中
- 使用 `fs.Sub()` 去除路径前缀
- 通过 `http.FS()` 转换为 HTTP 文件系统

### 8.2 SPA 路由处理

由于是单页应用，所有非 API 和非静态文件的请求都应该返回 `index.html`：

```go
if r.URL.Path != "/" {
    if _, err := staticFS.Open(strings.TrimPrefix(r.URL.Path, "/")); err != nil {
        r.URL.Path = "/"
    }
}
```

### 8.3 开发环境跨域处理

Vite 配置中的 proxy 设置：
```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:9090',
      changeOrigin: true,
    },
  },
}
```

### 8.4 构建优化

- 代码分割（React、Ant Design 分别打包）
- 资源压缩（Vite 自动处理）
- Tree shaking（移除未使用的代码）

## 九、验证清单

- [ ] 前端项目初始化成功
- [ ] 所有 API 调用正常工作
- [ ] 所有功能组件正常显示和交互
- [ ] 开发模式下前端热重载正常
- [ ] 生产构建成功生成 dist 目录
- [ ] Go 编译成功嵌入静态文件
- [ ] 单个可执行文件运行正常
- [ ] 所有平台构建成功（Linux/Windows/macOS）
- [ ] Windows GUI 模式托盘功能正常
- [ ] 所有原有功能保持完整

## 十、优势总结

1. **开发体验**: 热重载、TypeScript 类型检查、现代化开发工具
2. **可维护性**: 组件化、模块化、清晰的代码结构
3. **UI 一致性**: Ant Design 提供统一的设计语言
4. **性能**: Vite 快速构建、代码分割、按需加载
5. **扩展性**: 易于添加新功能和页面
6. **部署简单**: 仍然是单个可执行文件，无需额外依赖

## 十一、注意事项

1. **Node.js 版本**: 建议使用 Node.js 16+ 和 npm 8+
2. **构建顺序**: 必须先构建前端，再编译 Go
3. **路径问题**: 确保 embed 路径正确（`dist/*`）
4. **CORS**: 生产环境无需 CORS（同源），开发环境通过 Vite proxy 解决
5. **静态资源**: 所有静态资源（图片、字体等）都会被嵌入
6. **文件大小**: 嵌入前端后二进制文件会增大（约 1-2MB）

## 十二、后续优化建议

1. **国际化**: 使用 react-i18n 支持多语言
2. **主题定制**: 自定义 Ant Design 主题
3. **PWA**: 添加 Service Worker 支持离线访问
4. **单元测试**: 使用 Vitest 添加前端测试
5. **E2E 测试**: 使用 Playwright 添加端到端测试
6. **性能监控**: 添加前端性能监控
7. **错误追踪**: 集成错误追踪服务（如 Sentry）
