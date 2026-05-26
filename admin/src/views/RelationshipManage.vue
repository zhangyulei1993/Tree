<template>
  <div class="card">
    <div class="toolbar">
      <div>
        <h2 class="page-title">&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x7BA1;&#x7406;</h2>
        <select v-model.number="selectedFamilyId" class="select family-select" @change="reload">
          <option :value="0">&#x8BF7;&#x9009;&#x62E9;&#x5BB6;&#x65CF;</option>
          <option v-for="family in families" :key="family.id" :value="family.id">
            {{ family.name }}
          </option>
        </select>
      </div>

      <button class="btn" @click="openCreate">&#x65B0;&#x589E;&#x5173;&#x7CFB;</button>
    </div>

    <div class="hint card">
      <p><strong>&#x7236;&#x6BCD;&#x5B50;&#x5973;&#xFF1A;</strong>&#x524D;&#x4E00;&#x4F4D;&#x6210;&#x5458;&#x662F;&#x7236;&#x6BCD;&#xFF0C;&#x540E;&#x4E00;&#x4F4D;&#x6210;&#x5458;&#x662F;&#x5B50;&#x5973;&#x3002;</p>
      <p><strong>&#x914D;&#x5076;&#xFF1A;</strong>&#x4E24;&#x4F4D;&#x6210;&#x5458;&#x4E3A;&#x592B;&#x59BB;&#x6216;&#x914D;&#x5076;&#x5173;&#x7CFB;&#x3002;</p>
    </div>

    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>&#x6210;&#x5458; A</th>
            <th>&#x5173;&#x7CFB;</th>
            <th>&#x6210;&#x5458; B</th>
            <th>&#x5907;&#x6CE8;</th>
            <th>&#x64CD;&#x4F5C;</th>
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
              <button class="btn secondary small" @click="openEdit(item)">&#x7F16;&#x8F91;</button>
              <button class="btn danger small" @click="remove(item.id)">&#x5220;&#x9664;</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">&#x7F16;&#x8F91;&#x5173;&#x7CFB;</h3>
        <h3 v-else>&#x65B0;&#x589E;&#x5173;&#x7CFB;</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>&#x6240;&#x5C5E;&#x5BB6;&#x65CF;</label>
            <select v-model.number="form.family_id" class="select" @change="loadMembersForForm">
              <option :value="0">&#x8BF7;&#x9009;&#x62E9;&#x5BB6;&#x65CF;</option>
              <option v-for="family in families" :key="family.id" :value="family.id">
                {{ family.name }}
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x6210;&#x5458; A</label>
            <select v-model.number="form.from_member_id" class="select">
              <option :value="0">&#x8BF7;&#x9009;&#x62E9;&#x6210;&#x5458;</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - &#x7B2C; {{ member.generation }} &#x4EE3;
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x5173;&#x7CFB;&#x7C7B;&#x578B;</label>
            <select v-model="form.relation_type" class="select">
              <option value="parent_child">&#x7236;&#x6BCD;&#x5B50;&#x5973;</option>
              <option value="spouse">&#x914D;&#x5076;</option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x6210;&#x5458; B</label>
            <select v-model.number="form.to_member_id" class="select">
              <option :value="0">&#x8BF7;&#x9009;&#x62E9;&#x6210;&#x5458;</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - &#x7B2C; {{ member.generation }} &#x4EE3;
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>&#x5907;&#x6CE8;</label>
            <input v-model="form.remark" class="input" />
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
import { api, type Family, type FamilyMember, type FamilyRelationship } from '../api'

const families = ref<Family[]>([])
const members = ref<FamilyMember[]>([])
const relationships = ref<FamilyRelationship[]>([])
const selectedFamilyId = ref(0)
const showModal = ref(false)

function emptyForm(): FamilyRelationship {
  return {
    family_id: selectedFamilyId.value || 0,
    from_member_id: 0,
    to_member_id: 0,
    relation_type: 'parent_child',
    remark: ''
  }
}

const form = reactive<FamilyRelationship>(emptyForm())

onMounted(async () => {
  families.value = await api.familyList()

  if (families.value.length && families.value[0].id) {
    selectedFamilyId.value = families.value[0].id
  }

  await reload()
})

async function reload() {
  if (!selectedFamilyId.value) {
    members.value = []
    relationships.value = []
    return
  }

  members.value = await api.memberList(selectedFamilyId.value)
  relationships.value = await api.relationshipList(selectedFamilyId.value)
}

async function loadMembersForForm() {
  if (!form.family_id) {
    members.value = []
    return
  }

  members.value = await api.memberList(form.family_id)
}

function memberName(id: number) {
  return members.value.find((item) => item.id === id)?.name || '-'
}

function relationLabel(type: string) {
  if (type === 'parent_child') return '\u7236\u6BCD\u5B50\u5973'
  if (type === 'spouse') return '\u914D\u5076'
  return type
}

function openCreate() {
  Object.assign(form, emptyForm())
  showModal.value = true
}

async function openEdit(item: FamilyRelationship) {
  Object.assign(form, item)
  await loadMembersForForm()
  showModal.value = true
}

async function save() {
  if (!form.family_id) {
    alert('\u8BF7\u9009\u62E9\u5BB6\u65CF')
    return
  }

  if (!form.from_member_id || !form.to_member_id) {
    alert('\u8BF7\u9009\u62E9\u5173\u7CFB\u6210\u5458')
    return
  }

  if (form.from_member_id === form.to_member_id) {
    alert('\u4E24\u4F4D\u6210\u5458\u4E0D\u80FD\u662F\u540C\u4E00\u4EBA')
    return
  }

  if (form.id) {
    await api.relationshipUpdate(form)
  } else {
    await api.relationshipCreate(form)
  }

  selectedFamilyId.value = form.family_id
  showModal.value = false
  await reload()
}

async function remove(id?: number) {
  if (!id) return
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u6761\u4EB2\u5C5E\u5173\u7CFB\u5417\uFF1F')) return

  await api.relationshipDelete(id)
  await reload()
}
</script>

<style scoped>
.family-select {
  width: 260px;
}

.hint {
  margin-bottom: 18px;
  padding: 14px 18px;
}

.hint p {
  margin: 6px 0;
  color: #7c4a2a;
}
</style>