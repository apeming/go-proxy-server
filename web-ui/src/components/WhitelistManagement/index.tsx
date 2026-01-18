import React, { useState, useEffect } from 'react';
import { Card, Button, Space, message, Typography, Row, Col, Modal } from 'antd';
import { ReloadOutlined, PlusOutlined, SafetyOutlined } from '@ant-design/icons';
import WhitelistTable from './WhitelistTable';
import AddIPForm from './AddIPForm';
import { getWhitelist, addWhitelistIP, deleteWhitelistIP } from '../../api/whitelist';

const { Title } = Typography;

const WhitelistManagement: React.FC = () => {
  const [whitelist, setWhitelist] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);

  const loadWhitelist = async () => {
    try {
      setLoading(true);
      const response = await getWhitelist();
      setWhitelist(response.data);
    } catch (error) {
      console.error('Failed to load whitelist:', error);
      message.error('加载白名单失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWhitelist();
  }, []);

  const handleAddIP = async (ip: string) => {
    try {
      await addWhitelistIP(ip);
      message.success('IP 添加成功');
      setModalVisible(false);
      loadWhitelist();
    } catch (error) {
      console.error('Failed to add IP:', error);
      message.error('IP 添加失败');
    }
  };

  const handleDeleteIP = async (ip: string) => {
    try {
      await deleteWhitelistIP(ip);
      message.success('IP 删除成功');
      loadWhitelist();
    } catch (error) {
      console.error('Failed to delete IP:', error);
      message.error('IP 删除失败');
    }
  };

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        <SafetyOutlined style={{ marginRight: 8, color: '#1890ff' }} />
        IP 白名单管理
      </Title>

      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Card
            title={
              <Space>
                <SafetyOutlined style={{ fontSize: '18px', color: '#52c41a' }} />
                <span style={{ fontSize: '16px', fontWeight: 600 }}>IP 白名单列表 ({whitelist.length})</span>
              </Space>
            }
            bordered={false}
            style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={loadWhitelist} loading={loading}>
                  刷新
                </Button>
                <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalVisible(true)}>
                  添加 IP
                </Button>
              </Space>
            }
          >
            <WhitelistTable
              whitelist={whitelist}
              loading={loading}
              onDelete={handleDeleteIP}
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title={
          <Space>
            <PlusOutlined style={{ color: '#1890ff' }} />
            <span>添加 IP 白名单</span>
          </Space>
        }
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <AddIPForm onSubmit={handleAddIP} />
      </Modal>
    </div>
  );
};

export default WhitelistManagement;
