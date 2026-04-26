import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export const cn = (...inputs) => twMerge(clsx(inputs))

export const toIso = (value) => {
  if (!value) return '—'
  const n = Number(value)
  if (Number.isNaN(n)) return String(value)
  return new Date(n * 1000).toISOString()
}

export const asTrimmed = (value) => String(value ?? '').trim()

export const mapStatusVariant = (status) => {
  const v = String(status ?? '').toLowerCase()
  if (v.includes('active') || v.includes('completed') || v.includes('ok')) return 'ok'
  if (v.includes('consumed') || v.includes('reserved') || v.includes('cancelled') || v.includes('warn') || v.includes('stale')) return 'warn'
  if (v.includes('expired') || v.includes('error') || v.includes('failed') || v.includes('err')) return 'err'
  return 'neutral'
}

export const extractError = (json) => ({
  code: String(json?.error?.code ?? json?.code ?? 'UNKNOWN_ERROR'),
  message: String(json?.error?.message ?? json?.message ?? 'unknown error'),
})

export const safeJson = async (resp) => {
  const raw = await resp.text()
  if (!raw) return {}
  try { return JSON.parse(raw) } catch { return { raw } }
}
