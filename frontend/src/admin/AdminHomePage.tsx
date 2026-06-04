const adminLinks = [
  {
    href: '/admin/submissions',
    title: '申込み一覧',
    description: '月別の申込み確認と請求書ZIPのダウンロードを行います。',
  },
  {
    href: '/admin/products',
    title: '商品管理',
    description: '商品名、説明、カテゴリ、税込単価を管理します。',
  },
  {
    href: '/admin/forms',
    title: 'フォーム管理',
    description: '公開フォームと、フォームに表示する商品を管理します。',
  },
  {
    href: '/admin/price-rules',
    title: '価格ルール管理',
    description: '数量割引や段階単価を管理します。',
  },
]

function AdminHomePage() {
  return (
    <main>
      <h1>管理者ホーム</h1>
      <p>管理したい項目を選んでください。</p>
      <div className="product-list">
        {adminLinks.map((link) => (
          <section className="product-card" key={link.href}>
            <h2><a href={link.href}>{link.title}</a></h2>
            <p>{link.description}</p>
          </section>
        ))}
      </div>
    </main>
  )
}

export default AdminHomePage
