import React, { useState, useEffect } from 'react';
import { Card, Form, InputNumber, Button, Row, Col, message, Typography, Alert, Switch, Space } from 'antd';
import { SettingOutlined, SaveOutlined, ClockCircleOutlined, ApiOutlined, WindowsOutlined, CheckCircleOutlined, WarningOutlined } from '@ant-design/icons';
import { getConfig, saveConfig } from '../../api/config';
import type { UnifiedConfig } from '../../types/api';

const { Title, Text, Paragraph } = Typography;

const ConfigManagement: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [config, setConfig] = useState<UnifiedConfig | null>(null);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const response = await getConfig();
      setConfig(response.data);
      form.setFieldsValue({
        connect: response.data.timeout.connect,
        idleRead: response.data.timeout.idleRead,
        idleWrite: response.data.timeout.idleWrite,
        maxConcurrentConnections: response.data.limiter.maxConcurrentConnections,
        maxConcurrentConnectionsPerIP: response.data.limiter.maxConcurrentConnectionsPerIP,
        autostartEnabled: response.data.system.autostartEnabled,
      });
    } catch (error) {
      console.error('Failed to load config:', error);
      message.error('加载配置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConfig();
  }, []);

  const handleSave = async (values: any) => {
    try {
      setLoading(true);
      await saveConfig({
        timeout: {
          connect: values.connect,
          idleRead: values.idleRead,
          idleWrite: values.idleWrite,
        },
        limiter: {
          maxConcurrentConnections: values.maxConcurrentConnections,
          maxConcurrentConnectionsPerIP: values.maxConcurrentConnectionsPerIP,
        },
        system: {
          autostartEnabled: values.autostartEnabled,
        },
      });
      message.success('配置保存成功');
      loadConfig();
    } catch (error) {
      console.error('Failed to save config:', error);
      message.error('配置保存失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        <SettingOutlined style={{ marginRight: 8, color: '#1890ff' }} />
        系统配置
      </Title>

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        initialValues={{
          connect: 30,
          idleRead: 300,
          idleWrite: 300,
          maxConcurrentConnections: 100000,
          maxConcurrentConnectionsPerIP: 1000,
          autostartEnabled: false,
        }}
      >
        <Row gutter={[24, 24]}>
          {/* 超时配置 */}
          <Col span={24}>
            <Card
              title={
                <Space>
                  <ClockCircleOutlined style={{ color: '#1890ff' }} />
                  <Text strong>超时配置</Text>
                </Space>
              }
              bordered={false}
              style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}
              loading={loading}
            >
              <Alert
                message="配置说明"
                description={
                  <div>
                    <Paragraph style={{ marginBottom: 8 }}>
                      <Text strong>连接超时：</Text>建立连接时的最大等待时间，建议范围 10-60 秒
                    </Paragraph>
                    <Paragraph style={{ marginBottom: 8 }}>
                      <Text strong>空闲读取超时：</Text>连接空闲时等待读取数据的最大时间，建议范围 60-600 秒
                    </Paragraph>
                    <Paragraph style={{ marginBottom: 0 }}>
                      <Text strong>空闲写入超时：</Text>连接空闲时等待写入数据的最大时间，建议范围 60-600 秒
                    </Paragraph>
                  </div>
                }
                type="info"
                showIcon
                style={{ marginBottom: 24 }}
              />

              <Row gutter={24}>
                <Col xs={24} md={8}>
                  <Form.Item
                    label={<Text strong>连接超时（秒）</Text>}
                    name="connect"
                    rules={[
                      { required: true, message: '请输入连接超时时间' },
                      { type: 'number', min: 1, max: 300, message: '范围: 1-300 秒' },
                    ]}
                    tooltip="建立新连接时的最大等待时间"
                  >
                    <InputNumber
                      min={1}
                      max={300}
                      style={{ width: '100%' }}
                      placeholder="推荐: 30"
                      size="large"
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item
                    label={<Text strong>空闲读取超时（秒）</Text>}
                    name="idleRead"
                    rules={[
                      { required: true, message: '请输入空闲读取超时时间' },
                      { type: 'number', min: 1, max: 3600, message: '范围: 1-3600 秒' },
                    ]}
                    tooltip="连接空闲时等待读取数据的最大时间"
                  >
                    <InputNumber
                      min={1}
                      max={3600}
                      style={{ width: '100%' }}
                      placeholder="推荐: 300"
                      size="large"
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item
                    label={<Text strong>空闲写入超时（秒）</Text>}
                    name="idleWrite"
                    rules={[
                      { required: true, message: '请输入空闲写入超时时间' },
                      { type: 'number', min: 1, max: 3600, message: '范围: 1-3600 秒' },
                    ]}
                    tooltip="连接空闲时等待写入数据的最大时间"
                  >
                    <InputNumber
                      min={1}
                      max={3600}
                      style={{ width: '100%' }}
                      placeholder="推荐: 300"
                      size="large"
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </Col>

          {/* 连接限制配置 */}
          <Col span={24}>
            <Card
              title={
                <Space>
                  <ApiOutlined style={{ color: '#1890ff' }} />
                  <Text strong>连接限制配置</Text>
                </Space>
              }
              bordered={false}
              style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}
              loading={loading}
            >
              <Alert
                message="配置说明"
                description={
                  <div>
                    <Paragraph style={{ marginBottom: 8 }}>
                      <Text strong>最大并发连接数：</Text>整个代理服务器允许的最大并发连接数，防止资源耗尽
                    </Paragraph>
                    <Paragraph style={{ marginBottom: 0 }}>
                      <Text strong>单IP最大并发连接数：</Text>每个客户端IP允许的最大并发连接数，防止单个IP占用所有资源
                    </Paragraph>
                  </div>
                }
                type="info"
                showIcon
                style={{ marginBottom: 24 }}
              />

              <Row gutter={24}>
                <Col xs={24} md={12}>
                  <Form.Item
                    label={<Text strong>最大并发连接数</Text>}
                    name="maxConcurrentConnections"
                    rules={[
                      { required: true, message: '请输入最大并发连接数' },
                      { type: 'number', min: 1, max: 1000000, message: '范围: 1-1000000' },
                    ]}
                    tooltip="整个代理服务器允许的最大并发连接数"
                  >
                    <InputNumber
                      min={1}
                      max={1000000}
                      style={{ width: '100%' }}
                      placeholder="推荐: 100000"
                      size="large"
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    label={<Text strong>单IP最大并发连接数</Text>}
                    name="maxConcurrentConnectionsPerIP"
                    rules={[
                      { required: true, message: '请输入单IP最大并发连接数' },
                      { type: 'number', min: 1, max: 100000, message: '范围: 1-100000' },
                    ]}
                    tooltip="单个客户端IP允许的最大并发连接数"
                  >
                    <InputNumber
                      min={1}
                      max={100000}
                      style={{ width: '100%' }}
                      placeholder="推荐: 1000"
                      size="large"
                    />
                  </Form.Item>
                </Col>
              </Row>

              <Alert
                message="重要提示"
                description="修改连接限制配置后，需要重启代理服务器才能生效"
                type="warning"
                showIcon
                style={{ marginTop: 16 }}
              />
            </Card>
          </Col>

          {/* 系统设置 */}
          <Col span={24}>
            <Card
              title={
                <Space>
                  <WindowsOutlined style={{ color: '#1890ff' }} />
                  <Text strong>系统设置</Text>
                </Space>
              }
              bordered={false}
              style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}
              loading={loading}
            >
              {!config?.system.autostartSupported && (
                <Alert
                  message="平台限制"
                  description="开机自启功能仅在 Windows 平台可用，当前平台不支持此功能。"
                  type="info"
                  showIcon
                  icon={<WindowsOutlined />}
                  style={{ marginBottom: 24 }}
                />
              )}

              <div style={{ marginBottom: 24 }}>
                <Space direction="vertical" size="small">
                  <Text strong style={{ fontSize: '16px' }}>
                    <WindowsOutlined style={{ marginRight: 8, color: '#1890ff' }} />
                    Windows 开机自启
                  </Text>
                  <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                    启用后，应用程序将在 Windows 系统启动时自动运行。此功能会修改 Windows 注册表中的启动项。
                  </Paragraph>
                </Space>
              </div>

              <div style={{
                padding: '20px',
                background: '#fafafa',
                borderRadius: '8px',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
                <Space>
                  <SettingOutlined style={{ fontSize: '24px', color: config?.system.autostartSupported ? '#1890ff' : '#d9d9d9' }} />
                  <div>
                    <Text strong style={{ fontSize: '16px' }}>开机自启动</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: '13px' }}>
                      {config?.system.autostartSupported ? '系统启动时自动运行应用' : '当前平台不支持'}
                    </Text>
                  </div>
                </Space>
                <Form.Item
                  name="autostartEnabled"
                  valuePropName="checked"
                  style={{ marginBottom: 0 }}
                >
                  <Switch
                    disabled={!config?.system.autostartSupported}
                    checkedChildren="已启用"
                    unCheckedChildren="未启用"
                    size="default"
                  />
                </Form.Item>
              </div>

              {config?.system.registryEnabled !== undefined && (
                <Alert
                  message="注册表状态"
                  description={
                    <Space>
                      {config.system.registryEnabled ? (
                        <>
                          <CheckCircleOutlined style={{ color: '#52c41a' }} />
                          <Text>Windows 注册表启动项已配置</Text>
                        </>
                      ) : (
                        <>
                          <WarningOutlined style={{ color: '#faad14' }} />
                          <Text>Windows 注册表启动项未配置</Text>
                        </>
                      )}
                    </Space>
                  }
                  type={config.system.registryEnabled ? 'success' : 'warning'}
                  showIcon
                  style={{ marginTop: 16 }}
                />
              )}
            </Card>
          </Col>

          {/* 保存按钮 */}
          <Col span={24}>
            <Card bordered={false} style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}>
              <Form.Item style={{ marginBottom: 0 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  icon={<SaveOutlined />}
                  size="large"
                  block
                >
                  保存所有配置
                </Button>
              </Form.Item>
            </Card>
          </Col>
        </Row>
      </Form>
    </div>
  );
};

export default ConfigManagement;
