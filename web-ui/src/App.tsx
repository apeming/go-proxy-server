import React, { useState } from 'react';
import { Layout, Menu, theme } from 'antd';
import {
  DashboardOutlined,
  ControlOutlined,
  UserOutlined,
  SafetyOutlined,
  SettingOutlined,
  ClockCircleOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import Dashboard from './components/Dashboard';
import ProxyControl from './components/ProxyControl';
import UserManagement from './components/UserManagement';
import WhitelistManagement from './components/WhitelistManagement';
import SystemSettings from './components/SystemSettings';
import TimeoutConfig from './components/TimeoutConfig';
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
      key: 'timeout',
      icon: <ClockCircleOutlined />,
      label: '超时配置',
    },
    {
      key: 'system',
      icon: <SettingOutlined />,
      label: '系统设置',
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
      case 'timeout':
        return <TimeoutConfig />;
      case 'system':
        return <SystemSettings />;
      default:
        return <Dashboard />;
    }
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
            boxShadow: '0 1px 4px rgba(0,21,41,.08)',
          }}
        >
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
