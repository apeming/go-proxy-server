import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Badge, Space, Typography } from 'antd';
import {
  ApiOutlined,
  UserOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ThunderboltOutlined,
  CloudUploadOutlined,
  CloudDownloadOutlined,
  LinkOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { Line } from '@ant-design/charts';
import { getProxyStatus } from '../../api/proxy';
import { getUsers } from '../../api/user';
import { getWhitelist } from '../../api/whitelist';
import { getRealtimeMetrics, getMetricsHistory } from '../../api/metrics';
import type { ProxyStatus } from '../../types/proxy';
import type { MetricsSnapshot, MetricsHistory } from '../../types/metrics';

const { Title, Text } = Typography;

const Dashboard: React.FC = () => {
  const [proxyStatus, setProxyStatus] = useState<ProxyStatus | null>(null);
  const [userCount, setUserCount] = useState(0);
  const [whitelistCount, setWhitelistCount] = useState(0);
  const [metrics, setMetrics] = useState<MetricsSnapshot | null>(null);
  const [metricsHistory, setMetricsHistory] = useState<MetricsHistory[]>([]);
  const [loading, setLoading] = useState(true);

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
  };

  const formatSpeed = (bytesPerSec: number): string => {
    return formatBytes(bytesPerSec) + '/s';
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (days > 0) return `${days}天 ${hours}小时`;
    if (hours > 0) return `${hours}小时 ${minutes}分钟`;
    return `${minutes}分钟`;
  };

  const loadData = async (isInitialLoad = false) => {
    try {
      if (isInitialLoad) {
        setLoading(true);
      }
      const [statusRes, usersRes, whitelistRes, metricsRes] = await Promise.all([
        getProxyStatus(),
        getUsers(),
        getWhitelist(),
        getRealtimeMetrics(),
      ]);
      setProxyStatus(statusRes.data);
      setUserCount(usersRes.data.length);
      setWhitelistCount(whitelistRes.data.length);
      setMetrics(metricsRes);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
    } finally {
      if (isInitialLoad) {
        setLoading(false);
      }
    }
  };

  const loadHistory = async () => {
    try {
      const endTime = Math.floor(Date.now() / 1000);
      const startTime = endTime - 3600; // Last hour
      const history = await getMetricsHistory(startTime, endTime, 60);
      setMetricsHistory(history);
    } catch (error) {
      console.error('Failed to load metrics history:', error);
    }
  };

  useEffect(() => {
    loadData(true); // Initial load with loading state
    loadHistory();
    const interval = setInterval(() => {
      loadData(false); // Subsequent updates without loading state
      loadHistory();
    }, 5000); // Update every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const runningProxies = [
    proxyStatus?.socks5?.running,
    proxyStatus?.http?.running,
  ].filter(Boolean).length;

  // Prepare chart data
  const bandwidthData = metricsHistory.map((h) => ([
    {
      time: new Date(h.Timestamp * 1000).toLocaleTimeString(),
      value: h.UploadSpeed / 1024 / 1024, // Convert to MB/s
      type: '上传速度',
    },
    {
      time: new Date(h.Timestamp * 1000).toLocaleTimeString(),
      value: h.DownloadSpeed / 1024 / 1024, // Convert to MB/s
      type: '下载速度',
    },
  ])).flat();

  const connectionsData = metricsHistory.map((h) => ({
    time: new Date(h.Timestamp * 1000).toLocaleTimeString(),
    value: h.ActiveConnections,
  }));

  const bandwidthConfig = {
    data: bandwidthData,
    xField: 'time',
    yField: 'value',
    seriesField: 'type',
    smooth: true,
    animation: false, // Disable animation to prevent flashing on data updates
    yAxis: {
      label: {
        formatter: (v: string) => `${v} MB/s`,
      },
    },
  };

  const connectionsConfig = {
    data: connectionsData,
    xField: 'time',
    yField: 'value',
    smooth: true,
    color: '#5B8FF9',
    animation: false, // Disable animation to prevent flashing on data updates
    yAxis: {
      label: {
        formatter: (v: string) => `${v} 个`,
      },
    },
  };

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
        实时监控
      </Title>

      <Row gutter={[24, 24]}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="活跃连接数"
              value={metrics?.activeConnections || 0}
              prefix={<LinkOutlined />}
              suffix="个"
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="上传速度"
              value={metrics ? formatSpeed(metrics.uploadSpeed) : '0 B/s'}
              prefix={<CloudUploadOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="下载速度"
              value={metrics ? formatSpeed(metrics.downloadSpeed) : '0 B/s'}
              prefix={<CloudDownloadOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="错误计数"
              value={metrics?.errorCount || 0}
              prefix={<WarningOutlined />}
              suffix="次"
              valueStyle={{ color: metrics && metrics.errorCount > 0 ? '#ff4d4f' : '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="总连接数"
              value={metrics?.totalConnections || 0}
              suffix="次"
              valueStyle={{ fontSize: '20px' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="接收流量"
              value={metrics ? formatBytes(metrics.bytesReceived) : '0 B'}
              valueStyle={{ fontSize: '20px', color: '#1890ff' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="发送流量"
              value={metrics ? formatBytes(metrics.bytesSent) : '0 B'}
              valueStyle={{ fontSize: '20px', color: '#722ed1' }}
            />
          </Card>
        </Col>

        <Col xs={24} sm={12} lg={6}>
          <Card loading={loading} bordered={false}>
            <Statistic
              title="运行时长"
              value={metrics ? formatUptime(metrics.uptime) : '0分钟'}
              valueStyle={{ fontSize: '20px', color: '#52c41a' }}
            />
          </Card>
        </Col>
      </Row>

      <Title level={3} style={{ marginTop: 32, marginBottom: 24 }}>
        流量趋势（最近1小时）
      </Title>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card title="带宽使用情况" bordered={false}>
            <Line {...bandwidthConfig} height={300} />
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="连接数变化" bordered={false}>
            <Line {...connectionsConfig} height={300} />
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
