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
  member_count?: number
  relationship_count?: number
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

export interface PublicTreeConfig {
  id?: number
  title: string
  subtitle: string
  description: string
  status: number
}

export interface PublicTreeNode {
  id?: number
  name: string
  role: string
  generation: number
  description: string
  sort: number
  status: number
}

export interface PublicTreeRelationship {
  id?: number
  from_node_id: number
  to_node_id: number
  relation_type: string
  relation_name: string
  remark: string
  sort: number
  status: number
}

export interface RelationshipType {
  id?: number
  code: string
  name: string
  category: string
  direction: string
  description: string
  sort: number
  status: number
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

  familyList(params?: { keyword?: string; status?: string | number; page?: number; page_size?: number }) {
    return request.get<Family[]>('/admin/family/list', { params })
  },

  familyCreate(data: Partial<Family>) {
    return request.post<Family>('/admin/family/create', data)
  },

  familyUpdate(data: Partial<Family>) {
    return request.post<Family>('/admin/family/update', data)
  },

  familyDelete(id: number) {
    return request.delete<null>('/admin/family/delete/' + id)
  },

  memberList(familyId?: number) {
    const params = familyId ? { family_id: familyId } : {}
    return request.get<FamilyMember[]>('/admin/member/list', { params })
  },

  memberCreate(data: Partial<FamilyMember>) {
    return request.post<FamilyMember>('/admin/member/create', data)
  },

  memberUpdate(data: Partial<FamilyMember>) {
    return request.post<FamilyMember>('/admin/member/update', data)
  },

  memberDelete(id: number) {
    return request.delete<null>('/admin/member/delete/' + id)
  },

  relationshipList(familyId?: number) {
    const params = familyId ? { family_id: familyId } : {}
    return request.get<FamilyRelationship[]>('/admin/relationship/list', { params })
  },

  relationshipCreate(data: Partial<FamilyRelationship>) {
    return request.post<FamilyRelationship>('/admin/relationship/create', data)
  },

  relationshipUpdate(data: Partial<FamilyRelationship>) {
    return request.post<FamilyRelationship>('/admin/relationship/update', data)
  },

  relationshipDelete(id: number) {
    return request.delete<null>('/admin/relationship/delete/' + id)
  },

  publicTreeConfig() {
    return request.get<PublicTreeConfig>('/admin/public-tree/config')
  },

  publicTreeConfigUpdate(data: PublicTreeConfig) {
    return request.post<PublicTreeConfig>('/admin/public-tree/config/update', data)
  },

  relationshipTypeList() {
    return request.get<RelationshipType[]>('/admin/relationship-type/list')
  },

  relationshipTypeCreate(data: RelationshipType) {
    return request.post<RelationshipType>('/admin/public-tree/relationship-type/create', data)
  },

  relationshipTypeUpdate(data: RelationshipType) {
    return request.post<RelationshipType>('/admin/public-tree/relationship-type/update', data)
  },

  relationshipTypeDelete(id: number) {
    return request.delete<null>('/admin/public-tree/relationship-type/delete/' + id)
  },

  publicTreeNodeList() {
    return request.get<PublicTreeNode[]>('/admin/public-tree/node/list')
  },

  publicTreeNodeCreate(data: PublicTreeNode) {
    return request.post<PublicTreeNode>('/admin/public-tree/node/create', data)
  },

  publicTreeNodeUpdate(data: PublicTreeNode) {
    return request.post<PublicTreeNode>('/admin/public-tree/node/update', data)
  },

  publicTreeNodeDelete(id: number) {
    return request.delete<null>('/admin/public-tree/node/delete/' + id)
  },

  publicTreeRelationshipList() {
    return request.get<PublicTreeRelationship[]>('/admin/public-tree/relationship/list')
  },

  publicTreeRelationshipCreate(data: PublicTreeRelationship) {
    return request.post<PublicTreeRelationship>('/admin/public-tree/relationship/create', data)
  },

  publicTreeRelationshipUpdate(data: PublicTreeRelationship) {
    return request.post<PublicTreeRelationship>('/admin/public-tree/relationship/update', data)
  },

  publicTreeRelationshipDelete(id: number) {
    return request.delete<null>('/admin/public-tree/relationship/delete/' + id)
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
