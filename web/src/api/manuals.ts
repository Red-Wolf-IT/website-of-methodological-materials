import { apiGet, apiPost } from './client'
import type { CreateManualPayload, Manual, ManualListParams, ManualListResult } from '@/types'

function buildQuery(params: ManualListParams): string {
  const search = new URLSearchParams()
  if (params.q) search.set('q', params.q)
  if (params.author) search.set('author', params.author)
  if (params.tag_id) search.set('tag_id', String(params.tag_id))
  if (params.sort) search.set('sort', params.sort)
  if (params.page) search.set('page', String(params.page))
  if (params.limit) search.set('limit', String(params.limit))
  const qs = search.toString()
  return qs ? `?${qs}` : ''
}

export function fetchManuals(params: ManualListParams = {}): Promise<ManualListResult> {
  return apiGet<ManualListResult>(`/manuals${buildQuery(params)}`)
}

export function fetchManual(id: string): Promise<Manual> {
  return apiGet<Manual>(`/manuals/${id}`)
}

export function createManual(payload: CreateManualPayload): Promise<Manual> {
  return apiPost<Manual>('/manuals', payload)
}

export function attachTags(manualId: string, tagIds: number[]): Promise<Manual> {
  return apiPost<Manual>(`/manuals/${manualId}/tags`, { tag_ids: tagIds })
}
