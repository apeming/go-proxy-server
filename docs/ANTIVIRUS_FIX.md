# Windows误报问题 - 快速解决指南

## 问题现象

编译的Windows程序被Windows Defender报告为病毒或恶意软件，特别是在重启系统后被删除。

## 根本原因

Windows Defender对以下行为组合非常敏感：
1. ❌ **修改注册表自动启动项**（最敏感）
2. ❌ 网络监听和代理功能
3. ❌ 未签名的可执行文件

## 已实施的解决方案

本项目已经实施了以下改进来降低误报率：

### ✅ 改进1：使用启动文件夹代替注册表（v1.1+）

**旧方式（高风险）：**
- 修改注册表：`HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
- 被Windows Defender视为典型恶意软件行为

**新方式（安全）：**
- 在启动文件夹创建快捷方式：`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`
- 不修改注册表，Windows Defender更宽容
- 用户可见且易于管理

### ✅ 改进2：添加Windows资源文件

- 完整的版本信息（产品名称、版本号、公司信息）
- 应用程序清单（权限声明、兼容性）
- 让程序看起来更"正规"

### ✅ 改进3：使用纯Go实现替代VBScript（v1.2+，重要！）

**旧方式（极高风险）：**
- 创建临时VBScript文件（.vbs）
- 使用 `cscript` 执行脚本
- 这是典型的恶意软件行为模式，触发 **Trojan:Win32/Bearfoos.A!ml** 等误报

**新方式（安全）：**
- 使用 `github.com/go-ole/go-ole` 纯Go库
- 直接调用Windows COM接口创建快捷方式
- 不创建临时脚本文件，不调用外部命令
- **显著降低误报率**

## 解决方案（已集成到项目）

本项目已经集成了完整的Windows资源文件支持，可以显著降低误报率。

### 第一步：安装资源编译工具

**选择以下任一方法安装**（推荐方法1）：

```bash
# 方法1：goversioninfo（推荐，纯Go实现）
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest

# 方法2：go-winres（替代方案，纯Go实现）
go install github.com/tc-hib/go-winres@latest

# 方法3：mingw-w64（传统方法，需要C编译器）
# Ubuntu/Debian: sudo apt-get install mingw-w64
# macOS: brew install mingw-w64
```

**如果网络问题无法安装Go工具**，请先配置代理：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 第二步：编译Windows版本

安装工具后，正常使用Make命令即可：

```bash
# 构建系统托盘版本（推荐用于发布）
make build-windows-gui

# 构建控制台版本（推荐用于调试）
make build-windows
```

资源文件会自动编译并嵌入到exe中。

### 第三步：验证

在Windows上，右键点击生成的exe文件 → "属性" → "详细信息"，确认能看到：

- ✅ 文件描述：SOCKS5 and HTTP Proxy Server
- ✅ 产品名称：Go Proxy Server
- ✅ 版本号：1.0.0.0
- ✅ 公司名称：Go Proxy Server Project
- ✅ 版权信息：Copyright (C) 2025

如果能看到这些信息，说明资源文件已正确嵌入，误报率会大幅降低。

## 如果仍然被误报

### 方案1：代码签名（最有效）

购买代码签名证书（约$100-500/年）并签名exe文件。这是最有效的方法。

### 方案2：提交误报申诉

访问 https://www.microsoft.com/en-us/wdsi/filesubmission 提交文件分析申请。

### 方案3：用户手动添加排除

指导用户在Windows安全中心添加排除项：
1. 打开"Windows安全中心"
2. 点击"病毒和威胁防护" → "管理设置"
3. 向下滚动到"排除项" → "添加或删除排除项"
4. 添加程序路径

### 方案4：VirusTotal验证

上传到 https://www.virustotal.com/ 检测。如果误报率<5/70，说明程序是安全的。

## 项目改进清单

已完成的改进：

- ✅ **使用启动文件夹代替注册表**（重要！）
  - 不再修改 `HKEY_CURRENT_USER\...\Run` 注册表
  - 改为在 `%APPDATA%\...\Startup` 创建快捷方式
  - 显著降低Windows Defender敏感度
- ✅ **使用纯Go实现替代VBScript**（v1.2+，最重要！）
  - 移除临时VBScript文件创建和执行
  - 使用 `go-ole` 库直接调用COM接口
  - 不再触发 Trojan:Win32/Bearfoos.A!ml 误报
  - 这是降低误报率的**关键改进**
- ✅ 添加Windows版本信息（versioninfo.json）
- ✅ 添加应用程序清单（scripts/manifest.xml）
- ✅ 添加资源文件定义（scripts/resource.rc）
- ✅ 创建自动化构建脚本（scripts/build_resources.sh）
- ✅ 集成到Makefile（make build-windows自动嵌入）
- ✅ 支持3种编译工具（goversioninfo/go-winres/windres）
- ✅ 更新文档说明（README.md和docs/WINDOWS_BUILD.md）

## 常见问题

### Q: 构建时提示找不到资源编译器？
**A**: 按照上面"第一步"安装任一工具即可。

### Q: 网络超时无法安装go工具？
**A**: 配置国内代理：`go env -w GOPROXY=https://goproxy.cn,direct`

### Q: 可以跳过资源文件直接构建吗？
**A**: 可以，但强烈不推荐。没有资源文件的exe误报率很高。

### Q: 如何自定义版本号？
**A**: 编辑根目录的`versioninfo.json`文件。

## 详细文档

- [完整Windows构建指南](docs/WINDOWS_BUILD.md)
- [脚本使用说明](scripts/README.md)

## 文件说明

```
项目根目录/
├── versioninfo.json           # 版本信息配置（goversioninfo）
├── scripts/
│   ├── resource.rc            # Windows资源定义（windres）
│   ├── manifest.xml           # 应用程序清单
│   ├── build_resources.sh     # 自动构建脚本（支持3种工具）
│   └── README.md              # 脚本说明
└── docs/
    └── WINDOWS_BUILD.md       # 详细构建指南
```

## 技术原理

Windows会检查可执行文件的：
1. 数字签名（最重要）
2. 版本信息资源
3. 应用程序清单
4. 行为特征（网络监听、脚本执行等）
5. SmartScreen信誉

我们通过以下方式提高可信度：
- ✅ 添加版本信息和清单（2和3）
- ✅ 移除VBScript执行行为（4）- **关键改进**
- ⏳ 代码签名（1）- 需要购买证书
- ⏳ 建立SmartScreen信誉（5）- 需要时间积累

**v1.2版本的关键改进**：移除VBScript执行是降低误报率的最重要改进，因为这是Windows Defender最敏感的行为模式之一。
