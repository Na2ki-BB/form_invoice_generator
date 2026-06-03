import { useEffect, useState } from 'react'
import { apiUrl } from '../config'

type SubmissionSummary = {
  id: number
  invoiceNumber: string
  formTitle: string
  formSlug: string
  customerName: string
  customerPhone: string
  totalAmount: number
  submittedAt: string
  status: string
}

const currentMonth = new Date().toISOString().slice(0, 7)
const formatPrice = (price: number) => `${price.toLocaleString('ja-JP')}円`

function SubmissionsPage() {
  const [month, setMonth] = useState(currentMonth)
  const [submissions, setSubmissions] = useState<SubmissionSummary[]>([])
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState('')
  const [downloadError, setDownloadError] = useState('')

  useEffect(() => {
    const loadSubmissions = async () => {
      setIsLoading(true)
      setError('')

      try {
        const response = await fetch(apiUrl(`/admin/submissions?month=${month}`), { headers: { 'X-Local-Admin': 'true' } })
        if (!response.ok) throw new Error('submissions loading failed')

        setSubmissions(await response.json())
        setSelectedIds([])
      } catch {
        setError('申込み一覧の読み込みに失敗しました。')
      } finally {
        setIsLoading(false)
      }
    }

    void loadSubmissions()
  }, [month])

  const allSelected = submissions.length > 0 && selectedIds.length === submissions.length

  const toggleAll = () => {
    setSelectedIds(allSelected ? [] : submissions.map((submission) => submission.id))
  }

  const toggleSubmission = (submissionId: number) => {
    setSelectedIds((currentIds) =>
      currentIds.includes(submissionId)
        ? currentIds.filter((id) => id !== submissionId)
        : [...currentIds, submissionId],
    )
  }

  const downloadSelectedInvoices = async () => {
    setDownloadError('')

    try {
      const response = await fetch(apiUrl('/admin/invoices/bulk-download'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'X-Local-Admin': 'true' },
        body: JSON.stringify({ submissionIds: selectedIds }),
      })
      if (!response.ok) throw new Error('invoice download failed')

      const archive = await response.blob()
      const downloadUrl = URL.createObjectURL(archive)
      const link = document.createElement('a')
      link.href = downloadUrl
      link.download = `${month}-invoices.zip`
      link.click()
      URL.revokeObjectURL(downloadUrl)
    } catch {
      setDownloadError('請求書ZIPのダウンロードに失敗しました。')
    }
  }

  return (
    <main>
      <h1>月別申込み一覧</h1>
      <label className="month-field">
        表示する月
        <input type="month" value={month} onChange={(event) => setMonth(event.target.value)} />
      </label>

      {isLoading && <p>申込み一覧を読み込んでいます。</p>}
      {error && <p className="error-message" role="alert">{error}</p>}
      {!isLoading && !error && submissions.length === 0 && <p>この月の申込みはありません。</p>}

      {!isLoading && !error && submissions.length > 0 && (
        <>
          <button type="button" onClick={toggleAll}>
            {allSelected ? '選択をすべて解除' : '表示中をすべて選択'}
          </button>
          <p>{selectedIds.length}件を選択中</p>
          <button type="button" disabled={selectedIds.length === 0} onClick={downloadSelectedInvoices}>
            選択した請求書をダウンロード
          </button>
          {downloadError && <p className="error-message" role="alert">{downloadError}</p>}
          <table className="submissions-table">
          <thead>
            <tr>
              <th>選択</th>
              <th>請求書番号</th>
              <th>フォーム</th>
              <th>受付日時</th>
              <th>氏名</th>
              <th>電話番号</th>
              <th>合計金額</th>
              <th>状態</th>
            </tr>
          </thead>
          <tbody>
            {submissions.map((submission) => (
              <tr key={submission.id}>
                <td>
                  <input
                    type="checkbox"
                    checked={selectedIds.includes(submission.id)}
                    onChange={() => toggleSubmission(submission.id)}
                    aria-label={`${submission.customerName}を選択`}
                  />
                </td>
                <td><a href={`/admin/submissions/${submission.id}`}>{submission.invoiceNumber}</a></td>
                <td>{submission.formTitle} ({submission.formSlug})</td>
                <td>{new Date(submission.submittedAt).toLocaleString('ja-JP')}</td>
                <td>{submission.customerName}</td>
                <td>{submission.customerPhone}</td>
                <td>{formatPrice(submission.totalAmount)}</td>
                <td>{submission.status}</td>
              </tr>
            ))}
          </tbody>
          </table>
        </>
      )}
    </main>
  )
}

export default SubmissionsPage
