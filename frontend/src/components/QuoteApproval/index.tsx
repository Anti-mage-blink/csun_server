import React, { useState, useEffect } from 'react'
import { 
  Table, 
  Tag, 
  Button, 
  Card, 
  Space, 
  Input, 
  message, 
  Empty, 
  Descriptions, 
  Steps, 
  Divider, 
  Typography,
  Popconfirm,
  Modal
} from 'antd'
import { 
  ArrowLeftOutlined, 
  CheckCircleOutlined, 
  CloseCircleOutlined, 
  ClockCircleOutlined,
  PlayCircleOutlined,
  FileTextOutlined,
  PrinterOutlined,
  RollbackOutlined
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { QuoteProcess, QuoteProcessNode, Quote, QuoteItem } from '@/api/quote'
import { parseFilenameFromPath, downloadCosFile } from '@/utils/file'
import companyLogo from './公司logo.png'
import './index.css'

const { TextArea } = Input
const { Title } = Typography

export interface QuoteApprovalProps {
  // 数据源
  quotes: Quote[];
  quoteItems: QuoteItem[];
  quoteProcesses: QuoteProcess[];
  quoteProcessNodes: QuoteProcessNode[];
  
  // 模式 / 变体
  // 'approve': 待我审批，过滤出当前用户需要审批的流程，显示审批操作面板并能触发 onApprove
  // 'my-submit': 我发起的审批，过滤出当前用户发起的流程，只读，不显示审批操作面板
  // 'filing': 备案查看，不过滤（全量），只读，不显示审批操作面板
  mode: 'approve' | 'my-submit' | 'filing';
  
  // 当前登录用户ID，用于 'approve' 过滤我的待审批节点，以及 'my-submit' 过滤我的发起记录
  currentUserId?: number;
  
  // 外部加载状态
  loading?: boolean;
  
  // 触发刷新数据回调
  onRefresh?: () => void;
  
  // 提交审批回调（仅在 mode === 'approve' 且点击同意/退回时触发）
  onApprove?: (
    process: QuoteProcess,
    currentNode: QuoteProcessNode,
    status: '已通过' | '已拒绝',
    comment: string
  ) => Promise<void> | void;

  // 撤回报价单回调（仅在 mode === 'my-submit' 且点击撤回时触发）
  onWithdraw?: (process: QuoteProcess) => Promise<void> | void;

  // 组件自定义大标题，若不提供则根据 mode 自动选择默认值
  title?: string;
}

const QuoteApproval: React.FC<QuoteApprovalProps> = ({
  quotes = [],
  quoteItems = [],
  quoteProcesses = [],
  quoteProcessNodes = [],
  mode,
  currentUserId,
  loading = false,
  onRefresh,
  onApprove,
  onWithdraw,
  title
}) => {
  // 页面切换控制 ('list' | 'detail')
  const [view, setView] = useState<'list' | 'detail'>('list')
  const [selectedProcess, setSelectedProcess] = useState<QuoteProcess | null>(null)
  
  // 审批/撤回操作状态
  const [approveComment, setApproveComment] = useState<string>('')
  const [submitting, setSubmitting] = useState<boolean>(false)
  const [withdrawing, setWithdrawing] = useState<boolean>(false)

  // 报价单打印预览 Modal 状态
  const [printModalVisible, setPrintModalVisible] = useState<boolean>(false)

  useEffect(() => {
    if (printModalVisible) {
      document.body.classList.add('modal-print-active')
    } else {
      document.body.classList.remove('modal-print-active')
    }
    return () => {
      document.body.classList.remove('modal-print-active')
    }
  }, [printModalVisible])

  // --------------------------------------------------------
  // 数据过滤与逻辑关联
  // --------------------------------------------------------
  
  // 1. 获取关联的 Quote 
  const getProcessQuote = (process: QuoteProcess): Quote | undefined => {
    return quotes.find(q => q.id === process.quote_id)
  }

  // 2. 获取该流程对应的当前审批/进行节点
  const getProcessCurrentNode = (process: QuoteProcess): QuoteProcessNode | undefined => {
    // 优先寻找该流程下状态为 "待审批" 的节点
    const pendingNode = quoteProcessNodes.find(n => n.process_id === process.id && n.status === '待审批')
    if (pendingNode) return pendingNode
    
    // 如果没有待审批节点，说明流程已全部走完或退回，返回 ID 最大的节点（代表最新节点状态）
    const processNodes = quoteProcessNodes.filter(n => n.process_id === process.id)
    if (processNodes.length > 0) {
      return [...processNodes].sort((a, b) => b.id - a.id)[0]
    }
    return undefined
  }

  // 3. 根据 mode 过滤与排序需要展示的审批流主表列表
  const getFilteredProcesses = (): QuoteProcess[] => {
    let list: QuoteProcess[] = []

    if (mode === 'my-submit' && currentUserId !== undefined) {
      // 我发起的审批：过滤出发起人是当前用户的流程
      list = quoteProcesses.filter(p => p.creator_id === currentUserId)
    } else if (mode === 'approve' && currentUserId !== undefined) {
      // 我的审批：后端已按节点的 approver_id == user.id 查出相关流程，全量展示
      list = quoteProcesses
    } else {
      // 备案查看（或其他情况）：不过滤，全量展示
      list = quoteProcesses
    }

    // 辅助获取报价日期毫秒戳
    const getQuoteTime = (process: QuoteProcess): number => {
      const quote = getProcessQuote(process)
      if (!quote || !quote.quote_date) return 0
      const t = dayjs(quote.quote_date).valueOf()
      return isNaN(t) ? 0 : t
    }

    // 辅助获取更新时间毫秒戳
    const getUpdateTime = (process: QuoteProcess): number => {
      const rawUpdate = process.updated_at || process.created_at
      if (!rawUpdate) {
        const currentNode = getProcessCurrentNode(process)
        const nodeTime = currentNode?.approve_at || currentNode?.created_at
        if (nodeTime) {
          const nt = dayjs(nodeTime).valueOf()
          if (!isNaN(nt)) return nt
        }
        return 0
      }
      const t = dayjs(rawUpdate).valueOf()
      return isNaN(t) ? 0 : t
    }

    // 按报价日期降序 -> 更新时间降序
    const compareByDateAndUpdated = (a: QuoteProcess, b: QuoteProcess): number => {
      const timeQuoteA = getQuoteTime(a)
      const timeQuoteB = getQuoteTime(b)
      if (timeQuoteA !== timeQuoteB) {
        return timeQuoteB - timeQuoteA // 降序（越新的日期靠前）
      }

      const timeUpdateA = getUpdateTime(a)
      const timeUpdateB = getUpdateTime(b)
      if (timeUpdateA !== timeUpdateB) {
        return timeUpdateB - timeUpdateA // 降序（越新的时间靠前）
      }

      return b.id - a.id
    }

    return [...list].sort((a, b) => {
      if (mode === 'filing') {
        // 备案查看 mode：默认排序先按报价日期（quote.quote_date）降序、对于报价日期相同再按更新时间（quote_process.updated_at）降序
        return compareByDateAndUpdated(a, b)
      } else {
        // 我的发起 (my-submit) 和 我的审批 (approve) mode：
        // 先分为“待审批”（quote_process.present_status）部分和其余部分（待审批部分在其余部分之前）
        const isPendingA = (a.present_status || '待审批') === '待审批'
        const isPendingB = (b.present_status || '待审批') === '待审批'

        if (isPendingA && !isPendingB) return -1
        if (!isPendingA && isPendingB) return 1

        // 内部再按报价日期降序、对于报价日期相同再按更新时间降序
        return compareByDateAndUpdated(a, b)
      }
    })
  }

  const filteredProcesses = getFilteredProcesses()

  // --------------------------------------------------------
  // 审批操作处理
  // --------------------------------------------------------
  const handleApproveAction = async (status: '已通过' | '已拒绝') => {
    if (!selectedProcess) return
    
    const processNodes = quoteProcessNodes.filter(n => n.process_id === selectedProcess.id)
    const isProcessPending = (selectedProcess.present_status || '待审批') === '待审批'
    const pendingNode = processNodes.find(n => n.status === '待审批')

    if (!isProcessPending || !pendingNode || (currentUserId !== undefined && pendingNode.approver_id !== currentUserId)) {
      message.error('当前流程非待您审批状态，无法操作')
      return
    }

    if (!onApprove) {
      message.warning('未配置审批回调函数')
      return
    }

    setSubmitting(true)
    try {
      await onApprove(selectedProcess, pendingNode, status, approveComment)
      
      // 审批成功后返回列表并清空输入
      setView('list')
      setSelectedProcess(null)
      setApproveComment('')
    } catch (err: any) {
      message.error(err.message || '提交审批失败，请重试')
    } finally {
      setSubmitting(false)
    }
  }

  // --------------------------------------------------------
  // 撤回操作处理
  // --------------------------------------------------------
  const handleWithdrawAction = async (process: QuoteProcess) => {
    if (!onWithdraw) return
    setWithdrawing(true)
    try {
      await onWithdraw(process)
    } finally {
      setWithdrawing(false)
    }
  }

  // --------------------------------------------------------
  // 页面一：渲染审批页面（列表总览）
  // --------------------------------------------------------
  const renderListView = () => {
    const mapProcessToRow = (process: QuoteProcess, isPendingSection: boolean = true) => {
      const quote = getProcessQuote(process)
      const currentNode = getProcessCurrentNode(process)
      const rawQuoteDate = quote?.quote_date || '';

      const rawTime = mode === 'approve' 
        ? (currentNode?.created_at || process.created_at || process.updated_at)
        : (process.updated_at || process.created_at || currentNode?.approve_at || currentNode?.created_at)

      const displayTime = rawTime ? dayjs(rawTime).format('YYYY/M/D HH:mm:ss') : '未知时间'

      return {
        key: process.id,
        process,
        quote,
        currentNode,
        isPendingSection,
        quote_code: quote?.quote_code || '未知编号',
        customer_name: quote?.customer_name || '未知客户',
        creator_name: quote?.creator_name || process.creator_name || '未知经办人',
        quote_date: rawQuoteDate ? dayjs(rawQuoteDate).format('YYYY-MM-DD') : '未知日期',
        currentNodeName: currentNode?.name || '审批完成',
        currentNodeStatus: currentNode?.status || '已结束',
        created_at: displayTime
      }
    }

    const columns: any[] = [
      {
        title: '报价单编号',
        dataIndex: 'quote_code',
        key: 'quote_code',
        align: 'center',
        render: (text: string) => <span className="highlight-code">{text}</span>
      },
      {
        title: '客户名称',
        dataIndex: 'customer_name',
        key: 'customer_name',
        align: 'center',
      },
      {
        title: '市场部经办人',
        dataIndex: 'creator_name',
        key: 'creator_name',
        align: 'center',
        render: (text: string) => <Tag color="blue">{text}</Tag>
      },
      {
        title: '报价日期',
        dataIndex: 'quote_date',
        key: 'quote_date',
        align: 'center',
      },
      {
        title: '审批阶段',
        dataIndex: 'currentNodeName',
        key: 'currentNodeName',
        align: 'center',
        render: (_: string, record: any) => {
          const status = record.process?.present_status || record.currentNodeStatus || '待审批'
          let color = 'orange'
          let icon = <ClockCircleOutlined />
          let className = ''
          if (status === '已通过' || status === '已同意') {
            color = 'green'
            icon = <CheckCircleOutlined />
          } else if (status === '已拒绝') {
            color = 'red'
            icon = <CloseCircleOutlined />
          } else if (status === '已撤回') {
            color = 'default'
            icon = <RollbackOutlined />
            className = 'tag-withdrawn'
          }
          return <Tag color={color} icon={icon} className={className}>{status}</Tag>
        }
      },
      {
        title: mode === 'approve' ? '接收时间' : '最后更新时间',
        dataIndex: 'created_at',
        key: 'created_at',
        align: 'center',
      },
      {
        title: '操作',
        key: 'action',
        align: 'center',
        render: (_: any, record: any) => {
          let btnText = '查看详情'
          if (mode === 'approve') {
            btnText = record.isPendingSection ? '去审批' : '查看详情'
          }
          return (
            <Button 
              type="primary" 
              size="small" 
              onClick={() => {
                setSelectedProcess(record.process)
                setView('detail')
              }}
            >
              {btnText}
            </Button>
          )
        }
      }
    ]

    // 确定默认标题
    const displayTitle = title || (
      mode === 'approve' ? '我的审批' :
      mode === 'my-submit' ? '我发起的审批' : '备案查看'
    )

    if (mode === 'approve') {
      // 在“我的审批”模式下，对于每一簇报价数据，先从 quote_process_node 列表中找到 approver_id == user.id 的节点：
      // 若该节点的 status === '待审批' →→ 归入 “待审批” 列表。
      // 若该节点的 status !== '待审批'（如 '已通过'、'已拒绝'）→→ 归入 “已审批 (非待审批)” 列表。
      const isUserPendingProcess = (process: QuoteProcess): boolean => {
        const userNodes = quoteProcessNodes.filter(n => 
          n.process_id === process.id && 
          (currentUserId === undefined || n.approver_id === currentUserId)
        )
        return userNodes.some(n => n.status === '待审批')
      }

      const pendingProcesses = filteredProcesses.filter(p => isUserPendingProcess(p))
      const approvedProcesses = filteredProcesses.filter(p => !isUserPendingProcess(p))

      const pendingDataSource = pendingProcesses.map(p => mapProcessToRow(p, true))
      const approvedDataSource = approvedProcesses.map(p => mapProcessToRow(p, false))

      return (
        <div className="need-approve-view">
          <div className="header-bar">
            <Title level={4}>
              <FileTextOutlined className="icon-margin" /> {displayTitle}
              <span className="total-badge">{pendingProcesses.length}</span>
            </Title>
            {onRefresh && (
              <Button type="default" onClick={onRefresh} loading={loading}>
                刷新
              </Button>
            )}
          </div>

          <Card title="待审批" className="table-card mb-16">
            <Table 
              dataSource={pendingDataSource} 
              columns={columns} 
              loading={loading}
              scroll={{ x: 'max-content' }}
              locale={{ 
                emptyText: <Empty description="当前没有待审批的报价单" /> 
              }}
              pagination={{ pageSize: 10 }}
            />
          </Card>

          <Card title="已审批" className="table-card">
            <Table 
              dataSource={approvedDataSource} 
              columns={columns} 
              loading={loading}
              scroll={{ x: 'max-content' }}
              locale={{ 
                emptyText: <Empty description="当前没有已审批的记录" /> 
              }}
              pagination={{ pageSize: 10 }}
            />
          </Card>
        </div>
      )
    }

    const listDataSource = filteredProcesses.map(p => mapProcessToRow(p, false))

    return (
      <div className="need-approve-view">
        <div className="header-bar">
          <Title level={4}>
            <FileTextOutlined className="icon-margin" /> {displayTitle}
            <span className="total-badge">{filteredProcesses.length}</span>
          </Title>
          {onRefresh && (
            <Button type="default" onClick={onRefresh} loading={loading}>
              刷新
            </Button>
          )}
        </div>

        <Card className="table-card">
          <Table 
            dataSource={listDataSource} 
            columns={columns} 
            loading={loading}
            scroll={{ x: 'max-content' }}
            locale={{ 
              emptyText: <Empty description="当前没有相关的报价审批单数据" /> 
            }}
            pagination={{ pageSize: 10 }}
          />
        </Card>
      </div>
    )
  }

  // --------------------------------------------------------
  // 页面二：渲染审批详情及审批流页面
  // --------------------------------------------------------
  const renderDetailView = () => {
    if (!selectedProcess) return null

    // 优先从最新列表数据中查寻最新状态的流程记录，保持页面位置不变
    const currentProcess = quoteProcesses.find(p => p.id === selectedProcess.id) || selectedProcess
    const quote = getProcessQuote(currentProcess)
    
    // 获取当前报价单关联的所有明细单
    const items = quoteItems.filter(item => item.quote_id === currentProcess.quote_id)
    
    // 获取该审批流程的所有节点列表，并优先按 seq_num 顺序排序，展现完整的审批流轨迹
    const nodes = quoteProcessNodes
      .filter(node => node.process_id === currentProcess.id)
      .sort((a, b) => {
        const seqA = a.seq_num !== undefined && a.seq_num !== null ? a.seq_num : a.id;
        const seqB = b.seq_num !== undefined && b.seq_num !== null ? b.seq_num : b.id;
        return seqA - seqB;
      })

    // 确定 Steps 步骤条当前所在的位置
    const currentStepIndex = nodes.findIndex(node => node.status === '待审批')
    
    // 判断流程是否处于待审批状态
    const isProcessPending = (currentProcess.present_status || '待审批') === '待审批'
    
    // 找到该流程中处于“待审批”状态的节点
    const currentPendingNode = nodes.find(node => node.status === '待审批')

    // 当前我的审批节点（仅在 mode === 'approve'、流程处于待审批状态、且当前待审批节点的审批人是当前登录用户时才成立）
    const myCurrentPendingNode = (
      mode === 'approve' && 
      isProcessPending && 
      currentPendingNode && 
      currentUserId !== undefined && 
      currentPendingNode.approver_id === currentUserId
    ) ? currentPendingNode : undefined

    // 判断是否满足“打印报价单”按钮显示条件：仅在我的发起 mode，且 present_status 为“已通过”或“已同意”
    const processStatus = currentProcess.present_status || ''
    const canPrintQuoteSheet = mode === 'my-submit' && (processStatus === '已通过' || processStatus === '已同意')

    // 格式化日期：quote_process.updated_at 转化为 YYYY-MM-DD
    const rawDate = currentProcess.updated_at || currentProcess.created_at || quote?.quote_date
    const formattedDate = rawDate ? dayjs(rawDate).format('YYYY-MM-DD') : ''
    const formattedChineseDate = rawDate ? dayjs(rawDate).format('YYYY年M月D日') : ''

    // 渲染审批单状态 Tag（待审批: 黄色 / 已通过: 绿色 / 已拒绝: 红色 / 已撤回: 灰色）
    const renderProcessStatusTag = (status?: string | null) => {
      const statusVal = status || '待审批'
      let color = 'gold'
      let className = ''
      if (statusVal === '已通过' || statusVal === '已同意') {
        color = 'green'
      } else if (statusVal === '已拒绝') {
        color = 'red'
      } else if (statusVal === '已撤回') {
        color = 'default'
        className = 'tag-withdrawn'
      }
      return <Tag color={color} className={className}>{statusVal}</Tag>
    }

    // 明细表列定义
    const itemColumns = [
      {
        title: '产品大类',
        dataIndex: 'product_category_name',
        key: 'product_category_name',
      },
      {
        title: '产品名称',
        dataIndex: 'product_name',
        key: 'product_name',
      },
      {
        title: '规格型号',
        dataIndex: 'product_spec',
        key: 'product_spec',
      },
      {
        title: '批量档位',
        dataIndex: 'order_batch_tier',
        key: 'order_batch_tier',
        render: (tier: string) => {
          let color = 'default'
          if (tier === '大批量') color = 'green'
          if (tier === '中小批量') color = 'blue'
          if (tier === '样品/小单') color = 'orange'
          return <Tag color={color}>{tier}</Tag>
        }
      },
      {
        title: '目录基准价',
        dataIndex: 'catalog_base_price',
        key: 'catalog_base_price',
        render: (val: number) => `￥${val ? val.toFixed(2) : '0.00'}`
      },
      {
        title: '报价浮动比例',
        dataIndex: 'quote_float_rate',
        key: 'quote_float_rate',
        render: (val: number) => {
          if (val === undefined || val === null) return '0.00%'
          const percent = (val * 100).toFixed(2)
          return (
            <span style={{ color: val >= 0 ? '#52c41a' : '#f5222d', fontWeight: 'bold' }}>
              {val >= 0 ? `+${percent}%` : `${percent}%`}
            </span>
          )
        }
      },
      {
        title: '报价单价',
        dataIndex: 'quote_unit_price',
        key: 'quote_unit_price',
        render: (val: number) => <span className="price-text">￥{val ? val.toFixed(2) : '0.00'}</span>
      },
      {
        title: '数量',
        dataIndex: 'quantity',
        key: 'quantity',
      },
      {
        title: '总金额',
        dataIndex: 'total_amount',
        key: 'total_amount',
        render: (val: number) => (
          <span className="total-price">
            ￥{val ? val.toLocaleString(undefined, { minimumFractionDigits: 2 }) : '0.00'}
          </span>
        )
      },
      {
        title: '是否低于底线价',
        dataIndex: 'is_below_floor_price',
        key: 'is_below_floor_price',
        render: (isBelow: boolean) => isBelow ? <Tag color="error">低于底价</Tag> : <Tag color="success">正常</Tag>
      }
    ]

    return (
      <div className="approve-details-view animate-fade-in">
        <div className="back-bar no-print">
          <Button 
            type="link" 
            icon={<ArrowLeftOutlined />} 
            onClick={() => {
              setView('list')
              setSelectedProcess(null)
              setApproveComment('')
            }}
          >
            返回列表
          </Button>
          <span className="detail-title-code">审批详情 - {quote?.quote_code || '未知编号'}</span>
          <Space className="back-bar-actions" style={{ marginLeft: 'auto' }}>
            {mode === 'my-submit' && onWithdraw && (!currentProcess.present_status || currentProcess.present_status === '待审批') && (
              <Popconfirm
                title="确定撤回该报价单吗？"
                description="撤回后当前审批流将被终止"
                onConfirm={() => handleWithdrawAction(currentProcess)}
                okText="确定"
                cancelText="取消"
              >
                <Button 
                  type="default"
                  className="btn-withdraw"
                  icon={<RollbackOutlined />} 
                  loading={withdrawing}
                >
                  撤回
                </Button>
              </Popconfirm>
            )}
            {canPrintQuoteSheet && (
              <Button 
                type="primary" 
                icon={<PrinterOutlined />} 
                onClick={() => setPrintModalVisible(true)}
              >
                打印报价单
              </Button>
            )}
            <Button 
              type="primary" 
              icon={<PrinterOutlined />} 
              onClick={() => window.print()}
            >
              打印
            </Button>
          </Space>
        </div>

        <div className="detail-layout">
          {/* 左侧：报价及审批详情 */}
          <div className="detail-left">
            <Card title="报价单基本信息" className="mb-16 card-shadow">
              <Descriptions bordered column={{ xs: 1, sm: 2, md: 2 }} size="small">
                <Descriptions.Item label="报价单编号">
                  <span className="highlight-code">{quote?.quote_code}</span>
                </Descriptions.Item>
                <Descriptions.Item label="客户名称">{quote?.customer_name}</Descriptions.Item>
                <Descriptions.Item label="联系人">{quote?.contact_name || '无'}</Descriptions.Item>
                <Descriptions.Item label="职位">{quote?.contact_title || '无'}</Descriptions.Item>
                <Descriptions.Item label="市场部经办人">{quote?.creator_name || selectedProcess.creator_name}</Descriptions.Item>
                <Descriptions.Item label="报价日期">
                  {quote?.quote_date ? dayjs(quote.quote_date).format('YYYY-MM-DD') : '未知日期'}
                </Descriptions.Item>
                <Descriptions.Item label="报价有效期">{quote?.valid_days ? `${quote.valid_days} 天` : '无'}</Descriptions.Item>
                <Descriptions.Item label="付款方式">{quote?.pay_way || '无'}</Descriptions.Item>
                <Descriptions.Item label="账期">{quote?.credit_period || '无'}</Descriptions.Item>
                <Descriptions.Item label="说明" span={2}>{quote?.remarks || '无'}</Descriptions.Item>
                <Descriptions.Item label="附件" span={2}>
                  {(() => {
                    const raw = quote?.attachment_path_array;
                    let list: string[] = [];
                    if (Array.isArray(raw)) {
                      list = raw;
                    } else if (typeof raw === 'string' && raw.trim()) {
                      try {
                        const parsed = JSON.parse(raw);
                        if (Array.isArray(parsed)) list = parsed;
                        else list = [raw];
                      } catch {
                        list = [raw];
                      }
                    }
                    if (list.length === 0) {
                      return <span style={{ color: '#8c8c8c' }}>无</span>;
                    }
                    return (
                      <Space direction="vertical" size={2}>
                        {list.map((item, idx) => {
                          const fileName = parseFilenameFromPath(item);
                          return (
                            <a
                              key={idx}
                              onClick={(e) => {
                                e.preventDefault();
                                downloadCosFile(item);
                              }}
                              style={{ cursor: 'pointer', textDecoration: 'underline' }}
                              title={`点击下载附件: ${fileName}`}
                            >
                              {fileName}
                            </a>
                          );
                        })}
                      </Space>
                    );
                  })()}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="报价明细" className="mb-16 card-shadow detail-quote-items-card">
              <Table 
                dataSource={items} 
                columns={itemColumns} 
                rowKey="id" 
                pagination={false} 
                size="small"
              />
            </Card>

            {/* 仅在待审批模式，且确实有待处理节点时，才显示审批动作卡片 */}
            {mode === 'approve' && myCurrentPendingNode && (
              <Card title={`审批操作 (${myCurrentPendingNode.name})`} className="card-shadow">
                <div className="approve-action-section">
                  <div className="textarea-label">填写审批意见：</div>
                  <TextArea 
                    rows={4} 
                    value={approveComment}
                    onChange={(e) => setApproveComment(e.target.value)}
                    placeholder="请输入您的审批意见或退回原因..."
                    maxLength={200}
                    showCount
                  />
                  <Divider />
                  <Space size="middle" className="action-buttons">
                    <Button 
                      type="primary" 
                      icon={<CheckCircleOutlined />} 
                      className="btn-pass"
                      onClick={() => handleApproveAction('已通过')}
                      loading={submitting}
                    >
                      同意通过
                    </Button>
                    <Button 
                      danger 
                      type="primary"
                      icon={<CloseCircleOutlined />} 
                      onClick={() => handleApproveAction('已拒绝')}
                      loading={submitting}
                    >
                      拒绝退回
                    </Button>
                  </Space>
                </div>
              </Card>
            )}
          </div>

          {/* 右侧：整个审批流节点跟踪 */}
          <div className="detail-right-steps">
            <Card 
              title={
                <Space align="center">
                  <span>报价审批流跟踪</span>
                  {renderProcessStatusTag(currentProcess.present_status)}
                </Space>
              } 
              className="card-shadow full-height-card"
            >
              {nodes.length === 0 ? (
                <Empty description="暂无审批流程记录" />
              ) : (
                <Steps
                  direction="vertical"
                  current={currentStepIndex === -1 ? nodes.length : currentStepIndex}
                  status={currentProcess.present_status === '已撤回' ? 'wait' : (currentStepIndex === -1 ? 'finish' : 'process')}
                  size="small"
                >
                  {nodes.map((node) => {
                    let stepStatus: 'wait' | 'process' | 'finish' | 'error' = 'wait'
                    let icon = <ClockCircleOutlined />
                    let titleColor = 'gray'

                    const isWithdrawn = currentProcess.present_status === '已撤回'
                    const statusStr = node.status || ''
                    const isPassed = statusStr === '已通过' || statusStr === '已同意' || statusStr.includes('通过') || statusStr.includes('同意')
                    const isPending = statusStr === '待审批' || statusStr.includes('待')
                    const isRejected = statusStr === '已拒绝' || statusStr.includes('拒绝') || statusStr.includes('退回')

                    if (isWithdrawn) {
                      stepStatus = 'wait'
                      titleColor = '#8c8c8c'
                      if (isPassed) {
                        icon = <CheckCircleOutlined style={{ color: '#8c8c8c' }} />
                      } else if (isRejected) {
                        icon = <CloseCircleOutlined style={{ color: '#8c8c8c' }} />
                      } else if (isPending) {
                        icon = <PlayCircleOutlined style={{ color: '#8c8c8c' }} />
                      } else {
                        icon = <ClockCircleOutlined style={{ color: '#8c8c8c' }} />
                      }
                    } else if (isPassed) {
                      stepStatus = 'finish'
                      icon = <CheckCircleOutlined style={{ color: '#52c41a' }} />
                      titleColor = '#52c41a'
                    } else if (isPending) {
                      stepStatus = 'process'
                      icon = <PlayCircleOutlined style={{ color: '#1890ff' }} />
                      titleColor = '#1890ff'
                    } else if (isRejected) {
                      stepStatus = 'error'
                      icon = <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                      titleColor = '#ff4d4f'
                    }

                    return (
                      <Steps.Step
                        key={node.id}
                        status={stepStatus}
                        icon={icon}
                        title={
                          <span style={{ color: titleColor, fontWeight: 'bold' }}>
                            {node.name}
                          </span>
                        }
                        description={
                          <div className="step-description-box">
                            <p><strong>处理人：</strong>{node.approver_name} <Tag>{node.status}</Tag></p>
                            {node.created_at && (
                              <p className="step-time"><strong>到达时间：</strong>{new Date(node.created_at).toLocaleString()}</p>
                            )}
                            {node.approve_at && (
                              <p className="step-time"><strong>处理时间：</strong>{new Date(node.approve_at).toLocaleString()}</p>
                            )}
                            {node.approve_comment && (
                              <div className="step-comment">
                                <strong>批注意见：</strong>{node.approve_comment}
                              </div>
                            )}
                          </div>
                        }
                      />
                    )
                  })}
                </Steps>
              )}
            </Card>
          </div>
        </div>

        {/* 报价单格式打印预览 Modal */}
        <Modal
          title="报价单预览"
          open={printModalVisible}
          onCancel={() => setPrintModalVisible(false)}
          width="92%"
          style={{ maxWidth: 850, top: 16 }}
          centered
          destroyOnClose
          footer={[
            <Button key="cancel" onClick={() => setPrintModalVisible(false)}>
              取消
            </Button>,
            <Button 
              key="print" 
              type="primary" 
              icon={<PrinterOutlined />} 
              onClick={() => window.print()}
            >
              打印
            </Button>
          ]}
          className="quote-print-preview-modal"
        >
          <div className="quote-print-sheet">
            {/* 页头：Logo & 标语 */}
            <div className="quote-sheet-header">
              <div className="logo-container">
                <img src={companyLogo} alt="公司Logo" className="company-logo-img" />
                <div className="slogan-box">
                  <div className="slogan-cn">使全球交通更安全、更节能</div>
                  <div className="slogan-en">MAKING GLOBAL-TRAFFIC MORE SAFER AND ENERGY-EFFICIENT</div>
                </div>
              </div>
            </div>

            <div className="quote-sheet-divider" />

            {/* 基本信息网格 */}
            <div className="quote-sheet-info-grid">
              <div className="info-row">
                <div className="info-cell">
                  <span className="info-label">收件单位：</span>
                  <span className="info-value underline">{quote?.customer_name || ''}</span>
                </div>
                <div className="info-cell">
                  <span className="info-label">发件人：</span>
                  <span className="info-value underline">{quote?.creator_name || currentProcess.creator_name || ''}</span>
                </div>
              </div>

              <div className="info-row">
                <div className="info-cell">
                  <span className="info-label">收件人：</span>
                  <span className="info-value underline">{quote?.contact_name || ''}</span>
                </div>
                <div className="info-cell">
                  <span className="info-label">页数：</span>
                  <span className="info-value underline">1页</span>
                </div>
              </div>

              <div className="info-row">
                <div className="info-cell">
                  <span className="info-label">传真：</span>
                  <span className="info-value underline"></span>
                </div>
                <div className="info-cell">
                  <span className="info-label">日期：</span>
                  <span className="info-value underline">{formattedDate}</span>
                </div>
              </div>

              <div className="info-row">
                <div className="info-cell">
                  <span className="info-label">电话：</span>
                  <span className="info-value underline"></span>
                </div>
                <div className="info-cell">
                  <span className="info-label">抄送：</span>
                  <span className="info-value underline"></span>
                </div>
              </div>

              <div className="info-row">
                <div className="info-cell">
                  <span className="info-label">关于：</span>
                  <span className="info-value underline">产品报价</span>
                </div>
                <div className="info-cell">
                  <span className="info-label">签发：</span>
                  <span className="info-value underline"></span>
                </div>
              </div>
            </div>

            <div className="quote-sheet-divider" />

            {/* 产品名称、型号、数量、金额 */}
            <div className="quote-sheet-table-title">
              产品名称、型号、数量、金额：
            </div>

            {/* 报价明细表格 */}
            <table className="quote-sheet-table">
              <thead>
                <tr>
                  <th style={{ width: '50px' }}>序号</th>
                  <th>产品名称</th>
                  <th>规格型号</th>
                  <th style={{ width: '60px' }}>单位</th>
                  <th style={{ width: '60px' }}>数量</th>
                  <th>含税单价（元/件）</th>
                  <th>含税总价（元）</th>
                </tr>
              </thead>
              <tbody>
                {items.map((item, idx) => (
                  <tr key={item.id || idx}>
                    <td>{idx + 1}</td>
                    <td>{item.product_name || ''}</td>
                    <td>{item.product_spec || ''}</td>
                    <td>件</td>
                    <td>{item.quantity ?? ''}</td>
                    <td>{item.quote_unit_price !== undefined && item.quote_unit_price !== null ? item.quote_unit_price : ''}</td>
                    <td>{item.total_amount !== undefined && item.total_amount !== null ? item.total_amount : ''}</td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* 备注 */}
            <div className="quote-sheet-remarks">
              <div className="remarks-row">
                <span className="remarks-label">备注：</span>
                <div className="remarks-list">
                  <div>1、本次报价为含税价（税率 13%）</div>
                  <div>2、付款方式：{quote?.pay_way || ''}</div>
                  <div>3、报价有效期：{quote?.valid_days ? `${quote.valid_days} 天` : ''}</div>
                </div>
              </div>
            </div>

            {/* 右下角落款 */}
            <div className="quote-sheet-footer">
              <div className="company-name">湖南世鑫新材料有限公司</div>
              <div className="footer-date">{formattedChineseDate}</div>
            </div>
          </div>
        </Modal>
      </div>
    )
  }

  // --------------------------------------------------------
  // 渲染视图总控
  // --------------------------------------------------------
  return (
    <div className="approve-container">
      {view === 'list' ? renderListView() : renderDetailView()}
    </div>
  )
}

export default QuoteApproval
