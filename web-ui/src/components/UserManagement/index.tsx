import React, { useState, useEffect } from 'react';
import { Card, Button, Space, message, Typography, Row, Col, Modal } from 'antd';
import { ReloadOutlined, UserAddOutlined, TeamOutlined } from '@ant-design/icons';
import UserTable from './UserTable';
import AddUserForm from './AddUserForm';
import { getUsers, addUser, deleteUser } from '../../api/user';
import type { User, AddUserRequest } from '../../types/user';

const { Title } = Typography;

const UserManagement: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await getUsers();
      setUsers(response.data);
    } catch (error) {
      console.error('Failed to load users:', error);
      message.error('加载用户列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
  }, []);

  const handleAddUser = async (values: AddUserRequest) => {
    try {
      await addUser(values);
      message.success('用户添加成功');
      setModalVisible(false);
      loadUsers();
    } catch (error) {
      console.error('Failed to add user:', error);
      message.error('用户添加失败');
    }
  };

  const handleDeleteUser = async (username: string) => {
    try {
      await deleteUser({ username });
      message.success('用户删除成功');
      loadUsers();
    } catch (error) {
      console.error('Failed to delete user:', error);
      message.error('用户删除失败');
    }
  };

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24 }}>
        <TeamOutlined style={{ marginRight: 8, color: '#1890ff' }} />
        用户管理
      </Title>

      <Row gutter={[24, 24]}>
        <Col span={24}>
          <Card
            title={
              <Space>
                <TeamOutlined style={{ fontSize: '18px', color: '#52c41a' }} />
                <span style={{ fontSize: '16px', fontWeight: 600 }}>用户列表 ({users.length})</span>
              </Space>
            }
            bordered={false}
            style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.1)' }}
            extra={
              <Space>
                <Button icon={<ReloadOutlined />} onClick={loadUsers} loading={loading}>
                  刷新
                </Button>
                <Button type="primary" icon={<UserAddOutlined />} onClick={() => setModalVisible(true)}>
                  添加用户
                </Button>
              </Space>
            }
          >
            <UserTable
              users={users}
              loading={loading}
              onDelete={handleDeleteUser}
            />
          </Card>
        </Col>
      </Row>

      <Modal
        title={
          <Space>
            <UserAddOutlined style={{ color: '#1890ff' }} />
            <span>添加新用户</span>
          </Space>
        }
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <AddUserForm onSubmit={handleAddUser} />
      </Modal>
    </div>
  );
};

export default UserManagement;
