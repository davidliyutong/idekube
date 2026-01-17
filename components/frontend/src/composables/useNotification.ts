import { useToast } from 'vuestic-ui'

export type NotificationOptions =
  | string
  | {
      message: string
      title?: string
      color?: 'primary' | 'info' | 'warning' | 'danger' | 'success' | 'secondary'
      duration?: number
      position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left'
    }

export interface UseNotificationReturn {
  showNotification: (options: NotificationOptions) => void
  success: (options: NotificationOptions) => void
  error: (options: NotificationOptions | null) => void
  warning: (options: NotificationOptions) => void
  info: (options: NotificationOptions) => void
  showSuccess: (options: NotificationOptions) => void
  showError: (options: NotificationOptions | null) => void
}

export function useNotification(): UseNotificationReturn {
  const toast = useToast()

  const show = (input: NotificationOptions | null, defaultColor?: string) => {
    if (input == null) return
    const opts = typeof input === 'string' ? { message: input as string, color: defaultColor || 'info' } : { ...input, color: input.color || (defaultColor as string) }
    toast.init(opts)
  }

  return {
    showNotification: (options: NotificationOptions) => show(options),
    success: (options: NotificationOptions) => show(options, 'success'),
    error: (options: NotificationOptions | null) => show(options ?? '发生错误', 'danger'),
    warning: (options: NotificationOptions) => show(options, 'warning'),
    info: (options: NotificationOptions) => show(options, 'info'),
    // Aliases for backward compatibility
    showSuccess: (options: NotificationOptions) => show(options, 'success'),
    showError: (options: NotificationOptions | null) => show(options ?? '发生错误', 'danger'),
  }
}
