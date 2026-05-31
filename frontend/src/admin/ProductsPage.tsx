import { useEffect, useState } from 'react'

type Product = {
  id: string
  name: string
  description: string
  category: string
  baseUnitPrice: number
  isActive: boolean
  sortOrder: number
}

const emptyProduct: Product = {
  id: '',
  name: '',
  description: '',
  category: 'その他',
  baseUnitPrice: 0,
  isActive: true,
  sortOrder: 0,
}

function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [editingProduct, setEditingProduct] = useState<Product>(emptyProduct)
  const [isNew, setIsNew] = useState(true)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const loadProducts = async () => {
    const response = await fetch('http://127.0.0.1:8080/admin/products', { headers: { 'X-Local-Admin': 'true' } })
    if (!response.ok) throw new Error('products loading failed')
    setProducts(await response.json())
  }

  useEffect(() => {
    loadProducts().catch(() => setError('商品一覧の読み込みに失敗しました。'))
  }, [])

  const updateField = <K extends keyof Product>(field: K, value: Product[K]) => {
    setEditingProduct((currentProduct) => ({ ...currentProduct, [field]: value }))
  }

  const saveProduct = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setMessage('')
    setError('')

    try {
      const response = await fetch('http://127.0.0.1:8080/admin/products', {
        method: isNew ? 'POST' : 'PUT',
        headers: { 'Content-Type': 'application/json', 'X-Local-Admin': 'true' },
        body: JSON.stringify(editingProduct),
      })
      if (!response.ok) throw new Error('product saving failed')

      await loadProducts()
      setEditingProduct(emptyProduct)
      setIsNew(true)
      setMessage('商品を保存しました。')
    } catch {
      setError('商品の保存に失敗しました。入力内容を確認してください。')
    }
  }

  return (
    <main>
      <h1>商品管理</h1>
      {message && <p>{message}</p>}
      {error && <p className="error-message" role="alert">{error}</p>}

      <table className="submissions-table">
        <thead><tr><th>商品名</th><th>カテゴリ</th><th>税込単価</th><th>状態</th><th>操作</th></tr></thead>
        <tbody>
          {products.map((product) => (
            <tr key={product.id}>
              <td>{product.name}</td><td>{product.category}</td><td>{product.baseUnitPrice.toLocaleString('ja-JP')}円</td>
              <td>{product.isActive ? '有効' : '無効'}</td>
              <td><button type="button" onClick={() => { setEditingProduct(product); setIsNew(false) }}>編集</button></td>
            </tr>
          ))}
        </tbody>
      </table>

      <h2>{isNew ? '商品を登録' : '商品を編集'}</h2>
      <form className="customer-fields" onSubmit={saveProduct}>
        <label>商品ID<input required disabled={!isNew} value={editingProduct.id} onChange={(event) => updateField('id', event.target.value)} /></label>
        <label>商品名<input required value={editingProduct.name} onChange={(event) => updateField('name', event.target.value)} /></label>
        <label>説明<textarea value={editingProduct.description} onChange={(event) => updateField('description', event.target.value)} /></label>
        <label>カテゴリ<input required value={editingProduct.category} onChange={(event) => updateField('category', event.target.value)} /></label>
        <label>税込単価<input required type="number" min="0" value={editingProduct.baseUnitPrice} onChange={(event) => updateField('baseUnitPrice', Number(event.target.value))} /></label>
        <label>並び順<input required type="number" value={editingProduct.sortOrder} onChange={(event) => updateField('sortOrder', Number(event.target.value))} /></label>
        <label><input type="checkbox" checked={editingProduct.isActive} onChange={(event) => updateField('isActive', event.target.checked)} /> 有効</label>
        <div className="button-row">
          {!isNew && <button type="button" className="secondary-button" onClick={() => { setEditingProduct(emptyProduct); setIsNew(true) }}>新規登録へ戻る</button>}
          <button type="submit">保存する</button>
        </div>
      </form>
    </main>
  )
}

export default ProductsPage
