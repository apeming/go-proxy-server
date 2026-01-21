# Windows 10 代理服务器性能优化指南

本指南专门针对在 Windows 10 上运行的 go-proxy-server 进行性能排查和优化。

## 目录
1. [资源问题排查](#资源问题排查)
2. [代理服务器配置优化](#代理服务器配置优化)
3. [Windows 系统优化](#windows-系统优化)
4. [常见问题解决](#常见问题解决)

---

## 资源问题排查

### 快速排查（使用监控脚本）

```powershell
# 在项目根目录运行
.\scripts\monitor.ps1 -Port 12223 -Interval 5
```

监控脚本会实时显示：
- 进程资源使用（CPU、内存、句柄、线程）
- 网络连接统计（ESTABLISHED、TIME_WAIT、CLOSE_WAIT）
- 系统资源使用率
- 动态端口使用情况
- 自动告警提示

### 手动排查方法

#### 1. 查看进程资源使用

**任务管理器**（最简单）：
1. 按 `Ctrl + Shift + Esc` 打开任务管理器
2. 切换到"详细信息"标签
3. 找到 `go-proxy-server.exe` 或 `go-proxy-server-gui.exe`
4. 查看 CPU、内存、句柄数

**PowerShell**（更详细）：
```powershell
# 查看基本信息
Get-Process go-proxy-server* | Format-Table Name, CPU, WorkingSet, Handles, Threads -AutoSize

# 持续监控（每5秒刷新）
while ($true) {
    Clear-Host
    Get-Process go-proxy-server* | Format-Table Name, CPU, @{Name="Memory(MB)";Expression={[math]::Round($_.WorkingSet64/1MB,2)}}, Handles, Threads -AutoSize
    Start-Sleep -Seconds 5
}
```

#### 2. 查看网络连接

```powershell
# 查看指定端口的连接数
netstat -an | findstr :12223 | find /c /v ""

# 查看各状态的连接数
netstat -an | findstr :12223 | findstr ESTABLISHED | find /c /v ""
netstat -an | findstr :12223 | findstr TIME_WAIT | find /c /v ""
netstat -an | findstr :12223 | findstr CLOSE_WAIT | find /c /v ""

# 查看详细连接信息
netstat -ano | findstr :12223
```

#### 3. 使用资源监视器

```powershell
# 打开资源监视器
resmon
```

在资源监视器中：
- **CPU** 标签：查看进程 CPU 使用率
- **内存** 标签：查看内存使用详情
- **网络** 标签：查看网络连接和流量
- **磁盘** 标签：查看磁盘 I/O

#### 4. 使用性能监视器

```powershell
# 打开性能监视器
perfmon
```

添加以下计数器：
- `Process(go-proxy-server*)\% Processor Time`
- `Process(go-proxy-server*)\Working Set`
- `Process(go-proxy-server*)\Handle Count`
- `Process(go-proxy-server*)\Thread Count`
- `TCPv4\Connections Established`

---

## 代理服务器配置优化

### 1. 通过 Web 管理界面调整

1. 打开浏览器访问: `http://localhost:9090`
2. 进入"系统配置"页面
3. 调整以下参数：

#### 连接限制配置

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| `MaxConcurrentConnections` | 10000 | 全局最大并发连接数 |
| `MaxConcurrentConnectionsPerIP` | 500 | 单IP最大并发连接数 |

**调整建议**：
- 如果出现 "connection reset" 错误，逐步增加这两个值
- 根据服务器内存调整：每1000个连接约需 100-200MB 内存
- 监控实际连接数，设置为峰值的 1.5-2 倍

#### 超时配置

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| `ConnectTimeout` | 30s | 连接建立超时 |
| `IdleReadTimeout` | 300s | 空闲读超时（5分钟） |
| `IdleWriteTimeout` | 300s | 空闲写超时（5分钟） |
| `MaxConnectionAge` | 3600s | 最大连接存活时间（1小时） |
| `CleanupTimeout` | 5s | 连接清理超时 |

**调整建议**：
- 如果客户端连接经常超时，增加 `IdleReadTimeout` 和 `IdleWriteTimeout`
- 如果内存持续增长，减少 `MaxConnectionAge`
- 长连接场景建议增加超时时间，短连接场景建议减少

### 2. 通过数据库直接修改（高级）

```powershell
# 需要安装 sqlite3 命令行工具
# 下载地址: https://www.sqlite.org/download.html

# 数据库路径
$dbPath = "$env:APPDATA\go-proxy-server\data.db"

# 查看当前配置
sqlite3 $dbPath "SELECT * FROM system_configs;"

# 修改配置示例
sqlite3 $dbPath "UPDATE system_configs SET value='10000' WHERE key='MaxConcurrentConnections';"
sqlite3 $dbPath "UPDATE system_configs SET value='500' WHERE key='MaxConcurrentConnectionsPerIP';"
sqlite3 $dbPath "UPDATE system_configs SET value='300s' WHERE key='IdleReadTimeout';"

# 修改后需要重启代理服务器
```

---

## Windows 系统优化

### 自动优化（推荐）

```powershell
# 以管理员权限运行 PowerShell，然后执行
.\scripts\optimize.ps1
```

### 手动优化步骤

#### 1. 扩展动态端口范围

**问题**：Windows 默认动态端口范围较小（49152-65535，约16000个端口），高并发时容易耗尽。

**解决方法**（需要管理员权限）：
```powershell
# 查看当前配置
netsh int ipv4 show dynamicport tcp

# 扩展端口范围（10000-65535，约55000个端口）
netsh int ipv4 set dynamicport tcp start=10000 num=55535

# 验证修改
netsh int ipv4 show dynamicport tcp
```

#### 2. 调整 TIME_WAIT 超时时间

**问题**：Windows 默认 TIME_WAIT 超时为 120 秒，导致大量端口处于 TIME_WAIT 状态。

**解决方法**（需要管理员权限）：
```powershell
# 将 TIME_WAIT 超时调整为 30 秒
reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v TcpTimedWaitDelay /t REG_DWORD /d 30 /f

# 注意：需要重启系统生效
```

#### 3. 启用 TCP 时间戳

**解决方法**（需要管理员权限）：
```powershell
# 启用 TCP 时间戳
netsh int tcp set global timestamps=enabled

# 查看当前 TCP 全局设置
netsh int tcp show global
```

#### 4. 调整 TCP 连接参数

```powershell
# 启用 TCP 窗口自动调优
netsh int tcp set global autotuninglevel=normal

# 启用 TCP Chimney Offload（如果网卡支持）
netsh int tcp set global chimney=enabled

# 启用接收端合并（RSC）
netsh int tcp set global rsc=enabled
```

#### 5. 增加用户进程限制

Windows 对单个进程的资源限制较宽松，但可以通过注册表调整：

```powershell
# 增加最大用户端口数（需要管理员权限）
reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v MaxUserPort /t REG_DWORD /d 65534 /f

# 重启系统生效
```

#### 6. 防火墙配置

```powershell
# 添加防火墙规则（需要管理员权限）
New-NetFirewallRule -DisplayName "Go Proxy Server" -Direction Inbound -Protocol TCP -LocalPort 12223 -Action Allow

# 查看规则
Get-NetFirewallRule -DisplayName "Go Proxy Server"
```

#### 7. 禁用不必要的服务

为了释放系统资源，可以禁用一些不必要的 Windows 服务：
- Windows Search（如果不需要文件索引）
- Superfetch/SysMain（如果内存充足）
- Windows Update（在测试期间临时禁用）

---

## 常见问题解决

### 问题 1: "connection reset by peer" 错误

**症状**：压测时出现大量连接重置错误

**可能原因**：
1. 连接数限制达到上限
2. 系统资源不足（内存、句柄、端口）
3. 超时配置过短

**解决步骤**：

1. **检查连接数限制**
   ```powershell
   # 运行监控脚本查看当前连接数
   .\scripts\monitor.ps1 -Port 12223
   ```

2. **增加连接限制**
   - 通过 Web 界面增加 `MaxConcurrentConnections` 和 `MaxConcurrentConnectionsPerIP`
   - 建议从当前值的 2 倍开始尝试

3. **检查系统资源**
   - 内存使用是否超过 80%
   - 句柄数是否超过 10000
   - TIME_WAIT 连接是否超过 5000

4. **应用系统优化**
   ```powershell
   # 以管理员权限运行
   .\scripts\optimize.ps1
   ```

5. **重启代理服务器**
   - 关闭当前运行的代理服务器
   - 重新启动

6. **逐步增加压测并发数**
   ```bash
   # 从小并发开始
   ./bin/benchmark -host 服务器IP -port 12223 -c 50 -n 1000
   ./bin/benchmark -host 服务器IP -port 12223 -c 100 -n 1000
   ./bin/benchmark -host 服务器IP -port 12223 -c 200 -n 1000
   ```

### 问题 2: 内存持续增长

**症状**：代理服务器内存使用持续增长，不释放

**可能原因**：
1. 连接未正确关闭（连接泄漏）
2. `MaxConnectionAge` 设置过大
3. 大量 CLOSE_WAIT 连接

**解决步骤**：

1. **检查 CLOSE_WAIT 连接**
   ```powershell
   netstat -an | findstr :12223 | findstr CLOSE_WAIT
   ```

2. **减少连接存活时间**
   - 将 `MaxConnectionAge` 从 3600s 减少到 1800s（30分钟）

3. **重启代理服务器**
   - 定期重启可以释放累积的资源

### 问题 3: CPU 使用率过高

**症状**：代理服务器 CPU 使用率持续在 80% 以上

**可能原因**：
1. 并发连接数过多
2. 频繁的连接建立和断开
3. 加密/解密操作过多（HTTPS）

**解决步骤**：

1. **限制并发连接数**
   - 减少 `MaxConcurrentConnections`

2. **增加连接复用**
   - 增加 `MaxConnectionAge` 以减少连接重建

3. **使用多个代理实例**
   - 在不同端口运行多个代理实例
   - 使用负载均衡分发流量

### 问题 4: 端口耗尽

**症状**：大量 TIME_WAIT 连接，新连接无法建立

**解决步骤**：

1. **扩展动态端口范围**
   ```powershell
   netsh int ipv4 set dynamicport tcp start=10000 num=55535
   ```

2. **减少 TIME_WAIT 超时**
   ```powershell
   reg add HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters /v TcpTimedWaitDelay /t REG_DWORD /d 30 /f
   ```

3. **重启系统使配置生效**

---

## 性能测试建议

### 1. 基准测试

```bash
# 测试基准性能（小并发）
./bin/benchmark -host 服务器IP -port 12223 -c 10 -n 1000
```

### 2. 逐步增加负载

```bash
# 逐步增加并发数，找到性能瓶颈
./bin/benchmark -host 服务器IP -port 12223 -c 50 -n 5000
./bin/benchmark -host 服务器IP -port 12223 -c 100 -n 5000
./bin/benchmark -host 服务器IP -port 12223 -c 200 -n 5000
./bin/benchmark -host 服务器IP -port 12223 -c 500 -n 5000
```

### 3. 持续负载测试

```bash
# 持续测试 5 分钟，观察稳定性
./bin/benchmark -host 服务器IP -port 12223 -c 100 -d 5m
```

### 4. 监控资源使用

在压测期间，同时运行监控脚本：
```powershell
.\scripts\monitor.ps1 -Port 12223 -Interval 5
```

---

## 优化效果验证

### 优化前后对比

记录以下指标：
- 最大并发连接数
- 请求成功率
- 平均响应时间
- CPU 使用率
- 内存使用量
- TIME_WAIT 连接数

### 预期优化效果

| 指标 | 优化前 | 优化后 |
|------|--------|--------|
| 最大并发连接数 | ~1000 | ~5000-10000 |
| TIME_WAIT 连接数 | >10000 | <5000 |
| 端口耗尽风险 | 高 | 低 |
| 连接成功率 | 80-90% | >95% |

---

## 快速参考

### 常用命令

```powershell
# 监控资源
.\scripts\monitor.ps1 -Port 12223

# 应用优化
.\scripts\optimize.ps1

# 查看连接数
netstat -an | findstr :12223 | find /c /v ""

# 查看进程资源
Get-Process go-proxy-server* | Format-Table Name, CPU, WorkingSet, Handles

# 重启代理服务器（如果是系统托盘模式）
# 右键托盘图标 -> 退出 -> 重新启动程序
```

### 推荐配置（高性能场景）

```
MaxConcurrentConnections: 10000
MaxConcurrentConnectionsPerIP: 500
ConnectTimeout: 30s
IdleReadTimeout: 300s
IdleWriteTimeout: 300s
MaxConnectionAge: 1800s
CleanupTimeout: 5s
```

### 推荐配置（稳定性优先）

```
MaxConcurrentConnections: 5000
MaxConcurrentConnectionsPerIP: 200
ConnectTimeout: 30s
IdleReadTimeout: 180s
IdleWriteTimeout: 180s
MaxConnectionAge: 900s
CleanupTimeout: 5s
```

---

## 技术支持

如果遇到问题，请提供以下信息：
1. 监控脚本的输出截图
2. 压测工具的错误信息
3. 代理服务器的日志文件（`%APPDATA%\go-proxy-server\app.log`）
4. Windows 版本和系统配置
