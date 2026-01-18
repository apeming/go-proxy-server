import React, { useState, useEffect } from 'react';
import { Card, Button, InputNumber, Switch, Space, Form, message, Badge, Divider, Typography } from 'antd';
import { PlayCircleOutlined, StopOutlined, SaveOutlined, ApiOutlined } from '@ant-design/icons';
import { startProxy, stopProxy, saveProxyConfig } from '../../api/proxy';
import type { ProxyServerStatus } from '../../types/proxy';

const { Text } = Typography;

interface ProxyCardProps {
  type: 'socks5' | 'http';
  title: string;
  status?: ProxyServerStatus;
  onStatusChange: () => void;
  loading: boolean;
}

const ProxyCard: React.FC<ProxyCardProps> = ({ type, title, status, onStatusChange, loading }) => {
  const [form] = Form.useForm();
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => {
    if (status) {
      form.setFieldsValue({
        port: status.port,
        bindListen: status.bindListen,
        autoStart: status.autoStart,
      });
    }
  }, [status, form]);

  const handleStart = async () => {
    try {
      const values = form.getFieldsValue();
      setActionLoading(true);

      // Save configuration first (including autoStart setting)
      await saveProxyConfig({
        type,
        port: values.port,
        bindListen: values.bindListen,
        autoStart: values.autoStart,
      });

      // Then start the proxy
      await startProxy({
        type,
        port: values.port,
        bindListen: values.bindListen,
      });
      message.success(`${title}启动成功`);
      onStatusChange();
    } catch (error) {
      console.error('Failed to start proxy:', error);
      message.error(`${title}启动失败`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleStop = async () => {
    try {
      setActionLoading(true);
      await stopProxy({ type });
      message.success(`${title}停止成功`);
      onStatusChange();
    } catch (error) {
      console.error('Failed to stop proxy:', error);
      message.error(`${title}停止失败`);
    } finally {
      setActionLoading(false);
    }
  };

  const handleSaveConfig = async () => {
    try {
      const values = form.getFieldsValue();
      setActionLoading(true);
      await saveProxyConfig({
        type,
        port: values.port,
        bindListen: values.bindListen,
        autoStart: values.autoStart,
      });
      message.success('配置保存成功');
      onStatusChange();
    } catch (error) {
      console.error('Failed to save config:', error);
      message.error('配置保存失败');
    } finally {
      setActionLoading(false);
    }
  };

  const defaultPort = type === 'socks5' ? 1080 : 8080;
  const cardColor = type === 'socks5' ? '#1890ff' : '#722ed1';

  return (
    <Card
      loading={loading}
      bordered={false}
      style={{
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
        borderTop: `4px solid ${cardColor}`,
      }}
    >
      <div style={{ marginBottom: 16 }}>
        <Space align="center" size="middle">
          <div style={{
            width: 48,
            height: 48,
            borderRadius: '8px',
            background: `linear-gradient(135deg, ${cardColor}22, ${cardColor}44)`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <ApiOutlined style={{ fontSize: '24px', color: cardColor }} />
          </div>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Text strong style={{ fontSize: '18px' }}>{title}</Text>
              {status?.running ? (
                <Badge status="processing" text="运行中" />
              ) : (
                <Badge status="default" text="已停止" />
              )}
            </div>
            <Text type="secondary" style={{ fontSize: '12px' }}>
              {type === 'socks5' ? 'SOCKS5 协议代理服务' : 'HTTP/HTTPS 协议代理服务'}
            </Text>
          </div>
        </Space>
      </div>

      <Divider style={{ margin: '16px 0' }} />

      <Form
        form={form}
        layout="vertical"
        initialValues={{ port: defaultPort, bindListen: false, autoStart: false }}
      >
        <Form.Item
          label={<Text strong>监听端口</Text>}
          name="port"
          tooltip="代理服务监听的端口号 (1-65535)"
        >
          <InputNumber
            min={1}
            max={65535}
            style={{ width: '100%' }}
            disabled={status?.running}
            size="large"
            placeholder={`默认端口: ${defaultPort}`}
          />
        </Form.Item>

        <Form.Item
          label={<Text strong>Bind-Listen 模式</Text>}
          name="bindListen"
          valuePropName="checked"
          tooltip="启用后，服务器将使用客户端连接的本地 IP 作为出站连接的源地址"
        >
          <Switch
            disabled={status?.running}
            checkedChildren="已启用"
            unCheckedChildren="未启用"
          />
        </Form.Item>

        <Form.Item
          label={<Text strong>开机自启</Text>}
          name="autoStart"
          valuePropName="checked"
          tooltip="应用启动时自动启动此代理服务"
        >
          <Switch
            checkedChildren="已启用"
            unCheckedChildren="未启用"
          />
        </Form.Item>

        <Space size="middle" style={{ width: '100%', justifyContent: 'flex-end' }}>
          <Button
            icon={<SaveOutlined />}
            onClick={handleSaveConfig}
            loading={actionLoading}
            size="large"
          >
            保存配置
          </Button>
          {status?.running ? (
            <Button
              type="primary"
              danger
              icon={<StopOutlined />}
              onClick={handleStop}
              loading={actionLoading}
              size="large"
            >
              停止服务
            </Button>
          ) : (
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={handleStart}
              loading={actionLoading}
              size="large"
              style={{ background: cardColor, borderColor: cardColor }}
            >
              启动服务
            </Button>
          )}
        </Space>
      </Form>
    </Card>
  );
};

export default ProxyCard;
