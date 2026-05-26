import request from './utils/request'

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
  id: number
  title: string
  image_url: string
  link_url: string
  sort: number
  status: number
}

export interface Article {
  id: number
  title: string
  cover: string
  summary: string
  content: string
  category_id: number
  sort: number
  status: number
  created_at: string
}

export interface Family {
  id: number
  name: string
  surname: string
  origin: string
  cover: string
  description: string
  status: number
}

export interface FamilyMember {
  id: number
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
  id: number
  family_id: number
  from_member_id: number
  to_member_id: number
  relation_type: string
  remark: string
}

export interface PublicTreeConfig {
  id: number
  title: string
  subtitle: string
  description: string
  status: number
}

export interface PublicTreeNode {
  id: number
  name: string
  role: string
  generation: number
  description: string
  sort: number
  status: number
}

export interface PublicTreeRelationship {
  id: number
  from_node_id: number
  to_node_id: number
  relation_type: string
  relation_name: string
  remark: string
  sort: number
  status: number
}

export interface PublicTreeResult {
  config: PublicTreeConfig
  nodes: PublicTreeNode[]
  relationships: PublicTreeRelationship[]
  generation_keys: number[]
  generations: Record<string, PublicTreeNode[]>
}

export interface FamilyTreeResult {
  family: Family
  members: FamilyMember[]
  relationships: FamilyRelationship[]
  generation_keys: number[]
  generations: Record<string, FamilyMember[]>
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export const api = {
  siteConfig() {
    return request.get('/web/site/config') as Promise<SiteConfig>
  },

  bannerList() {
    return request.get('/web/banner/list') as Promise<Banner[]>
  },

  articleList(params: { page: number; page_size: number }) {
    return request.get('/web/article/list', { params }) as Promise<PageResult<Article>>
  },

  articleDetail(id: string | number) {
    return request.get('/web/article/detail/' + id) as Promise<Article>
  },

  submitContact(data: { name: string; phone: string; email: string; content: string }) {
    return request.post('/web/contact/submit', data) as Promise<null>
  },

  familyTree(familyId?: number) {
    const params = familyId ? { family_id: familyId } : {}
    return request.get('/web/family/tree', { params }) as Promise<FamilyTreeResult>
  },

  publicTree() {
    return request.get('/web/public-tree') as Promise<PublicTreeResult>
  }
}
