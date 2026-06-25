import type { ApiError } from '@/types'

const API_BASE = import.meta.env.VITE_API_URL ?? '/api'

export class ApiClientError extends Error {
  status: number
  fields?: Array<{ field: string; message: string }>

  constructor(status: number, message: string, fields?: Array<{ field: string; message: string }>) {
    super(message)
    this.status = status
    this.fields = fields
  }
}

async function parseError(response: Response): Promise<ApiClientError> {
  try {
    const body = (await response.json()) as ApiError
    return new ApiClientError(response.status, body.error?.message ?? 'Ошибка запроса', body.error?.fields)
  } catch {
    return new ApiClientError(response.status, 'Ошибка запроса')
  }
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`)
  if (!response.ok) throw await parseError(response)
  const body = await response.json()
  return body.data as T
}

export async function apiPost<T>(path: string, payload: unknown): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  if (!response.ok) throw await parseError(response)
  const body = await response.json()
  return body.data as T
}

export function fileUrl(filePath: string): string {
  if (filePath.startsWith('http')) return filePath
  const base = API_BASE.replace(/\/$/, '')
  return `${base}${filePath}`
}

export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(new Date(iso))
}
