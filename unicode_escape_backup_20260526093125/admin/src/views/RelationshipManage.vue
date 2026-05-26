<template>
  <div class="card">
    <div class="toolbar">
      <div>
        <h2 class="page-title">亲属关系管理</h2>
        <select v-model.number="selectedFamilyId" class="select family-select" @change="reload">
          <option :value="0">请选择家族</option>
          <option v-for="family in families" :key="family.id" :value="family.id">
            {{ family.name }}
          </option>
        </select>
      </div>

      <button class="btn" @click="openCreate">新增关系</button>
    </div>

    <div class="hint card">
      <p><strong>父母子女：</strong>前一位成员是父母，后一位成员是子女。</p>
      <p><strong>配偶：</strong>两位成员为夫妻或配偶关系。</p>
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
              <button class="btn secondary small" @click="openEdit(item)">编辑</button>
              <button class="btn danger small" @click="remove(item.id)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal" class="modal-mask">
      <div class="modal">
        <h3 v-if="form.id">编辑关系</h3>
        <h3 v-else>新增关系</h3>

        <div class="form-grid">
          <div class="form-row">
            <label>所属家族</label>
            <select v-model.number="form.family_id" class="select" @change="loadMembersForForm">
              <option :value="0">请选择家族</option>
              <option v-for="family in families" :key="family.id" :value="family.id">
                {{ family.name }}
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>成员 A</label>
            <select v-model.number="form.from_member_id" class="select">
              <option :value="0">请选择成员</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - 第 {{ member.generation }} 代
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>关系类型</label>
            <select v-model="form.relation_type" class="select">
              <option value="parent_child">父母子女</option>
              <option value="spouse">配偶</option>
            </select>
          </div>

          <div class="form-row">
            <label>成员 B</label>
            <select v-model.number="form.to_member_id" class="select">
              <option :value="0">请选择成员</option>
              <option v-for="member in members" :key="member.id" :value="member.id">
                {{ member.name }} - 第 {{ member.generation }} 代
              </option>
            </select>
          </div>

          <div class="form-row">
            <label>备注</label>
            <input v-model="form.remark" class="input" />
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