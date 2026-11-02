import { apiClient, isRealApiMode } from '@/api/client'
import { mockContentArticles, mockContentCategories } from '@/mock/content'
import type { ApiResponse, ContentArticleDetail, ContentArticleSummary, ContentCategory, PaginatedResult } from '@/types/api'

export interface ContentArticleQuery {
  categoryKey?: string
  keyword?: string
  featured?: boolean
  page?: number
  pageSize?: number
}

export async function listContentCategories(): Promise<ContentCategory[]> {
  if (!isRealApiMode) return mockContentCategories
  const response = await apiClient.get<ApiResponse<ContentCategory[]>>('/content/categories')
  return response.data.data
}

export async function listContentArticles(query: ContentArticleQuery = {}): Promise<PaginatedResult<ContentArticleSummary>> {
  if (!isRealApiMode) {
    const page = query.page || 1
    const pageSize = query.pageSize || 20
    let items = mockContentArticles
    if (query.categoryKey) items = items.filter((item) => item.categoryKey === query.categoryKey)
    if (query.featured !== undefined) items = items.filter((item) => item.isFeatured === query.featured)
    if (query.keyword) {
      const keyword = query.keyword.toLowerCase()
      items = items.filter((item) => `${item.title}${item.summary || ''}${item.body}`.toLowerCase().includes(keyword))
    }
    return {
      items: items.slice((page - 1) * pageSize, page * pageSize),
      page,
      pageSize,
      total: items.length
    }
  }
  const response = await apiClient.get<ApiResponse<PaginatedResult<ContentArticleSummary>>>('/content/articles', { params: query })
  return response.data.data
}

export async function getContentArticle(articleId: number | string): Promise<ContentArticleDetail> {
  if (!isRealApiMode) {
    const article = mockContentArticles.find((item) => String(item.id) === String(articleId) || item.slug === String(articleId))
    if (!article) throw new Error('内容不存在')
    return article
  }
  const response = await apiClient.get<ApiResponse<ContentArticleDetail>>(`/content/articles/${articleId}`)
  return response.data.data
}
