<template>
  <div class="family-page">
    <section class="card">
      <div class="toolbar">
        <div>
          <h2 class="page-title">家族管理</h2>
          <p class="page-desc">先选择家族，再维护关系树和家庭成员。</p>
        </div>
        <button class="btn" @click="openFamilyCreate">新增家族</button>
      </div>

      <div class="family-grid">
        <button
          v-for="item in families"
          :key="item.id"
          class="family-item"
          :class="{ active: selectedFamily?.id === item.id }"
          @click="selectFamily(item)"
        >
          <img v-if="item.cover" :src="item.cover" class="family-cover" />
          <div v-else class="family-cover empty-cover">{{ item.surname || item.name.slice(0, 1) }}</div>
          <div class="family-info">
            <strong>{{ item.name }}</strong>
            <span>{{ item.surname || '-' }} · {{ item.origin || '-' }}</span>
          </div>
          <span class="status">{{ statusText(item.status) }}</span>
        </button>
      </div>
    </section>

    <section v-if="selectedFamily" class="card">
      <div class="detail-head">
        <div>
          <h3>{{ selectedFamily.name }}</h3>
          <p>{{ familyDescription }}</p>
        </div>
        <div class="actions">
          <button class="btn secondary small" @click="openFamilyEdit(selectedFamily)">编辑家族</button>
          <button class="btn danger small" @click="removeFamily(selectedFamily.id)">删除家族</button>
        </div>
      </div>

      <div class="tabs">
        <button :class="{ active: activeTab === 'tree' }" @click="activeTab = 'tree'">关系树</button>
        <button :class="{ active: activeTab === 'members' }" @click="activeTab = 'members'">家庭成员</button>
        <button :class="{ active: activeTab === 'relations' }" @click="activeTab = 'relations'">亲属关系</button>
      </div>

      <div v-if="activeTab === 'tree'" class="tree-panel">
        <div v-if="members.length === 0" class="empty-state">暂无成员，请先添加家庭成员。</div>
        <div v-else class="generation-list">
          <div v-for="group in generationGroups" :key="group.generation" class="generation-row">
            <div class="generation-label">第 {{ group.generation }} 代</div>
            <div class="member-cards">
              <article v-for="member in group.members" :key="member.id" class="tree-member">
                <img v-if="member.avatar" :src="member.avatar" class="member-avatar" />
                <div v-else class="member-avatar placeholder">{{ member.name.slice(0, 1) }}</div>
                <strong>{{ member.name }}</strong>
                <span>{{ genderText(member.gender) }}</span>
              </article>
            </div>
          </div>
        </div>

        <div v-if="relationships.length" class="relation-summary">
          <h4>关系连接</h4>
          <div class="relation-tags">
            <span v-for="item in relationships" :key="item.id" class="relation-tag">
              {{ memberName(item.from_member_id) }} {{ relationLabel(item.relation_type) }} {{ memberName(item.to_member_id) }}
            </span>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'members'" class="section-panel">
        <div class="section-toolbar">
          <h4>家庭成员</h4>
          <button class="btn small" @click="openMemberCreate">新增成员</button>
        </div>

        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>头像</th>
                <th>姓名</th>
                <th>性别</th>
                <th>代际</th>
                <th>生日</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in members" :key="item.id">
                <td>{{ item.id }}</td>
                <td><img v-if="item.avatar" :src="item.avatar" class="cover avatar" /></td>
                <td>{{ item.name }}</td>
                <td>{{ genderText(item.gender) }}</td>
                <td>第 {{ item.generation }} 代</td>
                <td>{{ item.birthday }}</td>
                <td>{{ statusText(item.status) }}</td>
                <td>
                  <button class="btn secondary small" @click="openMemberEdit(item)">编辑</button>
                  <button class="btn danger small" @click="removeMember(item.id)">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div v-if="activeTab === 'relations'" class="section-panel">
        <div class="section-toolbar">
          <div>
            <h4>亲属关系</h4>
            <p class="hint">父母子女：A 是父母，B 是子女。配偶：A 和 B 为配偶关系。</p>
          </div>
          <button class="btn small" :disabled="members.length < 2" @click="openRelationCreate">新增关系</button>
        </div>

        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>成员 A</th>
                <th>关系</th>
                <th>成员 B</th>
                <th>备注</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in relationships" :key="item.id">
                <td>{{ item.id }}</td>
                <td>{{ memberName(item.from_member_id) }}</td>
                <td>{{ relationLabel(item.relation_type) }}</td>
                <td>{{ memberName(item.to_member_id) }}</td>
                <td>{{ item.remark }}</td>
                <td>
                  <button class="btn secondary small" @click="openRelationEdit(item)">编辑</button>
                  <button class="btn danger small" @click="removeRelation(item.id)">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <section v-else class="card empty-state">请先新增或选择一个家族。</section>

    <div v-if="showFamilyModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="familyForm.id">编辑家族</h3>
        <h3 v-else>新增家族</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>家族名称</label>
            <input v-model="familyForm.name" class="input" placeholder="张氏家族" />
          </div>
          <div class="form-row">
            <label>姓氏</label>
            <input v-model="familyForm.surname" class="input" placeholder="张" />
          </div>
          <div class="form-row">
            <label>祖籍</label>
            <input v-model="familyForm.origin" class="input" placeholder="浙江杭州" />
          </div>
          <div class="form-row">
            <label>封面</label>
            <input type="file" accept="image/*" @change="uploadFamilyCover" />
            <img v-if="familyForm.cover" :src="familyForm.cover" class="preview" />
          </div>
          <div class="form-row">
            <label>家族简介</label>
            <textarea v-model="familyForm.description" class="textarea"></textarea>
          </div>
          <div class="form-row">
            <label>状态</label>
            <select v-model.number="familyForm.status" class="select">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showFamilyModal = false">取消</button>
          <button class="btn" @click="saveFamily">保存</button>
        </div>
      </div>
    </div>

    <div v-if="showMemberModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="memberForm.id">编辑成员</h3>
        <h3 v-else>新增成员</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>姓名</label>
            <input v-model="memberForm.name" class="input" />
          </div>
          <div class="form-row">
            <label>性别</label>
            <select v-model="memberForm.gender" class="select">
              <option value="male">男</option>
              <option value="female">女</option>
              <option value="unknown">未知</option>
            </select>
          </div>
          <div class="form-row">
            <label>头像</label>
            <input type="file" accept="image/*" @change="uploadMemberAvatar" />
            <img v-if="memberForm.avatar" :src="memberForm.avatar" class="preview" />
          </div>
          <div class="form-row">
            <label>生日</label>
            <input v-model="memberForm.birthday" class="input" placeholder="1958-01-01" />
          </div>
          <div class="form-row">
            <label>出生地</label>
            <input v-model="memberForm.birth_place" class="input" />
          </div>
          <div class="form-row">
            <label>代际</label>
            <input v-model.number="memberForm.generation" type="number" class="input" />
          </div>
          <div class="form-row">
            <label>排序</label>
            <input v-model.number="memberForm.sort" type="number" class="input" />
          </div>
          <div class="form-row">
            <label>人物简介</label>
            <textarea v-model="memberForm.biography" class="textarea"></textarea>
          </div>
          <div class="form-row">
            <label>状态</label>
            <select v-model.number="memberForm.status" class="select">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showMemberModal = false">取消</button>
          <button class="btn" @click="saveMember">保存</button>
        </div>
      </div>
    </div>

    <div v-if="showRelationModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="relationForm.id">编辑关系</h3>
        <h3 v-else>新增关系</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>成员 A</label>
            <select v-model.number="relationForm.from_member_id" class="select">
              <option :value="0">请选择成员</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - 第 {{ member.generation }} 代
              </option>
            </select>
          </div>
          <div class="form-row">
            <label>关系类型</label>
            <select v-model="relationForm.relation_type" class="select">
              <option value="parent_child">父母子女</option>
              <option value="spouse">配偶</option>
            </select>
          </div>
          <div class="form-row">
            <label>成员 B</label>
            <select v-model.number="relationForm.to_member_id" class="select">
              <option :value="0">请选择成员</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - 第 {{ member.generation }} 代
              </option>
            </select>
          </div>
          <div class="form-row">
            <label>备注</label>
            <input v-model="relationForm.remark" class="input" />
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showRelationModal = false">取消</button>
          <button class="btn" @click="saveRelation">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type Family, type FamilyMember, type FamilyRelationship } from '../api'

const families = ref<Family[]>([])
const members = ref<FamilyMember[]>([])
const relationships = ref<FamilyRelationship[]>([])
const selectedFamily = ref<Family | null>(null)
const activeTab = ref<'tree' | 'members' | 'relations'>('tree')

const showFamilyModal = ref(false)
const showMemberModal = ref(false)
const showRelationModal = ref(false)

function emptyFamily(): Family {
  return {
    name: '',
    surname: '',
    origin: '',
    cover: '',
    description: '',
    status: 1
  }
}

function emptyMember(): FamilyMember {
  return {
    family_id: selectedFamily.value?.id || 0,
    name: '',
    gender: 'unknown',
    avatar: '',
    birthday: '',
    birth_place: '',
    generation: 1,
    biography: '',
    sort: 0,
    status: 1
  }
}

function emptyRelation(): FamilyRelationship {
  return {
    family_id: selectedFamily.value?.id || 0,
    from_member_id: 0,
    to_member_id: 0,
    relation_type: 'parent_child',
    remark: ''
  }
}

const familyForm = reactive<Family>(emptyFamily())
const memberForm = reactive<FamilyMember>(emptyMember())
const relationForm = reactive<FamilyRelationship>(emptyRelation())

const generationGroups = computed(() => {
  const map = new Map<number, FamilyMember[]>()

  for (const member of members.value) {
    const generation = member.generation || 1
    const group = map.get(generation) || []
    group.push(member)
    map.set(generation, group)
  }

  return Array.from(map.entries())
    .sort(([a], [b]) => a - b)
    .map(([generation, group]) => ({
      generation,
      members: group.sort((a, b) => (a.sort || 0) - (b.sort || 0))
    }))
})

const familyDescription = computed(() => selectedFamily.value?.description || '\u6682\u65E0\u5BB6\u65CF\u7B80\u4ECB')

onMounted(loadFamilies)

async function loadFamilies() {
  families.value = await api.familyList()

  if (!selectedFamily.value && families.value.length) {
    await selectFamily(families.value[0])
    return
  }

  if (selectedFamily.value) {
    selectedFamily.value = families.value.find((item) => item.id === selectedFamily.value?.id) || null
  }
}

async function selectFamily(family: Family) {
  selectedFamily.value = family
  activeTab.value = 'tree'
  await loadFamilyData()
}

async function loadFamilyData() {
  if (!selectedFamily.value?.id) {
    members.value = []
    relationships.value = []
    return
  }

  const familyId = selectedFamily.value.id
  members.value = await api.memberList(familyId)
  relationships.value = await api.relationshipList(familyId)
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function genderText(gender: string) {
  if (gender === 'male') return '\u7537'
  if (gender === 'female') return '\u5973'
  return '\u672A\u77E5'
}

function memberName(id: number) {
  return members.value.find((item) => item.id === id)?.name || '-'
}

function relationLabel(type: string) {
  if (type === 'parent_child') return '\u7236\u6BCD\u5B50\u5973'
  if (type === 'spouse') return '\u914D\u5076'
  return type
}

function openFamilyCreate() {
  Object.assign(familyForm, emptyFamily())
  showFamilyModal.value = true
}

function openFamilyEdit(item: Family) {
  Object.assign(familyForm, item)
  showFamilyModal.value = true
}

async function uploadFamilyCover(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  familyForm.cover = res.url
}

async function saveFamily() {
  if (!familyForm.name) {
    alert('\u8BF7\u8F93\u5165\u5BB6\u65CF\u540D\u79F0')
    return
  }

  if (familyForm.id) {
    await api.familyUpdate(familyForm)
  } else {
    await api.familyCreate(familyForm)
  }

  showFamilyModal.value = false
  await loadFamilies()
}

async function removeFamily(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u4E2A\u5BB6\u65CF\u5417\uFF1F\u76F8\u5173\u6210\u5458\u548C\u5173\u7CFB\u4E5F\u4F1A\u4E00\u8D77\u5220\u9664\u3002')) return

  await api.familyDelete(id)
  selectedFamily.value = null
  members.value = []
  relationships.value = []
  await loadFamilies()
}

function openMemberCreate() {
  Object.assign(memberForm, emptyMember())
  showMemberModal.value = true
}

function openMemberEdit(item: FamilyMember) {
  Object.assign(memberForm, item)
  showMemberModal.value = true
}

async function uploadMemberAvatar(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  memberForm.avatar = res.url
}

async function saveMember() {
  if (!selectedFamily.value?.id) return
  memberForm.family_id = selectedFamily.value.id

  if (!memberForm.name) {
    alert('\u8BF7\u8F93\u5165\u59D3\u540D')
    return
  }

  if (memberForm.id) {
    await api.memberUpdate(memberForm)
  } else {
    await api.memberCreate(memberForm)
  }

  showMemberModal.value = false
  await loadFamilyData()
}

async function removeMember(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u4F4D\u6210\u5458\u5417\uFF1F\u76F8\u5173\u5173\u7CFB\u4E5F\u4F1A\u4E00\u8D77\u5220\u9664\u3002')) return

  await api.memberDelete(id)
  await loadFamilyData()
}

function openRelationCreate() {
  Object.assign(relationForm, emptyRelation())
  showRelationModal.value = true
}

function openRelationEdit(item: FamilyRelationship) {
  Object.assign(relationForm, item)
  showRelationModal.value = true
}

async function saveRelation() {
  if (!selectedFamily.value?.id) return
  relationForm.family_id = selectedFamily.value.id

  if (!relationForm.from_member_id || !relationForm.to_member_id) {
    alert('\u8BF7\u9009\u62E9\u5173\u7CFB\u6210\u5458')
    return
  }

  if (relationForm.from_member_id === relationForm.to_member_id) {
    alert('\u4E24\u4F4D\u6210\u5458\u4E0D\u80FD\u662F\u540C\u4E00\u4EBA')
    return
  }

  if (relationForm.id) {
    await api.relationshipUpdate(relationForm)
  } else {
    await api.relationshipCreate(relationForm)
  }

  showRelationModal.value = false
  await loadFamilyData()
}

async function removeRelation(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u6761\u4EB2\u5C5E\u5173\u7CFB\u5417\uFF1F')) return

  await api.relationshipDelete(id)
  await loadFamilyData()
}
</script>

<style scoped>
.family-page {
  display: grid;
  gap: 18px;
}

.page-desc {
  margin: -8px 0 0;
  color: #7c4a2a;
}

.family-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.family-item {
  min-height: 86px;
  border: 1px solid #ead8c0;
  border-radius: 12px;
  background: #fff;
  padding: 12px;
  display: grid;
  grid-template-columns: 64px 1fr auto;
  align-items: center;
  gap: 12px;
  text-align: left;
}

.family-item.active,
.family-item:hover {
  border-color: #9f1d1d;
  background: #fff4e5;
}

.family-cover {
  width: 64px;
  height: 64px;
  border-radius: 10px;
  object-fit: cover;
}

.empty-cover {
  display: grid;
  place-items: center;
  background: #f3dfc7;
  color: #6b2f16;
  font-size: 24px;
  font-weight: 800;
}

.family-info {
  display: grid;
  gap: 6px;
}

.family-info span,
.status,
.hint {
  color: #7c4a2a;
  font-size: 13px;
}

.detail-head,
.section-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.detail-head h3,
.section-toolbar h4 {
  margin: 0;
}

.detail-head p,
.hint {
  margin: 8px 0 0;
}

.actions {
  display: flex;
  gap: 8px;
}

.tabs {
  display: flex;
  gap: 8px;
  margin: 20px 0;
  border-bottom: 1px solid #ead8c0;
}

.tabs button {
  border: none;
  background: transparent;
  padding: 12px 16px;
  color: #7c4a2a;
  font-weight: 700;
}

.tabs button.active {
  color: #9f1d1d;
  border-bottom: 3px solid #9f1d1d;
}

.tree-panel,
.section-panel {
  display: grid;
  gap: 18px;
}

.generation-list {
  display: grid;
  gap: 16px;
}

.generation-row {
  display: grid;
  grid-template-columns: 86px 1fr;
  gap: 14px;
  align-items: start;
}

.generation-label {
  padding: 10px;
  border-radius: 10px;
  background: #fff4e5;
  color: #6b2f16;
  font-weight: 700;
  text-align: center;
}

.member-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.tree-member {
  width: 128px;
  min-height: 146px;
  border: 1px solid #ead8c0;
  border-radius: 12px;
  background: #fff;
  padding: 12px;
  display: grid;
  justify-items: center;
  gap: 8px;
}

.member-avatar {
  width: 58px;
  height: 58px;
  border-radius: 50%;
  object-fit: cover;
  background: #f3dfc7;
}

.member-avatar.placeholder {
  display: grid;
  place-items: center;
  color: #6b2f16;
  font-size: 22px;
  font-weight: 800;
}

.relation-summary h4 {
  margin: 0 0 10px;
}

.relation-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.relation-tag {
  padding: 8px 10px;
  border-radius: 999px;
  background: #fff4e5;
  color: #6b2f16;
}

.empty-state {
  color: #7c4a2a;
}

.avatar {
  width: 58px;
  height: 58px;
  border-radius: 50%;
}

@media (max-width: 900px) {
  .detail-head,
  .section-toolbar {
    display: grid;
  }

  .generation-row {
    grid-template-columns: 1fr;
  }
}
</style>
