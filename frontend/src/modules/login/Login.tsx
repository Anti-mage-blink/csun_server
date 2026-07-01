import React, { useState, useEffect } from 'react'
import { Form, Input, Button, Modal, Select, message, notification } from 'antd'
import { UserOutlined, LockOutlined, CheckCircleOutlined } from '@ant-design/icons'
import { loginApi, getUsersListApi, submitForgotPasswordApi } from '@/api/auth'
import { useAuth, type User } from '@/context/AuthContext'
import './Login.css'

interface LoginProps {
  onLoginSuccess?: () => void
}

const { Option } = Select

const Login: React.FC<LoginProps> = ({ onLoginSuccess }) => {
  const { login } = useAuth()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState<boolean>(false)

  // 忘记密码 Modal 相关状态
  const [isForgotModalVisible, setIsForgotModalVisible] = useState<boolean>(false)
  const [users, setUsers] = useState<User[]>([])
  const [selectedUser, setSelectedUser] = useState<number | undefined>(undefined)
  const [isSubmitSuccess, setIsSubmitSuccess] = useState<boolean>(false)
  const [submitLoading, setSubmitLoading] = useState<boolean>(false)

  // 登录提交
  const onFinish = async (values: any) => {
    setLoading(true)
    try {
      const res = await loginApi(values.username.trim(), values.password.trim())
      if (res.code === 200) {
        message.success('登录成功，欢迎回来！')
        // 调用 context 里的登录方法，保存状态
        login(res.data.user, res.data.token)
        
        // 如果是管理员登录，提醒有待处理的密码重置申请（完全对标原 HTML 中的系统逻辑）
        if (res.data.user.role === '管理员') {
          const requestsStr = localStorage.getItem('passwordResetRequests') || '[]'
          const requests = JSON.parse(requestsStr)
          const pending = requests.filter((r: any) => r.status === 'pending')
          if (pending.length > 0) {
            const usersStr = localStorage.getItem('sys_users') || '[]'
            const sysUsers = JSON.parse(usersStr)
            const names = pending
              .map((r: any) => {
                const u = sysUsers.find((x: any) => x.id === r.userId)
                return u ? u.username : '未知'
              })
              .join('、')
            
            notification.info({
              message: '系统提示',
              description: `有 ${pending.length} 条密码重置申请待处理（用户：${names}），请在用户管理中重置。`,
              duration: 8,
              placement: 'topRight'
            })
          }
        }

        // 回调成功
        if (onLoginSuccess) {
          onLoginSuccess()
        }
      } else {
        message.error(res.message || '登录失败')
      }
    } catch (err: any) {
      message.error(err.message || '账号或密码错误')
    } finally {
      setLoading(false)
    }
  }

  // 加载用户列表（忘记密码选择用）
  const fetchUsers = async () => {
    try {
      const data = await getUsersListApi()
      // 排除 admin 账号，管理员密码不允许自助发起重置
      const filtered = data.filter(u => u.username !== 'admin')
      setUsers(filtered)
    } catch (err) {
      console.error('获取用户列表失败', err)
    }
  }

  const handleShowForgotPassword = () => {
    setIsForgotModalVisible(true)
    setIsSubmitSuccess(false)
    setSelectedUser(undefined)
    fetchUsers()
  }

  const handleForgotPasswordSubmit = async () => {
    if (!selectedUser) {
      message.error('请选择需要重置密码的账号')
      return
    }
    setSubmitLoading(true)
    try {
      await submitForgotPasswordApi(selectedUser)
      setIsSubmitSuccess(true)
    } catch (err: any) {
      message.error(err.message || '提交失败')
    } finally {
      setSubmitLoading(false)
    }
  }

  return (
    <div className="login-container">
      {/* 左侧大图/艺术文案分栏 */}
      <div className="login-left">
        <div className="login-left-content">
          <h1>世鑫产品报价系统</h1>
          <p>
            Shixin Product Quotation Management System
            <br />
            致力于为世鑫新材料提供高效、精准、专业的产品报价计算与工作流协作平台。
          </p>
        </div>
      </div>

      {/* 右侧登录表单分栏 */}
      <div className="login-right">
        <div className="login-box">
          <h2>账号登录</h2>
          <div className="subtitle">CSUN Product Quotation Management System</div>

          <Form
            form={form}
            name="login_form"
            layout="vertical"
            requiredMark={false}
            onFinish={onFinish}
            autoComplete="off"
          >
            <Form.Item
              label="账号"
              name="username"
              rules={[{ required: true, message: '请输入账号!' }]}
            >
              <Input
                prefix={<UserOutlined style={{ color: '#bfbfbf' }} />}
                placeholder="请输入账号"
                size="large"
              />
            </Form.Item>

            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码!' }]}
            >
              <Input.Password
                prefix={<LockOutlined style={{ color: '#bfbfbf' }} />}
                placeholder="请输入密码"
                size="large"
              />
            </Form.Item>

            <div className="login-extra">
              <span></span>
              <a onClick={handleShowForgotPassword}>忘记密码？</a>
            </div>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                size="large"
                loading={loading}
                style={{ width: '100%', height: '45px', fontSize: '16px', borderRadius: '6px', backgroundColor: '#2d6a9f' }}
              >
                登 录
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>

      {/* 忘记密码弹窗 */}
      <Modal
        title="忘记密码"
        open={isForgotModalVisible}
        onCancel={() => setIsForgotModalVisible(false)}
        footer={null}
        width={480}
        centered
        destroyOnClose
      >
        {!isSubmitSuccess ? (
          <div style={{ paddingTop: '10px' }}>
            <p style={{ fontSize: '13px', color: '#555', lineHeight: '1.8', marginBottom: '16px' }}>
              如果您忘记密码，可以在此提交密码重置申请。提交后，
              <strong style={{ color: '#ff4d4f' }}>管理员将收到通知并为您重置新密码</strong>。
            </p>
            
            <div style={{ marginBottom: '24px' }}>
              <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600 }}>您的账号 *</label>
              <Select
                placeholder="-- 请选择需要重置密码的账号 --"
                style={{ width: '100%', height: '40px' }}
                value={selectedUser}
                onChange={(val) => setSelectedUser(val)}
              >
                {users.map((u) => (
                  <Option key={u.id} value={u.id}>
                    {u.username}（{u.role}）
                  </Option>
                ))}
              </Select>
            </div>

            <Button
              type="primary"
              onClick={handleForgotPasswordSubmit}
              loading={submitLoading}
              style={{ width: '100%', height: '42px', fontSize: '15px', borderRadius: '6px' }}
            >
              提交重置申请
            </Button>
          </div>
        ) : (
          <div className="forgot-result">
            <span className="forgot-success-icon">
              <CheckCircleOutlined />
            </span>
            <p style={{ fontSize: '16px', fontWeight: 600, color: '#52c41a', marginBottom: '12px' }}>
              提交成功！
            </p>
            <p>您的密码重置申请已成功提交。</p>
            <p>
              <strong>管理员将收到通知</strong>，处理完成后您可使用新密码登录。
            </p>
            <p style={{ color: '#8c8c8c', fontSize: '12px', marginTop: '12px' }}>
              如有疑问，请及时联系系统管理员。
            </p>
            <Button
              onClick={() => setIsForgotModalVisible(false)}
              style={{ marginTop: '20px', padding: '0 30px', height: '36px', borderRadius: '4px' }}
            >
              关闭
            </Button>
          </div>
        )}
      </Modal>
    </div>
  )
}

export default Login
