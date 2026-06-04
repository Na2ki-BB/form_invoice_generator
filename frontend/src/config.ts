const DEFAULT_API_BASE_URL = 'http://127.0.0.1:8080'

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL

export const apiUrl = (path: string) => `${API_BASE_URL}${path}`

export const COGNITO_DOMAIN = import.meta.env.VITE_COGNITO_DOMAIN
export const COGNITO_CLIENT_ID = import.meta.env.VITE_COGNITO_CLIENT_ID
export const COGNITO_REDIRECT_URI = import.meta.env.VITE_COGNITO_REDIRECT_URI
export const COGNITO_LOGOUT_URI = import.meta.env.VITE_COGNITO_LOGOUT_URI
