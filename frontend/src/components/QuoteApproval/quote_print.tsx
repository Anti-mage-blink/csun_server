import React from 'react'
import dayjs from 'dayjs'
import { Quote, QuoteItem, QuoteProcess } from '@/api/quote'
import companyLogo from './公司logo.png'

export interface QuotePrintProps {
  quote?: Quote | null
  items?: QuoteItem[]
  process?: QuoteProcess | null
}

const QuotePrint: React.FC<QuotePrintProps> = ({
  quote,
  items = [],
  process
}) => {
  const rawDate = process?.updated_at || process?.created_at || quote?.quote_date
  const formattedDate = rawDate ? dayjs(rawDate).format('YYYY-MM-DD') : ''
  const formattedChineseDate = rawDate ? dayjs(rawDate).format('YYYY年M月D日') : ''
  const creatorName = quote?.creator_name || process?.creator_name || ''

  return (
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
            <span className="info-value underline">{creatorName}</span>
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
            {quote?.remarks?.trim() ? <div>4、备注：{quote.remarks}</div> : null}
          </div>
        </div>
      </div>

      {/* 右下角落款 */}
      <div className="quote-sheet-footer">
        <div className="company-name">湖南世鑫新材料有限公司</div>
        <div className="footer-date">{formattedChineseDate}</div>
      </div>
    </div>
  )
}

export default QuotePrint
