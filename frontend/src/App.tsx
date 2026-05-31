import { useState } from 'react'
import type { CustomerInfo, Product } from './types'

const initialProducts: Product[] = [
  { id: 'prayer-a', name: '祈祷A', unitPrice: 5000, quantity: 0 },
  { id: 'ofuda', name: '御札', unitPrice: 1000, quantity: 0 },
  { id: 'omamori', name: 'お守り', unitPrice: 800, quantity: 0 },
]

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

function App() {
  const [products, setProducts] = useState(initialProducts)
  const [customerInfo, setCustomerInfo] = useState(initialCustomerInfo)
  const [errors, setErrors] = useState<string[]>([])
  const [screen, setScreen] = useState<'input' | 'confirm' | 'complete'>('input')
  const [submitError, setSubmitError] = useState('')
  const selectedProducts = products.filter((product) => product.quantity > 0)
  const totalAmount = products.reduce(
    (total, product) => total + product.unitPrice * product.quantity,
    0,
  )

  const handleQuantityChange = (productId: string, quantity: number) => {
    const normalizedQuantity = Math.min(10, Math.max(0, quantity || 0))
    setProducts((currentProducts) =>
      currentProducts.map((product) =>
        product.id === productId
          ? { ...product, quantity: normalizedQuantity }
          : product,
      ),
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
        <p>現在は画面確認用のため、データの保存とExcel出力はまだ行いません。</p>
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
                  {formatPrice(product.unitPrice)} × {product.quantity} ={' '}
                  {formatPrice(product.unitPrice * product.quantity)}
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

  return (
    <main>
      <h1>申込み請求フォーム</h1>
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
                  <p>小計: {formatPrice(product.unitPrice * product.quantity)}</p>
                </div>
                <label>
                  数量
                  <input
                    type="number"
                    min="0"
                    max="10"
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
