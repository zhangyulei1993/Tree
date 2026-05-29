
<template>
  <div class="kinship-page ultra-compact-kinship">
    <section class="hero-card">
      <div>
        <p class="eyebrow">亲属关系确认工具</p>
        <h1>输入基础关系，自动推导中文亲属称谓</h1>
        <p>
          这里只保存最基础的关系：父母、子女、兄弟姐妹、配偶。
          系统会根据性别、出生日期、年龄和关系路径，自动判断爷爷、奶奶、哥哥、弟弟、姐姐、妹妹、嫂子、弟媳、堂哥、堂弟、表姐、表妹、大舅子、小舅子等称谓。
        </p>
      </div>

      <div class="result-card">
        <span>推导结果</span>
        <strong>{{ result.name }}</strong>
        <small>{{ result.matched ? '已匹配规则' : '需要补充性别或年龄信息' }}</small>
      </div>
    </section>

    <section class="panel top-process-panel compact-process-panel">
      <div class="panel-head">
        <div>
          <h2>推导过程</h2>
          <p>检查每一层基础关系被系统转换成了什么。系统只根据基础关系、性别、出生日期和年龄计算，不根据姓名计算。</p>
        </div>
        <button class="light-btn" @click="resetAll">清空</button>
      </div>

      <div class="summary-grid">
        <div>
          <span>基础路径</span>
          <strong>{{ basePathText }}</strong>
        </div>
        <div>
          <span>细分路径</span>
          <strong>{{ detailPathText }}</strong>
        </div>
        <div>
          <span>最终结果</span>
          <strong>{{ result.name }}</strong>
        </div>
      </div>

      <div class="debug-list">
        <div v-for="(item, index) in debugSteps" :key="index" class="debug-item">
          <span>{{ item.from }}</span>
          <b>{{ item.baseRelation }}</b>
          <span>{{ item.to }}</span>
          <em>转换为：{{ item.detailName }}</em>
        </div>
      </div>

      <p v-if="!result.matched" class="notice">
        提示：如果要精确判断哥哥、弟弟、姐姐、妹妹、堂哥、堂弟、表姐、表妹，请至少填写本人和对方的出生日期或年龄。
      </p>
    </section>

<section class="main-grid">
      <article class="panel self-panel">
        <h2>本人信息</h2>
        <p>从谁的角度计算亲属关系。</p>

        <div class="form-grid">
          <label>
            本人名称
            <input v-model="self.name" placeholder="我" />
          </label>

          <label>
            本人性别
            <select v-model="self.gender">
              <option value="male">男</option>
              <option value="female">女</option>
              <option value="unknown">未知</option>
            </select>
          </label>

          <label>
            本人出生日期
            <input v-model="self.birthday" type="date" />
          </label>

          <label>
            本人年龄
            <input v-model.number="self.age" type="number" placeholder="不知道生日时填写" />
          </label>
        </div>

        <div class="quick-box">
          <button @click="useExample('grandFather')">父亲的父亲</button>
          <button @click="useExample('grandMother')">父亲的母亲</button>
          <button @click="useExample('maternalGrandFather')">母亲的父亲</button>
          <button @click="useExample('youngerBrotherWife')">弟弟的妻子</button>
          <button @click="useExample('elderSisterHusband')">姐姐的丈夫</button>
          <button @click="useExample('wifeElderBrother')">妻子的哥哥</button>
          <button @click="useExample('wifeYoungerBrother')">妻子的弟弟</button>
          <button @click="useExample('tangGe')">堂哥</button>
          <button @click="useExample('tangDi')">堂弟</button>
          <button @click="useExample('biaoJie')">表姐</button>
          <button @click="useExample('biaoMei')">表妹</button>
        </div>
      </article>

      <article class="panel path-panel">
        <div class="panel-head">
          <div>
            <h2>关系路径</h2>
            <p>从本人开始，一层一层选择基础关系。对方姓名只用于展示，不参与称谓计算。</p>
          </div>
          <button class="primary-btn" @click="addRow">添加一层关系</button>
        </div>

        <div class="path-list">
          <div v-for="(row, index) in rows" :key="row.id" class="path-row">
            <div class="step-index">{{ index + 1 }}</div>

            <label>
              基础关系
              <select v-model="row.relation">
                <option value="parent">父母</option>
                <option value="child">子女</option>
                <option value="sibling">兄弟姐妹</option>
                <option value="spouse">配偶</option>
              </select>
            </label>

            <label>
              对方姓名
              <input v-model="row.targetName" placeholder="例如：父亲、哥哥、妻子" />
            </label>

            <label>
              对方性别
              <select v-model="row.targetGender">
                <option value="male">男</option>
                <option value="female">女</option>
                <option value="unknown">未知</option>
              </select>
            </label>

            <label>
              对方出生日期
              <input v-model="row.targetBirthday" type="date" />
            </label>

            <label>
              对方年龄
              <input v-model.number="row.targetAge" type="number" placeholder="可选" />
            </label>

            <button class="remove-btn" @click="removeRow(index)">删除</button>
          </div>

          <div v-if="!rows.length" class="empty">
            还没有关系路径，请点击“添加一层关系”。
          </div>
        </div>
      </article>
    </section>

    
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { api } from '../api'

type Gender = 'male' | 'female' | 'unknown'

interface Person {
  name: string
  gender: Gender | string
  birthday: string
  age: number | null
}

interface PathRow {
  id: number
  relation: string
  targetName: string
  targetGender: Gender | string
  targetBirthday: string
  targetAge: number | null
}

const self = reactive<Person>({
  name: '我',
  gender: 'male',
  birthday: '1990-01-01',
  age: null
})

const rows = ref<PathRow[]>([
  createRow('parent', '父亲', 'male', '1968-01-02'),
  createRow('sibling', '大伯', 'male', '1964-01-01'),
  createRow('child', '堂哥', 'male', '1988-01-01'),
  createRow('spouse', '堂嫂', 'female', '1989-01-01')
])

const result = reactive({
  name: '待计算',
  matched: false
})

const debugSteps = ref<Array<{
  from: string
  to: string
  baseRelation: string
  detailName: string
}>>([])

const basePath = ref<string[]>([])
const detailPath = ref<string[]>([])
const resolving = ref(false)

const basePathText = computed(() => {
  return basePath.value.length ? basePath.value.join(' 的 ') : '本人'
})

const detailPathText = computed(() => {
  return detailPath.value.length ? detailPath.value.join(' 的 ') : '本人'
})

function createRow(
  relation = 'parent',
  targetName = '',
  targetGender: Gender | string = 'unknown',
  targetBirthday = '',
  targetAge: number | null = null
): PathRow {
  return {
    id: Date.now() + Math.floor(Math.random() * 100000),
    relation,
    targetName,
    targetGender,
    targetBirthday,
    targetAge
  }
}

function addRow() {
  rows.value.push(createRow())
}

function removeRow(index: number) {
  rows.value.splice(index, 1)
}

function resetAll() {
  self.name = '我'
  self.gender = 'male'
  self.birthday = '1990-01-01'
  self.age = null
  rows.value = []
  result.name = '待计算'
  result.matched = false
  basePath.value = []
  detailPath.value = []
  debugSteps.value = []
}

function useExample(type: string) {
  self.name = '我'
  self.gender = 'male'
  self.birthday = '1990-01-01'
  self.age = null

  const examples: Record<string, PathRow[]> = {
    grandFather: [
      createRow('parent', '父亲', 'male', '1965-01-01'),
      createRow('parent', '爷爷', 'male', '1940-01-01')
    ],
    grandMother: [
      createRow('parent', '父亲', 'male', '1965-01-01'),
      createRow('parent', '奶奶', 'female', '1942-01-01')
    ],
    maternalGrandFather: [
      createRow('parent', '母亲', 'female', '1968-01-01'),
      createRow('parent', '外公', 'male', '1940-01-01')
    ],
    maternalGrandMother: [
      createRow('parent', '母亲', 'female', '1968-01-01'),
      createRow('parent', '外婆', 'female', '1942-01-01')
    ],
    youngerBrotherWife: [
      createRow('sibling', '弟弟', 'male', '1995-01-01'),
      createRow('spouse', '弟弟的妻子', 'female', '1996-01-01')
    ],
    elderSisterHusband: [
      createRow('sibling', '姐姐', 'female', '1986-01-01'),
      createRow('spouse', '姐姐的丈夫', 'male', '1985-01-01')
    ],
    wifeElderBrother: [
      createRow('spouse', '妻子', 'female', '1992-01-01'),
      createRow('sibling', '妻子的哥哥', 'male', '1988-01-01')
    ],
    wifeYoungerBrother: [
      createRow('spouse', '妻子', 'female', '1992-01-01'),
      createRow('sibling', '妻子的弟弟', 'male', '1998-01-01')
    ],
    tangGe: [
      createRow('parent', '父亲', 'male', '1965-01-01'),
      createRow('sibling', '伯父', 'male', '1960-01-01'),
      createRow('child', '堂哥', 'male', '1988-01-01')
    ],
    tangDi: [
      createRow('parent', '父亲', 'male', '1965-01-01'),
      createRow('sibling', '叔叔', 'male', '1970-01-01'),
      createRow('child', '堂弟', 'male', '1996-01-01')
    ],
    biaoJie: [
      createRow('parent', '母亲', 'female', '1968-01-01'),
      createRow('sibling', '姨妈', 'female', '1965-01-01'),
      createRow('child', '表姐', 'female', '1988-01-01')
    ],
    biaoMei: [
      createRow('parent', '父亲', 'male', '1965-01-01'),
      createRow('sibling', '姑姑', 'female', '1968-01-01'),
      createRow('child', '表妹', 'female', '1996-01-01')
    ]
  }

  rows.value = examples[type] || examples.grandFather
}

async function resolveRemote() {
  if (resolving.value) return

  resolving.value = true
  try {
    const payload = await api.kinshipResolve({
      self: {
        name: self.name || '我',
        gender: String(self.gender || 'unknown'),
        birthday: self.birthday || '',
        age: Number(self.age || 0)
      },
      rows: rows.value.map((row) => ({
        relation: row.relation,
        target_name: row.targetName || '',
        target_gender: String(row.targetGender || 'unknown'),
        target_birthday: row.targetBirthday || '',
        target_age: Number(row.targetAge || 0)
      }))
    })

    result.name = payload.name || '未匹配'
    result.matched = Boolean(payload.matched)
    basePath.value = payload.base_path || []
    detailPath.value = payload.debug_steps?.map((item) => item.detail_name) || []
    debugSteps.value = (payload.debug_steps || []).map((item) => ({
      from: item.from,
      to: item.to,
      baseRelation: item.base_relation,
      detailName: item.detail_name
    }))
  } catch (error) {
    result.name = '计算失败'
    result.matched = false
    debugSteps.value = []
    console.warn(error)
  } finally {
    resolving.value = false
  }
}

watch(
  () => ({
    self: { ...self },
    rows: rows.value.map((row) => ({ ...row }))
  }),
  () => {
    resolveRemote()
  },
  { deep: true, immediate: true }
)
</script>

<style scoped>
.kinship-page {
  min-height: 100vh;
  padding: 28px;
  color: #3b1608;
  background:
    radial-gradient(circle at 12% 0%, rgba(251, 191, 36, 0.14), transparent 28%),
    linear-gradient(180deg, #fff7ed 0%, #f8efe3 100%);
}

.hero-card,
.panel {
  background: rgba(255, 250, 243, 0.96);
  border: 1px solid #ead3b5;
  border-radius: 22px;
  box-shadow: 0 18px 46px rgba(100, 56, 20, 0.08);
}

.hero-card {
  display: grid;
  grid-template-columns: 1fr 260px;
  gap: 24px;
  padding: 28px;
  margin-bottom: 18px;
}

.eyebrow {
  margin: 0 0 8px;
  color: #a6171c;
  font-weight: 900;
}

.hero-card h1 {
  margin: 0;
  font-size: 34px;
  line-height: 1.25;
}

.hero-card p,
.panel p {
  color: #7c3f1d;
  line-height: 1.8;
}

.result-card {
  display: grid;
  align-content: center;
  gap: 8px;
  padding: 22px;
  border-radius: 18px;
  background: #a6171c;
  color: #fff7ed;
}

.result-card strong {
  font-size: 34px;
}

.main-grid {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 18px;
  margin-bottom: 18px;
}

.panel {
  padding: 22px;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
}

.panel h2 {
  margin: 0;
  font-size: 22px;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 14px;
}

label {
  display: grid;
  gap: 7px;
  color: #7c2d12;
  font-size: 13px;
  font-weight: 900;
}

input,
select {
  width: 100%;
  height: 40px;
  border: 1px solid #ead3b5;
  border-radius: 12px;
  background: #fffaf5;
  color: #3b1608;
  padding: 0 12px;
  outline: none;
}

input[type="date"] {
  font-family: inherit;
}

button {
  cursor: pointer;
  border: 0;
  font-weight: 900;
}

.primary-btn,
.light-btn {
  height: 40px;
  padding: 0 16px;
  border-radius: 12px;
}

.primary-btn {
  color: #fffaf0;
  background: #a6171c;
}

.light-btn {
  color: #7c2d12;
  background: #f3dec3;
}

.quick-box {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 18px;
}

.quick-box button {
  height: 32px;
  padding: 0 10px;
  border-radius: 999px;
  color: #7c2d12;
  background: #f3dec3;
}

.path-list {
  display: grid;
  gap: 12px;
}

.path-row {
  display: grid;
  grid-template-columns: 38px 110px minmax(160px, 1fr) 92px 150px 100px 64px;
  gap: 10px;
  align-items: end;
  padding: 14px;
  border: 1px solid #ead3b5;
  border-radius: 16px;
  background: #fffaf5;
}

.step-index {
  width: 32px;
  height: 32px;
  border-radius: 999px;
  display: grid;
  place-items: center;
  background: #a6171c;
  color: #fffaf0;
  font-weight: 900;
  margin-bottom: 4px;
}

.remove-btn {
  height: 40px;
  border-radius: 12px;
  color: #fff;
  background: #b91c1c;
}

.empty {
  padding: 28px;
  text-align: center;
  color: #8a5b3c;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.summary-grid div {
  padding: 16px;
  border-radius: 16px;
  background: #fff7ed;
  border: 1px solid #ead3b5;
}

.summary-grid span {
  display: block;
  color: #8a5b3c;
  font-size: 13px;
  margin-bottom: 8px;
}

.summary-grid strong {
  font-size: 18px;
}

.debug-list {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.debug-item {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 12px 14px;
  border-radius: 14px;
  background: #fffaf5;
  border: 1px dashed #ead3b5;
}

.debug-item b {
  color: #a6171c;
}

.debug-item em {
  margin-left: auto;
  font-style: normal;
  color: #7c3f1d;
  font-weight: 900;
}

.notice {
  margin-top: 16px;
  color: #a6171c;
  font-weight: 900;
}

@media (max-width: 1080px) {
  .hero-card,
  .main-grid {
    grid-template-columns: 1fr;
  }

  .path-row {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: 1fr;
  }

  .debug-item em {
    margin-left: 0;
    width: 100%;
  }
}

/* kinship process top compact */
.top-process-panel {
  margin: 0 0 18px;
  padding: 18px 22px;
}

.top-process-panel .panel-head {
  margin-bottom: 12px;
}

.top-process-panel .panel-head h2 {
  font-size: 20px;
}

.top-process-panel .panel-head p {
  margin-top: 4px;
}

.top-process-panel .summary-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.top-process-panel .summary-grid div {
  padding: 12px 14px;
}

.top-process-panel .summary-grid span {
  margin-bottom: 5px;
}

.top-process-panel .summary-grid strong {
  font-size: 17px;
  line-height: 1.45;
}

.top-process-panel .debug-list {
  margin-top: 12px;
  display: grid;
  gap: 8px;
  max-height: 160px;
  overflow: auto;
  padding-right: 4px;
}

.top-process-panel .debug-item {
  padding: 8px 12px;
  font-size: 14px;
}

.top-process-panel .notice {
  margin-top: 10px;
}

@media (max-width: 1080px) {
  .top-process-panel .summary-grid {
    grid-template-columns: 1fr;
  }

  .top-process-panel .debug-list {
    max-height: none;
  }
}






</style>


<style>

</style>


<style>
/* ultra compact kinship start */

.ultra-compact-kinship {
  padding: 8px 12px !important;
  font-size: 12px !important;
  line-height: 1.35 !important;
}

/* 顶部区域整体压扁 */
.ultra-compact-kinship .hero-card {
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) 156px !important;
  gap: 10px !important;
  padding: 12px 16px !important;
  margin-bottom: 8px !important;
  border-radius: 14px !important;
  min-height: 0 !important;
}

.ultra-compact-kinship .hero-card h1 {
  font-size: 22px !important;
  line-height: 1.15 !important;
  margin: 2px 0 6px !important;
  letter-spacing: 0 !important;
}

.ultra-compact-kinship .hero-card p {
  font-size: 12px !important;
  line-height: 1.45 !important;
  margin: 0 !important;
}

.ultra-compact-kinship .eyebrow {
  font-size: 12px !important;
  margin: 0 0 4px !important;
}

/* 右侧结果卡片缩小 */
.ultra-compact-kinship .result-card {
  width: 156px !important;
  min-height: 84px !important;
  padding: 12px 14px !important;
  border-radius: 12px !important;
  align-self: center !important;
}

.ultra-compact-kinship .result-card span {
  font-size: 12px !important;
}

.ultra-compact-kinship .result-card strong {
  font-size: 26px !important;
  line-height: 1.1 !important;
  margin: 6px 0 !important;
}

.ultra-compact-kinship .result-card small {
  font-size: 11px !important;
}

/* 通用卡片 */
.ultra-compact-kinship .panel {
  padding: 12px 16px !important;
  margin-bottom: 8px !important;
  border-radius: 14px !important;
}

.ultra-compact-kinship .panel-head {
  margin-bottom: 8px !important;
}

.ultra-compact-kinship .panel h2,
.ultra-compact-kinship h2 {
  font-size: 18px !important;
  line-height: 1.2 !important;
  margin: 0 0 4px !important;
}

.ultra-compact-kinship .panel p,
.ultra-compact-kinship p {
  font-size: 12px !important;
  line-height: 1.45 !important;
  margin: 0 0 6px !important;
}

/* 推导过程压缩 */
.ultra-compact-kinship .top-process-panel {
  padding: 12px 16px !important;
  margin-bottom: 8px !important;
}

.ultra-compact-kinship .top-process-panel .panel-head {
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  gap: 10px !important;
}

.ultra-compact-kinship .summary-grid {
  display: grid !important;
  grid-template-columns: repeat(3, minmax(0, 1fr)) !important;
  gap: 8px !important;
}

.ultra-compact-kinship .summary-grid > div {
  padding: 8px 10px !important;
  border-radius: 10px !important;
}

.ultra-compact-kinship .summary-grid span {
  font-size: 11px !important;
  margin-bottom: 2px !important;
}

.ultra-compact-kinship .summary-grid strong {
  font-size: 15px !important;
  line-height: 1.25 !important;
}

.ultra-compact-kinship .debug-list {
  max-height: 86px !important;
  overflow: auto !important;
  margin-top: 8px !important;
  gap: 4px !important;
}

.ultra-compact-kinship .debug-item {
  min-height: 0 !important;
  padding: 5px 8px !important;
  border-radius: 9px !important;
  font-size: 12px !important;
  line-height: 1.3 !important;
}

.ultra-compact-kinship .notice {
  font-size: 12px !important;
  margin-top: 6px !important;
}

/* 本人信息 + 关系路径 */
.ultra-compact-kinship .main-grid {
  display: grid !important;
  grid-template-columns: 280px minmax(0, 1fr) !important;
  gap: 10px !important;
  margin-top: 0 !important;
}

.ultra-compact-kinship .self-panel {
  align-self: start !important;
}

.ultra-compact-kinship .form-grid {
  gap: 7px !important;
}

/* 表单控件全部变矮 */
.ultra-compact-kinship label {
  font-size: 11px !important;
  gap: 3px !important;
}

.ultra-compact-kinship input,
.ultra-compact-kinship select {
  height: 28px !important;
  min-height: 28px !important;
  padding: 0 8px !important;
  font-size: 12px !important;
  border-radius: 8px !important;
}

.ultra-compact-kinship button {
  font-size: 12px !important;
}

.ultra-compact-kinship .primary-btn,
.ultra-compact-kinship .light-btn,
.ultra-compact-kinship .remove-btn {
  height: 28px !important;
  min-height: 28px !important;
  padding: 0 10px !important;
  border-radius: 8px !important;
}

/* 关系路径每行进一步压缩 */
.ultra-compact-kinship .path-list {
  gap: 6px !important;
}

.ultra-compact-kinship .path-row {
  grid-template-columns: 28px 100px minmax(160px, 1fr) 82px 128px 86px 52px !important;
  gap: 6px !important;
  padding: 7px 8px !important;
  border-radius: 10px !important;
}

.ultra-compact-kinship .step-index {
  width: 24px !important;
  height: 24px !important;
  font-size: 12px !important;
}

/* 快捷按钮 */
.ultra-compact-kinship .quick-box {
  gap: 4px !important;
  margin-top: 8px !important;
}

.ultra-compact-kinship .quick-box button {
  height: 24px !important;
  min-height: 24px !important;
  padding: 0 7px !important;
  font-size: 11px !important;
}

/* 去掉多余空白 */
.ultra-compact-kinship * {
  box-sizing: border-box !important;
}

.ultra-compact-kinship h1,
.ultra-compact-kinship h2,
.ultra-compact-kinship h3,
.ultra-compact-kinship p {
  margin-top: 0 !important;
}

/* 小屏适配 */
@media (max-width: 1080px) {
  .ultra-compact-kinship {
    padding: 8px !important;
  }

  .ultra-compact-kinship .hero-card {
    grid-template-columns: 1fr !important;
  }

  .ultra-compact-kinship .result-card {
    width: 100% !important;
  }

  .ultra-compact-kinship .main-grid {
    grid-template-columns: 1fr !important;
  }

  .ultra-compact-kinship .summary-grid {
    grid-template-columns: 1fr !important;
  }

  .ultra-compact-kinship .path-row {
    grid-template-columns: 1fr !important;
  }
}

/* ultra compact kinship end */
</style>

