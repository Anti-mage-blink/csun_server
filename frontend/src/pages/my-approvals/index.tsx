import React, { useEffect, useState } from 'react'
import { Spin, Result } from 'antd'
import { useAuth } from '@/AuthContext'
import { 
  myApproveQueryApi, 
  approveHandleApi,
  QuoteProcess, 
  QuoteProcessNode, 
  Quote, 
  QuoteItem 
} from '@/api/quote'
import QuoteApproval from '@/components/QuoteApproval'
import Feedback from '@/components/Feedback'
import './index.css'

const MyApprovals: React.FC = () => {
  const { user } = useAuth()
  
  // 状态管理
  const [loading, setLoading] = useState<boolean>(false)
  const [data, setData] = useState<{
    total: number
    quote_processes: QuoteProcess[]
    quote_process_nodes: QuoteProcessNode[]
    quotes: Quote[]
    quote_items: QuoteItem[]
  } | null>(null)

  // 判断是否为审批角色
  const isApproverRole = (role: string | undefined): boolean => {
    if (!role) return false
    const approverRoles = ['工作小组组长-光伏热场', '工作小组组长-摩擦', '领导小组副组长', '领导小组组长', '组长', '副组长', '管理员', '上帝']
    return approverRoles.includes(role)
  }

  // 加载待审批数据
  const loadData = async (active = true) => {
    if (!user) return
    
    if (!isApproverRole(user.role)) {
      setLoading(false)
      return
    }

    setLoading(true)
    try {
      const res = await myApproveQueryApi(user.id)
      if (!active) return
      setData(res.data)
    } catch (err: any) {
      if (!active) return
      Feedback.handle(err, undefined, '获取待审批列表失败')
    } finally {
      if (active) setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    loadData(active)
    return () => {
      active = false
    }
  }, [user])

  // 处理审批操作提交 (将审批逻辑保留在子页面中，通过回调传递给通用组件)
  const handleApprove = async (
    process: QuoteProcess,
    currentNode: QuoteProcessNode,
    status: '已通过' | '已拒绝',
    comment: string
  ) => {
    try {
      // 调用真实后端接口，action 只能是 'approve' 或 'reject'
      const action: 'approve' | 'reject' = status === '已通过' ? 'approve' : 'reject'
      const response = await approveHandleApi({
        action,
        node_id: currentNode.id,
        process_id: process.id,
        comment: comment
      })

      // 使用统一反馈组件展示结果
      Feedback.handle(response)
      
      // 更新本地状态，完成审批后更新对应的 node 批注与状态，并将 present_node_id 记录在 process 上
      if (data) {
        const updatedNodes = data.quote_process_nodes.map(node => {
          if (node.id === currentNode.id) {
            return {
              ...node,
              status: status,
              approve_comment: comment, // comment 对应到 quote_process_node 的 approve_comment 字段
              approve_at: new Date().toISOString()
            }
          }
          return node
        })

        const updatedProcesses = data.quote_processes.map(p => {
          if (p.id === process.id) {
            return {
              ...p,
              present_node_id: currentNode.id // currentNode 对应到 quote_process 的 present_node_id 字段
            }
          }
          return p
        })

        setData({
          ...data,
          quote_process_nodes: updatedNodes,
          quote_processes: updatedProcesses,
          total: updatedProcesses.length
        })
      }

      // 重新加载全量待审批数据以保持数据同步
      loadData()
    } catch (err: any) {
      throw new Error(err.message || '提交审批操作失败')
    }
  }

  // 如果非审批人，显示友情提示
  if (user && !isApproverRole(user.role)) {
    return (
      <div className="my-approvals-container">
        <Result
          status="info"
          title="无需处理审批"
          subTitle={`您的角色为「${user.role}」，该角色不属于报价审批人范畴。目前仅支持「领导小组组长、领导小组副组长、工作小组组长」处理审批业务。`}
        />
      </div>
    )
  }

  return (
    <>
      {loading && !data ? (
        <div className="my-approvals-container">
          <div className="loading-container">
            <Spin size="large" tip="正在加载待审批数据，请稍候..." />
          </div>
        </div>
      ) : (
        <QuoteApproval
          quotes={data?.quotes || []}
          quoteItems={data?.quote_items || []}
          quoteProcesses={data?.quote_processes || []}
          quoteProcessNodes={data?.quote_process_nodes || []}
          mode="approve"
          currentUserId={user?.id}
          loading={loading}
          onRefresh={loadData}
          onApprove={handleApprove}
        />
      )}
    </>
  )
}

export default MyApprovals
