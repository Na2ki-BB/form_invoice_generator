import { useState } from 'react'
import type { Product } from './types'

const initialProducts: Product[] = [
  { id: 'prayer-a', name: '祈祷A', unitPrice: 5000, quantity: 0 },
  { id: 'ofuda', name: '御札', unitPrice: 1000, quantity: 0 },
  { id: 'omamori', name: 'お守り', unitPrice: 800, quantity: 0 },
]

const formatPrice = (price: number) => `${price.toLocaleString('ja-JP')}円`

function App() {
  const [products, setProducts] = useState(initialProducts)
  const totalAmount = products.reduce(
    (total, product) => total + product.unitPrice * product.quantity,
    0,
  )

  const handleQuantityChange = (productId: string, quantity: number) => {
    setProducts((currentProducts) =>
      currentProducts.map((product) =>
        product.id === productId ? { ...product, quantity } : product,
      ),
    )
  }
  return (
    <main>
      <h1>申込み請求フォーム</h1>
      <p>商品ごとに数量を入力してください。</p>

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
            <input type="text" name="customerName" required />
          </label>
          <label>
            ふりがな
            <input type="text" name="customerKana" />
          </label>
          <label>
            郵便番号 <span>必須</span>
            <input type="text" name="postalCode" required />
          </label>
          <label>
            住所 <span>必須</span>
            <input type="text" name="address" required />
          </label>
          <label>
            電話番号 <span>必須</span>
            <input type="tel" name="phone" required />
          </label>
          <label>
            メールアドレス
            <input type="email" name="email" />
          </label>
          <label>
            備考
            <textarea name="note" rows={4} />
          </label>
        </div>
      </section>
    </main>
  )
}

export default App
