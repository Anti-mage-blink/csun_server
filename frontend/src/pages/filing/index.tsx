import React, { useEffect, useState } from 'react'
import { Spin } from 'antd'
import { filingLookApi, Quote, QuoteItem, QuoteProcess, QuoteProcessNode } from '@/api/quote'
import QuoteApproval from '@/components/QuoteApproval'
import Feedback from '@/components/Feedback'
import './index.css'

const Filing: React.FC = () => {
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
    setLoading(true)
    try {
      const res = await filingLookApi()
      if (!active) return
      setData({
        quotes: res.data.quotes,
        quoteItems: res.data.quote_items,
        quoteProcesses: res.data.quote_processes,
        quoteProcessNodes: res.data.quote_process_nodes
      })
    } catch (err: any) {
      if (!active) return
      Feedback.handle(err, undefined, '获取全量备案数据失败')
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
  }, [])

  return (
    <div className="filing-page-wrapper">
      {loading && data.quotes.length === 0 ? (
        <div className="filing-loading-container">
          <Spin size="large" tip="正在加载备案数据，请稍候..." />
        </div>
      ) : (
        <QuoteApproval
          quotes={data.quotes}
          quoteItems={data.quoteItems}
          quoteProcesses={data.quoteProcesses}
          quoteProcessNodes={data.quoteProcessNodes}
          mode="filing"
          loading={loading}
          onRefresh={loadData}
        />
      )}
    </div>
  )
}

export default Filing
