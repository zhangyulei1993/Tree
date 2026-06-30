export type ContentCategory = 'tutorial' | 'story' | 'surname' | 'article'

export interface ContentCategoryMeta {
  key: ContentCategory
  title: string
  desc: string
}

export interface ContentArticle {
  id: string
  category: ContentCategory
  title: string
  summary: string
  body: string
  featured?: boolean
}

export const contentCategories: ContentCategoryMeta[] = [
  { key: 'tutorial', title: '使用教程', desc: '创建家庭、维护关系与公开主页' },
  { key: 'story', title: '家族故事', desc: '阅读家族记忆与人物故事' },
  { key: 'surname', title: '姓氏典故', desc: '了解姓氏源流、堂号与家训' },
  { key: 'article', title: '宗亲文章', desc: '平台精选的家族文化文章' }
]

export const contentArticles: ContentArticle[] = [
  {
    id: 'tutorial-create-family',
    category: 'tutorial',
    title: '如何创建第一个家庭',
    summary: '从填写姓氏与家庭名称开始，建立你的家庭空间。',
    featured: true,
    body: `在 Tree 中，每个家庭对应一个独立的家庭空间。创建时建议先确认本族主要姓氏与家庭名称，便于后续邀请亲人识别。

创建完成后，你可以先添加自己或长辈作为首批成员，再逐步补充父母、配偶与子女关系。成员资料与账号绑定分开管理，没有账号的亲属也可以先录入为成员节点。

如果暂时只有核心成员，也建议先创建家庭并录入关键节点，后续再邀请家人一起完善。`
  },
  {
    id: 'tutorial-add-members-tree',
    category: 'tutorial',
    title: '如何添加成员并生成家谱',
    summary: '录入成员信息，维护父母子女与配偶关系。',
    featured: true,
    body: `添加成员时，至少填写显示名称与性别，有助于后续称谓推导与关系展示。

父母子女关系用于构建家谱纵向脉络，配偶关系用于横向连接。添加兄弟姐妹时，系统会通过共同父母节点建立联系，因此若缺少父母节点，需先补充上一代成员。

在私有家谱页面，你可以查看当前成员列表与关系描述。每增加或调整关系，家谱结构会随之更新，便于家人共同核对。`
  },
  {
    id: 'tutorial-invite-family',
    category: 'tutorial',
    title: '如何邀请亲人一起维护',
    summary: '通过邀请链接让家人加入并绑定成员。',
    featured: true,
    body: `家庭创建者或管理员可以发起邀请，将邀请链接发送给亲人。对方接受邀请后，会加入该家庭并绑定到对应成员节点。

邀请前建议先创建好被邀请人的成员资料，或在接受邀请时完成绑定，避免重复录入。若对方暂时无法注册账号，成员节点仍可保留，待其方便时再绑定。

共同维护时，建议由一到两位家人负责关系核对，其他人补充生平与备注，减少信息冲突。`
  },
  {
    id: 'tutorial-public-homepage',
    category: 'tutorial',
    title: '如何申请公开家族主页',
    summary: '在审核通过后对外展示家族简介与联系方式。',
    body: `公开家族主页用于向同宗或访客展示已审核通过的基本信息与简介，不等于公开全部私有资料。

申请前请确认家庭简介、籍贯与公开联系方式是否准确，并了解哪些字段会对外展示。提交后需等待平台或家庭管理员审核，通过后方可在「找家族」中被浏览。

若暂不公开，家庭仍可仅在授权成员范围内使用，不影响日常维护家谱与邀请亲人。`
  },
  {
    id: 'story-old-photo',
    category: 'story',
    title: '一张老照片背后的迁徙记忆',
    summary: '从一张泛黄合影读出三代人的迁居轨迹。',
    featured: true,
    body: `许多家庭的第一条「故事」，往往来自一张没有标注年份的老照片。照片里站着祖父与两位兄长，背景是旧时县衙门口的石阶。

后来家人比对口述，才发现拍摄地并非祖籍所在，而是迁往省城后的第一个落脚点。照片背面用铅笔写着「癸卯年冬」，成为推算迁徙时间的重要线索。

把照片与成员节点关联，并记下已知人物与地点，即使细节尚未齐全，也能为后来的亲人留下可查证的起点。`
  },
  {
    id: 'story-handwritten-names',
    category: 'story',
    title: '祖父手写族谱里的名字',
    summary: '手抄名谱中的别字与排行，藏着核对关系的钥匙。',
    body: `祖父留下的手抄名谱只有薄薄几页，却记着从曾祖到孙辈的排行字辈。部分名字与身份证用字不同，核对时需结合小名与亲属称呼交叉验证。

在整理时，可先把手抄谱中的名字录入为备注，再与现有成员关系对照。遇到缺失的一代，不必强行补全，可标注「待考证」，留给后续家人继续完善。

手抄材料的价值，在于它提示了关系线索，而不是要求一次性全部定论。`
  },
  {
    id: 'surname-origin',
    category: 'surname',
    title: '姓氏源流可以如何查证',
    summary: '从文献、族谱与口述中交叉印证姓氏来历。',
    featured: true,
    body: `查证姓氏源流，通常需要结合地方志、旧谱牒与家族口述，不宜仅凭单一网络条目下结论。

可从本族主要聚居地、堂号记载与祖先传说三条线入手，记录来源文献与可信度。若家族有南迁或分支合并经历，源流叙述可能存在多个版本，建议并列记录而非择一删改。

在 Tree 中，可将考证结论写入家庭简介或备注，并注明「待进一步核实」，方便后人接续研究。`
  },
  {
    id: 'surname-hall-name',
    category: 'surname',
    title: '堂号与郡望是什么意思',
    summary: '理解堂号、郡望在族谱中的常见用法。',
    body: `堂号是家族或分支的标志，常见于祠堂匾额与族谱扉页；郡望则多指祖先发祥或聚居之地，有时与堂号并用。

不同分支堂号可能不同，并不必然表示没有亲缘关系。阅读旧谱时，宜将堂号、郡望与具体人物世系对照，而不是单独作为唯一依据。

记录到家庭资料时，可简要写明本族常用堂号与地区，有助于同宗辨认。`
  },
  {
    id: 'surname-family-motto',
    category: 'surname',
    title: '家训为什么值得记录',
    summary: '家训反映价值观，适合作为家族文化的简短注记。',
    body: `家训不一定是长篇阔论，一句关于勤俭、读书或待人的告诫，都可能成为家族记忆的一部分。

记录家训时，建议保留原文出处（如谁传授、何时记下），并避免将现代解读混入原文。若家训来自不同分支，可分别注明来源。

这类内容适合放在家庭简介或文化注记中，供家人阅读，而不必与成员关系数据混为一谈。`
  },
  {
    id: 'article-genealogy-prep',
    category: 'article',
    title: '修谱前需要准备什么',
    summary: '整理旧谱、照片与口述，明确本次修谱范围。',
    featured: true,
    body: `修谱前先界定范围：是续写近三代，还是自开基祖重理世系。范围不同，所需材料与工作量差异很大。

建议准备现有谱牒复印件、重要成员生平摘要、迁徙与堂号资料，并列出家中有谁熟悉旧谱。没有材料时，也可从健在长辈口述开始，逐条标注待核实项。

使用数字化工具时，先把已确认的成员与关系录入，再逐步扩展，比一次性追求完整更可持续。`
  },
  {
    id: 'article-verify-kinship',
    category: 'article',
    title: '如何核对亲属关系信息',
    summary: '用关系路径与称谓交叉验证，减少录入错误。',
    body: `录入父母子女与配偶关系后，可请另一位家人按称呼复述一遍，检查是否与系统展示一致。

对于姻亲、继亲等复杂关系，建议在备注中写明性质，避免后人误解。兄弟姐妹需通过共同父母节点关联，若父母缺失，应先补上一代。

定期小幅核对，比积累大量未验证信息后再返工更省力。`
  }
]

export function getCategoryMeta(key: ContentCategory) {
  return contentCategories.find((item) => item.key === key)
}

export function getArticleById(id: string) {
  return contentArticles.find((item) => item.id === id) || null
}

export function getFeaturedArticles() {
  return contentArticles.filter((item) => item.featured)
}

const categoryRoundOrder: ContentCategory[] = ['tutorial', 'story', 'surname', 'article']

function interleaveArticlesByCategory(articles: ContentArticle[], excludeId?: string) {
  const pool = articles.filter((item) => item.id !== excludeId)
  const buckets = new Map<ContentCategory, ContentArticle[]>()

  for (const key of categoryRoundOrder) {
    buckets.set(key, [])
  }

  for (const article of pool) {
    buckets.get(article.category)?.push(article)
  }

  const result: ContentArticle[] = []
  let hasMore = true

  while (hasMore) {
    hasMore = false
    for (const key of categoryRoundOrder) {
      const bucket = buckets.get(key)
      if (bucket && bucket.length > 0) {
        result.push(bucket.shift()!)
        hasMore = true
      }
    }
  }

  return result
}

export function getSpotlightArticle() {
  return getFeaturedArticles()[0] || contentArticles[0] || null
}

export function getFeaturedListArticles() {
  const spotlight = getSpotlightArticle()
  return interleaveArticlesByCategory(getFeaturedArticles(), spotlight?.id)
}

export function getHomeReadHighlights() {
  return contentCategories.map((category) => {
    const article =
      contentArticles.find((item) => item.category === category.key && item.featured) ||
      contentArticles.find((item) => item.category === category.key) ||
      null
    return { category, article }
  })
}

export function getArticlesByCategory(category: ContentCategory | 'all') {
  if (category === 'all') return contentArticles
  return contentArticles.filter((item) => item.category === category)
}
