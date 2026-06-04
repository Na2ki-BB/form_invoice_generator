import type React from 'react'
import { createLogoutUrl } from './auth'

function AdminLogoutLink() {
  const handleLogout = (event: React.MouseEvent<HTMLAnchorElement>) => {
    event.preventDefault()
    window.location.href = createLogoutUrl()
  }

  return <p><a href="/admin/login" onClick={handleLogout}>ログアウト</a></p>
}

export default AdminLogoutLink
