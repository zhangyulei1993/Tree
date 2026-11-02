import { isRealApiMode, request } from '@/api/client'
import {
  contentArticles,
  contentCategories,
  type ContentArticle as MockArticle,
  type ContentCategory as MockCategory
} from '@/mock/content'
import type { ContentArticleDetail, ContentArticleSummary, ContentCategory, PaginatedResult } from '@/types/api'

export interface ContentArticleQuery {
  categoryKey?: string
  keyword?: string
  featured?: boolean
  page?: number
  pageSize?: number
}

function mockCategory(item: MockCategory, index: number): ContentCategory {
  return {
    id: index + 1,
    key: item.key,
    name: item.title,
    description: item.desc,
    sortOrder: (index + 1) * 10,
    isActive: true,
    createdAt: '',
    updatedAt: ''
  }
}

function mockArticle(item: MockArticle): ContentArticleDetail {
  const category = contentCategories.find((cat) => cat.key === item.category)
  return {
    id: Number(item.id) || Math.abs(hashCode(item.id)),
    categoryId: 0,
    categoryKey: item.category,
    categoryName: category?.title || '内容',
    contentType: 'INTERNAL',
    title: item.title,
    slug: item.id,
    summary: item.summary,
    body: item.body,
    status: 'PUBLISHED',
    isFeatured: Boolean(item.featured),
    sortOrder: 0,
    publishedAt: '',
    createdAt: '',
    updatedAt: ''
  }
}

export async function listContentCategories(): Promise<ContentCategory[]> {
  if (!isRealApiMode) return contentCategories.map(mockCategory)
  return request<ContentCategory[]>('/content/categories', { public: true })
}

export async function listContentArticles(query: ContentArticleQuery = {}): Promise<PaginatedResult<ContentArticleSummary>> {
  if (!isRealApiMode) {
    const page = query.page || 1
    const pageSize = query.pageSize || 20
    let items = contentArticles.map(mockArticle)
    if (query.categoryKey) items = items.filter((item) => item.categoryKey === query.categoryKey)
    if (query.featured !== undefined) items = items.filter((item) => item.isFeatured === query.featured)
    if (query.keyword) {
      const keyword = query.keyword.toLowerCase()
      items = items.filter((item) => `${item.title}${item.summary || ''}${item.body}`.toLowerCase().includes(keyword))
    }
    return { items: items.slice((page - 1) * pageSize, page * pageSize), page, pageSize, total: items.length }
  }
  const params = buildQuery({
    categoryKey: query.categoryKey,
    keyword: query.keyword,
    featured: query.featured === undefined ? undefined : String(query.featured),
    page: String(query.page || 1),
    pageSize: String(query.pageSize || 20)
  })
  return request<PaginatedResult<ContentArticleSummary>>(`/content/articles?${params}`, { public: true })
}

export async function getContentArticle(articleId: number | string): Promise<ContentArticleDetail> {
  if (!isRealApiMode) {
    const article = contentArticles.find((item) => item.id === String(articleId))
    if (!article) throw new Error('内容不存在')
    return mockArticle(article)
  }
  return request<ContentArticleDetail>(`/content/articles/${articleId}`, { public: true })
}

function hashCode(value: string) {
  let hash = 0
  for (let index = 0; index < value.length; index += 1) {
    hash = ((hash << 5) - hash) + value.charCodeAt(index)
    hash |= 0
  }
  return hash
}

function buildQuery(params: Record<string, string | undefined>) {
  return Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value as string)}`)
    .join('&')
}
