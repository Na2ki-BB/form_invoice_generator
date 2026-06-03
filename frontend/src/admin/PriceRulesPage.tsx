import { useEffect, useState } from 'react'

type Product = { id: string; name: string }
type RuleType = 'fixed_total' | 'tier_unit'
type PriceRule = {
  id: number
  productId: string
  ruleType: RuleType
  minQuantity: number
  maxQuantity: number | null
  unitPrice: number | null
  totalPrice: number | null
  priority: number
  isActive: boolean
}

const emptyRule = (productId: string): PriceRule => ({
  id: 0,
  productId,
  ruleType: 'fixed_total',
  minQuantity: 1,
  maxQuantity: 1,
  unitPrice: null,
  totalPrice: 0,
  priority: 0,
  isActive: true,
})

function PriceRulesPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [selectedProductId, setSelectedProductId] = useState('')
  const [rules, setRules] = useState<PriceRule[]>([])
  const [editingRule, setEditingRule] = useState<PriceRule>(emptyRule(''))
  const [isNew, setIsNew] = useState(true)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    const loadProducts = async () => {
      const response = await fetch('http://127.0.0.1:8080/admin/products', { headers: { 'X-Local-Admin': 'true' } })
      if (!response.ok) throw new Error('products loading failed')
      const loadedProducts: Product[] = await response.json()
      setProducts(loadedProducts)
      if (loadedProducts.length > 0) setSelectedProductId(loadedProducts[0].id)
    }
    loadProducts().catch(() => setError('商品一覧の読み込みに失敗しました。'))
  }, [])

  useEffect(() => {
    if (!selectedProductId) return
    const loadRules = async () => {
      const response = await fetch(`http://127.0.0.1:8080/admin/price-rules?productId=${encodeURIComponent(selectedProductId)}`, { headers: { 'X-Local-Admin': 'true' } })
      if (!response.ok) throw new Error('price rules loading failed')
      setRules(await response.json())
      setEditingRule(emptyRule(selectedProductId))
      setIsNew(true)
    }
    loadRules().catch(() => setError('価格ルールの読み込みに失敗しました。'))
  }, [selectedProductId])

  const updateRule = <K extends keyof PriceRule>(field: K, value: PriceRule[K]) => {
    setEditingRule((currentRule) => ({ ...currentRule, [field]: value }))
  }

  const saveRule = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setMessage('')
    setError('')

    const ruleToSave: PriceRule = {
      ...editingRule,
      productId: selectedProductId,
      maxQuantity: editingRule.ruleType === 'fixed_total' ? editingRule.minQuantity : editingRule.maxQuantity,
      unitPrice: editingRule.ruleType === 'tier_unit' ? editingRule.unitPrice : null,
      totalPrice: editingRule.ruleType === 'fixed_total' ? editingRule.totalPrice : null,
    }

    try {
      const response = await fetch('http://127.0.0.1:8080/admin/price-rules', {
        method: isNew ? 'POST' : 'PUT',
        headers: { 'Content-Type': 'application/json', 'X-Local-Admin': 'true' },
        body: JSON.stringify(ruleToSave),
      })
      if (!response.ok) throw new Error(await response.text())
      const savedRule: PriceRule = await response.json()
      setRules((currentRules) => isNew
        ? [...currentRules, savedRule]
        : currentRules.map((rule) => (rule.id === savedRule.id ? savedRule : rule)))
      setEditingRule(emptyRule(selectedProductId))
      setIsNew(true)
      setMessage('価格ルールを保存しました。')
    } catch {
      setError('価格ルールの保存に失敗しました。入力内容を確認してください。')
    }
  }

  return (
    <main>
      <h1>価格ルール管理</h1>
      <p><a href="/admin/products">商品管理へ戻る</a></p>
      {message && <p>{message}</p>}
      {error && <p className="error-message" role="alert">{error}</p>}

      <label>
        商品
        <select value={selectedProductId} onChange={(event) => setSelectedProductId(event.target.value)}>
          {products.map((product) => <option key={product.id} value={product.id}>{product.name}</option>)}
        </select>
      </label>

      <table className="submissions-table">
        <thead><tr><th>種別</th><th>数量</th><th>単価</th><th>固定合計</th><th>優先度</th><th>状態</th><th>操作</th></tr></thead>
        <tbody>
          {rules.map((rule) => (
            <tr key={rule.id}>
              <td>{rule.ruleType}</td>
              <td>{rule.maxQuantity ? `${rule.minQuantity}〜${rule.maxQuantity}` : `${rule.minQuantity}〜`}</td>
              <td>{rule.unitPrice ?? '-'}</td>
              <td>{rule.totalPrice ?? '-'}</td>
              <td>{rule.priority}</td>
              <td>{rule.isActive ? '有効' : '無効'}</td>
              <td><button type="button" onClick={() => { setEditingRule(rule); setIsNew(false) }}>編集</button></td>
            </tr>
          ))}
        </tbody>
      </table>

      <h2>{isNew ? '価格ルールを登録' : '価格ルールを編集'}</h2>
      <form className="customer-fields" onSubmit={saveRule}>
        <label>種別
          <select value={editingRule.ruleType} onChange={(event) => updateRule('ruleType', event.target.value as RuleType)}>
            <option value="fixed_total">数量ごとの固定合計</option>
            <option value="tier_unit">数量による単価変更</option>
          </select>
        </label>
        <label>最小数量<input required type="number" min="1" value={editingRule.minQuantity} onChange={(event) => updateRule('minQuantity', Number(event.target.value))} /></label>
        {editingRule.ruleType === 'tier_unit' && <label>最大数量<input type="number" min={editingRule.minQuantity} value={editingRule.maxQuantity ?? ''} onChange={(event) => updateRule('maxQuantity', event.target.value === '' ? null : Number(event.target.value))} /></label>}
        {editingRule.ruleType === 'fixed_total' && <label>固定合計<input required type="number" min="0" value={editingRule.totalPrice ?? 0} onChange={(event) => updateRule('totalPrice', Number(event.target.value))} /></label>}
        {editingRule.ruleType === 'tier_unit' && <label>単価<input required type="number" min="0" value={editingRule.unitPrice ?? 0} onChange={(event) => updateRule('unitPrice', Number(event.target.value))} /></label>}
        <label>優先度<input required type="number" value={editingRule.priority} onChange={(event) => updateRule('priority', Number(event.target.value))} /></label>
        <label><input type="checkbox" checked={editingRule.isActive} onChange={(event) => updateRule('isActive', event.target.checked)} /> 有効</label>
        <div className="button-row">
          {!isNew && <button type="button" className="secondary-button" onClick={() => { setEditingRule(emptyRule(selectedProductId)); setIsNew(true) }}>新規登録へ戻る</button>}
          <button type="submit">保存する</button>
        </div>
      </form>
    </main>
  )
}

export default PriceRulesPage
