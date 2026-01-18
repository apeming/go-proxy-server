package web

const IndexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Proxy Server - 管理后台</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: #333;
            margin-bottom: 30px;
            text-align: center;
        }
        .section {
            background: white;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .section h2 {
            color: #555;
            margin-bottom: 15px;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        .proxy-control {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
        }
        .proxy-card {
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 15px;
        }
        .proxy-card h3 {
            color: #666;
            margin-bottom: 10px;
        }
        .status {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
        }
        .status.running {
            background: #4CAF50;
            color: white;
        }
        .status.stopped {
            background: #f44336;
            color: white;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #666;
            font-weight: 500;
        }
        input[type="text"],
        input[type="number"],
        input[type="password"] {
            width: 100%;
            padding: 8px 12px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        input[type="checkbox"] {
            margin-right: 5px;
        }
        button {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: background 0.3s;
        }
        .btn-primary {
            background: #4CAF50;
            color: white;
        }
        .btn-primary:hover {
            background: #45a049;
        }
        .btn-danger {
            background: #f44336;
            color: white;
        }
        .btn-danger:hover {
            background: #da190b;
        }
        .btn-secondary {
            background: #2196F3;
            color: white;
        }
        .btn-secondary:hover {
            background: #0b7dda;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 15px;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background: #f5f5f5;
            font-weight: 600;
            color: #666;
        }
        .actions {
            display: flex;
            gap: 10px;
        }
        .message {
            padding: 10px 15px;
            border-radius: 4px;
            margin-bottom: 15px;
            display: none;
        }
        .message.success {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .message.error {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Go Proxy Server 管理后台</h1>

        <div id="message" class="message"></div>

        <!-- 系统设置 -->
        <div class="section">
            <h2>系统设置</h2>
            <div class="form-group">
                <label>
                    <input type="checkbox" id="system-autostart">
                    Windows 开机自启（程序随系统启动）
                </label>
                <p style="margin: 5px 0 0 24px; font-size: 0.9em; color: #666;">
                    注意：此功能仅在 Windows 平台有效。启用后，程序会在 Windows 启动时自动运行。
                </p>
            </div>
        </div>

        <!-- 超时配置 -->
        <div class="section">
            <h2>超时配置</h2>
            <p style="margin-bottom: 15px; color: #666; font-size: 0.9em;">
                配置代理服务器的超时时间（单位：秒）。修改后立即生效，无需重启代理服务。
            </p>
            <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px;">
                <div class="form-group">
                    <label>连接超时</label>
                    <input type="number" id="timeout-connect" value="30" min="1" max="300">
                    <p style="margin-top: 5px; font-size: 0.85em; color: #888;">
                        建立TCP连接的最大时间（1-300秒）
                    </p>
                </div>
                <div class="form-group">
                    <label>空闲读取超时</label>
                    <input type="number" id="timeout-idle-read" value="300" min="1" max="3600">
                    <p style="margin-top: 5px; font-size: 0.85em; color: #888;">
                        读取数据时的空闲超时（1-3600秒）
                    </p>
                </div>
                <div class="form-group">
                    <label>空闲写入超时</label>
                    <input type="number" id="timeout-idle-write" value="120" min="1" max="3600">
                    <p style="margin-top: 5px; font-size: 0.85em; color: #888;">
                        写入数据时的空闲超时（1-3600秒）
                    </p>
                </div>
            </div>
            <button onclick="saveTimeout()" class="btn">保存超时配置</button>
        </div>

        <!-- 代理服务控制 -->
        <div class="section">
            <h2>代理服务控制</h2>
            <div class="proxy-control">
                <!-- SOCKS5 代理 -->
                <div class="proxy-card">
                    <h3>SOCKS5 代理 <span id="socks5-status" class="status stopped">已停止</span></h3>
                    <div class="form-group">
                        <label>端口号</label>
                        <input type="number" id="socks5-port" value="1080" min="1" max="65535">
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="socks5-bind">
                            启用 Bind-Listen 模式
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="socks5-autostart">
                            开机自启
                        </label>
                    </div>
                    <div class="actions">
                        <button class="btn-primary" onclick="startProxy('socks5')">启动</button>
                        <button class="btn-danger" onclick="stopProxy('socks5')">停止</button>
                    </div>
                </div>

                <!-- HTTP 代理 -->
                <div class="proxy-card">
                    <h3>HTTP 代理 <span id="http-status" class="status stopped">已停止</span></h3>
                    <div class="form-group">
                        <label>端口号</label>
                        <input type="number" id="http-port" value="8080" min="1" max="65535">
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="http-bind">
                            启用 Bind-Listen 模式
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="http-autostart">
                            开机自启
                        </label>
                    </div>
                    <div class="actions">
                        <button class="btn-primary" onclick="startProxy('http')">启动</button>
                        <button class="btn-danger" onclick="stopProxy('http')">停止</button>
                    </div>
                </div>
            </div>
        </div>

        <!-- 用户管理 -->
        <div class="section">
            <h2>用户管理</h2>
            <div class="form-group">
                <label>用户名</label>
                <input type="text" id="user-username" placeholder="输入用户名">
            </div>
            <div class="form-group">
                <label>密码</label>
                <input type="password" id="user-password" placeholder="输入密码">
            </div>
            <div class="form-group">
                <label>IP 地址（可选）</label>
                <input type="text" id="user-ip" placeholder="留空表示不限制 IP">
            </div>
            <button class="btn-secondary" onclick="addUser()">添加用户</button>
            <button class="btn-secondary" onclick="loadUsers()">刷新列表</button>

            <table id="users-table">
                <thead>
                    <tr>
                        <th>用户名</th>
                        <th>IP 地址</th>
                        <th>创建时间</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody id="users-tbody">
                    <tr><td colspan="4" style="text-align:center;">加载中...</td></tr>
                </tbody>
            </table>
        </div>

        <!-- IP 白名单管理 -->
        <div class="section">
            <h2>IP 白名单管理</h2>
            <div class="form-group">
                <label>IP 地址</label>
                <input type="text" id="whitelist-ip" placeholder="输入 IP 地址">
            </div>
            <button class="btn-secondary" onclick="addWhitelistIP()">添加 IP</button>
            <button class="btn-secondary" onclick="loadWhitelist()">刷新列表</button>

            <table id="whitelist-table">
                <thead>
                    <tr>
                        <th>IP 地址</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody id="whitelist-tbody">
                    <tr><td colspan="2" style="text-align:center;">加载中...</td></tr>
                </tbody>
            </table>
        </div>
    </div>

    <script>
        // 显示消息
        function showMessage(text, type = 'success') {
            const msg = document.getElementById('message');
            msg.textContent = text;
            msg.className = 'message ' + type;
            msg.style.display = 'block';
            setTimeout(() => {
                msg.style.display = 'none';
            }, 3000);
        }

        // API 调用封装
        async function apiCall(url, method = 'GET', data = null) {
            try {
                const options = {
                    method: method,
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                if (data) {
                    options.body = JSON.stringify(data);
                }
                const response = await fetch(url, options);
                if (!response.ok) {
                    const text = await response.text();
                    throw new Error(text || 'Request failed');
                }
                return await response.json();
            } catch (error) {
                showMessage('错误: ' + error.message, 'error');
                throw error;
            }
        }

        // 更新状态
        async function updateStatus() {
            try {
                const status = await apiCall('/api/status');

                // 更新 SOCKS5 状态
                const socks5Status = document.getElementById('socks5-status');
                if (status.socks5.running) {
                    socks5Status.textContent = '运行中';
                    socks5Status.className = 'status running';
                    document.getElementById('socks5-port').value = status.socks5.port;
                    document.getElementById('socks5-bind').checked = status.socks5.bindListen;
                } else {
                    socks5Status.textContent = '已停止';
                    socks5Status.className = 'status stopped';
                }
                document.getElementById('socks5-autostart').checked = status.socks5.autoStart;

                // 更新 HTTP 状态
                const httpStatus = document.getElementById('http-status');
                if (status.http.running) {
                    httpStatus.textContent = '运行中';
                    httpStatus.className = 'status running';
                    document.getElementById('http-port').value = status.http.port;
                    document.getElementById('http-bind').checked = status.http.bindListen;
                } else {
                    httpStatus.textContent = '已停止';
                    httpStatus.className = 'status stopped';
                }
                document.getElementById('http-autostart').checked = status.http.autoStart;
            } catch (error) {
                console.error('Failed to update status:', error);
            }
        }

        // 启动代理
        async function startProxy(type) {
            const port = parseInt(document.getElementById(type + '-port').value);
            const bindListen = document.getElementById(type + '-bind').checked;

            try {
                await apiCall('/api/proxy/start', 'POST', {
                    type: type,
                    port: port,
                    bindListen: bindListen
                });
                showMessage(type.toUpperCase() + ' 代理启动成功');
                updateStatus();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 停止代理
        async function stopProxy(type) {
            try {
                await apiCall('/api/proxy/stop', 'POST', { type: type });
                showMessage(type.toUpperCase() + ' 代理已停止');
                updateStatus();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 保存代理配置
        async function saveProxyConfig(type) {
            const port = parseInt(document.getElementById(type + '-port').value);
            const bindListen = document.getElementById(type + '-bind').checked;
            const autoStart = document.getElementById(type + '-autostart').checked;

            try {
                await apiCall('/api/proxy/config', 'POST', {
                    type: type,
                    port: port,
                    bindListen: bindListen,
                    autoStart: autoStart
                });
                showMessage('配置已保存');
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 加载用户列表
        async function loadUsers() {
            try {
                const users = await apiCall('/api/users');
                const tbody = document.getElementById('users-tbody');
                if (users.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="4" style="text-align:center;">暂无用户</td></tr>';
                    return;
                }
                tbody.innerHTML = users.map(user => ` + "`" + `
                    <tr>
                        <td>${user.Username}</td>
                        <td>${user.IP || '不限制'}</td>
                        <td>${new Date(user.CreatedAt).toLocaleString()}</td>
                        <td>
                            <button class="btn-danger" onclick="deleteUser('${user.IP}', '${user.Username}')">删除</button>
                        </td>
                    </tr>
                ` + "`" + `).join('');
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 添加用户
        async function addUser() {
            const username = document.getElementById('user-username').value.trim();
            const password = document.getElementById('user-password').value;
            const ip = document.getElementById('user-ip').value.trim();

            if (!username || !password) {
                showMessage('请输入用户名和密码', 'error');
                return;
            }

            try {
                await apiCall('/api/users', 'POST', {
                    username: username,
                    password: password,
                    ip: ip
                });
                showMessage('用户添加成功');
                document.getElementById('user-username').value = '';
                document.getElementById('user-password').value = '';
                document.getElementById('user-ip').value = '';
                loadUsers();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 删除用户
        async function deleteUser(ip, username) {
            if (!confirm('确定要删除用户 ' + username + ' 吗？')) {
                return;
            }

            try {
                await apiCall('/api/users', 'DELETE', {
                    username: username,
                    ip: ip
                });
                showMessage('用户删除成功');
                loadUsers();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 加载白名单
        async function loadWhitelist() {
            try {
                const ips = await apiCall('/api/whitelist');
                const tbody = document.getElementById('whitelist-tbody');
                if (ips.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="2" style="text-align:center;">暂无白名单 IP</td></tr>';
                    return;
                }
                tbody.innerHTML = ips.map(ip => ` + "`" + `
                    <tr>
                        <td>${ip}</td>
                        <td><button class="btn-danger" onclick="deleteWhitelistIP('${ip}')">删除</button></td>
                    </tr>
                ` + "`" + `).join('');
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 添加白名单 IP
        async function addWhitelistIP() {
            const ip = document.getElementById('whitelist-ip').value.trim();

            if (!ip) {
                showMessage('请输入 IP 地址', 'error');
                return;
            }

            try {
                await apiCall('/api/whitelist', 'POST', { ip: ip });
                showMessage('IP 添加成功');
                document.getElementById('whitelist-ip').value = '';
                loadWhitelist();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 删除白名单 IP
        async function deleteWhitelistIP(ip) {
            if (!confirm('确定要删除 IP ' + ip + ' 吗？')) {
                return;
            }

            try {
                await apiCall('/api/whitelist', 'DELETE', { ip: ip });
                showMessage('IP 删除成功');
                loadWhitelist();
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 加载系统设置
        async function loadSystemSettings() {
            try {
                const settings = await apiCall('/api/system/settings');
                document.getElementById('system-autostart').checked = settings.autostartEnabled;
            } catch (error) {
                console.error('Failed to load system settings:', error);
            }
        }

        // 保存系统设置
        async function saveSystemSettings() {
            const autostartEnabled = document.getElementById('system-autostart').checked;

            try {
                await apiCall('/api/system/settings', 'POST', {
                    autostartEnabled: autostartEnabled
                });
                showMessage('系统设置已保存');
            } catch (error) {
                // Error already shown by apiCall
                // Revert checkbox on error
                document.getElementById('system-autostart').checked = !autostartEnabled;
            }
        }

        // 加载超时配置
        async function loadTimeout() {
            try {
                const timeout = await apiCall('/api/timeout');
                document.getElementById('timeout-connect').value = timeout.connect;
                document.getElementById('timeout-idle-read').value = timeout.idleRead;
                document.getElementById('timeout-idle-write').value = timeout.idleWrite;
            } catch (error) {
                console.error('Failed to load timeout configuration:', error);
            }
        }

        // 保存超时配置
        async function saveTimeout() {
            const connect = parseInt(document.getElementById('timeout-connect').value);
            const idleRead = parseInt(document.getElementById('timeout-idle-read').value);
            const idleWrite = parseInt(document.getElementById('timeout-idle-write').value);

            // 验证输入
            if (connect < 1 || connect > 300) {
                showMessage('连接超时必须在 1-300 秒之间', true);
                return;
            }
            if (idleRead < 1 || idleRead > 3600) {
                showMessage('空闲读取超时必须在 1-3600 秒之间', true);
                return;
            }
            if (idleWrite < 1 || idleWrite > 3600) {
                showMessage('空闲写入超时必须在 1-3600 秒之间', true);
                return;
            }

            try {
                await apiCall('/api/timeout', 'POST', {
                    connect: connect,
                    idleRead: idleRead,
                    idleWrite: idleWrite
                });
                showMessage('超时配置已保存，立即生效');
            } catch (error) {
                // Error already shown by apiCall
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            updateStatus();
            loadUsers();
            loadWhitelist();
            loadSystemSettings();
            loadTimeout();
            // 每 5 秒更新一次状态
            setInterval(updateStatus, 5000);

            // 添加自启开关事件监听器
            document.getElementById('socks5-autostart').addEventListener('change', function() {
                saveProxyConfig('socks5');
            });
            document.getElementById('http-autostart').addEventListener('change', function() {
                saveProxyConfig('http');
            });

            // 添加系统设置事件监听器
            document.getElementById('system-autostart').addEventListener('change', function() {
                saveSystemSettings();
            });
        };
    </script>
</body>
</html>
` + "`"
