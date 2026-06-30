import { apiClient, apiMode, unwrapData } from '@/api/client'
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

const mockCategories: ContentCategory[] = [
  {
    id: 1,
    key: 'tutorial',
    name: '使用教程',
    description: '创建家庭、维护成员、邀请亲友的操作说明。',
    sortOrder: 10,
    isActive: true,
    createdAt: '2026-06-01 09:00',
    updatedAt: '2026-06-08 09:00'
  },
  {
    id: 2,
    key: 'family-story',
    name: '家族故事',
    description: '公开展示页可使用的家族故事与人物资料。',
    sortOrder: 20,
    isActive: true,
    createdAt: '2026-06-01 09:10',
    updatedAt: '2026-06-08 09:10'
  },
  {
    id: 3,
    key: 'surname-culture',
    name: '姓氏文化',
    description: '姓氏来源、迁徙线索与地方文化资料。',
    sortOrder: 30,
    isActive: false,
    createdAt: '2026-06-01 09:20',
    updatedAt: '2026-06-08 09:20'
  }
]

const mockArticles: ContentArticleDetail[] = [
  {
    id: 1,
    categoryId: 1,
    categoryKey: 'tutorial',
    categoryName: '使用教程',
    title: '如何创建第一个家庭空间',
    slug: 'create-first-family',
    summary: '从填写姓氏、籍贯到邀请成员，说明创建家庭的基本流程。',
    body: '创建家庭前先确认主姓氏和创建者身份。创建完成后，可以继续补充成员节点和邀请家人绑定账号。',
    authorName: 'Tree 编辑部',
    source: '平台内容',
    status: 'PUBLISHED',
    isFeatured: true,
    sortOrder: 10,
    publishedAt: '2026-06-06 10:20',
    createdAt: '2026-06-02 11:00',
    updatedAt: '2026-06-06 10:20'
  },
  {
    id: 2,
    categoryId: 1,
    categoryKey: 'tutorial',
    categoryName: '使用教程',
    title: '成员节点和账号绑定有什么区别',
    slug: 'member-node-account-binding',
    summary: '解释 family_member 与 user 的关系，避免把家谱节点和平台账号混为一谈。',
    body: '成员节点是家谱中的人，平台账号是登录系统的人。一个成员节点可以暂不绑定账号，也可以标记为无需绑定。',
    authorName: 'Tree 编辑部',
    source: '平台内容',
    status: 'DRAFT',
    isFeatured: false,
    sortOrder: 20,
    publishedAt: null,
    createdAt: '2026-06-03 14:15',
    updatedAt: '2026-06-07 16:40'
  },
  {
    id: 3,
    categoryId: 2,
    categoryKey: 'family-story',
    categoryName: '家族故事',
    title: '公开家谱资料应如何整理',
    slug: 'public-family-profile-guide',
    summary: '整理公开展示内容时，优先使用可公开、可核对、无敏感联系方式的资料。',
    body: '公开展示适合放置家族简介、迁徙脉络和寻亲说明，不应展示完整手机号、身份证号或其他敏感信息。',
    authorName: 'Tree 编辑部',
    source: '平台内容',
    status: 'PUBLISHED',
    isFeatured: false,
    sortOrder: 30,
    publishedAt: '2026-06-07 09:30',
    createdAt: '2026-06-04 09:40',
    updatedAt: '2026-06-07 09:30'
  }
]

function paginate<T>(items: T[], page = 1, pageSize = 20): PageResult<T> {
  const start = (page - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    page,
    pageSize,
    total: items.length
  }
}

function findMockCategory(key: string) {
  return mockCategories.find((item) => item.key === key) || mockCategories[0]
}

function toArticleDetail(id: number, input: Partial<ContentArticleInput>): ContentArticleDetail {
  const category = findMockCategory(input.categoryKey || 'tutorial')
  const now = '2026-06-08 10:00'
  return {
    id,
    categoryId: category.id,
    categoryKey: category.key,
    categoryName: category.name,
    title: input.title || '未命名文章',
    slug: input.slug || `article-${id}`,
    summary: input.summary || '',
    coverUrl: input.coverUrl || null,
    body: input.body || '',
    authorName: input.authorName || 'Tree 编辑部',
    source: input.source || '平台内容',
    status: input.status || 'DRAFT',
    isFeatured: Boolean(input.isFeatured),
    sortOrder: input.sortOrder || 0,
    publishedAt: input.status === 'PUBLISHED' ? now : null,
    createdAt: now,
    updatedAt: now
  }
}

export async function listContentCategories(includeInactive = true): Promise<ContentCategory[]> {
  if (apiMode === 'mock') {
    return includeInactive ? mockCategories : mockCategories.filter((item) => item.isActive)
  }
  const response = await apiClient.get('/admin/content/categories', {
    params: { includeInactive }
  })
  return unwrapData<ContentCategory[]>(response)
}

export async function createContentCategory(input: ContentCategoryInput): Promise<ContentCategory> {
  if (apiMode === 'mock') {
    return {
      id: mockCategories.length + 1,
      key: input.key || `category-${mockCategories.length + 1}`,
      name: input.name,
      description: input.description || '',
      sortOrder: input.sortOrder || 0,
      isActive: input.isActive ?? true,
      createdAt: '2026-06-08 10:00',
      updatedAt: '2026-06-08 10:00'
    }
  }
  const response = await apiClient.post('/admin/content/categories', input)
  return unwrapData<ContentCategory>(response)
}

export async function updateContentCategory(id: number, input: ContentCategoryInput): Promise<ContentCategory> {
  if (apiMode === 'mock') {
    const current = mockCategories.find((item) => item.id === id) || mockCategories[0]
    return {
      ...current,
      ...input,
      key: current.key,
      updatedAt: '2026-06-08 10:00'
    }
  }
  const response = await apiClient.put(`/admin/content/categories/${id}`, input)
  return unwrapData<ContentCategory>(response)
}

export async function deleteContentCategory(id: number): Promise<{ deleted: boolean }> {
  if (apiMode === 'mock') {
    return { deleted: Boolean(id) }
  }
  const response = await apiClient.delete(`/admin/content/categories/${id}`)
  return unwrapData<{ deleted: boolean }>(response)
}

export async function listContentArticles(query: ContentArticleQuery): Promise<PageResult<ContentArticleSummary>> {
  if (apiMode === 'mock') {
    const items = mockArticles.filter((item) => {
      const hitCategory = !query.categoryKey || item.categoryKey === query.categoryKey
      const hitStatus = !query.status || item.status === query.status
      const keyword = query.keyword?.trim()
      const hitKeyword = !keyword || [item.title, item.summary, item.body].some((value) => String(value || '').includes(keyword))
      return hitCategory && hitStatus && hitKeyword
    })
    return paginate(items, query.page, query.pageSize)
  }
  const response = await apiClient.get('/admin/content/articles', { params: query })
  return unwrapData<PageResult<ContentArticleSummary>>(response)
}

export async function getContentArticle(id: number): Promise<ContentArticleDetail> {
  if (apiMode === 'mock') {
    return mockArticles.find((item) => item.id === id) || mockArticles[0]
  }
  const response = await apiClient.get(`/admin/content/articles/${id}`)
  return unwrapData<ContentArticleDetail>(response)
}

export async function createContentArticle(input: ContentArticleInput): Promise<ContentArticleDetail> {
  if (apiMode === 'mock') {
    return toArticleDetail(mockArticles.length + 1, input)
  }
  const response = await apiClient.post('/admin/content/articles', input)
  return unwrapData<ContentArticleDetail>(response)
}

export async function updateContentArticle(id: number, input: Partial<ContentArticleInput>): Promise<ContentArticleDetail> {
  if (apiMode === 'mock') {
    const current = mockArticles.find((item) => item.id === id) || mockArticles[0]
    return {
      ...current,
      ...toArticleDetail(id, {
        categoryKey: current.categoryKey,
        title: current.title,
        slug: current.slug,
        summary: current.summary || '',
        coverUrl: current.coverUrl || '',
        body: current.body,
        authorName: current.authorName || '',
        source: current.source || '',
        status: current.status,
        isFeatured: current.isFeatured,
        sortOrder: current.sortOrder,
        ...input
      }),
      id
    }
  }
  const response = await apiClient.put(`/admin/content/articles/${id}`, input)
  return unwrapData<ContentArticleDetail>(response)
}

export async function publishContentArticle(id: number): Promise<ContentArticleDetail> {
  if (apiMode === 'mock') {
    const current = mockArticles.find((item) => item.id === id) || mockArticles[0]
    return { ...current, status: 'PUBLISHED', publishedAt: '2026-06-08 10:00', updatedAt: '2026-06-08 10:00' }
  }
  const response = await apiClient.post(`/admin/content/articles/${id}/publish`, {})
  return unwrapData<ContentArticleDetail>(response)
}

export async function unpublishContentArticle(id: number): Promise<ContentArticleDetail> {
  if (apiMode === 'mock') {
    const current = mockArticles.find((item) => item.id === id) || mockArticles[0]
    return { ...current, status: 'ARCHIVED', updatedAt: '2026-06-08 10:00' }
  }
  const response = await apiClient.post(`/admin/content/articles/${id}/unpublish`, {})
  return unwrapData<ContentArticleDetail>(response)
}

export async function deleteContentArticle(id: number): Promise<{ deleted: boolean }> {
  if (apiMode === 'mock') {
    return { deleted: Boolean(id) }
  }
  const response = await apiClient.delete(`/admin/content/articles/${id}`)
  return unwrapData<{ deleted: boolean }>(response)
}
