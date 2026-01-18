import React from 'react';
import { Form, Input, Button } from 'antd';
import type { AddUserRequest } from '../../types/user';

interface AddUserFormProps {
  onSubmit: (values: AddUserRequest) => void;
}

const AddUserForm: React.FC<AddUserFormProps> = ({ onSubmit }) => {
  const [form] = Form.useForm();

  const handleSubmit = (values: AddUserRequest) => {
    onSubmit(values);
    form.resetFields();
  };

  return (
    <Form
      form={form}
      layout="inline"
      onFinish={handleSubmit}
    >
      <Form.Item
        name="username"
        rules={[{ required: true, message: '请输入用户名' }]}
      >
        <Input placeholder="用户名" />
      </Form.Item>

      <Form.Item
        name="password"
        rules={[{ required: true, message: '请输入密码' }]}
      >
        <Input.Password placeholder="密码" />
      </Form.Item>

      <Form.Item name="ip">
        <Input placeholder="IP 地址（可选）" />
      </Form.Item>

      <Form.Item>
        <Button type="primary" htmlType="submit">
          添加用户
        </Button>
      </Form.Item>
    </Form>
  );
};

export default AddUserForm;
