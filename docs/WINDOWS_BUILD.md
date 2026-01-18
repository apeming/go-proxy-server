# Windows构建指南 - 避免误报问题

## 问题说明

Go编译的网络代理程序在Windows上运行时，可能被Windows Defender或其他安全软件误报为恶意程序。这是因为：

1. **修改注册表自动启动**（最敏感） - Windows Defender对此高度警惕
2. **未签名的可执行文件** - Windows对未签名程序更加警惕
3. **网络监听行为** - 代理服务器的网络行为与某些恶意软件相似
4. **缺少版本信息** - 可执行文件缺少标准的Windows元数据

## 已实施的解决方案

本项目已经实施了以下改进来降低误报率：

### ✅ 改进1：使用启动文件夹代替注册表（v1.1+）

**重要变更：** 自动启动功能不再修改注册表，改为使用Windows启动文件夹。

- **旧方式（高风险）：** 修改注册表 `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`
- **新方式（安全）：** 在 `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup` 创建快捷方式
- **优势：** 不修改注册表，Windows Defender更宽容，用户可见且易于管理

### ✅ 改进2：添加Windows资源文件

集成了完整的版本信息和应用程序清单，让程序看起来更"正规"。

## 构建步骤

### 1. 安装编译工具

在编译Windows版本前，需要安装资源编译工具。推荐使用**goversioninfo**（纯Go方案，最稳定）：

**方法1：goversioninfo（推荐）**
```bash
go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
```

**方法2：go-winres（纯Go替代方案）**
```bash
go install github.com/tc-hib/go-winres@latest
```

**方法3：mingw-w64（传统方法，需要C编译器）**

Ubuntu/Debian/WSL:
```bash
sudo apt-get update
sudo apt-get install mingw-w64
```

macOS:
```bash
brew install mingw-w64
```

Windows (MSYS2):
```bash
pacman -S mingw-w64-x86_64-toolchain
```

**注意**：只需要安装其中一种工具即可。如果你已经安装了Go环境，推荐使用方法1（goversioninfo）。

### 2. 构建Windows版本

使用Makefile构建（会自动编译资源文件）：

```bash
# 构建Windows控制台版本（带版本信息）
make build-windows

# 构建Windows GUI版本（系统托盘，带版本信息）
make build-windows-gui

# 构建所有平台版本
make build-all
```

资源文件会自动编译并嵌入到可执行文件中。

### 3. 验证资源信息

构建完成后，在Windows上右键点击`.exe`文件 → "属性" → "详细信息"，应该能看到：

- 文件描述：SOCKS5 and HTTP Proxy Server
- 产品名称：Go Proxy Server
- 版本号：1.0.0.0
- 公司名称：Go Proxy Server Project
- 版权信息：Copyright (C) 2025

### 4. 手动构建资源文件（可选）

如果只想更新资源文件：

```bash
# 单独编译资源文件
make build-resources

# 清理资源文件
make clean-resources
```

## 进一步降低误报的方法

### 方法1：代码签名（最有效，但需要成本）

购买代码签名证书并签名可执行文件：

```bash
# 使用signtool签名（需要证书）
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com go-proxy-server.exe
```

**证书来源：**
- DigiCert、Sectigo等CA机构（约$100-500/年）
- 开源项目可申请免费的SignPath证书

### 方法2：向Microsoft提交误报申诉

如果Windows Defender误报：

1. 访问 https://www.microsoft.com/en-us/wdsi/filesubmission
2. 选择"Submit a file for malware analysis"
3. 上传`.exe`文件并说明是合法的代理服务器软件
4. 通常1-3个工作日会得到反馈

### 方法3：用户手动添加例外

指导用户将程序添加到Windows Defender排除列表：

1. 打开"Windows安全中心"
2. 点击"病毒和威胁防护"
3. 点击"管理设置"
4. 向下滚动到"排除项" → "添加或删除排除项"
5. 添加程序路径或整个文件夹

### 方法4：使用VirusTotal验证

构建完成后，可以上传到VirusTotal检查：

```bash
# 获取文件哈希
sha256sum bin/go-proxy-server.exe
```

访问 https://www.virustotal.com/ 上传文件检测。如果误报率较低（<5/70），说明构建成功。

## 项目中的改进

本项目已做的改进：

1. **版本信息** (`scripts/resource.rc`)
   - 添加产品名称、版本号、公司信息
   - 添加文件描述和版权信息

2. **应用程序清单** (`scripts/manifest.xml`)
   - 声明Windows兼容性（Win7-Win11）
   - 设置执行级别为`asInvoker`（不需要管理员权限）
   - 声明DPI感知能力

3. **自动化构建** (`Makefile`)
   - Windows构建自动包含资源文件
   - 简化构建流程

## 资源文件说明

### resource.rc
定义Windows资源，包括：
- 版本信息（VERSIONINFO）
- 产品信息（公司名称、描述、版权等）
- 应用程序清单引用

### manifest.xml
Windows应用程序清单，包括：
- 系统兼容性声明
- UAC执行级别（不需要管理员权限）
- DPI感知设置

### build_resources.sh
自动编译脚本：
- 检测`windres`工具
- 编译`.rc`文件为`.syso`文件
- `.syso`文件会被Go自动包含到Windows构建中

## 常见问题

### Q: 编译时提示找不到windres？
A: 安装mingw-w64工具链（参考上面的"安装编译工具"部分）

### Q: 添加资源后仍然被误报？
A:
1. 确认资源文件已正确嵌入（右键属性查看详细信息）
2. 考虑代码签名（最有效的方法）
3. 向Microsoft提交误报申诉
4. 提供用户手动添加排除的指导

### Q: 如何自定义版本号？
A: 编辑`scripts/resource.rc`文件，修改`FILEVERSION`和`ProductVersion`字段

### Q: 能否添加自定义图标？
A: 可以。准备一个`.ico`文件，取消注释`resource.rc`中的`IDI_ICON1 ICON "icon.ico"`行，并将图标文件放在`scripts/`目录

## 后续改进建议

1. **添加图标** - 创建专业的应用图标
2. **数字签名** - 获取代码签名证书
3. **SmartScreen信誉** - 随着下载量增加，SmartScreen会逐渐信任程序
4. **发布哈希值** - 在官方网站发布SHA256哈希值供用户验证

## 参考资料

- [Microsoft File Submission](https://www.microsoft.com/en-us/wdsi/filesubmission)
- [Go Windows Resources](https://pkg.go.dev/cmd/link)
- [Windows Application Manifest](https://docs.microsoft.com/en-us/windows/win32/sbscs/application-manifests)
