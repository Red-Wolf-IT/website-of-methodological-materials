import { apiGet, apiPost } from './client'
import type { Tag } from '@/types'

export function fetchTags(): Promise<Tag[]> {
  return apiGet<Tag[]>('/tags')
}

export function createTag(name: string): Promise<Tag> {
  return apiPost<Tag>('/tags', { name })
}
