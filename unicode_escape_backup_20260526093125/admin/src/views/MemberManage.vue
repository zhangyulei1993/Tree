<template>
  <div class="card">
    <div class="toolbar">
      <div>
        <h2 class="page-title">成员管理</h2>
        <select v-model.number="selectedFamilyId" class="select family-select" @change="loadMembers">
          <option :value="0">全部家族</option>
          <option v-for="family in families" :key="family.id" :value="family.id">
            {{ family.name }}
          </option>
        </select>
      </div>

      <button class="btn" @click="openCreate">新增成员</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>头像</th>
            <th>姓名</th>
            <th>所属家族</th>
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
            <td>{{ familyName(item.family_id) }}</td>
            <td>{{ genderText(item.gender) }}</td>
            <td>第 {{ item.generation }} 代</td>
            <td>{{ item.birthday }}</td>
            <td>{{ statusText(item.status) }}</td>
            <td>
              <button class="btn secondary small" @click="openEdit(item)">编辑</button>
              <button class="btn danger small" @click="remove(item.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">编辑成员</h3>
        <h3 v-else>新增成员</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>所属家族</label>
            <select v-model.number="form.family_id" class="select">
              <option :value="0">请选择家族</option>
              <option v-for="family in families" :key="family.id" :value="family.id">
                {{ family.name }}
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>姓名</label>
            <input v-model="form.name" class="input" />
          </div>

          <div class="form-row">
            <label>性别</label>
            <select v-model="form.gender" class="select">
              <option value="male">男</option>
              <option value="female">女</option>
              <option value="unknown">未知</option>
            </select>
          </div>

          <div class="form-row">
            <label>头像</label>
            <input type="file" accept="image/*" @change="uploadAvatar" />
            <img v-if="form.avatar" :src="form.avatar" class="preview" />
          </div>

          <div class="form-row">
            <label>生日</label>
            <input v-model="form.birthday" class="input" placeholder="1958-01-01" />
          </div>

          <div class="form-row">
            <label>出生地</label>
            <input v-model="form.birth_place" class="input" />
          </div>

          <div class="form-row">
            <label>代际</label>
            <input v-model.number="form.generation" type="number" class="input" />
          </div>

          <div class="form-row">
            <label>排序</label>
            <input v-model.number="form.sort" type="number" class="input" />
          </div>

          <div class="form-row">
            <label>人物简介</label>
            <textarea v-model="form.biography" class="textarea"></textarea>
          </div>

          <div class="form-row">
            <label>状态</label>
            <select v-model.number="form.status" class="select">
              <option :value="1">显示</option>
              <option :value="0">隐藏</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">取消</button>
          <button class="btn" @click="save">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Family, type FamilyMember } from '../api'

const families = ref<Family[]>([])
const members = ref<FamilyMember[]>([])
const selectedFamilyId = ref(0)
const showModal = ref(false)

function emptyForm(): FamilyMember {
  return {
    family_id: selectedFamilyId.value || 0,
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

const form = reactive<FamilyMember>(emptyForm())

onMounted(async () => {
  families.value = await api.familyList()

  if (families.value.length && families.value[0].id) {
    selectedFamilyId.value = families.value[0].id
  }

  await loadMembers()
})

async function loadMembers() {
  members.value = await api.memberList(selectedFamilyId.value || undefined)
}

function familyName(id: number) {
  return families.value.find((item) => item.id === id)?.name || '-'
}

function genderText(gender: string) {
  if (gender === 'male') return '\u7537'
  if (gender === 'female') return '\u5973'
  return '\u672A\u77E5'
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

function openEdit(item: FamilyMember) {
  Object.assign(form, item)
  showModal.value = true
}

async function uploadAvatar(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const res = await api.upload(file)
  form.avatar = res.url
}

async function save() {
  if (!form.family_id) {
    alert('\u8BF7\u9009\u62E9\u5BB6\u65CF')
    return
  }

  if (!form.name) {
    alert('\u8BF7\u8F93\u5165\u59D3\u540D')
    return
  }

  if (form.id) {
    await api.memberUpdate(form)
  } else {
    await api.memberCreate(form)
  }

  showModal.value = false
  await loadMembers()
}

async function remove(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u4F4D\u6210\u5458\u5417\uFF1F\u76F8\u5173\u5173\u7CFB\u4E5F\u4F1A\u4E00\u8D77\u5220\u9664\u3002')) return

  await api.memberDelete(id)
  await loadMembers()
}
</script>

<style scoped>
.family-select {
  width: 260px;
}

.avatar {
  width: 58px;
  height: 58px;
  border-radius: 50%;
}
</style>