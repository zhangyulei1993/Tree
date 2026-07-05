<template>
  <div class="page-stack">
    <PageHeader title="内容中心" description="管理小程序与 Web 端展示的教程、故事、典故和文章。" />

    <el-alert v-if="operationError" :title="`操作失败：${operationError}`" type="error" show-icon @close="operationError = ''" />

    <el-tabs v-model="activeTab" class="content-tabs">
      <el-tab-pane label="文章管理" name="articles">
        <SearchPanel>
          <el-form>
            <el-form-item label="关键词">
              <el-input v-model="articleKeyword" clearable placeholder="标题 / 摘要 / 正文" style="width: 220px" />
            </el-form-item>
            <el-form-item label="分类">
              <el-select v-model="articleCategory" clearable placeholder="全部" style="width: 150px">
                <el-option v-for="item in categories" :key="item.key" :label="item.name" :value="item.key" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="articleStatus" clearable placeholder="全部" style="width: 150px">
                <el-option label="草稿" value="DRAFT" />
                <el-option label="已发布" value="PUBLISHED" />
                <el-option label="已归档" value="ARCHIVED" />
              </el-select>
            </el-form-item>
            <el-button type="primary" :loading="articleLoading" @click="reloadArticles">查询</el-button>
            <el-button :disabled="articleLoading" @click="resetArticleSearch">重置</el-button>
            <el-button type="success" @click="openArticleCreate">新建文章</el-button>
          </el-form>
        </SearchPanel>

        <el-alert v-if="articleLoadError" :title="`文章列表加载失败：${articleLoadError}`" type="error" show-icon :closable="false" />

        <DataTable>
          <el-table v-loading="articleLoading" :data="articles" stripe empty-text="暂无文章">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column label="类型" width="110">
              <template #default="{ row }">
                <el-tag :type="row.contentType === 'WECHAT_OFFICIAL' ? 'success' : 'info'">
                  {{ row.contentType === 'WECHAT_OFFICIAL' ? '公众号文章' : '站内文章' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="categoryName" label="分类" width="100" />
            <el-table-column prop="slug" label="路径" min-width="150" show-overflow-tooltip />
            <el-table-column label="状态" width="90">
              <template #default="{ row }"><StatusTag :status="row.status" /></template>
            </el-table-column>
            <el-table-column label="精选" width="70">
              <template #default="{ row }">
                <el-tag v-if="row.isFeatured" type="success">精选</el-tag>
                <span v-else class="muted">否</span>
              </template>
            </el-table-column>
            <el-table-column label="发布时间" width="150">
              <template #default="{ row }">{{ formatTime(row.publishedAt) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="230">
              <template #default="{ row }">
                <div class="table-actions">
                  <el-button size="small" @click="openArticleEdit(row.id)">编辑</el-button>
                  <el-button v-if="row.status !== 'PUBLISHED'" size="small" type="primary" @click="publishArticle(row.id)">发布</el-button>
                  <el-button v-else size="small" @click="unpublishArticle(row.id)">下架</el-button>
                  <el-popconfirm title="确认删除这篇文章？" @confirm="removeArticle(row.id)">
                    <template #reference>
                      <el-button size="small" type="danger" plain>删除</el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>
          <el-pagination
            class="pager"
            layout="total, sizes, prev, pager, next"
            :total="articleTotal"
            :current-page="articlePage"
            :page-size="articlePageSize"
            :page-sizes="[10, 20, 50]"
            @current-change="changeArticlePage"
            @size-change="changeArticlePageSize"
          />
        </DataTable>
      </el-tab-pane>

      <el-tab-pane label="分类管理" name="categories">
        <div class="toolbar">
          <el-button type="primary" @click="openCategoryCreate">新建分类</el-button>
          <el-button :loading="categoryLoading" @click="loadCategories">刷新</el-button>
        </div>
        <DataTable>
          <el-table v-loading="categoryLoading" :data="categories" stripe empty-text="暂无分类">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="key" label="标识" width="140" />
            <el-table-column prop="name" label="名称" width="140" />
            <el-table-column prop="description" label="说明" min-width="220" show-overflow-tooltip />
            <el-table-column prop="sortOrder" label="排序" width="90" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.isActive ? 'success' : 'info'">{{ row.isActive ? '启用' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="180">
              <template #default="{ row }">
                <el-button size="small" @click="openCategoryEdit(row)">编辑</el-button>
                <el-popconfirm title="确认删除该分类？已有文章可能无法继续归类。" @confirm="removeCategory(row.id)">
                  <template #reference>
                    <el-button size="small" type="danger" plain>删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </DataTable>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="articleDialogVisible" :title="articleEditingId ? '编辑文章' : '新建文章'" width="760px">
      <el-form label-width="88px">
        <el-form-item label="文章类型">
          <el-radio-group v-model="articleForm.contentType">
            <el-radio-button value="INTERNAL">站内文章</el-radio-button>
            <el-radio-button value="WECHAT_OFFICIAL">公众号文章</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="articleForm.categoryKey" placeholder="请选择分类">
            <el-option v-for="item in categories" :key="item.key" :label="item.name" :value="item.key" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题"><el-input v-model="articleForm.title" /></el-form-item>
        <el-form-item :label="articleForm.contentType === 'WECHAT_OFFICIAL' ? '系统路径' : '路径'">
          <el-input
            v-model="articleForm.slug"
            :disabled="articleForm.contentType === 'WECHAT_OFFICIAL'"
            :placeholder="articleForm.contentType === 'WECHAT_OFFICIAL' ? '保存时自动生成' : '如 tutorial-create-family'"
          />
        </el-form-item>
        <el-form-item label="摘要"><el-input v-model="articleForm.summary" type="textarea" :rows="2" /></el-form-item>
        <el-form-item v-if="articleForm.contentType === 'INTERNAL'" label="正文">
          <el-input v-model="articleForm.body" type="textarea" :rows="10" />
        </el-form-item>
        <template v-else>
          <el-form-item label="文章链接">
            <el-input
              v-model="articleForm.externalUrl"
              placeholder="https://mp.weixin.qq.com/s/..."
              clearable
            />
          </el-form-item>
          <el-form-item label=" ">
            <el-alert
              type="info"
              :closable="false"
              show-icon
              title="请填写公众号已发布文章的永久链接；小程序内将通过微信官方能力打开。"
            />
          </el-form-item>
        </template>
        <el-form-item label="作者"><el-input v-model="articleForm.authorName" /></el-form-item>
        <el-form-item label="来源"><el-input v-model="articleForm.source" /></el-form-item>
        <el-form-item label="属性">
          <el-switch v-model="articleForm.isFeatured" active-text="精选" />
          <el-input-number v-model="articleForm.sortOrder" class="sort-input" :min="0" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="articleForm.status">
            <el-option label="草稿" value="DRAFT" />
            <el-option label="已发布" value="PUBLISHED" />
            <el-option label="已归档" value="ARCHIVED" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="articleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitArticle">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="categoryDialogVisible" :title="categoryEditingId ? '编辑分类' : '新建分类'" width="520px">
      <el-form label-width="88px">
        <el-form-item label="标识"><el-input v-model="categoryForm.key" :disabled="Boolean(categoryEditingId)" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="categoryForm.name" /></el-form-item>
        <el-form-item label="说明"><el-input v-model="categoryForm.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="categoryForm.sortOrder" :min="0" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="categoryForm.isActive" active-text="启用" inactive-text="停用" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'

import {
  createContentArticle,
  createContentCategory,
  deleteContentArticle,
  deleteContentCategory,
  getContentArticle,
  listContentArticles,
  listContentCategories,
  publishContentArticle,
  unpublishContentArticle,
  updateContentArticle,
  updateContentCategory
} from '@/api/content'
import { getApiErrorMessage } from '@/api/client'
import DataTable from '@/components/DataTable.vue'
import PageHeader from '@/components/PageHeader.vue'
import SearchPanel from '@/components/SearchPanel.vue'
import StatusTag from '@/components/StatusTag.vue'
import type { ContentArticleInput, ContentArticleSummary, ContentCategory } from '@/types/api'

const activeTab = ref('articles')
const categories = ref<ContentCategory[]>([])
const categoryLoading = ref(false)
const categoryDialogVisible = ref(false)
const categoryEditingId = ref<number | null>(null)
const operationError = ref('')
const submitting = ref(false)

const categoryForm = reactive({
  key: '',
  name: '',
  description: '',
  sortOrder: 0,
  isActive: true
})

const articles = ref<ContentArticleSummary[]>([])
const articleKeyword = ref('')
const articleCategory = ref('')
const articleStatus = ref('')
const articlePage = ref(1)
const articlePageSize = ref(20)
const articleTotal = ref(0)
const articleLoading = ref(false)
const articleLoadError = ref('')
const articleDialogVisible = ref(false)
const articleEditingId = ref<number | null>(null)

const articleForm = reactive<ContentArticleInput>({
  categoryKey: '',
  contentType: 'INTERNAL',
  title: '',
  slug: '',
  summary: '',
  body: '',
  externalUrl: '',
  authorName: '',
  source: '',
  status: 'DRAFT',
  isFeatured: false,
  sortOrder: 0
})

onMounted(async () => {
  await loadCategories()
  await loadArticles()
})

async function loadCategories() {
  categoryLoading.value = true
  operationError.value = ''
  try {
    categories.value = await listContentCategories(true)
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    categoryLoading.value = false
  }
}

async function loadArticles() {
  articleLoading.value = true
  articleLoadError.value = ''
  try {
    const result = await listContentArticles({
      keyword: articleKeyword.value || undefined,
      categoryKey: articleCategory.value || undefined,
      status: articleStatus.value || undefined,
      page: articlePage.value,
      pageSize: articlePageSize.value
    })
    articles.value = result.items
    articleTotal.value = result.total
  } catch (error) {
    articleLoadError.value = getApiErrorMessage(error)
  } finally {
    articleLoading.value = false
  }
}

function reloadArticles() {
  articlePage.value = 1
  void loadArticles()
}

function resetArticleSearch() {
  articleKeyword.value = ''
  articleCategory.value = ''
  articleStatus.value = ''
  articlePage.value = 1
  void loadArticles()
}

function changeArticlePage(value: number) {
  articlePage.value = value
  void loadArticles()
}

function changeArticlePageSize(value: number) {
  articlePageSize.value = value
  articlePage.value = 1
  void loadArticles()
}

function openArticleCreate() {
  articleEditingId.value = null
  Object.assign(articleForm, {
    categoryKey: categories.value[0]?.key || '',
    contentType: 'INTERNAL',
    title: '',
    slug: '',
    summary: '',
    body: '',
    externalUrl: '',
    authorName: 'Tree 编辑部',
    source: '平台内容',
    status: 'DRAFT',
    isFeatured: false,
    sortOrder: 0
  })
  articleDialogVisible.value = true
}

async function openArticleEdit(id: number) {
  operationError.value = ''
  try {
    const article = await getContentArticle(id)
    articleEditingId.value = id
    Object.assign(articleForm, {
      categoryKey: article.categoryKey,
      contentType: article.contentType || 'INTERNAL',
      title: article.title,
      slug: article.slug,
      summary: article.summary || '',
      body: article.body,
      externalUrl: article.externalUrl || '',
      authorName: article.authorName || '',
      source: article.source || '',
      status: article.status,
      isFeatured: article.isFeatured,
      sortOrder: article.sortOrder
    })
    articleDialogVisible.value = true
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  }
}

async function submitArticle() {
  if (!articleForm.categoryKey) {
    ElMessage.warning('请选择文章分类')
    return
  }
  if (!articleForm.title.trim()) {
    ElMessage.warning('请填写文章标题')
    return
  }
  if (articleForm.contentType === 'INTERNAL' && !articleForm.body.trim()) {
    ElMessage.warning('请填写站内文章正文')
    return
  }
  if (articleForm.contentType === 'WECHAT_OFFICIAL' && !isWechatArticleUrl(articleForm.externalUrl || '')) {
    ElMessage.warning('请填写有效的微信公众号文章永久链接')
    return
  }
  if (articleForm.contentType === 'INTERNAL' && !articleForm.slug.trim()) {
    ElMessage.warning('请填写站内文章路径')
    return
  }
  submitting.value = true
  operationError.value = ''
  try {
    if (articleEditingId.value) {
      await updateContentArticle(articleEditingId.value, articleForm)
      ElMessage.success('文章已更新')
    } else {
      await createContentArticle(articleForm)
      ElMessage.success('文章已创建')
    }
    articleDialogVisible.value = false
    await loadArticles()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function isWechatArticleUrl(value: string) {
  return /^https:\/\/mp\.weixin\.qq\.com\/s(?:\/[^?\s#]+|\?[^#\s]+)$/i.test(value.trim())
}

async function publishArticle(id: number) {
  await runOperation(async () => {
    await publishContentArticle(id)
    ElMessage.success('文章已发布')
    await loadArticles()
  })
}

async function unpublishArticle(id: number) {
  await runOperation(async () => {
    await unpublishContentArticle(id)
    ElMessage.success('文章已下架')
    await loadArticles()
  })
}

async function removeArticle(id: number) {
  await runOperation(async () => {
    await deleteContentArticle(id)
    ElMessage.success('文章已删除')
    await loadArticles()
  })
}

function openCategoryCreate() {
  categoryEditingId.value = null
  Object.assign(categoryForm, { key: '', name: '', description: '', sortOrder: 0, isActive: true })
  categoryDialogVisible.value = true
}

function openCategoryEdit(row: ContentCategory) {
  categoryEditingId.value = row.id
  Object.assign(categoryForm, {
    key: row.key,
    name: row.name,
    description: row.description || '',
    sortOrder: row.sortOrder,
    isActive: row.isActive
  })
  categoryDialogVisible.value = true
}

async function submitCategory() {
  submitting.value = true
  operationError.value = ''
  try {
    if (categoryEditingId.value) {
      await updateContentCategory(categoryEditingId.value, categoryForm)
      ElMessage.success('分类已更新')
    } else {
      await createContentCategory(categoryForm)
      ElMessage.success('分类已创建')
    }
    categoryDialogVisible.value = false
    await loadCategories()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

async function removeCategory(id: number) {
  await runOperation(async () => {
    await deleteContentCategory(id)
    ElMessage.success('分类已删除')
    await loadCategories()
  })
}

async function runOperation(fn: () => Promise<void>) {
  operationError.value = ''
  try {
    await fn()
  } catch (error) {
    operationError.value = getApiErrorMessage(error)
  }
}

function formatTime(value?: string | null) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'
}
</script>

<style scoped>
.content-tabs {
  border-radius: 8px;
  background: #fff;
  padding: 16px;
}

.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.table-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.pager {
  justify-content: flex-end;
  padding: 14px;
}

.sort-input {
  margin-left: 18px;
}
</style>
