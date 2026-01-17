import { useToast } from 'vuestic-ui'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface NotificationOptions {
  message: string
  duration?: number
  closeable?: boolean
}

let toastInstance: ReturnType<typeof useToast> | null = null

export function useNotification() {
  if (!toastInstance) {
    toastInstance = useToast()
  }

  const notify = (type: NotificationType, options: string | NotificationOptions) => {
    const config = typeof options === 'string' 
      ? { message: options, duration: 3000, closeable: true }
      : { duration: 3000, closeable: true, ...options }

    if (toastInstance) {
      toastInstance.init({
        message: config.message,
        duration: config.duration,
        closeable: config.closeable,
        color: type,
      })
    }
  }

  return {
    success: (options: string | NotificationOptions) => notify('success', options),
    error: (options: string | NotificationOptions) => notify('error', options),
    warning: (options: string | NotificationOptions) => notify('warning', options),
    info: (options: string | NotificationOptions) => notify('info', options),
  }
}
