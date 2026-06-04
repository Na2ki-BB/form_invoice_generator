import { useEffect, useState } from 'react'
import AdminHomeLink from './AdminHomeLink'
import { apiUrl } from '../config'

type Item = { productId: string; name: string; unitPrice: number; quantity: number; amount: number }
type Detail = {
  id: number
  invoiceNumber: string
  customerName: string
  postalCode: string
  address: string
  note: string
  submittedAt: string
  items: Item[]
}

const formatPrice = (price: number) => `${price.toLocaleString('ja-JP')}円`

function SubmissionDetailPage({ submissionId }: { submissionId: number }) {
  const [detail, setDetail] = useState<Detail | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    const loadDetail = async () => {
      const response = await fetch(apiUrl(`/admin/submissions/${submissionId}`), { headers: { 'X-Local-Admin': 'true' } })
      if (!response.ok) throw new Error('submission loading failed')
      setDetail(await response.json())
    }
    loadDetail().catch(() => setError('申込み詳細の読み込みに失敗しました。'))
  }, [submissionId])

  if (error) return <main><p className="error-message" role="alert">{error}</p></main>
  if (!detail) return <main><p>申込み詳細を読み込んでいます。</p></main>

  return (
    <main>
      <h1>申込み詳細</h1>
      <AdminHomeLink />
      <p><a href="/admin/submissions">申込み一覧へ戻る</a></p>
      <dl className="confirmation-list">
        <dt>請求書番号</dt><dd>{detail.invoiceNumber}</dd>
        <dt>受付日時</dt><dd>{new Date(detail.submittedAt).toLocaleString('ja-JP')}</dd>
        <dt>氏名</dt><dd>{detail.customerName}</dd>
        <dt>郵便番号</dt><dd>{detail.postalCode}</dd>
        <dt>住所</dt><dd>{detail.address}</dd>
        <dt>備考</dt><dd>{detail.note || '未入力'}</dd>
      </dl>
      <h2>明細</h2>
      <table className="submissions-table">
        <thead><tr><th>商品</th><th>数量</th><th>単価</th><th>金額</th></tr></thead>
        <tbody>{detail.items.map((item) => <tr key={item.productId}><td>{item.name}</td><td>{item.quantity}</td><td>{formatPrice(item.unitPrice)}</td><td>{formatPrice(item.amount)}</td></tr>)}</tbody>
      </table>
    </main>
  )
}

export default SubmissionDetailPage
