import { format, formatDistanceToNow, parseISO } from 'date-fns'

export function formatDate(date: string | Date, formatStr = 'PPP'): string {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    return format(dateObj, formatStr)
  } catch (error) {
    console.error('Failed to format date:', error)
    return String(date)
  }
}

export function formatDateTime(date: string | Date): string {
  return formatDate(date, 'PPP p')
}

export function formatRelativeTime(date: string | Date): string {
  try {
    const dateObj = typeof date === 'string' ? parseISO(date) : date
    return formatDistanceToNow(dateObj, { addSuffix: true })
  } catch (error) {
    console.error('Failed to format relative time:', error)
    return String(date)
  }
}
