import React, { useEffect, useState } from 'react'
import { Spin } from 'antd'
import { useAuth } from '@/AuthContext'
import { myApplyQueryApi, withdrawQuoteApi, Quote, QuoteItem, QuoteProcess, QuoteProcessNode } from '@/api/quote'
import QuoteApproval from '@/components/QuoteApproval'
import Feedback from '@/components/Feedback'
import './index.css'

const MyApplications: React.FC = () => {
  const { user } = useAuth()
  const [loading, setLoading] = useState<boolean>(false)
  const [data, setData] = useState<{
    quotes: Quote[]
    quoteItems: QuoteItem[]
    quoteProcesses: QuoteProcess[]
    quoteProcessNodes: QuoteProcessNode[]
  }>({
    quotes: [],
    quoteItems: [],
    quoteProcesses: [],
    quoteProcessNodes: []
  })

  const loadData = async (active = true) => {
    if (!user) return
    setLoading(true)
    try {
      const res = await myApplyQueryApi(user.id)
      if (!active) return
      
      // 使用 Feedback.handle 对成功获取进行可能的提示（可选，因为通常只提示错误或保存成功，但这里可遵循用户指定“调用返回传递给FeedBack组件”）
      // 这里我们在出错时进行提示，查询成功若有特定后端返回消息也可以自动用 Feedback 过滤
      Feedback.handle(res, undefined, '获取我的发起数据失败')
      
      setData({
        quotes: res.data.quotes,
        quoteItems: res.data.quote_items,
        quoteProcesses: res.data.quote_processes,
        quoteProcessNodes: res.data.quote_process_nodes
      })
    } catch (err: any) {
      if (!active) return
      Feedback.handle(err, undefined, '获取我的发起数据失败')
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

  const handleWithdraw = async (process: QuoteProcess) => {
    if (!user) return
    try {
      const userName = user.name || ''
      const res = await withdrawQuoteApi({
        process_id: process.id,
        user: {
          id: user.id,
          name: userName
        }
      })
      Feedback.handle(res, '撤回报价单成功', '撤回报价单失败')
      await loadData(true)
    } catch (err: any) {
      Feedback.handle(err, undefined, '撤回报价单失败')
    }
  }

  return (
    <div className="my-applications-page-wrapper">
      {loading && data.quotes.length === 0 ? (
        <div className="my-applications-loading-container">
          <Spin size="large" tip="正在加载我的发起数据，请稍候..." />
        </div>
      ) : (
        <QuoteApproval
          quotes={data.quotes}
          quoteItems={data.quoteItems}
          quoteProcesses={data.quoteProcesses}
          quoteProcessNodes={data.quoteProcessNodes}
          mode="my-submit"
          currentUserId={user?.id}
          loading={loading}
          onRefresh={loadData}
          onWithdraw={handleWithdraw}
          title="我发起的申请"
        />
      )}
    </div>
  )
}

export default MyApplications
