<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">家庭成员</span>
          <h1>{{ family?.familyName || '成员管理' }}</h1>
          <p class="muted">当前角色：{{ family?.role || '读取中' }}</p>
        </div>
        <div class="heading-actions">
          <RouterLink class="button secondary" :to="`/families/${familyId}`">家庭详情</RouterLink>
          <RouterLink class="button secondary" :to="`/families/${familyId}/tree`">查看家庭树</RouterLink>
        </div>
      </div>

      <section v-if="pageLoading" class="card state-panel">正在加载成员...</section>
      <section v-else-if="pageError" class="card state-panel error" role="alert">
        <strong>{{ forbidden ? '无权访问成员数据' : '成员数据加载失败' }}</strong>
        <span>{{ pageError }}</span>
        <button class="button secondary" @click="loadPage">重新加载</button>
      </section>
      <template v-else>
        <div v-if="canManage" class="forms-grid">
          <form class="card form-card" @submit.prevent="submitMember">
            <div>
              <h2>新增独立成员</h2>
              <p class="muted">只创建成员节点，不建立亲属关系。</p>
            </div>
            <label>
              <span>姓名</span>
              <input v-model.trim="memberForm.name" class="field" maxlength="100" placeholder="成员姓名" />
            </label>
            <div class="compact-grid">
              <label>
                <span>性别</span>
                <select v-model="memberForm.gender" class="field">
                  <option value="UNKNOWN">未知</option>
                  <option value="MALE">男</option>
                  <option value="FEMALE">女</option>
                </select>
              </label>
              <label>
                <span>出生年份</span>
                <input v-model.number="memberForm.birthYear" class="field" type="number" min="1" max="9999" placeholder="可选" />
              </label>
            </div>
            <label class="check-row">
              <input v-model="memberForm.isAlive" type="checkbox" />
              <span>当前健在</span>
            </label>
            <p v-if="memberError" class="feedback error" role="alert">{{ memberError }}</p>
            <button class="button" :disabled="memberSubmitting">
              {{ memberSubmitting ? '创建中...' : '新增成员' }}
            </button>
          </form>

          <form class="card form-card" @submit.prevent="submitRelationship">
            <div>
              <h2>添加新亲属</h2>
              <p class="notice">此操作会创建一个新成员并建立关系，不是连接两个已有成员。</p>
            </div>
            <label>
              <span>基础成员</span>
              <select v-model.number="relationshipForm.baseMemberId" class="field">
                <option :value="0" disabled>请选择成员</option>
                <option v-for="member in members" :key="member.memberId" :value="member.memberId">
                  #{{ member.memberId }} {{ member.name }}
                </option>
              </select>
            </label>
            <div class="compact-grid">
              <label>
                <span>亲属类型</span>
                <select v-model="relationshipForm.addType" class="field">
                  <option value="ADD_FATHER">添加父亲</option>
                  <option value="ADD_MOTHER">添加母亲</option>
                  <option value="ADD_CHILD">添加子女</option>
                  <option value="ADD_SPOUSE">添加配偶</option>
                  <option value="ADD_SIBLING">添加兄弟姐妹</option>
                </select>
              </label>
              <label>
                <span>新亲属性别</span>
                <select v-model="relationshipForm.gender" class="field">
                  <option value="UNKNOWN">未知</option>
                  <option value="MALE">男</option>
                  <option value="FEMALE">女</option>
                </select>
              </label>
            </div>
            <label>
              <span>新亲属姓名</span>
              <input v-model.trim="relationshipForm.name" class="field" maxlength="100" placeholder="将创建的新成员姓名" />
            </label>
            <label v-if="relationshipForm.addType !== 'ADD_SPOUSE'">
              <span>父子关系属性</span>
              <select v-model="relationshipForm.parentLinkType" class="field">
                <option value="PRIMARY">PRIMARY</option>
                <option value="STEP">STEP</option>
                <option value="ADOPTIVE">ADOPTIVE</option>
                <option value="SUCCESSION">SUCCESSION</option>
                <option value="NOTE_ONLY">NOTE_ONLY</option>
                <option value="OTHER">OTHER</option>
              </select>
            </label>
            <p v-if="relationshipError" class="feedback error" role="alert">{{ relationshipError }}</p>
            <p v-if="relationshipSuccess" class="feedback success" role="status">{{ relationshipSuccess }}</p>
            <button class="button" :disabled="relationshipSubmitting || members.length === 0">
              {{ relationshipSubmitting ? '创建中...' : '创建新亲属及关系' }}
            </button>
          </form>
        </div>

        <section v-else class="card readonly-note">
          当前角色为 MEMBER，仅可查看成员和家庭树。
        </section>

        <section class="member-section">
          <div class="section-heading">
            <h2>成员列表</h2>
            <button class="button secondary" :disabled="membersLoading" @click="loadMembers">刷新</button>
          </div>
          <section v-if="membersLoading" class="card state-panel">正在刷新成员列表...</section>
          <section v-else-if="members.length === 0" class="card state-panel">
            <strong>暂无成员</strong>
          </section>
          <div v-else class="member-list">
            <article v-for="member in members" :key="member.memberId" class="card member-card">
              <div class="member-title">
                <strong>#{{ member.memberId }} {{ member.name }}</strong>
                <span class="status-tag">{{ member.status }}</span>
              </div>
              <dl>
                <div><dt>性别</dt><dd>{{ genderLabel(member.gender) }}</dd></div>
                <div><dt>字辈/代次</dt><dd>后端未提供</dd></div>
                <div><dt>家庭角色</dt><dd>{{ member.boundFamilyRole || '未绑定角色' }}</dd></div>
                <div><dt>健在状态</dt><dd>{{ aliveLabel(member) }}</dd></div>
              </dl>
            </article>
          </div>
        </section>
      </template>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import axios from 'axios'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { getFamilyDetail } from '@/api/families'
import { createMember, listMembers } from '@/api/members'
import { createRelationship } from '@/api/relationships'
import PageShell from '@/components/PageShell.vue'
import type {
  ApiResponse,
  FamilyDetail,
  FamilyMember,
  Gender,
  RelationshipAddType
} from '@/types/api'

const route = useRoute()
const familyId = computed(() => String(route.params.familyId))
const family = ref<FamilyDetail | null>(null)
const members = ref<FamilyMember[]>([])
const pageLoading = ref(true)
const membersLoading = ref(false)
const pageError = ref('')
const forbidden = ref(false)
const memberSubmitting = ref(false)
const relationshipSubmitting = ref(false)
const memberError = ref('')
const relationshipError = ref('')
const relationshipSuccess = ref('')

const memberForm = reactive({
  name: '',
  gender: 'UNKNOWN' as Gender,
  birthYear: undefined as number | undefined,
  isAlive: true
})

const relationshipForm = reactive({
  baseMemberId: 0,
  addType: 'ADD_CHILD' as RelationshipAddType,
  name: '',
  gender: 'UNKNOWN' as Gender,
  parentLinkType: 'PRIMARY'
})

const canManage = computed(() => family.value?.role === 'FOUNDER' || family.value?.role === 'FAMILY_ADMIN')

function errorText(requestError: unknown, fallback: string) {
  if (axios.isAxiosError<ApiResponse<unknown>>(requestError)) {
    forbidden.value = requestError.response?.status === 403
    if (requestError.response?.data?.code) {
      return `${requestError.response.data.code}：${requestError.response.data.message}`
    }
  }
  return apiErrorMessage(requestError, fallback)
}

function genderLabel(gender: string) {
  return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未知'
}

function aliveLabel(member: FamilyMember) {
  if (member.isAlive === true) return '健在'
  if (member.isAlive === false) return '已故'
  return '未填写'
}

async function loadMembers() {
  membersLoading.value = true
  try {
    members.value = await listMembers(familyId.value)
    if (!relationshipForm.baseMemberId && members.value.length > 0) {
      relationshipForm.baseMemberId = members.value[0].memberId
    }
  } catch (requestError) {
    pageError.value = errorText(requestError, '无法加载成员列表。')
  } finally {
    membersLoading.value = false
  }
}

async function loadPage() {
  pageLoading.value = true
  pageError.value = ''
  forbidden.value = false
  try {
    const [familyResult, memberResult] = await Promise.all([
      getFamilyDetail(familyId.value),
      listMembers(familyId.value)
    ])
    family.value = familyResult
    members.value = memberResult
    if (memberResult.length > 0) relationshipForm.baseMemberId = memberResult[0].memberId
  } catch (requestError) {
    pageError.value = errorText(requestError, '无法加载家庭成员页面。')
  } finally {
    pageLoading.value = false
  }
}

async function submitMember() {
  memberError.value = ''
  if (!memberForm.name) {
    memberError.value = '成员姓名不能为空。'
    return
  }
  memberSubmitting.value = true
  try {
    await createMember(familyId.value, {
      name: memberForm.name,
      gender: memberForm.gender,
      birthYear: memberForm.birthYear || undefined,
      isAlive: memberForm.isAlive
    })
    memberForm.name = ''
    memberForm.gender = 'UNKNOWN'
    memberForm.birthYear = undefined
    memberForm.isAlive = true
    await loadMembers()
  } catch (requestError) {
    memberError.value = errorText(requestError, '新增成员失败。')
  } finally {
    memberSubmitting.value = false
  }
}

async function submitRelationship() {
  relationshipError.value = ''
  relationshipSuccess.value = ''
  if (!relationshipForm.baseMemberId) {
    relationshipError.value = '请选择基础成员。'
    return
  }
  if (!relationshipForm.name) {
    relationshipError.value = '新亲属姓名不能为空。'
    return
  }

  relationshipSubmitting.value = true
  try {
    const result = await createRelationship(familyId.value, {
      baseMemberId: relationshipForm.baseMemberId,
      addType: relationshipForm.addType,
      newMember: {
        name: relationshipForm.name,
        gender: relationshipForm.gender,
        isAlive: true
      },
      relationship: relationshipForm.addType === 'ADD_SPOUSE'
        ? { relationshipType: 'SPOUSE' }
        : {
            relationshipType: 'PARENT_CHILD',
            parentLinkType: relationshipForm.parentLinkType
          }
    })
    relationshipSuccess.value = `创建成功，当前 graphVersion：${result.graphVersion}`
    relationshipForm.name = ''
    relationshipForm.gender = 'UNKNOWN'
    await loadMembers()
  } catch (requestError) {
    relationshipError.value = errorText(requestError, '创建亲属关系失败。')
  } finally {
    relationshipSubmitting.value = false
  }
}

onMounted(loadPage)
</script>

<style scoped>
.page-section {
  padding: 32px 0;
}

.page-heading,
.section-heading,
.heading-actions,
.member-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.page-heading {
  align-items: flex-end;
  margin-bottom: 18px;
}

h1,
h2,
p {
  margin-top: 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.forms-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.form-card {
  display: grid;
  align-content: start;
  gap: 15px;
  padding: 20px;
}

.compact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

label {
  display: grid;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
}

.check-row {
  display: flex;
  align-items: center;
}

.notice,
.readonly-note {
  border-left: 3px solid var(--color-warm-gold);
  background: #fbf7ed;
  padding: 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

.readonly-note {
  margin-bottom: 18px;
}

.feedback {
  margin: 0;
  font-weight: 600;
}

.error {
  color: var(--color-danger);
}

.success {
  color: var(--color-success);
}

.state-panel {
  display: grid;
  min-height: 170px;
  place-items: center;
  gap: 12px;
  padding: 24px;
  text-align: center;
}

.state-panel.error {
  color: var(--color-danger);
}

.member-section {
  margin-top: 24px;
}

.member-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.member-card {
  padding: 18px;
}

.status-tag {
  border-radius: 6px;
  background: #edf4ef;
  padding: 4px 8px;
  color: var(--color-success);
  font-size: 12px;
  font-weight: 700;
}

dl {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 0;
}

dt {
  color: var(--color-text-secondary);
  font-size: 12px;
}

dd {
  margin: 4px 0 0;
  font-weight: 600;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 760px) {
  .page-heading,
  .heading-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .forms-grid,
  .member-list,
  .compact-grid,
  dl {
    grid-template-columns: 1fr;
  }
}
</style>
