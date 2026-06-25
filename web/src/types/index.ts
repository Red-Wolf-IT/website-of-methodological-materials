export interface Tag {
  id: number
  name: string
}

export interface Manual {
  id: string
  title: string
  author: string
  content: string
  file_path?: string
  views_count: number
  created_at: string
  updated_at?: string
  tags?: Tag[]
}

export interface ManualListResult {
  items: Manual[]
  total: number
  page: number
  limit: number
}

export interface ManualListParams {
  q?: string
  author?: string
  tag_id?: number
  sort?: 'popular' | ''
  page?: number
  limit?: number
}

export interface CreateManualPayload {
  title: string
  author: string
  content: string
}

export interface ApiError {
  error: {
    message: string
    fields?: Array<{ field: string; message: string }>
  }
}

export interface ValidationFieldError {
  field: string
  message: string
}
