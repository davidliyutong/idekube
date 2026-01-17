import { AxiosError } from 'axios'

export interface APIError {
  code?: string
  message: string
  details?: any
}

export function handleError(error: unknown): APIError {
  if (error instanceof AxiosError) {
    const data = error.response?.data as any
    return {
      code: data?.code || error.code,
      message: data?.message || error.message || 'An error occurred',
      details: data?.details || error.response?.data,
    }
  }

  if (error instanceof Error) {
    return {
      message: error.message,
    }
  }

  return {
    message: String(error) || 'Unknown error',
  }
}

export function getErrorMessage(error: unknown): string {
  return handleError(error).message
}
