<template>
  <div class="kinship-page">
    <section class="top-card">
      <div>
        <p class="eyebrow">亲属关系确认工具</p>
        <h1>按基础关系推导中文亲属称谓</h1>
        <p>只根据基础关系、性别、出生日期或年龄计算；姓名只用于展示，不参与推导。</p>
      </div>

      <div class="result-box" :class="{ pending: !result.matched }">
        <span>推导结果</span>
        <strong>{{ result.title }}</strong>
        <small>{{ result.matched ? (result.terminal ? '已匹配终点称谓' : '已匹配称谓规则') : result.reason }}</small>
      </div>
    </section>

    <section class="panel compact">
      <div class="panel-head">
        <div>
          <h2>推导过程</h2>
          <p>当前路径：{{ result.displayPath }}</p>
        </div>

        <button type="button" @click="resetAll">清空</button>
      </div>

      <div class="process">
        <div v-for="(item, index) in processList" :key="index">
          <b>{{ index + 1 }}</b>
          <span>{{ item }}</span>
        </div>
      </div>

      <p v-if="result.reason" class="tip">{{ result.reason }}</p>
    </section>

    <section class="panel self-card">
      <h2>本人信息</h2>

      <div class="form-grid self-grid">
        <label>
          名称
          <input v-model="self.name" placeholder="我" />
        </label>

        <label>
          性别
          <select v-model="self.gender" @change="syncFollowingSpouseGender(-1)">
            <option value="male">男</option>
            <option value="female">女</option>
            <option value="unknown">未知</option>
          </select>
        </label>

        <label>
          出生日期
          <input v-model="self.birthday" type="date" />
        </label>

        <label>
          年龄
          <input v-model="self.age" placeholder="不知道生日时填写" />
        </label>
      </div>

      <div class="quick">
        <button type="button" @click="loadExample('wifeBrotherWife')">内兄嫂/内弟媳</button>
        <button type="button" @click="loadExample('wifeSisterHusband')">连襟</button>
        <button type="button" @click="loadExample('husbandBrotherWife')">妯娌</button>
        <button type="button" @click="loadExample('husbandSisterHusband')">姑姐夫/姑妹夫</button>
        <button type="button" @click="loadExample('wifeFather')">岳父</button>
        <button type="button" @click="loadExample('fatherFather')">爷爷</button>
        <button type="button" @click="loadExample('cousin')">堂哥</button>
      </div>
    </section>

    <section class="panel">
      <div class="panel-head">
        <div>
          <h2>关系路径</h2>
          <p>最多 5 层。终点称谓或超出状态机边界后，继续添加会显示原因。</p>
        </div>

        <div class="add-area">
          <button
            type="button"
            :class="{ disabled: addDisabled }"
            :aria-disabled="addDisabled"
            @click="handleAddRow"
          >
            {{ addButtonText }}
          </button>
          <small v-if="addDisabled">{{ addTip }}</small>
        </div>
      </div>

      <div class="rows">
        <div v-if="!rows.length" class="empty">尚未添加关系。点击“添加一层关系”开始推导。</div>

        <div v-for="(row, index) in rows" :key="row.id" class="row-card">
          <div class="index">{{ index + 1 }}</div>

          <label>
            基础关系
            <select v-model="row.relation" @change="onRelationChange(index)">
              <option
                v-for="item in relationOptions(index)"
                :key="item.value"
                :value="item.value"
                :disabled="item.disabled"
                :title="item.reason"
              >
                {{ item.label }}{{ item.disabled ? '（不可选）' : '' }}
              </option>
            </select>
          </label>

          <label>
            姓名
            <input v-model="row.name" placeholder="只展示，不参与计算" />
          </label>

          <label>
            性别
            <select v-model="row.gender" @change="syncFollowingSpouseGender(index)">
              <option value="unknown">未知</option>
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </label>

          <label>
            出生日期
            <input v-model="row.birthday" type="date" />
          </label>

          <label>
            年龄
            <input v-model="row.age" placeholder="可选" />
          </label>

          <button type="button" class="delete" @click="removeRow(index)">删除</button>

          <p class="row-tip">{{ rowTip(index) }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import {
  allowedRelations,
  genderLabel,
  RELATION_LABELS,
  relationOptionsFor,
  resolveKinship,
  type BaseRelation,
  type Gender,
  type KinshipStep,
  type RelationOption
} from '../utils/kinship'

interface Row extends KinshipStep {
  id: number
}

let uid = 1

const self = reactive({
  name: '我',
  gender: 'male' as Gender,
  birthday: '1990-01-01',
  age: ''
})

const rows = reactive<Row[]>([])

const context = computed(() => ({
  self,
  steps: rows
}))

const result = computed(() => resolveKinship(context.value))

const processList = computed(() => {
  if (!rows.length) return result.value.process

  const perStep = rows.map((row, index) => {
    const prefix = rows.slice(0, index + 1)
    const current = resolveKinship({ self, steps: prefix })
    return `${RELATION_LABELS[row.relation]} ｜ ${genderLabel(row.gender)} ｜ 转换为：${current.title}`
  })

  return [...result.value.process, ...perStep]
})

const nextRelations = computed(() => allowedRelations(context.value))
const addDisabled = computed(() => nextRelations.value.length === 0)

const addButtonText = computed(() => {
  if (rows.length >= 5) return '最多 5 层关系'
  if (result.value.terminal) return '关系已明确'
  if (!nextRelations.value.length) return '当前路径到此为止'
  return '添加一层关系'
})

const addTip = computed(() => {
  if (rows.length >= 5) return '最多只能添加 5 层关系。'
  if (result.value.terminal) return result.value.reason || '当前关系已经是终点称谓。'
  return '当前路径在基础关系模式下不建议继续扩展。'
})

function previousGenderOf(index: number): Gender {
  if (index === 0) return self.gender
  return rows[index - 1]?.gender || 'unknown'
}

function oppositeGender(gender: Gender): Gender {
  if (gender === 'male') return 'female'
  if (gender === 'female') return 'male'
  return 'unknown'
}

function applySpouseDefault(index: number) {
  const row = rows[index]
  if (!row || row.relation !== 'spouse') return

  const gender = oppositeGender(previousGenderOf(index))
  if (gender !== 'unknown') row.gender = gender
}

function syncFollowingSpouseGender(prevIndex: number) {
  applySpouseDefault(prevIndex + 1)
}

function createRow(relation: BaseRelation, gender: Gender = 'unknown', age = '', birthday = ''): Row {
  return {
    id: uid++,
    relation,
    name: '',
    gender,
    birthday,
    age
  }
}

function relationOptions(index: number): RelationOption[] {
  return relationOptionsFor({ self, steps: rows.slice(0, index) })
}

function firstAllowedFor(index: number): BaseRelation | null {
  const allowed = allowedRelations({ self, steps: rows.slice(0, index) })
  return allowed[0] || null
}

function handleAddRow() {
  if (addDisabled.value) {
    window.alert(addTip.value)
    return
  }

  const relation = nextRelations.value[0]
  if (!relation) return

  const row = createRow(relation)
  rows.push(row)
  applySpouseDefault(rows.length - 1)
}

function removeRow(index: number) {
  rows.splice(index, 1)
}

function onRelationChange(index: number) {
  rows.splice(index + 1)

  const row = rows[index]
  const allowed = allowedRelations({ self, steps: rows.slice(0, index) })

  if (!allowed.includes(row.relation)) {
    const fallback = firstAllowedFor(index)
    if (fallback) row.relation = fallback
  }

  applySpouseDefault(index)
}

function rowTip(index: number) {
  const options = relationOptions(index)
  const available = options.filter((item) => !item.disabled)

  if (!available.length) {
    const reason = options.find((item) => item.reason)?.reason
    return reason || '当前层级不可继续选择。'
  }

  return '当前可选：' + available.map((item) => item.label).join('、')
}

function resetAll() {
  rows.splice(0, rows.length)
}

function loadExample(type: string) {
  rows.splice(0, rows.length)

  if (type === 'wifeBrotherWife') {
    self.gender = 'male'
    rows.push(createRow('spouse', 'female', '28'), createRow('sibling', 'male', '35'), createRow('spouse', 'female', '33'))
  }

  if (type === 'wifeSisterHusband') {
    self.gender = 'male'
    rows.push(createRow('spouse', 'female', '28'), createRow('sibling', 'female', '24'), createRow('spouse', 'male', '26'))
  }

  if (type === 'husbandBrotherWife') {
    self.gender = 'female'
    rows.push(createRow('spouse', 'male', '32'), createRow('sibling', 'male', '29'), createRow('spouse', 'female', '28'))
  }

  if (type === 'husbandSisterHusband') {
    self.gender = 'female'
    rows.push(createRow('spouse', 'male', '32'), createRow('sibling', 'female', '25'), createRow('spouse', 'male', '26'))
  }

  if (type === 'wifeFather') {
    self.gender = 'male'
    rows.push(createRow('spouse', 'female', '28'), createRow('parent', 'male', '60'))
  }

  if (type === 'fatherFather') {
    self.gender = 'male'
    rows.push(createRow('parent', 'male', '60'), createRow('parent', 'male', '85'))
  }

  if (type === 'cousin') {
    self.gender = 'male'
    rows.push(createRow('parent', 'male', '60'), createRow('sibling', 'male', '66'), createRow('child', 'male', '35'))
  }
}
</script>

<style scoped>
.kinship-page {
  min-height: 100vh;
  padding: 16px;
  background: #f9f1e6;
  color: #3d1707;
  font-size: 14px;
}

.top-card,
.panel {
  border: 1px solid #ead1ad;
  border-radius: 16px;
  background: rgba(255, 250, 242, 0.96);
  box-shadow: 0 10px 30px rgba(96, 50, 16, 0.08);
}

.top-card {
  display: grid;
  grid-template-columns: 1fr 260px;
  gap: 20px;
  padding: 22px;
  align-items: center;
}

.eyebrow {
  margin: 0 0 6px;
  color: #9b3a16;
  font-weight: 800;
}

h1,
h2,
p {
  margin-top: 0;
}

h1 {
  margin-bottom: 8px;
  font-size: 28px;
}

h2 {
  margin-bottom: 8px;
  font-size: 19px;
}

p {
  color: #8a3a16;
}

.result-box {
  display: grid;
  gap: 8px;
  padding: 18px;
  border-radius: 14px;
  color: #fff;
  background: #b4141d;
}

.result-box.pending {
  background: #8c3a1f;
}

.result-box strong {
  font-size: 28px;
}

.panel {
  margin-top: 14px;
  padding: 16px;
}

.compact {
  padding-bottom: 12px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
}

.process {
  display: grid;
  gap: 6px;
  max-height: 150px;
  overflow: auto;
}

.process > div {
  display: flex;
  gap: 8px;
  padding: 7px 10px;
  border-radius: 10px;
  background: #fff4e5;
}

.process b,
.index {
  display: inline-grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  color: #fff;
  background: #b4141d;
  flex: 0 0 auto;
}

.form-grid {
  display: grid;
  gap: 10px;
}

.self-grid {
  grid-template-columns: repeat(4, 1fr);
}

label {
  display: grid;
  gap: 5px;
  color: #7b2b0d;
  font-weight: 800;
}

input,
select {
  height: 34px;
  padding: 0 10px;
  border: 1px solid #efc99e;
  border-radius: 9px;
  background: #fffaf3;
  color: #3d1707;
}

button {
  height: 34px;
  padding: 0 14px;
  border: 0;
  border-radius: 10px;
  color: #fff;
  background: #b4141d;
  cursor: pointer;
  font-weight: 800;
}

button.disabled,
button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.quick {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.quick button {
  color: #85310f;
  background: #f3d9b8;
}

.add-area {
  display: grid;
  gap: 4px;
  justify-items: end;
}

.add-area small,
.tip,
.row-tip {
  color: #9b3a16;
  font-weight: 700;
}

.rows {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.empty {
  padding: 14px;
  border: 1px dashed #edcca5;
  border-radius: 12px;
  color: #9b3a16;
  background: #fff4e5;
  font-weight: 700;
}

.row-card {
  display: grid;
  grid-template-columns: 34px 130px minmax(140px, 1fr) 110px 150px 100px 70px;
  gap: 10px;
  align-items: end;
  padding: 10px;
  border: 1px solid #edcca5;
  border-radius: 12px;
  background: #fffaf3;
}

.delete {
  background: #c21b20;
}

.row-tip {
  grid-column: 2 / -1;
  margin: 0;
  font-size: 12px;
}

@media (max-width: 900px) {
  .top-card,
  .self-grid {
    grid-template-columns: 1fr;
  }

  .row-card {
    grid-template-columns: 30px 1fr;
  }

  .row-card label,
  .delete,
  .row-tip {
    grid-column: 2 / -1;
  }
}
</style>
