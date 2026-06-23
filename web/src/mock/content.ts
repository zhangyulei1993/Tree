import type { ContentArticleDetail, ContentCategory } from '@/types/api'

export const mockContentCategories: ContentCategory[] = [
  { id: 1, key: 'tutorial', name: '使用教程', description: '创建家庭、维护关系与公开主页', sortOrder: 10, isActive: true, createdAt: '', updatedAt: '' },
  { id: 2, key: 'story', name: '家族故事', description: '阅读家族记忆与人物故事', sortOrder: 20, isActive: true, createdAt: '', updatedAt: '' },
  { id: 3, key: 'surname', name: '姓氏典故', description: '了解姓氏源流、堂号与家训', sortOrder: 30, isActive: true, createdAt: '', updatedAt: '' },
  { id: 4, key: 'article', name: '宗亲文章', description: '平台精选的家族文化文章', sortOrder: 40, isActive: true, createdAt: '', updatedAt: '' }
]

export const mockContentArticles: ContentArticleDetail[] = [
  {
    id: 1,
    categoryId: 1,
    categoryKey: 'tutorial',
    categoryName: '使用教程',
    title: '如何创建第一个家庭',
    slug: 'tutorial-create-family',
    summary: '从填写姓氏与家庭名称开始，建立你的家族空间。',
    body: '在 Tree 中，每个家庭对应一个独立的家族空间。创建时建议先确认本族主要姓氏与家庭名称，便于后续邀请亲人识别。\n\n创建完成后，你可以先添加自己或长辈作为首批成员，再逐步补充父母、配偶与子女关系。成员资料与账号绑定分开管理，没有账号的亲属也可以先录入为成员节点。\n\n如果暂时只有核心成员，也建议先创建家庭并录入关键节点，后续再邀请家人一起完善。',
    authorName: 'Tree 编辑部',
    source: '平台精选',
    status: 'PUBLISHED',
    isFeatured: true,
    sortOrder: 10,
    publishedAt: '',
    createdAt: '',
    updatedAt: ''
  },
  {
    id: 2,
    categoryId: 2,
    categoryKey: 'story',
    categoryName: '家族故事',
    title: '一张老照片背后的迁徙记忆',
    slug: 'story-old-photo',
    summary: '从一张泛黄合影读出三代人的迁居轨迹。',
    body: '许多家庭的第一条故事，往往来自一张没有标注年份的老照片。\n\n后来家人比对口述，才发现拍摄地并非祖籍所在，而是迁往省城后的第一个落脚点。\n\n把照片与成员节点关联，并记下已知人物与地点，即使细节尚未齐全，也能为后来的亲人留下可查证的起点。',
    authorName: 'Tree 编辑部',
    source: '平台精选',
    status: 'PUBLISHED',
    isFeatured: true,
    sortOrder: 20,
    publishedAt: '',
    createdAt: '',
    updatedAt: ''
  },
  {
    id: 3,
    categoryId: 3,
    categoryKey: 'surname',
    categoryName: '姓氏典故',
    title: '姓氏源流可以如何查证',
    slug: 'surname-origin',
    summary: '从文献、族谱与口述中交叉印证姓氏来历。',
    body: '查证姓氏源流，通常需要结合地方志、旧谱牒与家族口述，不宜仅凭单一网络条目下结论。\n\n可从本族主要聚居地、堂号记载与祖先传说三条线入手，记录来源文献与可信度。\n\n在 Tree 中，可将考证结论写入家庭简介或备注，并注明待进一步核实，方便后人接续研究。',
    authorName: 'Tree 编辑部',
    source: '平台精选',
    status: 'PUBLISHED',
    isFeatured: true,
    sortOrder: 30,
    publishedAt: '',
    createdAt: '',
    updatedAt: ''
  }
]
