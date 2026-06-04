import {
  COGNITO_CLIENT_ID,
  COGNITO_DOMAIN,
  COGNITO_LOGOUT_URI,
  COGNITO_REDIRECT_URI,
} from '../config'

const TOKEN_KEY = 'adminAuthToken'
const PKCE_VERIFIER_KEY = 'adminPkceVerifier'
const OAUTH_STATE_KEY = 'adminOAuthState'
const OAUTH_SCOPE = 'openid email profile'

type StoredToken = {
  accessToken: string
  idToken?: string
  expiresAt: number
}

type TokenResponse = {
  access_token: string
  id_token?: string
  expires_in: number
}

export const isCognitoConfigured = () =>
  Boolean(COGNITO_DOMAIN && COGNITO_CLIENT_ID && COGNITO_REDIRECT_URI && COGNITO_LOGOUT_URI)

export const getStoredAccessToken = () => {
  const rawToken = sessionStorage.getItem(TOKEN_KEY)
  if (!rawToken) return ''

  try {
    const token = JSON.parse(rawToken) as StoredToken
    if (!token.accessToken || token.expiresAt <= Date.now()) {
      clearAdminSession()
      return ''
    }
    return token.accessToken
  } catch {
    clearAdminSession()
    return ''
  }
}

export const isAdminAuthenticated = () => !isCognitoConfigured() || Boolean(getStoredAccessToken())

export const clearAdminSession = () => {
  sessionStorage.removeItem(TOKEN_KEY)
  sessionStorage.removeItem(PKCE_VERIFIER_KEY)
  sessionStorage.removeItem(OAUTH_STATE_KEY)
}

export const createLoginUrl = async () => {
  assertCognitoConfigured()

  const verifier = createRandomString(64)
  const state = createRandomString(32)
  const challenge = await createCodeChallenge(verifier)
  sessionStorage.setItem(PKCE_VERIFIER_KEY, verifier)
  sessionStorage.setItem(OAUTH_STATE_KEY, state)

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: COGNITO_CLIENT_ID,
    redirect_uri: COGNITO_REDIRECT_URI,
    scope: OAUTH_SCOPE,
    code_challenge_method: 'S256',
    code_challenge: challenge,
    state,
  })
  return `${trimTrailingSlash(COGNITO_DOMAIN)}/oauth2/authorize?${params.toString()}`
}

export const exchangeCodeForToken = async (code: string, state: string) => {
  assertCognitoConfigured()

  const expectedState = sessionStorage.getItem(OAUTH_STATE_KEY)
  const verifier = sessionStorage.getItem(PKCE_VERIFIER_KEY)
  if (!expectedState || expectedState !== state || !verifier) {
    throw new Error('invalid auth callback')
  }

  const response = await fetch(`${trimTrailingSlash(COGNITO_DOMAIN)}/oauth2/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: COGNITO_CLIENT_ID,
      code,
      redirect_uri: COGNITO_REDIRECT_URI,
      code_verifier: verifier,
    }),
  })
  if (!response.ok) throw new Error('token exchange failed')

  const tokenResponse = await response.json() as TokenResponse
  sessionStorage.setItem(TOKEN_KEY, JSON.stringify({
    accessToken: tokenResponse.access_token,
    idToken: tokenResponse.id_token,
    expiresAt: Date.now() + tokenResponse.expires_in * 1000,
  } satisfies StoredToken))
  sessionStorage.removeItem(PKCE_VERIFIER_KEY)
  sessionStorage.removeItem(OAUTH_STATE_KEY)
}

export const createLogoutUrl = () => {
  clearAdminSession()
  if (!isCognitoConfigured()) return '/admin/login'

  const params = new URLSearchParams({
    client_id: COGNITO_CLIENT_ID,
    logout_uri: COGNITO_LOGOUT_URI,
  })
  return `${trimTrailingSlash(COGNITO_DOMAIN)}/logout?${params.toString()}`
}

export const adminFetch = async (input: RequestInfo | URL, init: RequestInit = {}) => {
  const headers = new Headers(init.headers)
  if (isCognitoConfigured()) {
    const accessToken = getStoredAccessToken()
    if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`)
  } else {
    headers.set('X-Local-Admin', 'true')
  }

  const response = await fetch(input, { ...init, headers })
  if (response.status === 401 && isCognitoConfigured()) {
    clearAdminSession()
  }
  return response
}

const assertCognitoConfigured = () => {
  if (!isCognitoConfigured()) {
    throw new Error('Cognito is not configured')
  }
}

const createRandomString = (length: number) => {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  return base64UrlEncode(bytes)
}

const createCodeChallenge = async (verifier: string) => {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier))
  return base64UrlEncode(new Uint8Array(digest))
}

const base64UrlEncode = (bytes: Uint8Array) =>
  btoa(String.fromCharCode(...bytes))
    .replaceAll('+', '-')
    .replaceAll('/', '_')
    .replaceAll('=', '')

const trimTrailingSlash = (value: string) => value.replace(/\/+$/, '')
