import { jwtDecode } from 'jwt-decode'

export interface JWTPayload {
  sub: string
  exp: number
  iat: number
  user_id?: string
  username?: string
  role?: string
  [key: string]: any
}

export function decodeToken(token: string): JWTPayload | null {
  try {
    return jwtDecode<JWTPayload>(token)
  } catch (error) {
    console.error('Failed to decode token:', error)
    return null
  }
}

export function isTokenExpired(token: string): boolean {
  const decoded = decodeToken(token)
  if (!decoded || !decoded.exp) {
    return true
  }
  
  // Check if token expires in less than 5 minutes
  const expirationTime = decoded.exp * 1000
  const currentTime = Date.now()
  const bufferTime = 5 * 60 * 1000 // 5 minutes

  return expirationTime - currentTime < bufferTime
}

export function getTokenExpirationTime(token: string): Date | null {
  const decoded = decodeToken(token)
  if (!decoded || !decoded.exp) {
    return null
  }
  return new Date(decoded.exp * 1000)
}
