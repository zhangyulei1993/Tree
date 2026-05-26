<template>
  <div class="card">
    <div class="toolbar">
      <div>
        <h2 class="page-title">&#x6210;&#x5458;&#x7BA1;&#x7406;</h2>
        <select v-model.number="selectedFamilyId" class="select family-select" @change="loadMembers">
          <option :value="0">&#x5168;&#x90E8;&#x5BB6;&#x65CF;</option>
          <option v-for="family in families" :key="family.id" :value="family.id">
            {{ family.name }}
          </option>
        </select>
      </div>

      <button class="btn" @click="openCreate">&#x65B0;&#x589E;&#x6210;&#x5458;</button>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x5934;&#x50CF;</th>
            <th>&#x59D3;&#x540D;</th>
            <th>&#x6240;&#x5C5E;&#x5BB6;&#x65CF;</th>
            <th>&#x6027;&#x522B;</th>
            <th>&#x4EE3;&#x9645;</th>
            <th>&#x751F;&#x65E5;</th>
            <th>&#x72B6;&#x6001;</th>
            <th>&#x64CD;&#x4F5C;</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="item in members" :key="item.id">
            <td>{{ item.id }}</td>
            <td><img v-if="item.avatar" :src="item.avatar" class="cover avatar" /></td>
            <td>{{ item.name }}</td>
            <td>{{ familyName(item.family_id) }}</td>
            <td>{{ genderText(item.gender) }}</td>
            <td>&#x7B2C; {{ item.generation }} &#x4EE3;</td>
            <td>{{ item.birthday }}</td>
            <td>{{ statusText(item.status) }}</td>
            <td>
              <button class="btn secondary small" @click="openEdit(item)">&#x7F16;&#x8F91;</button>
              <button class="btn danger small" @click="remove(item.id)">&#x5220;&#x9664;</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">&#x7F16;&#x8F91;&#x6210;&#x5458;</h3>
        <h3 v-else>&#x65B0;&#x589E;&#x6210;&#x5458;</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>&#x6240;&#x5C5E;&#x5BB6;&#x65CF;</label>
            <select v-model.number="form.family_id" class="select">
              <option :value="0">&#x8BF7;&#x9009;&#x62E9;&#x5BB6;&#x65CF;</option>
              <option v-for="family in families" :key="family.id" :value="family.id">
                {{ family.name }}
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x59D3;&#x540D;</label>
            <input v-model="form.name" class="input" />
          </div>

          <div class="form-row">
            <label>&#x6027;&#x522B;</label>
            <select v-model="form.gender" class="select">
              <option value="male">&#x7537;</option>
              <option value="female">&#x5973;</option>
              <option value="unknown">&#x672A;&#x77E5;</option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x5934;&#x50CF;</label>
            <input type="file" accept="image/*" @change="uploadAvatar" />
            <img v-if="form.avatar" :src="form.avatar" class="preview" />
          </div>

          <div class="form-row">
            <label>&#x751F;&#x65E5;</label>
            <input v-model="form.birthday" class="input" placeholder="1958-01-01" />
          </div>

          <div class="form-row">
            <label>&#x51FA;&#x751F;&#x5730;</label>
            <input v-model="form.birth_place" class="input" />
          </div>

          <div class="form-row">
            <label>&#x4EE3;&#x9645;</label>
            <input v-model.number="form.generation" type="number" class="input" />
          </div>

          <div class="form-row">
            <label>&#x6392;&#x5E8F;</label>
            <input v-model.number="form.sort" type="number" class="input" />
          </div>

          <div class="form-row">
            <label>&#x4EBA;&#x7269;&#x7B80;&#x4ECB;</label>
            <textarea v-model="form.biography" class="textarea"></textarea>
          </div>

          <div class="form-row">
            <label>&#x72B6;&#x6001;</label>
            <select v-model.number="form.status" class="select">
              <option :value="1">&#x663E;&#x793A;</option>
              <option :value="0">&#x9690;&#x85CF;</option>
            </select>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn secondary" @click="showModal = false">&#x53D6;&#x6D88;</button>
          <button class="btn" @click="save">&#x4FDD;&#x5B58;</button>
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