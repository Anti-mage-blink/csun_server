import { Typography } from 'antd'
import styles from './index.module.css'

const { Title, Paragraph } = Typography

/**
 * 新建报价单模块（占位）
 * 待后续迭代按需求实现基本信息表单等具体功能
 */
const QuoteCreate = () => {
  return (
    <div className={styles.container}>
      <Title level={4}>新建报价单</Title>
      <Paragraph type="secondary">模块开发中，待后续迭代实现。</Paragraph>
    </div>
  )
}

export default QuoteCreate
