import React, { useState } from 'react';
import { Layout, Menu, theme, Button, Modal, Space } from 'antd';
import {
  DashboardOutlined,
  ControlOutlined,
  UserOutlined,
  SafetyOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ApiOutlined,
  PoweroffOutlined,
} from '@ant-design/icons';
import Dashboard from './components/Dashboard';
import ProxyControl from './components/ProxyControl';
import UserManagement from './components/UserManagement';
import WhitelistManagement from './components/WhitelistManagement';
import ConfigManagement from './components/ConfigManagement';
import { shutdownApplication } from './api/system';
import './App.css';

const { Header, Sider, Content } = Layout;

const App: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [selectedKey, setSelectedKey] = useState(() => {
    // Load saved page from localStorage, default to 'dashboard'
    return localStorage.getItem('selectedPage') || 'dashboard';
  });
  const {
    token: { colorBgContainer },
  } = theme.useToken();

  const menuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      key: 'proxy',
      icon: <ControlOutlined />,
      label: '代理控制',
    },
    {
      key: 'users',
      icon: <UserOutlined />,
      label: '用户管理',
    },
    {
      key: 'whitelist',
      icon: <SafetyOutlined />,
      label: 'IP 白名单',
    },
    {
      key: 'config',
      icon: <SettingOutlined />,
      label: '系统配置',
    },
  ];

  const renderContent = () => {
    switch (selectedKey) {
      case 'dashboard':
        return <Dashboard />;
      case 'proxy':
        return <ProxyControl />;
      case 'users':
        return <UserManagement />;
      case 'whitelist':
        return <WhitelistManagement />;
      case 'config':
        return <ConfigManagement />;
      default:
        return <Dashboard />;
    }
  };

  const handleShutdown = () => {
    Modal.confirm({
      title: '确认退出',
      content: '确定要退出应用程序吗? 所有代理服务将停止。',
      okText: '确认退出',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          await shutdownApplication();
          // Show success message briefly, then close the window
          const modal = Modal.success({
            title: '应用程序正在关闭',
            content: '应用程序正在安全退出,窗口将在2秒后自动关闭...',
          });

          // Close the browser window/tab after 2 seconds
          setTimeout(() => {
            modal.destroy();
            window.close();
            // If window.close() doesn't work (some browsers restrict it),
            // redirect to a blank page
            setTimeout(() => {
              window.location.href = 'about:blank';
            }, 100);
          }, 2000);
        } catch (error) {
          // Ignore error as server might close before responding
          console.log('Application is shutting down');
          // Still try to close the window
          setTimeout(() => {
            window.close();
            setTimeout(() => {
              window.location.href = 'about:blank';
            }, 100);
          }, 1000);
        }
      },
    });
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '0 16px',
          color: '#fff',
          fontSize: collapsed ? '20px' : '18px',
          fontWeight: 'bold',
          transition: 'all 0.2s',
        }}>
          <ApiOutlined style={{ fontSize: '24px', marginRight: collapsed ? 0 : '8px' }} />
          {!collapsed && <span>Proxy Server</span>}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => {
            setSelectedKey(key);
            localStorage.setItem('selectedPage', key);
          }}
        />
      </Sider>
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'all 0.2s' }}>
        <Header
          style={{
            padding: '0 24px',
            background: colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            boxShadow: '0 1px 4px rgba(0,21,41,.08)',
          }}
        >
          <Space>
            {React.createElement(collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
              className: 'trigger',
              onClick: () => setCollapsed(!collapsed),
              style: { fontSize: '18px', cursor: 'pointer', marginRight: '24px' },
            })}
            <h1 style={{
              margin: 0,
              fontSize: '20px',
              fontWeight: 600,
              color: '#1890ff',
            }}>
              Go Proxy Server 管理后台
            </h1>
          </Space>
          <Button
            danger
            icon={<PoweroffOutlined />}
            onClick={handleShutdown}
          >
            退出应用
          </Button>
        </Header>
        <Content
          style={{
            margin: '24px',
            padding: '24px',
            background: '#f0f2f5',
            minHeight: 'calc(100vh - 112px)',
          }}
        >
          <div style={{ maxWidth: '1600px', margin: '0 auto' }}>
            {renderContent()}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default App;
