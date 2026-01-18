import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Badge, Space, Typography } from 'antd';
import {
  ApiOutlined,
  UserOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { getProxyStatus } from '../../api/proxy';
import { getUsers } from '../../api/user';
import { getWhitelist } from '../../api/whitelist';
import type { ProxyStatus } from '../../types/proxy';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  const [proxyStatus, setProxyStatus] = useState<ProxyStatus | null>(null);
  const [userCount, setUserCount] = useState(0);
  const [whitelistCount, setWhitelistCount] = useState(0);
  const [loading, setLoading] = useState(true);

  const loadData = async () => {
    try {
      setLoading(true);
      const [statusRes, usersRes, whitelistRes] = await Promise.all([
        getProxyStatus(),
        getUsers(),
        getWhitelist(),
      ]);
      setProxyStatus(statusRes.data);
      setUserCount(usersRes.data.length);
      setWhitelistCount(whitelistRes.data.length);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 10000);
    return () => clearInterval(interval);
  }, []);

  const runningProxies = [
    proxyStatus?.socks5?.running,
    proxyStatus?.http?.running,
  ].filter(Boolean).length;

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        <ThunderboltOutlined style={{ marginRight: 8, color: '#1890ff' }} />
        系统概览
      </Title>

      <Row gutter={[24, 24]}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false} style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', height: '140px' }}>
            <Statistic
              title={<span style={{ color: 'rgba(255,255,255,0.85)', fontSize: '14px' }}>运行中的代理</span>}
              value={runningProxies}
              suffix="/ 2"
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#fff', fontSize: '32px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false} style={{ background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)', height: '140px' }}>
            <Statistic
              title={<span style={{ color: 'rgba(255,255,255,0.85)', fontSize: '14px' }}>用户总数</span>}
              value={userCount}
              prefix={<UserOutlined />}
              valueStyle={{ color: '#fff', fontSize: '32px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false} style={{ background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', height: '140px' }}>
            <Statistic
              title={<span style={{ color: 'rgba(255,255,255,0.85)', fontSize: '14px' }}>白名单 IP</span>}
              value={whitelistCount}
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#fff', fontSize: '32px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false} style={{ background: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)', height: '140px' }}>
            <Statistic
              title={<span style={{ color: 'rgba(255,255,255,0.85)', fontSize: '14px' }}>系统状态</span>}
              value="正常"
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#fff', fontSize: '24px', fontWeight: 'bold' }}
            />
          </Card>
        </Col>
      </Row>

      <Title level={3} style={{ marginTop: 32, marginBottom: 24 }}>
        代理服务状态
      </Title>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <ApiOutlined style={{ fontSize: '18px', color: '#1890ff' }} />
                <span style={{ fontSize: '16px', fontWeight: 600 }}>SOCKS5 代理</span>
              </Space>
            }
            loading={loading}
            extra={
              proxyStatus?.socks5?.running ? (
                <Badge status="processing" text="运行中" />
              ) : (
                <Badge status="default" text="已停止" />
              )
            }
            style={{ height: '100%' }}
          >
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">运行状态</Text>
                <Space>
                  {proxyStatus?.socks5?.running ? (
                    <>
                      <CheckCircleOutlined style={{ color: '#52c41a', fontSize: '18px' }} />
                      <Text strong style={{ color: '#52c41a' }}>运行中</Text>
                    </>
                  ) : (
                    <>
                      <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: '18px' }} />
                      <Text strong style={{ color: '#ff4d4f' }}>已停止</Text>
                    </>
                  )}
                </Space>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">监听端口</Text>
                <Text strong style={{ fontSize: '16px', color: '#1890ff' }}>
                  {proxyStatus?.socks5?.port || '-'}
                </Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">Bind-Listen 模式</Text>
                <Text strong>
                  {proxyStatus?.socks5?.bindListen ? '已启用' : '未启用'}
                </Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">开机自启</Text>
                <Text strong>
                  {proxyStatus?.socks5?.autoStart ? '已启用' : '未启用'}
                </Text>
              </div>
            </Space>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title={
              <Space>
                <ApiOutlined style={{ fontSize: '18px', color: '#722ed1' }} />
                <span style={{ fontSize: '16px', fontWeight: 600 }}>HTTP 代理</span>
              </Space>
            }
            loading={loading}
            extra={
              proxyStatus?.http?.running ? (
                <Badge status="processing" text="运行中" />
              ) : (
                <Badge status="default" text="已停止" />
              )
            }
            style={{ height: '100%' }}
          >
            <Space direction="vertical" size="large" style={{ width: '100%' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">运行状态</Text>
                <Space>
                  {proxyStatus?.http?.running ? (
                    <>
                      <CheckCircleOutlined style={{ color: '#52c41a', fontSize: '18px' }} />
                      <Text strong style={{ color: '#52c41a' }}>运行中</Text>
                    </>
                  ) : (
                    <>
                      <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: '18px' }} />
                      <Text strong style={{ color: '#ff4d4f' }}>已停止</Text>
                    </>
                  )}
                </Space>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">监听端口</Text>
                <Text strong style={{ fontSize: '16px', color: '#722ed1' }}>
                  {proxyStatus?.http?.port || '-'}
                </Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">Bind-Listen 模式</Text>
                <Text strong>
                  {proxyStatus?.http?.bindListen ? '已启用' : '未启用'}
                </Text>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '12px 0' }}>
                <Text type="secondary">开机自启</Text>
                <Text strong>
                  {proxyStatus?.http?.autoStart ? '已启用' : '未启用'}
                </Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
