
<template>
  <div class="family-page">
    <section class="panel family-list-panel">
      <div class="panel-head">
        <div>
          <h2>家族管理</h2>
          <p>先选择家族，再维护该家族下的关系树、家庭成员和亲属关系。</p>
        </div>
        <button class="primary-btn" @click="openFamilyCreate">新增家族</button>
      </div>

      <div class="filter-bar">
        <input
          v-model="query.keyword"
          class="input"
          placeholder="搜索家族名称、姓氏、籍贯"
          @keyup.enter="loadFamilies"
        />
        <select v-model="query.status" class="select">
          <option value="">全部状态</option>
          <option value="1">显示中</option>
          <option value="0">已隐藏</option>
          <option value="2">已归档</option>
        </select>
        <button class="light-btn" @click="loadFamilies">查询</button>
      </div>

      <div v-if="recentFamilies.length" class="recent-row">
        <span>最近管理</span>
        <button
          v-for="family in recentFamilies"
          :key="family.id"
          class="recent-chip"
          :class="{ active: currentFamily && String(currentFamily.id) === String(family.id) }"
          @click="selectFamily(family)"
        >
          {{ family.name }}
        </button>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>家族名称</th>
              <th>姓氏</th>
              <th>籍贯</th>
              <th>成员数</th>
              <th>关系数</th>
              <th>状态</th>
              <th>更新时间</th>
              <th class="op-col">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="family in families"
              :key="family.id"
              :class="{ selected: currentFamily && String(currentFamily.id) === String(family.id) }"
            >
              <td>
                <div class="family-name-cell">
                  <div class="avatar">{{ firstChar(family.surname || family.name) }}</div>
                  <div>
                    <strong>{{ family.name }}</strong>
                    <small>{{ family.description || '暂无家族简介' }}</small>
                  </div>
                </div>
              </td>
              <td>{{ family.surname || '-' }}</td>
              <td>{{ family.origin || '-' }}</td>
              <td>{{ family.member_count ?? family.memberCount ?? '-' }}</td>
              <td>{{ family.relationship_count ?? family.relationshipCount ?? '-' }}</td>
              <td>
                <span class="status" :class="'s-' + Number(family.status ?? 1)">
                  {{ statusText(family.status) }}
                </span>
              </td>
              <td>{{ formatDate(family.updated_at || family.updatedAt || family.created_at || family.createdAt) }}</td>
              <td>
                <div class="ops">
                  <button class="primary-mini" @click="selectFamily(family)">管理</button>
                  <button class="light-mini" @click="openFamilyEdit(family)">编辑</button>
                  <button class="danger-mini" @click="deleteFamily(family)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="!families.length && !loading">
              <td colspan="8" class="empty">暂无家族数据</td>
            </tr>
            <tr v-if="loading">
              <td colspan="8" class="empty">加载中...</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="panel workspace-panel" v-if="currentFamily">
      <div class="workspace-head">
        <div>
          <h2>{{ currentFamily.name }}</h2>
          <p>
            {{ currentFamily.surname || '-' }} · {{ currentFamily.origin || '暂无籍贯' }}
          </p>
        </div>
        <div class="head-actions">
          <button class="light-btn" @click="openFamilyEdit(currentFamily)">编辑家族</button>
          <button class="danger-btn" @click="deleteFamily(currentFamily)">删除家族</button>
        </div>
      </div>

      <div class="tabs">
        <button :class="{ active: activeTab === 'tree' }" @click="activeTab = 'tree'">关系树</button>
        <button :class="{ active: activeTab === 'members' }" @click="activeTab = 'members'">家庭成员</button>
        <button :class="{ active: activeTab === 'relations' }" @click="activeTab = 'relations'">亲属关系</button>
      </div>

      <div v-if="activeTab === 'tree'" class="tab-body">
        <div v-if="members.length" class="simple-tree">
          <div
            v-for="member in members"
            :key="member.id"
            class="member-node"
          >
            <div class="member-avatar">{{ firstChar(member.name) }}</div>
            <div>
              <strong>{{ member.name }}</strong>
              <small>{{ member.title || member.generation_name || member.generationName || '家庭成员' }}</small>
            </div>
          </div>
        </div>
        <div v-else class="empty-card">
          暂无成员，请先添加家庭成员。
        </div>
      </div>

      <div v-if="activeTab === 'members'" class="tab-body">
        <div class="sub-head">
          <div>
            <h3>家庭成员</h3>
            <p>维护当前家族下的成员信息。</p>
          </div>
          <button class="primary-btn" @click="openMemberCreate">新增成员</button>
        </div>

        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>姓名</th>
                <th>性别</th>
                <th>称谓/身份</th>
                <th>出生地</th>
                <th>排序</th>
                <th>状态</th>
                <th class="op-col">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="member in members" :key="member.id">
                <td>
                  <div class="family-name-cell">
                    <div class="avatar small">{{ firstChar(member.name) }}</div>
                    <strong>{{ member.name }}</strong>
                  </div>
                </td>
                <td>{{ genderText(member.gender) }}</td>
                <td>{{ member.title || member.generation_name || member.generationName || '-' }}</td>
                <td>{{ member.birth_place || member.birthPlace || '-' }}</td>
                <td>{{ member.sort ?? 0 }}</td>
                <td>
                  <span class="status" :class="'s-' + Number(member.status ?? 1)">
                    {{ statusText(member.status) }}
                  </span>
                </td>
                <td>
                  <div class="ops">
                    <button class="light-mini" @click="openMemberEdit(member)">编辑</button>
                    <button class="danger-mini" @click="deleteMember(member)">删除</button>
                  </div>
                </td>
              </tr>
              <tr v-if="!members.length">
                <td colspan="7" class="empty">暂无成员</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="activeTab === 'relations'" class="tab-body">
        <div class="sub-head">
          <div>
            <h3>亲属关系</h3>
            <p>维护当前家族成员之间的父子、母子、配偶、兄妹等关系。</p>
          </div>
          <button class="primary-btn" @click="openRelationCreate">新增关系</button>
        </div>

        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>成员一</th>
                <th>成员二</th>
                <th>关系</th>
                <th>类型</th>
                <th>说明</th>
                <th>状态</th>
                <th class="op-col">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="relation in relations" :key="relation.id">
                <td>{{ memberName(relation.from_member_id || relation.fromMemberId || relation.from_node_id || relation.fromNodeId) }}</td>
                <td>{{ memberName(relation.to_member_id || relation.toMemberId || relation.to_node_id || relation.toNodeId) }}</td>
                <td>{{ relationTypeName(relation.relation_type || relation.relationType) }}</td>
                <td>{{ relation.relation_type || relation.relationType || '-' }}</td>
                <td>{{ relation.remark || relation.description || '-' }}</td>
                <td>
                  <span class="status" :class="'s-' + Number(relation.status ?? 1)">
                    {{ statusText(relation.status) }}
                  </span>
                </td>
                <td>
                  <div class="ops">
                    <button class="light-mini" @click="openRelationEdit(relation)">编辑</button>
                    <button class="danger-mini" @click="deleteRelation(relation)">删除</button>
                  </div>
                </td>
              </tr>
              <tr v-if="!relations.length">
                <td colspan="7" class="empty">暂无亲属关系</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <section class="panel empty-workspace" v-else>
      <h2>请选择一个家族</h2>
      <p>从上方家族列表点击“管理”，即可进入该家族的成员和关系维护。</p>
    </section>

    <div v-if="familyDialog" class="modal-mask">
      <form class="modal" @submit.prevent="saveFamily">
        <h3>{{ familyForm.id ? '编辑家族' : '新增家族' }}</h3>
        <div class="form-grid">
          <label>
            家族名称
            <input v-model="familyForm.name" required placeholder="例如：王家一家人" />
          </label>
          <label>
            姓氏
            <input v-model="familyForm.surname" placeholder="例如：王" />
          </label>
          <label>
            籍贯
            <input v-model="familyForm.origin" placeholder="例如：湖北沙市" />
          </label>
          <label>
            状态
            <select v-model="familyForm.status">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
              <option :value="2">归档</option>
            </select>
          </label>
          <label class="full">
            简介
            <textarea v-model="familyForm.description" rows="4" placeholder="家族简介"></textarea>
          </label>
        </div>
        <div class="modal-actions">
          <button type="button" class="light-btn" @click="familyDialog = false">取消</button>
          <button type="submit" class="primary-btn">保存</button>
        </div>
      </form>
    </div>

    <div v-if="memberDialog" class="modal-mask">
      <form class="modal" @submit.prevent="saveMember">
        <h3>{{ memberForm.id ? '编辑成员' : '新增成员' }}</h3>
        <div class="form-grid">
          <label>
            姓名
            <input v-model="memberForm.name" required placeholder="成员姓名" />
          </label>
          <label>
            性别
            <select v-model="memberForm.gender">
              <option value="">未知</option>
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </label>
          <label>
            称谓/身份
            <input v-model="memberForm.title" placeholder="例如：父亲、母亲、长子" />
          </label>
          <label>
            出生地
            <input v-model="memberForm.birth_place" placeholder="出生地" />
          </label>
          <label>
            排序
            <input v-model.number="memberForm.sort" type="number" />
          </label>
          <label>
            状态
            <select v-model="memberForm.status">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </label>
          <label class="full">
            说明
            <textarea v-model="memberForm.description" rows="3" placeholder="成员说明"></textarea>
          </label>
        </div>
        <div class="modal-actions">
          <button type="button" class="light-btn" @click="memberDialog = false">取消</button>
          <button type="submit" class="primary-btn">保存</button>
        </div>
      </form>
    </div>

    <div v-if="relationDialog" class="modal-mask">
      <form class="modal" @submit.prevent="saveRelation">
        <h3>{{ relationForm.id ? '编辑关系' : '新增关系' }}</h3>
        <div class="form-grid">
          <label>
            成员一
            <select v-model="relationForm.from_member_id" required>
              <option value="">请选择</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }}
              </option>
            </select>
          </label>
          <label>
            成员二
            <select v-model="relationForm.to_member_id" required>
              <option value="">请选择</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }}
              </option>
            </select>
          </label>
          <label>
            关系名称
            <input v-model="relationForm.relation_name" required placeholder="例如：父子、母子、配偶" />
          </label>
          <label>
            关系类型
            <select v-model="relationForm.relation_type" required>
              <option value="">请选择关系类型</option>
              <option v-for="type in activeRelationshipTypes" :key="type.code" :value="type.code">
                {{ type.name }}
              </option>
            </select>
          </label>
          <label>
            状态
            <select v-model="relationForm.status">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </label>
          <label class="full">
            说明
            <textarea v-model="relationForm.remark" rows="3" placeholder="关系说明"></textarea>
          </label>
        </div>
        <div class="modal-actions">
          <button type="button" class="light-btn" @click="relationDialog = false">取消</button>
          <button type="submit" class="primary-btn">保存</button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type RelationshipType } from '../api'

type AnyItem = Record<string, any>

const loading = ref(false)
const families = ref<AnyItem[]>([])
const members = ref<AnyItem[]>([])
const relations = ref<AnyItem[]>([])
const relationshipTypes = ref<RelationshipType[]>([])
const currentFamily = ref<AnyItem | null>(null)
const activeTab = ref('tree')

const query = reactive({
  keyword: '',
  status: ''
})

const familyDialog = ref(false)
const memberDialog = ref(false)
const relationDialog = ref(false)

const familyForm = reactive<AnyItem>({
  id: '',
  name: '',
  surname: '',
  origin: '',
  description: '',
  status: 1
})

const memberForm = reactive<AnyItem>({
  id: '',
  family_id: '',
  name: '',
  gender: '',
  title: '',
  birth_place: '',
  description: '',
  sort: 0,
  status: 1
})

const relationForm = reactive<AnyItem>({
  id: '',
  family_id: '',
  from_member_id: '',
  to_member_id: '',
  relation_name: '',
  relation_type: '',
  remark: '',
  status: 1
})

const recentFamilies = computed(() => {
  const ids = getRecentFamilyIds()
  return ids
    .map((id) => families.value.find((item) => String(item.id) === String(id)))
    .filter(Boolean) as AnyItem[]
})

const activeRelationshipTypes = computed(() => relationshipTypes.value.filter((item) => item.status === 1))

async function loadRelationshipTypes() {
  relationshipTypes.value = await api.relationshipTypeList()
}

async function loadFamilies() {
  loading.value = true

  try {
    families.value = await api.familyList({
      keyword: query.keyword,
      status: query.status,
      page: 1,
      page_size: 100
    })

    const savedId = localStorage.getItem('currentFamilyId')
    const savedFamily = families.value.find((item) => String(item.id) === String(savedId))

    if (!currentFamily.value && savedFamily) {
      selectFamily(savedFamily, false)
    } else if (!currentFamily.value && families.value.length) {
      selectFamily(families.value[0], false)
    }
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  } finally {
    loading.value = false
  }
}

async function loadMembers() {
  if (!currentFamily.value) return

  members.value = await api.memberList(currentFamily.value.id)
}

async function loadRelations() {
  if (!currentFamily.value) return

  relations.value = await api.relationshipList(currentFamily.value.id)
}

async function reloadWorkspace() {
  if (!currentFamily.value) return

  try {
    await Promise.all([loadMembers(), loadRelations()])
  } catch (error) {
    console.warn(error)
  }
}

function selectFamily(family: AnyItem, saveRecent = true) {
  currentFamily.value = family
  localStorage.setItem('currentFamilyId', String(family.id))

  if (saveRecent) {
    saveRecentFamilyId(family.id)
  }

  reloadWorkspace()
}

function getRecentFamilyIds() {
  try {
    const raw = localStorage.getItem('recentFamilyIds')
    const value = raw ? JSON.parse(raw) : []
    return Array.isArray(value) ? value : []
  } catch {
    return []
  }
}

function saveRecentFamilyId(id: any) {
  const ids = getRecentFamilyIds().filter((item) => String(item) !== String(id))
  ids.unshift(id)
  localStorage.setItem('recentFamilyIds', JSON.stringify(ids.slice(0, 6)))
}

function resetFamilyForm() {
  Object.assign(familyForm, {
    id: '',
    name: '',
    surname: '',
    origin: '',
    description: '',
    status: 1
  })
}

function openFamilyCreate() {
  resetFamilyForm()
  familyDialog.value = true
}

function openFamilyEdit(row: AnyItem) {
  Object.assign(familyForm, {
    id: row.id,
    name: row.name || '',
    surname: row.surname || '',
    origin: row.origin || '',
    description: row.description || '',
    status: Number(row.status ?? 1)
  })

  familyDialog.value = true
}

async function saveFamily() {
  const payload = { ...familyForm }

  try {
    if (payload.id) {
      await api.familyUpdate(payload)
    } else {
      delete payload.id

      await api.familyCreate(payload)
    }

    familyDialog.value = false
    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

async function deleteFamily(row: AnyItem) {
  if (!confirm('确认删除家族「' + row.name + '」吗？')) return

  try {
    await api.familyDelete(row.id)

    if (currentFamily.value && String(currentFamily.value.id) === String(row.id)) {
      currentFamily.value = null
      members.value = []
      relations.value = []
      localStorage.removeItem('currentFamilyId')
    }

    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

function resetMemberForm() {
  Object.assign(memberForm, {
    id: '',
    family_id: currentFamily.value?.id || '',
    name: '',
    gender: '',
    title: '',
    birth_place: '',
    description: '',
    sort: 0,
    status: 1
  })
}

function openMemberCreate() {
  resetMemberForm()
  memberDialog.value = true
}

function openMemberEdit(row: AnyItem) {
  Object.assign(memberForm, {
    id: row.id,
    family_id: row.family_id || row.familyId || currentFamily.value?.id || '',
    name: row.name || '',
    gender: row.gender || '',
    title: row.title || '',
    birth_place: row.birth_place || row.birthPlace || '',
    description: row.description || '',
    sort: Number(row.sort ?? 0),
    status: Number(row.status ?? 1)
  })

  memberDialog.value = true
}

async function saveMember() {
  const payload: AnyItem = {
    ...memberForm,
    family_id: currentFamily.value?.id
  }

  try {
    if (payload.id) {
      await api.memberUpdate(payload)
    } else {
      delete payload.id

      await api.memberCreate(payload)
    }

    memberDialog.value = false
    await loadMembers()
    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

async function deleteMember(row: AnyItem) {
  if (!confirm('确认删除成员「' + row.name + '」吗？')) return

  try {
    await api.memberDelete(row.id)

    await loadMembers()
    await loadRelations()
    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

function resetRelationForm() {
  Object.assign(relationForm, {
    id: '',
    family_id: currentFamily.value?.id || '',
    from_member_id: '',
    to_member_id: '',
    relation_name: '',
    relation_type: '',
    remark: '',
    status: 1
  })
}

function openRelationCreate() {
  resetRelationForm()
  relationDialog.value = true
}

function openRelationEdit(row: AnyItem) {
  Object.assign(relationForm, {
    id: row.id,
    family_id: row.family_id || row.familyId || currentFamily.value?.id || '',
    from_member_id: row.from_member_id || row.fromMemberId || row.from_node_id || row.fromNodeId || '',
    to_member_id: row.to_member_id || row.toMemberId || row.to_node_id || row.toNodeId || '',
    relation_name: row.relation_name || row.relationName || '',
    relation_type: row.relation_type || row.relationType || '',
    remark: row.remark || row.description || '',
    status: Number(row.status ?? 1)
  })

  relationDialog.value = true
}

async function saveRelation() {
  if (String(relationForm.from_member_id) === String(relationForm.to_member_id)) {
    alert('两个成员不能相同')
    return
  }

  const payload: AnyItem = {
    ...relationForm,
    family_id: currentFamily.value?.id,
    from_node_id: relationForm.from_member_id,
    to_node_id: relationForm.to_member_id
  }

  try {
    if (payload.id) {
      await api.relationshipUpdate(payload)
    } else {
      delete payload.id

      await api.relationshipCreate(payload)
    }

    relationDialog.value = false
    await loadRelations()
    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

async function deleteRelation(row: AnyItem) {
  if (!confirm('确认删除这条亲属关系吗？')) return

  try {
    await api.relationshipDelete(row.id)

    await loadRelations()
    await loadFamilies()
  } catch (error) {
    alert(error instanceof Error ? error.message : String(error))
  }
}

function memberName(id: any) {
  const member = members.value.find((item) => String(item.id) === String(id))
  return member?.name || '-'
}

function relationTypeName(code: any) {
  return relationshipTypes.value.find((item) => item.code === code)?.name || code || '-'
}

function firstChar(value: any) {
  const text = String(value || '家')
  return text.slice(0, 1)
}

function statusText(value: any) {
  const status = Number(value ?? 1)

  if (status === 0) return '隐藏'
  if (status === 2) return '归档'

  return '显示'
}

function genderText(value: any) {
  if (value === 'male') return '男'
  if (value === 'female') return '女'

  return '未知'
}

function formatDate(value: any) {
  if (!value) return '-'
  return String(value).slice(0, 10)
}

onMounted(() => {
  loadRelationshipTypes()
  loadFamilies()
})
</script>

<style scoped>
.family-page {
  padding: 24px;
  color: #3b1608;
}

.panel {
  background: rgba(255, 250, 243, 0.92);
  border: 1px solid #ead3b5;
  border-radius: 18px;
  padding: 22px;
  box-shadow: 0 18px 45px rgba(100, 56, 20, 0.08);
  margin-bottom: 18px;
}

.panel-head,
.workspace-head,
.sub-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel h2,
.workspace-head h2,
.sub-head h3 {
  margin: 0;
  font-size: 24px;
  font-weight: 900;
}

.panel p,
.workspace-head p,
.sub-head p {
  margin: 8px 0 0;
  color: #7c3f1d;
  font-size: 14px;
}

.filter-bar {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) 160px 88px;
  gap: 12px;
  margin: 20px 0 14px;
}

.input,
.select,
.modal input,
.modal select,
.modal textarea {
  width: 100%;
  height: 38px;
  border: 1px solid #ead3b5;
  border-radius: 10px;
  background: #fffaf5;
  color: #3b1608;
  padding: 0 12px;
  outline: none;
}

.modal textarea {
  height: auto;
  padding: 10px 12px;
  resize: vertical;
}

.primary-btn,
.light-btn,
.danger-btn,
.primary-mini,
.light-mini,
.danger-mini {
  border: 0;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 800;
  white-space: nowrap;
}

.primary-btn {
  height: 42px;
  padding: 0 18px;
  background: #a6171c;
  color: #fffaf0;
}

.light-btn {
  height: 38px;
  padding: 0 16px;
  background: #f3dec3;
  color: #7c2d12;
}

.danger-btn {
  height: 38px;
  padding: 0 16px;
  background: #b91c1c;
  color: #fff;
}

.primary-mini,
.light-mini,
.danger-mini {
  height: 30px;
  padding: 0 10px;
  font-size: 12px;
}

.primary-mini {
  background: #a6171c;
  color: #fffaf0;
}

.light-mini {
  background: #f3dec3;
  color: #7c2d12;
}

.danger-mini {
  background: #b91c1c;
  color: #fff;
}

.recent-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0 16px;
  color: #7c3f1d;
  font-size: 13px;
}

.recent-chip {
  border: 1px solid #ead3b5;
  background: #fff7ed;
  color: #7c2d12;
  border-radius: 999px;
  padding: 6px 12px;
  cursor: pointer;
}

.recent-chip.active {
  background: #a6171c;
  border-color: #a6171c;
  color: #fffaf0;
}

.table-wrap {
  overflow-x: auto;
}

.table {
  width: 100%;
  border-collapse: collapse;
  min-width: 860px;
}

.table th {
  text-align: left;
  background: #fff2df;
  color: #7c2d12;
  padding: 12px 14px;
  font-size: 13px;
  border-bottom: 1px solid #ead3b5;
}

.table td {
  padding: 14px;
  border-bottom: 1px solid #f0ddc4;
  font-size: 14px;
  vertical-align: middle;
}

.table tr.selected td {
  background: rgba(166, 23, 28, 0.055);
}

.op-col {
  width: 210px;
}

.ops {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.family-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.family-name-cell small {
  display: block;
  margin-top: 4px;
  color: #8a5b3c;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.avatar,
.member-avatar {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: #f3dec3;
  color: #7c2d12;
  display: grid;
  place-items: center;
  font-weight: 900;
  font-size: 22px;
  flex: 0 0 auto;
}

.avatar.small {
  width: 32px;
  height: 32px;
  font-size: 16px;
}

.status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 48px;
  height: 24px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 800;
}

.status.s-1 {
  background: #fff2df;
  color: #9a3412;
}

.status.s-0 {
  background: #f1f5f9;
  color: #64748b;
}

.status.s-2 {
  background: #fee2e2;
  color: #991b1b;
}

.empty,
.empty-card {
  text-align: center;
  color: #8a5b3c;
  padding: 28px;
}

.tabs {
  display: flex;
  gap: 24px;
  margin-top: 26px;
  border-bottom: 1px solid #ead3b5;
}

.tabs button {
  border: 0;
  background: transparent;
  color: #7c3f1d;
  padding: 0 0 12px;
  font-weight: 900;
  cursor: pointer;
  border-bottom: 3px solid transparent;
}

.tabs button.active {
  color: #a6171c;
  border-bottom-color: #a6171c;
}

.tab-body {
  padding-top: 20px;
}

.simple-tree {
  min-height: 280px;
  border: 1px dashed #ead3b5;
  border-radius: 18px;
  background:
    linear-gradient(90deg, rgba(127, 29, 29, 0.04) 1px, transparent 1px),
    linear-gradient(180deg, rgba(127, 29, 29, 0.04) 1px, transparent 1px),
    #fffaf3;
  background-size: 28px 28px;
  padding: 24px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 14px;
}

.member-node {
  background: rgba(255, 250, 243, 0.95);
  border: 1px solid #ead3b5;
  border-radius: 14px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.member-node small {
  display: block;
  margin-top: 3px;
  color: #8a5b3c;
}

.empty-workspace {
  text-align: center;
  padding: 48px 24px;
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(20, 10, 6, 0.42);
  display: grid;
  place-items: center;
  z-index: 999;
  padding: 20px;
}

.modal {
  width: min(720px, 100%);
  background: #fffaf3;
  border-radius: 18px;
  border: 1px solid #ead3b5;
  box-shadow: 0 28px 80px rgba(0, 0, 0, 0.24);
  padding: 22px;
}

.modal h3 {
  margin: 0 0 18px;
  font-size: 22px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.form-grid label {
  display: grid;
  gap: 7px;
  font-weight: 800;
  color: #7c2d12;
  font-size: 13px;
}

.form-grid .full {
  grid-column: 1 / -1;
}

.modal-actions {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .family-page {
    padding: 12px;
  }

  .panel {
    padding: 16px;
    border-radius: 16px;
  }

  .panel-head,
  .workspace-head,
  .sub-head {
    flex-direction: column;
  }

  .filter-bar {
    grid-template-columns: 1fr;
  }

  .head-actions {
    display: flex;
    gap: 10px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .tabs {
    gap: 18px;
    overflow-x: auto;
  }
}
</style>
