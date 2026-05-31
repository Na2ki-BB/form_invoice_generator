import { useEffect, useState } from 'react'

type Form = { id: number; title: string; description: string; publicSlug: string; isActive: boolean; productIds: string[] }
type Product = { id: string; name: string }
const emptyForm: Form = { id: 0, title: '', description: '', publicSlug: '', isActive: true, productIds: [] }

function FormsPage() {
  const [forms, setForms] = useState<Form[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [editingForm, setEditingForm] = useState<Form>(emptyForm)
  const [isNew, setIsNew] = useState(true)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')

  const loadData = async () => {
    const [formsResponse, productsResponse] = await Promise.all([
      fetch('http://127.0.0.1:8080/admin/forms', { headers: { 'X-Local-Admin': 'true' } }), fetch('http://127.0.0.1:8080/admin/products', { headers: { 'X-Local-Admin': 'true' } }),
    ])
    if (!formsResponse.ok || !productsResponse.ok) throw new Error('loading failed')
    setForms(await formsResponse.json())
    setProducts(await productsResponse.json())
  }
  useEffect(() => { loadData().catch(() => setError('フォーム一覧の読み込みに失敗しました。')) }, [])

  const saveForm = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setMessage(''); setError('')
    try {
      const response = await fetch('http://127.0.0.1:8080/admin/forms', { method: isNew ? 'POST' : 'PUT', headers: { 'Content-Type': 'application/json', 'X-Local-Admin': 'true' }, body: JSON.stringify(editingForm) })
      if (!response.ok) throw new Error('saving failed')
      await loadData(); setEditingForm(emptyForm); setIsNew(true); setMessage('フォームを保存しました。')
    } catch { setError('フォームの保存に失敗しました。') }
  }
  const toggleProduct = (productId: string) => setEditingForm((form) => ({ ...form, productIds: form.productIds.includes(productId) ? form.productIds.filter((id) => id !== productId) : [...form.productIds, productId] }))

  return <main>
    <h1>フォーム管理</h1>{message && <p>{message}</p>}{error && <p className="error-message">{error}</p>}
    <table className="submissions-table"><thead><tr><th>タイトル</th><th>公開URL</th><th>商品数</th><th>操作</th></tr></thead><tbody>{forms.map((form) => <tr key={form.id}><td>{form.title}</td><td>/forms/{form.publicSlug}</td><td>{form.productIds.length}</td><td><button type="button" onClick={() => { setEditingForm(form); setIsNew(false) }}>編集</button></td></tr>)}</tbody></table>
    <h2>{isNew ? 'フォームを登録' : 'フォームを編集'}</h2>
    <form className="customer-fields" onSubmit={saveForm}>
      <label>タイトル<input required value={editingForm.title} onChange={(event) => setEditingForm({ ...editingForm, title: event.target.value })} /></label>
      <label>公開slug<input required value={editingForm.publicSlug} onChange={(event) => setEditingForm({ ...editingForm, publicSlug: event.target.value })} /></label>
      <label>説明<textarea value={editingForm.description} onChange={(event) => setEditingForm({ ...editingForm, description: event.target.value })} /></label>
      <fieldset><legend>表示商品</legend>{products.map((product) => <label key={product.id}><input type="checkbox" checked={editingForm.productIds.includes(product.id)} onChange={() => toggleProduct(product.id)} /> {product.name}</label>)}</fieldset>
      <label><input type="checkbox" checked={editingForm.isActive} onChange={(event) => setEditingForm({ ...editingForm, isActive: event.target.checked })} /> 有効</label>
      <div className="button-row">{!isNew && <button type="button" className="secondary-button" onClick={() => { setEditingForm(emptyForm); setIsNew(true) }}>新規登録へ戻る</button>}<button type="submit">保存する</button></div>
    </form>
  </main>
}
export default FormsPage
