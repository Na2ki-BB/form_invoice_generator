import { useEffect, useState } from 'react'
import FormsPage from './admin/FormsPage'
import PriceRulesPage from './admin/PriceRulesPage'
import ProductsPage from './admin/ProductsPage'
import SubmissionDetailPage from './admin/SubmissionDetailPage'
import SubmissionsPage from './admin/SubmissionsPage'
import type { CustomerInfo, Product, PublicForm, Quote } from './types'

const initialCustomerInfo: CustomerInfo = {
  customerName: '',
  customerKana: '',
  postalCode: '',
  address: '',
  phone: '',
  email: '',
  note: '',
}

const formatPrice = (price: number) => `${price.toLocaleString('ja-JP')}円`
const emptyQuote: Quote = { items: [], totalAmount: 0 }

const getPublicFormSlug = () => {
  const pathParts = window.location.pathname.split('/')
  return pathParts.length === 3 && pathParts[1] === 'forms' && pathParts[2]
    ? pathParts[2]
    : 'default'
}

function App() {
  if (window.location.pathname === '/admin/submissions') {
    return <SubmissionsPage />
  }
  if (window.location.pathname.startsWith('/admin/submissions/')) {
    const submissionId = Number(window.location.pathname.replace('/admin/submissions/', ''))
    return <SubmissionDetailPage submissionId={submissionId} />
  }
  if (window.location.pathname === '/admin/products') {
    return <ProductsPage />
  }
  if (window.location.pathname === '/admin/price-rules') {
    return <PriceRulesPage />
  }
  if (window.location.pathname === '/admin/forms') {
    return <FormsPage />
  }

  return <PublicFormPage />
}

function PublicFormPage() {
  const formSlug = getPublicFormSlug()
  const [products, setProducts] = useState<Product[]>([])
  const [formTitle, setFormTitle] = useState('申込み請求フォーム')
  const [isLoading, setIsLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [customerInfo, setCustomerInfo] = useState(initialCustomerInfo)
  const [errors, setErrors] = useState<string[]>([])
  const [screen, setScreen] = useState<'input' | 'confirm' | 'complete'>('input')
  const [submitError, setSubmitError] = useState('')
  const [quote, setQuote] = useState<Quote>(emptyQuote)
  const [isQuoteLoading, setIsQuoteLoading] = useState(false)
  const [quoteError, setQuoteError] = useState('')
  const selectedProducts = products.filter((product) => product.quantity > 0)
  const findQuoteItem = (product: Product) =>
    quote.items.find((item) => item.productId === product.id && item.quantity === product.quantity)
  const getDisplayedUnitPrice = (product: Product) => findQuoteItem(product)?.unitPrice ?? product.unitPrice
  const getDisplayedAmount = (product: Product) => findQuoteItem(product)?.amount ?? product.unitPrice * product.quantity
  const hasCurrentQuote = selectedProducts.every((product) => Boolean(findQuoteItem(product)))
  const totalAmount = hasCurrentQuote
    ? quote.totalAmount
    : products.reduce((total, product) => total + product.unitPrice * product.quantity, 0)

  useEffect(() => {
    const loadForm = async () => {
      try {
        const response = await fetch(`http://127.0.0.1:8080/public/forms/${encodeURIComponent(formSlug)}`)
        if (!response.ok) throw new Error('form loading failed')

        const publicForm: PublicForm = await response.json()
        setFormTitle(publicForm.title)
        setProducts(publicForm.products.map((product) => ({ ...product, quantity: 0 })))
      } catch {
        setLoadError('商品情報の読み込みに失敗しました。時間をおいてもう一度お試しください。')
      } finally {
        setIsLoading(false)
      }
    }

    void loadForm()
  }, [formSlug])

  useEffect(() => {
    const requestedItems = products
      .filter((product) => product.quantity > 0)
      .map((product) => ({ productId: product.id, quantity: product.quantity }))
    if (requestedItems.length === 0) {
      setQuote(emptyQuote)
      setQuoteError('')
      setIsQuoteLoading(false)
      return
    }

    const controller = new AbortController()
    const loadQuote = async () => {
      setIsQuoteLoading(true)
      setQuoteError('')
      try {
        const response = await fetch(`http://127.0.0.1:8080/public/forms/${encodeURIComponent(formSlug)}/quote`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ items: requestedItems }),
          signal: controller.signal,
        })
        if (!response.ok) throw new Error('quote loading failed')
        setQuote(await response.json())
      } catch {
        if (!controller.signal.aborted) {
          setQuoteError('金額の計算に失敗しました。時間をおいてもう一度お試しください。')
        }
      } finally {
        if (!controller.signal.aborted) setIsQuoteLoading(false)
      }
    }

    void loadQuote()
    return () => controller.abort()
  }, [formSlug, products])

  const handleQuantityChange = (productId: string, quantity: number) => {
    setProducts((currentProducts) =>
      currentProducts.map((product) => {
        if (product.id !== productId) return product

        const normalizedQuantity = Math.min(
          product.maxQuantity,
          Math.max(product.minQuantity, quantity || 0),
        )
        return { ...product, quantity: normalizedQuantity }
      }),
    )
  }

  const handleCustomerInfoChange = (field: keyof CustomerInfo, value: string) => {
    setCustomerInfo((currentCustomerInfo) => ({
      ...currentCustomerInfo,
      [field]: value,
    }))
  }

  const validate = () => {
    const validationErrors: string[] = []

    if (!customerInfo.customerName.trim()) validationErrors.push('氏名を入力してください。')
    if (!customerInfo.postalCode.trim()) validationErrors.push('郵便番号を入力してください。')
    if (!customerInfo.address.trim()) validationErrors.push('住所を入力してください。')
    if (!customerInfo.phone.trim()) validationErrors.push('電話番号を入力してください。')
    if (!products.some((product) => product.quantity > 0)) {
      validationErrors.push('少なくとも1つの商品を選んでください。')
    } else if (isQuoteLoading) {
      validationErrors.push('金額を計算しています。少し待ってから確認してください。')
    } else if (quoteError || !hasCurrentQuote) {
      validationErrors.push('金額を確認できませんでした。数量を確認してもう一度お試しください。')
    }

    return validationErrors
  }

  const handleSubmission = async () => {
    setSubmitError('')

    try {
      const response = await fetch('http://127.0.0.1:8080/submissions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...customerInfo,
          formSlug,
          items: selectedProducts.map((product) => ({
            productId: product.id,
            quantity: product.quantity,
          })),
        }),
      })

      if (!response.ok) throw new Error('submission failed')

      const invoice = await response.blob()
      const downloadUrl = URL.createObjectURL(invoice)
      const link = document.createElement('a')
      link.href = downloadUrl
      link.download = 'invoice.xlsx'
      link.click()
      URL.revokeObjectURL(downloadUrl)
      setScreen('complete')
    } catch {
      setSubmitError('送信に失敗しました。時間をおいてもう一度お試しください。')
    }
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const validationErrors = validate()
    setErrors(validationErrors)

    if (validationErrors.length === 0) {
      setScreen('confirm')
    }
  }

  if (screen === 'complete') {
    return (
      <main>
        <h1>申込みが完了しました</h1>
        <p>申込み内容を受け付けました。</p>
        <p>申込み内容を保存し、Excel請求書をダウンロードしました。</p>
      </main>
    )
  }

  if (screen === 'confirm') {
    return (
      <main>
        <h1>申込み内容の確認</h1>
        <section>
          <h2>申込者情報</h2>
          <dl className="confirmation-list">
            <dt>氏名</dt><dd>{customerInfo.customerName}</dd>
            <dt>ふりがな</dt><dd>{customerInfo.customerKana || '未入力'}</dd>
            <dt>郵便番号</dt><dd>{customerInfo.postalCode}</dd>
            <dt>住所</dt><dd>{customerInfo.address}</dd>
            <dt>電話番号</dt><dd>{customerInfo.phone}</dd>
            <dt>メールアドレス</dt><dd>{customerInfo.email || '未入力'}</dd>
            <dt>備考</dt><dd>{customerInfo.note || '未入力'}</dd>
          </dl>
        </section>

        <section>
          <h2>商品</h2>
          <div className="product-list">
            {selectedProducts.map((product) => (
              <div className="product-row" key={product.id}>
                <strong>{product.name}</strong>
                <span>
                  {formatPrice(getDisplayedUnitPrice(product))} × {product.quantity} ={' '}
                  {formatPrice(getDisplayedAmount(product))}
                </span>
              </div>
            ))}
          </div>
        </section>

        <p className="total-amount">合計金額: {formatPrice(totalAmount)}</p>
        <div className="button-row">
          <button type="button" className="secondary-button" onClick={() => setScreen('input')}>戻る</button>
          <button type="button" onClick={handleSubmission}>送信する</button>
        </div>
        {submitError && <p className="error-message" role="alert">{submitError}</p>}
      </main>
    )
  }

  if (isLoading) {
    return <main><p>商品情報を読み込んでいます。</p></main>
  }

  if (loadError) {
    return <main><p className="error-message" role="alert">{loadError}</p></main>
  }

  return (
    <main>
      <h1>{formTitle}</h1>
      <p>商品ごとに数量を入力してください。</p>

      <form onSubmit={handleSubmit}>
        <section>
          <h2>商品</h2>
          <div className="product-list">
            {products.map((product) => (
              <div className="product-row" key={product.id}>
                <div>
                  <strong>{product.name}</strong>
                  <p>{formatPrice(product.unitPrice)}（税込）</p>
                  <p>小計: {formatPrice(getDisplayedAmount(product))}</p>
                </div>
                <label>
                  数量
                  <input
                    type="number"
                    min={product.minQuantity}
                    max={product.maxQuantity}
                    value={product.quantity}
                    onChange={(event) =>
                      handleQuantityChange(product.id, Number(event.target.value))
                    }
                  />
                </label>
              </div>
            ))}
          </div>
        </section>

        <p className="total-amount">合計金額: {formatPrice(totalAmount)}</p>
        {isQuoteLoading && <p>金額を計算しています。</p>}
        {quoteError && <p className="error-message" role="alert">{quoteError}</p>}

        <section>
          <h2>申込者情報</h2>
          <div className="customer-fields">
            <label>
              氏名 <span>必須</span>
              <input
                type="text"
                value={customerInfo.customerName}
                onChange={(event) => handleCustomerInfoChange('customerName', event.target.value)}
              />
            </label>
            <label>
              ふりがな
              <input
                type="text"
                value={customerInfo.customerKana}
                onChange={(event) => handleCustomerInfoChange('customerKana', event.target.value)}
              />
            </label>
            <label>
              郵便番号 <span>必須</span>
              <input
                type="text"
                value={customerInfo.postalCode}
                onChange={(event) => handleCustomerInfoChange('postalCode', event.target.value)}
              />
            </label>
            <label>
              住所 <span>必須</span>
              <input
                type="text"
                value={customerInfo.address}
                onChange={(event) => handleCustomerInfoChange('address', event.target.value)}
              />
            </label>
            <label>
              電話番号 <span>必須</span>
              <input
                type="tel"
                value={customerInfo.phone}
                onChange={(event) => handleCustomerInfoChange('phone', event.target.value)}
              />
            </label>
            <label>
              メールアドレス
              <input
                type="email"
                value={customerInfo.email}
                onChange={(event) => handleCustomerInfoChange('email', event.target.value)}
              />
            </label>
            <label>
              備考
              <textarea
                rows={4}
                value={customerInfo.note}
                onChange={(event) => handleCustomerInfoChange('note', event.target.value)}
              />
            </label>
          </div>
        </section>

        {errors.length > 0 && (
          <div className="error-message" role="alert">
            <p>入力内容を確認してください。</p>
            <ul>
              {errors.map((error) => <li key={error}>{error}</li>)}
            </ul>
          </div>
        )}

        <button type="submit">入力内容を確認する</button>
      </form>
    </main>
  )
}

export default App
