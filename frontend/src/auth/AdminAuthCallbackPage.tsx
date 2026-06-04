import { useEffect, useState } from 'react'
import { exchangeCodeForToken, isCognitoConfigured } from './auth'

function AdminAuthCallbackPage() {
  const [error, setError] = useState('')

  useEffect(() => {
    const handleCallback = async () => {
      if (!isCognitoConfigured()) {
        setError('Cognitoログイン設定が未設定です。')
        return
      }

      const params = new URLSearchParams(window.location.search)
      const code = params.get('code')
      const state = params.get('state')
      if (!code || !state) {
        setError('ログイン結果を確認できませんでした。')
        return
      }

      try {
        await exchangeCodeForToken(code, state)
        window.history.replaceState(null, '', '/admin')
        window.location.href = '/admin'
      } catch {
        setError('ログイン処理に失敗しました。もう一度ログインしてください。')
      }
    }

    void handleCallback()
  }, [])

  return (
    <main>
      <h1>ログイン確認中</h1>
      {error ? (
        <>
          <p className="error-message" role="alert">{error}</p>
          <p><a href="/admin/login">ログイン画面へ戻る</a></p>
        </>
      ) : (
        <p>ログイン結果を確認しています。</p>
      )}
    </main>
  )
}

export default AdminAuthCallbackPage
