# 自定义 Windows 图标

## 方法一：使用 rcedit（推荐）

1. 下载 [rcedit](https://github.com/electron/rcedit/releases)

2. 准备你的图标文件（`.ico` 格式，建议包含多种尺寸：16x16, 32x32, 48x48, 256x256）

3. 运行命令：
```bash
rcedit go-proxy-server.exe --set-icon your-icon.ico
```

## 方法二：使用 Resource Hacker

1. 下载 [Resource Hacker](http://www.angusj.com/resourcehacker/)

2. 打开 `go-proxy-server.exe`

3. 找到 Icon Group → 替换为你的图标文件

4. 保存并退出

## 方法三：编译时嵌入图标（需要 rsrc 工具）

1. 安装 rsrc 工具：
```bash
go install github.com/akavel/rsrc@latest
```

2. 创建 manifest 文件 `rsrc.json`：
```json
{
  "IconPath": "icon.ico",
  "Manifest": ""
}
```

3. 生成资源文件：
```bash
rsrc -ico icon.ico -o rsrc.syso
```

4. 重新编译（rsrc.syso 会自动包含）：
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-H=windowsgui" -o go-proxy-server.exe
```

## 图标文件要求

- 格式：`.ico`
- 建议包含多种尺寸：
  - 16x16 (系统托盘)
  - 32x32 (小图标)
  - 48x48 (中等图标)
  - 256x256 (大图标)
- 支持透明背景

## 在线制作图标

推荐工具：
- [ICO Convert](https://icoconvert.com/) - PNG 转 ICO
- [Favicon.io](https://favicon.io/) - 在线生成图标
- [RealFaviconGenerator](https://realfavicongenerator.net/)
