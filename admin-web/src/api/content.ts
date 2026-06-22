import { apiClient, unwrapData } from '@/api/client'
import type {
  ContentArticleDetail,
  ContentArticleInput,
  ContentArticleSummary,
  ContentCategory,
  ContentCategoryInput,
  PageResult
} from '@/types/api'

export interface ContentArticleQuery {
  categoryKey?: string
  status?: string
  keyword?: string
  page?: number
  pageSize?: number
}

export async function listContentCategories(includeInactive = true): Promise<ContentCategory[]> {
  const response = await apiClient.get('/admin/content/categories', {
    params: { includeInactive }
  })
  return unwrapData<ContentCategory[]>(response)
}

export async function createContentCategory(input: ContentCategoryInput): Promise<ContentCategory> {
  const response = await apiClient.post('/admin/content/categories', input)
  return unwrapData<ContentCategory>(response)
}

export async function updateContentCategory(id: number, input: ContentCategoryInput): Promise<ContentCategory> {
  const response = await apiClient.put(`/admin/content/categories/${id}`, input)
  return unwrapData<ContentCategory>(response)
}

export async function deleteContentCategory(id: number): Promise<{ deleted: boolean }> {
  const response = await apiClient.delete(`/admin/content/categories/${id}`)
  return unwrapData<{ deleted: boolean }>(response)
}

export async function listContentArticles(query: ContentArticleQuery): Promise<PageResult<ContentArticleSummary>> {
  const response = await apiClient.get('/admin/content/articles', { params: query })
  return unwrapData<PageResult<ContentArticleSummary>>(response)
}

export async function getContentArticle(id: number): Promise<ContentArticleDetail> {
  const response = await apiClient.get(`/admin/content/articles/${id}`)
  return unwrapData<ContentArticleDetail>(response)
}

export async function createContentArticle(input: ContentArticleInput): Promise<ContentArticleDetail> {
  const response = await apiClient.post('/admin/content/articles', input)
  return unwrapData<ContentArticleDetail>(response)
}

export async function updateContentArticle(id: number, input: Partial<ContentArticleInput>): Promise<ContentArticleDetail> {
  const response = await apiClient.put(`/admin/content/articles/${id}`, input)
  return unwrapData<ContentArticleDetail>(response)
}

export async function publishContentArticle(id: number): Promise<ContentArticleDetail> {
  const response = await apiClient.post(`/admin/content/articles/${id}/publish`, {})
  return unwrapData<ContentArticleDetail>(response)
}

export async function unpublishContentArticle(id: number): Promise<ContentArticleDetail> {
  const response = await apiClient.post(`/admin/content/articles/${id}/unpublish`, {})
  return unwrapData<ContentArticleDetail>(response)
}

export async function deleteContentArticle(id: number): Promise<{ deleted: boolean }> {
  const response = await apiClient.delete(`/admin/content/articles/${id}`)
  return unwrapData<{ deleted: boolean }>(response)
}
