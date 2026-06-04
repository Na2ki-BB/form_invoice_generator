import { useState } from 'react'
import { createLoginUrl, isCognitoConfigured } from './auth'

function AdminLoginPage() {
  const [error, setError] = useState('')

  const handleLogin = async () => {
    setError('')
    try {
      window.location.href = await createLoginUrl()
    } catch {
      setError('Cognitoログイン設定が未設定です。')
    }
  }

  if (!isCognitoConfigured()) {
    return (
      <main>
        <h1>管理者ログイン</h1>
        <p>ローカル開発ではCognitoログインを使わず、管理画面を開けます。</p>
        <p><a href="/admin">管理者ホームへ</a></p>
      </main>
    )
  }

  return (
    <main>
      <h1>管理者ログイン</h1>
      <p>管理画面を使うにはログインしてください。</p>
      <button type="button" onClick={handleLogin}>Cognitoでログインする</button>
      {error && <p className="error-message" role="alert">{error}</p>}
    </main>
  )
}

export default AdminLoginPage
