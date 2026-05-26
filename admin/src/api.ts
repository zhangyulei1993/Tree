import request from './utils/request'

export interface LoginResult {
  token: string
  user: {
    id: number
    username: string
    nickname: string
    role: string
  }
}

export interface SiteConfig {
  id?: number
  site_name: string
  logo: string
  phone: string
  email: string
  address: string
  description: string
}

export interface Banner {
  id?: number
  title: string
  image_url: string
  link_url: string
  sort: number
  status: number
}

export interface Article {
  id?: number
  title: string
  cover: string
  summary: string
  content: string
  category_id: number
  sort: number
  status: number
  created_at?: string
}

export interface ContactMessage {
  id: number
  name: string
  phone: string
  email: string
  content: string
  status: number
  created_at: string
}

export interface Family {
  id?: number
  name: string
  surname: string
  origin: string
  cover: string
  description: string
  status: number
}

export interface FamilyMember {
  id?: number
  family_id: number
  name: string
  gender: string
  avatar: string
  birthday: string
  birth_place: string
  generation: number
  biography: string
  sort: number
  status: number
}

export interface FamilyRelationship {
  id?: number
  family_id: number
  from_member_id: number
  to_member_id: number
  relation_type: string
  remark: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export const api = {
  login(data: { username: string; password: string }) {
    return request.post<LoginResult>('/admin/login', data)
  },

  dashboard() {
    return request.get<{
      banner_count: number
      article_count: number
      contact_count: number
    }>('/admin/dashboard')
  },

  getSiteConfig() {
    return request.get<SiteConfig>('/admin/site/config')
  },

  updateSiteConfig(data: SiteConfig) {
    return request.post<SiteConfig>('/admin/site/config/update', data)
  },

  bannerList() {
    return request.get<Banner[]>('/admin/banner/list')
  },

  bannerCreate(data: Banner) {
    return request.post<Banner>('/admin/banner/create', data)
  },

  bannerUpdate(data: Banner) {
    return request.post<Banner>('/admin/banner/update', data)
  },

  bannerDelete(id: number) {
    return request.delete<null>('/admin/banner/delete/' + id)
  },

  articleList(params: { page: number; page_size: number }) {
    return request.get<PageResult<Article>>('/admin/article/list', { params })
  },

  articleCreate(data: Article) {
    return request.post<Article>('/admin/article/create', data)
  },

  articleUpdate(data: Article) {
    return request.post<Article>('/admin/article/update', data)
  },

  articleDelete(id: number) {
    return request.delete<null>('/admin/article/delete/' + id)
  },

  contactList() {
    return request.get<ContactMessage[]>('/admin/contact/list')
  },

  contactDelete(id: number) {
    return request.delete<null>('/admin/contact/delete/' + id)
  },

  familyList() {
    return request.get<Family[]>('/admin/family/list')
  },

  familyCreate(data: Family) {
    return request.post<Family>('/admin/family/create', data)
  },

  familyUpdate(data: Family) {
    return request.post<Family>('/admin/family/update', data)
  },

  familyDelete(id: number) {
    return request.delete<null>('/admin/family/delete/' + id)
  },

  memberList(familyId?: number) {
    const params = familyId ? { family_id: familyId } : {}
    return request.get<FamilyMember[]>('/admin/member/list', { params })
  },

  memberCreate(data: FamilyMember) {
    return request.post<FamilyMember>('/admin/member/create', data)
  },

  memberUpdate(data: FamilyMember) {
    return request.post<FamilyMember>('/admin/member/update', data)
  },

  memberDelete(id: number) {
    return request.delete<null>('/admin/member/delete/' + id)
  },

  relationshipList(familyId?: number) {
    const params = familyId ? { family_id: familyId } : {}
    return request.get<FamilyRelationship[]>('/admin/relationship/list', { params })
  },

  relationshipCreate(data: FamilyRelationship) {
    return request.post<FamilyRelationship>('/admin/relationship/create', data)
  },

  relationshipUpdate(data: FamilyRelationship) {
    return request.post<FamilyRelationship>('/admin/relationship/update', data)
  },

  relationshipDelete(id: number) {
    return request.delete<null>('/admin/relationship/delete/' + id)
  },

  upload(file: File) {
    const form = new FormData()
    form.append('file', file)

    return request.post<{ url: string; file: unknown }>('/admin/upload', form, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}