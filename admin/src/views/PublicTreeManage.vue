<template>
  <div class="public-tree-page compact-public-tree-admin public-tree-tab-page">

    <aside class="public-tree-tab-sidebar">
      <div class="public-tree-tab-title">
        <strong>关系树管理</strong>
        <span>公开展示数据</span>
      </div>

      
      <button
        type="button"
        class="public-tree-tab-btn"
        :class="{ active: publicTreeActiveTab === 'config' }"
        @click="publicTreeActiveTab = 'config'"
      >
        <strong>基础配置</strong>
        <span>官网关系树标题、说明</span>
      </button>

      <button
        type="button"
        class="public-tree-tab-btn"
        :class="{ active: publicTreeActiveTab === 'nodes' }"
        @click="publicTreeActiveTab = 'nodes'"
      >
        <strong>人物节点</strong>
        <span>添加和维护人物</span>
      </button>
    </aside>

    <main class="public-tree-tab-content">

      <section
        v-if="publicTreeActiveTab === 'config'"
        class="public-tree-tab-panel"
      >
        <section class="card">
              <div class="toolbar">
                <div>
                  <h2 class="page-title">公开树展示</h2>
                  <p class="page-desc">维护官网展示用的公开关系树，不会直接修改真实家族数据。</p>
                </div>
                <button class="btn" @click="saveConfig">保存配置</button>
              </div>
        
              <div class="form-grid">
                <div class="form-row">
                  <label>标题</label>
                  <input v-model="config.title" class="input" />
                </div>
                <div class="form-row">
                  <label>副标题</label>
                  <input v-model="config.subtitle" class="input" />
                </div>
                <div class="form-row">
                  <label>描述</label>
                  <textarea v-model="config.description" class="textarea"></textarea>
                </div>
                <div class="form-row">
                  <label>状态</label>
                  <select v-model.number="config.status" class="select">
                    <option :value="1">显示</option>
                    <option :value="0">隐藏</option>
                  </select>
                </div>
              </div>
            </section>

        <section class="card">
              <div class="section-toolbar">
                <div>
                  <h3>关系类型字典</h3>
                  <p class="page-desc">这里维护“父子、母子、配偶、姻亲”等关系定义，公开树关系可以重复使用。</p>
                </div>
                <button class="btn small" @click="openTypeCreate">新增类型</button>
              </div>
        
              <div class="table-wrap">
                <table class="table">
                  <thead>
                    <tr>
                      <th>编码</th>
                      <th>名称</th>
                      <th>分类</th>
                      <th>说明</th>
                      <th>排序</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in relationshipTypes" :key="item.id">
                      <td>{{ item.code }}</td>
                      <td>{{ item.name }}</td>
                      <td>{{ categoryText(item.category) }}</td>
                      <td>{{ item.description }}</td>
                      <td>{{ item.sort }}</td>
                      <td>{{ statusText(item.status) }}</td>
                      <td>
                        <button class="btn secondary small" @click="openTypeEdit(item)">编辑</button>
                        <button class="btn danger small" @click="removeType(item.id)">删除</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
      </section>


      <section
        v-if="publicTreeActiveTab === 'nodes'"
        class="public-tree-tab-panel"
      >
        <section class="card">
              <div class="section-toolbar">
                <h3>人物管理</h3>
                <button class="btn small" @click="openNodeCreate">新增人物</button>
              </div>
        
              <div class="table-wrap">
                <table class="table">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>名称</th>
                      <th>角色</th>
                      <th>代际</th>
                      <th>排序</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in nodes" :key="item.id">
                      <td>{{ item.id }}</td>
                      <td>{{ item.name }}</td>
                      <td>{{ item.role }}</td>
                      <td>第 {{ item.generation }} 代</td>
                      <td>{{ item.sort }}</td>
                      <td>{{ statusText(item.status) }}</td>
                      <td>
                        <button class="btn secondary small" @click="openNodeEdit(item)">编辑</button>
                        <button class="btn danger small" @click="removeNode(item.id)">删除</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

        <section class="card">
              <div class="section-toolbar">
                <div>
                  <h3>关系管理</h3>
                  <p class="page-desc">关系连接两个公开树节点，供官网展示。</p>
                </div>
                <button class="btn small" :disabled="nodes.length < 2" @click="openRelationshipCreate">新增关系</button>
              </div>
        
              <div class="table-wrap">
                <table class="table">
                  <thead>
                    <tr>
                      <th>ID</th>
                      <th>人物 A</th>
                      <th>关系</th>
                      <th>人物 B</th>
                      <th>备注</th>
                      <th>排序</th>
                      <th>状态</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in relationships" :key="item.id">
                      <td>{{ item.id }}</td>
                      <td>{{ nodeName(item.from_node_id) }}</td>
                      <td>{{ item.relation_name || item.relation_type }}</td>
                      <td>{{ nodeName(item.to_node_id) }}</td>
                      <td>{{ item.remark }}</td>
                      <td>{{ item.sort }}</td>
                      <td>{{ statusText(item.status) }}</td>
                      <td>
                        <button class="btn secondary small" @click="openRelationshipEdit(item)">编辑</button>
                        <button class="btn danger small" @click="removeRelationship(item.id)">删除</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
      </section>

    </main>

    <div v-if="showNodeModal" class="modal-mask">
          <div class="modal">
            <h3 v-if="nodeForm.id">编辑人物</h3>
            <h3 v-else>新增人物</h3>
    
            <div class="form-grid">
              <div class="form-row">
                <label>名称</label>
                <input v-model="nodeForm.name" class="input" />
              </div>
              <div class="form-row">
                <label>角色</label>
                <input v-model="nodeForm.role" class="input" />
              </div>
              <div class="form-row">
                <label>代际</label>
                <input v-model.number="nodeForm.generation" type="number" class="input" />
              </div>
              <div class="form-row">
                <label>排序</label>
                <input v-model.number="nodeForm.sort" type="number" class="input" />
              </div>
              <div class="form-row">
                <label>描述</label>
                <textarea v-model="nodeForm.description" class="textarea"></textarea>
              </div>
              <div class="form-row">
                <label>状态</label>
                <select v-model.number="nodeForm.status" class="select">
                  <option :value="1">显示</option>
                  <option :value="0">隐藏</option>
                </select>
              </div>
            </div>
    
            <div class="modal-actions">
              <button class="btn secondary" @click="showNodeModal = false">取消</button>
              <button class="btn" @click="saveNode">保存</button>
            </div>
          </div>
        </div>

    <div v-if="showRelationshipModal" class="modal-mask">
          <div class="modal">
            <h3 v-if="relationshipForm.id">编辑关系</h3>
            <h3 v-else>新增关系</h3>
    
            <div class="form-grid">
              <div class="form-row">
                <label>人物 A</label>
                <select v-model.number="relationshipForm.from_node_id" class="select">
                  <option :value="0">请选择人物</option>
                  <option v-for="node in nodes" :key="node.id" :value="node.id">
                    {{ node.name }} - 第 {{ node.generation }} 代
                  </option>
                </select>
              </div>
              <div class="form-row">
                <label>关系类型</label>
                <select v-model="relationshipForm.relation_type" class="select" @change="applySelectedType">
                  <option value="">请选择关系类型</option>
                  <option v-for="type in activeRelationshipTypes" :key="type.code" :value="type.code">
                    {{ type.name }} - {{ type.description }}
                  </option>
                </select>
              </div>
              <div class="form-row">
                <label>人物 B</label>
                <select v-model.number="relationshipForm.to_node_id" class="select">
                  <option :value="0">请选择人物</option>
                  <option v-for="node in nodes" :key="node.id" :value="node.id">
                    {{ node.name }} - 第 {{ node.generation }} 代
                  </option>
                </select>
              </div>
              <div class="form-row">
                <label>备注</label>
                <input v-model="relationshipForm.remark" class="input" />
              </div>
              <div class="form-row">
                <label>排序</label>
                <input v-model.number="relationshipForm.sort" type="number" class="input" />
              </div>
              <div class="form-row">
                <label>状态</label>
                <select v-model.number="relationshipForm.status" class="select">
                  <option :value="1">显示</option>
                  <option :value="0">隐藏</option>
                </select>
              </div>
            </div>
    
            <div class="modal-actions">
              <button class="btn secondary" @click="showRelationshipModal = false">取消</button>
              <button class="btn" @click="saveRelationship">保存</button>
            </div>
          </div>
        </div>

    <div v-if="showTypeModal" class="modal-mask">
          <div class="modal">
            <h3 v-if="typeForm.id">编辑关系类型</h3>
            <h3 v-else>新增关系类型</h3>
    
            <div class="form-grid">
              <div class="form-row">
                <label>编码</label>
                <input v-model="typeForm.code" class="input" placeholder="father_child" />
              </div>
              <div class="form-row">
                <label>名称</label>
                <input v-model="typeForm.name" class="input" placeholder="父子" />
              </div>
              <div class="form-row">
                <label>分类</label>
                <select v-model="typeForm.category" class="select">
                  <option value="blood">血亲</option>
                  <option value="marriage">姻亲</option>
                  <option value="other">其他</option>
                </select>
              </div>
              <div class="form-row">
                <label>说明</label>
                <input v-model="typeForm.description" class="input" />
              </div>
              <div class="form-row">
                <label>排序</label>
                <input v-model.number="typeForm.sort" type="number" class="input" />
              </div>
              <div class="form-row">
                <label>状态</label>
                <select v-model.number="typeForm.status" class="select">
                  <option :value="1">显示</option>
                  <option :value="0">隐藏</option>
                </select>
              </div>
            </div>
    
            <div class="modal-actions">
              <button class="btn secondary" @click="showTypeModal = false">取消</button>
              <button class="btn" @click="saveType">保存</button>
            </div>
          </div>
        </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import {
  api,
  type PublicTreeConfig,
  type PublicTreeNode,
  type PublicTreeRelationship,
  type RelationshipType
} from '../api'

const publicTreeActiveTab = ref('config')

const nodes = ref<PublicTreeNode[]>([])
const relationships = ref<PublicTreeRelationship[]>([])
const relationshipTypes = ref<RelationshipType[]>([])
const showNodeModal = ref(false)
const showRelationshipModal = ref(false)
const showTypeModal = ref(false)

const config = reactive<PublicTreeConfig>({
  title: '',
  subtitle: '',
  description: '',
  status: 1
})

function emptyNode(): PublicTreeNode {
  return {
    name: '',
    role: '',
    generation: 1,
    description: '',
    sort: 0,
    status: 1
  }
}

function emptyRelationship(): PublicTreeRelationship {
  return {
    from_node_id: 0,
    to_node_id: 0,
    relation_type: '',
    relation_name: '',
    remark: '',
    sort: 0,
    status: 1
  }
}

function emptyType(): RelationshipType {
  return {
    code: '',
    name: '',
    category: 'blood',
    direction: 'one_way',
    description: '',
    sort: 0,
    status: 1
  }
}

const nodeForm = reactive<PublicTreeNode>(emptyNode())
const relationshipForm = reactive<PublicTreeRelationship>(emptyRelationship())
const typeForm = reactive<RelationshipType>(emptyType())

const activeRelationshipTypes = computed(() => relationshipTypes.value.filter((item) => item.status === 1))

onMounted(load)

async function load() {
  const [configRes, nodeRes, relationshipRes, typeRes] = await Promise.all([
    api.publicTreeConfig(),
    api.publicTreeNodeList(),
    api.publicTreeRelationshipList(),
    api.relationshipTypeList()
  ])

  Object.assign(config, configRes)
  nodes.value = nodeRes
  relationships.value = relationshipRes
  relationshipTypes.value = typeRes
}

function statusText(status: number) {
  return status === 1 ? '显示' : '隐藏'
}

function nodeName(id: number) {
  return nodes.value.find((item) => item.id === id)?.name || '-'
}

function categoryText(category: string) {
  if (category === 'blood') return '血亲'
  if (category === 'marriage') return '姻亲'
  return '其他'
}

async function saveConfig() {
  await api.publicTreeConfigUpdate(config)
  alert('公开树配置已保存')
  await load()
}

function openNodeCreate() {
  Object.assign(nodeForm, emptyNode())
  showNodeModal.value = true
}

function openNodeEdit(item: PublicTreeNode) {
  Object.assign(nodeForm, item)
  showNodeModal.value = true
}

async function saveNode() {
  if (!nodeForm.name) {
    alert('请输入节点名称')
    return
  }

  if (nodeForm.id) {
    await api.publicTreeNodeUpdate(nodeForm)
  } else {
    await api.publicTreeNodeCreate(nodeForm)
  }

  showNodeModal.value = false
  await load()
}

async function removeNode(id?: number) {
  if (!id) return
  if (!confirm('确定删除这个节点吗？相关关系也会一起删除。')) return

  await api.publicTreeNodeDelete(id)
  await load()
}

function openRelationshipCreate() {
  Object.assign(relationshipForm, emptyRelationship())
  showRelationshipModal.value = true
}

function openRelationshipEdit(item: PublicTreeRelationship) {
  Object.assign(relationshipForm, item)
  showRelationshipModal.value = true
}

function applySelectedType() {
  const type = relationshipTypes.value.find((item) => item.code === relationshipForm.relation_type)

  if (!type) return

  relationshipForm.relation_name = type.name
}

async function saveRelationship() {
  if (!relationshipForm.from_node_id || !relationshipForm.to_node_id) {
    alert('请选择关系节点')
    return
  }

  if (!relationshipForm.relation_type) {
    alert('请填写关系类型')
    return
  }

  if (relationshipForm.id) {
    await api.publicTreeRelationshipUpdate(relationshipForm)
  } else {
    await api.publicTreeRelationshipCreate(relationshipForm)
  }

  showRelationshipModal.value = false
  await load()
}

async function removeRelationship(id?: number) {
  if (!id) return
  if (!confirm('确定删除这条公开树关系吗？')) return

  await api.publicTreeRelationshipDelete(id)
  await load()
}

function openTypeCreate() {
  Object.assign(typeForm, emptyType())
  showTypeModal.value = true
}

function openTypeEdit(item: RelationshipType) {
  Object.assign(typeForm, item)
  showTypeModal.value = true
}

async function saveType() {
  if (!typeForm.code || !typeForm.name) {
    alert('请填写关系编码和名称')
    return
  }

  if (typeForm.id) {
    await api.relationshipTypeUpdate(typeForm)
  } else {
    await api.relationshipTypeCreate(typeForm)
  }

  showTypeModal.value = false
  await load()
}

async function removeType(id?: number) {
  if (!id) return
  if (!confirm('确定删除这个关系类型吗？已经被公开树关系使用的类型不能删除。')) return

  await api.relationshipTypeDelete(id)
  await load()
}
</script>

<style scoped>
.public-tree-page {
  display: grid;
  gap: 18px;
}

.page-desc {
  margin: 6px 0 0;
  color: #7c4a2a;
}

.section-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

.section-toolbar h3 {
  margin: 0;
}

/* compact-public-tree-admin: 后台官网关系树管理页紧凑布局 */
.compact-public-tree-admin {
  position: relative !important;
  inset: auto !important;
  z-index: auto !important;
  width: 100% !important;
  max-width: 100% !important;
  min-height: auto !important;
  height: auto !important;
  overflow: visible !important;
  box-sizing: border-box !important;
  padding: 14px !important;
  display: grid;
  gap: 12px;
  background: transparent !important;
}

.compact-public-tree-admin *,
.compact-public-tree-admin *::before,
.compact-public-tree-admin *::after {
  box-sizing: border-box;
}

.compact-public-tree-admin .page,
.compact-public-tree-admin .content,
.compact-public-tree-admin .main,
.compact-public-tree-admin .panel,
.compact-public-tree-admin .card,
.compact-public-tree-admin .section {
  position: relative !important;
  inset: auto !important;
  z-index: auto !important;
  width: 100% !important;
  max-width: 100% !important;
}

.compact-public-tree-admin h1,
.compact-public-tree-admin h2,
.compact-public-tree-admin h3 {
  margin: 0 0 8px !important;
  line-height: 1.25;
}

.compact-public-tree-admin p {
  margin: 4px 0 8px !important;
  line-height: 1.55;
}

.compact-public-tree-admin .card,
.compact-public-tree-admin .panel,
.compact-public-tree-admin article {
  padding: 14px !important;
  border-radius: 16px !important;
  margin-bottom: 0 !important;
}

.compact-public-tree-admin .toolbar,
.compact-public-tree-admin .actions,
.compact-public-tree-admin .filter,
.compact-public-tree-admin .filters {
  display: flex !important;
  align-items: center !important;
  flex-wrap: wrap !important;
  gap: 8px !important;
  margin-bottom: 10px !important;
}

.compact-public-tree-admin button,
.compact-public-tree-admin .btn {
  min-height: 32px !important;
  height: 32px !important;
  padding: 0 12px !important;
  border-radius: 10px !important;
  font-size: 13px !important;
  line-height: 32px !important;
}

.compact-public-tree-admin input,
.compact-public-tree-admin select,
.compact-public-tree-admin textarea {
  min-height: 32px !important;
  padding: 6px 10px !important;
  border-radius: 10px !important;
  font-size: 13px !important;
}

.compact-public-tree-admin textarea {
  min-height: 58px !important;
}

.compact-public-tree-admin .form-grid,
.compact-public-tree-admin .grid,
.compact-public-tree-admin .form {
  gap: 10px !important;
}

.compact-public-tree-admin .table-wrap,
.compact-public-tree-admin .table-wrapper,
.compact-public-tree-admin .list-wrap {
  max-height: 330px !important;
  overflow: auto !important;
  border-radius: 14px !important;
}

.compact-public-tree-admin table {
  width: 100%;
  font-size: 13px !important;
  border-collapse: collapse;
}

.compact-public-tree-admin th,
.compact-public-tree-admin td {
  padding: 7px 9px !important;
  line-height: 1.35 !important;
  vertical-align: middle !important;
}

.compact-public-tree-admin .preview,
.compact-public-tree-admin .tree-preview,
.compact-public-tree-admin .tree-board,
.compact-public-tree-admin .relation-preview {
  max-height: 420px !important;
  overflow: auto !important;
}

.compact-public-tree-admin .modal,
.compact-public-tree-admin .dialog,
.compact-public-tree-admin .drawer {
  z-index: 2000 !important;
}

.compact-public-tree-admin .compact-two-columns,
.compact-public-tree-admin .manage-grid,
.compact-public-tree-admin .panel-grid,
.compact-public-tree-admin .content-grid {
  display: grid !important;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) !important;
  gap: 12px !important;
  align-items: start !important;
}

.compact-public-tree-admin .empty,
.compact-public-tree-admin .tip,
.compact-public-tree-admin .notice {
  padding: 12px 14px !important;
  border-radius: 14px !important;
  margin: 0 !important;
}

@media (max-width: 1200px) {
  .compact-public-tree-admin .compact-two-columns,
  .compact-public-tree-admin .manage-grid,
  .compact-public-tree-admin .panel-grid,
  .compact-public-tree-admin .content-grid {
    grid-template-columns: 1fr !important;
  }
}

@media (max-width: 768px) {
  .compact-public-tree-admin {
    padding: 10px !important;
  }

  .compact-public-tree-admin .card,
  .compact-public-tree-admin .panel,
  .compact-public-tree-admin article {
    padding: 12px !important;
  }

  .compact-public-tree-admin .table-wrap,
  .compact-public-tree-admin .table-wrapper,
  .compact-public-tree-admin .list-wrap {
    max-height: 280px !important;
  }
}

/* public tree left tabs layout */
.public-tree-tab-page {
  display: grid !important;
  grid-template-columns: 188px minmax(0, 1fr) !important;
  gap: 14px !important;
  align-items: start !important;
  width: 100% !important;
  max-width: 100% !important;
  padding: 14px !important;
  box-sizing: border-box !important;
  overflow: visible !important;
  background: transparent !important;
}

.public-tree-tab-sidebar {
  position: sticky;
  top: 14px;
  z-index: 5;
  display: grid;
  gap: 8px;
  padding: 12px;
  border-radius: 18px;
  background: #fffaf3;
  border: 1px solid #ead8c0;
  box-shadow: 0 12px 32px rgba(80, 45, 20, 0.06);
}

.public-tree-tab-title {
  display: grid;
  gap: 4px;
  padding: 6px 6px 10px;
  border-bottom: 1px solid #f0dfc9;
}

.public-tree-tab-title strong {
  color: #3f2418;
  font-size: 16px;
}

.public-tree-tab-title span {
  color: #9a6a45;
  font-size: 12px;
}

.public-tree-tab-btn {
  width: 100%;
  min-height: auto !important;
  height: auto !important;
  padding: 11px 12px !important;
  border: 1px solid transparent !important;
  border-radius: 14px !important;
  background: transparent !important;
  color: #7c4a2a !important;
  text-align: left !important;
  line-height: 1.35 !important;
  cursor: pointer;
}

.public-tree-tab-btn strong {
  display: block;
  font-size: 14px;
  color: inherit;
}

.public-tree-tab-btn span {
  display: block;
  margin-top: 3px;
  font-size: 12px;
  color: #9a6a45;
}

.public-tree-tab-btn.active {
  background: linear-gradient(135deg, #7f1d1d, #b91c1c) !important;
  border-color: rgba(253, 230, 138, 0.55) !important;
  color: #fde68a !important;
  box-shadow: 0 10px 24px rgba(127, 29, 29, 0.16);
}

.public-tree-tab-btn.active span {
  color: #ffedd5;
}

.public-tree-tab-content {
  min-width: 0;
  display: block;
}

.public-tree-tab-panel {
  display: grid;
  gap: 12px;
  min-width: 0;
}

.public-tree-tab-panel > * {
  max-width: 100% !important;
}

.public-tree-tab-panel .card,
.public-tree-tab-panel .panel,
.public-tree-tab-panel article,
.public-tree-tab-panel section {
  margin-bottom: 0 !important;
}

.public-tree-tab-page .table-wrap,
.public-tree-tab-page .table-wrapper,
.public-tree-tab-page .list-wrap {
  max-height: 430px !important;
  overflow: auto !important;
}

.public-tree-tab-page .toolbar,
.public-tree-tab-page .actions,
.public-tree-tab-page .filter,
.public-tree-tab-page .filters {
  display: flex !important;
  flex-wrap: wrap !important;
  gap: 8px !important;
  align-items: center !important;
}

.public-tree-tab-page button,
.public-tree-tab-page .btn {
  min-height: 32px !important;
}

@media (max-width: 900px) {
  .public-tree-tab-page {
    grid-template-columns: 1fr !important;
  }

  .public-tree-tab-sidebar {
    position: relative;
    top: auto;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .public-tree-tab-title {
    grid-column: 1 / -1;
  }
}


/* public tree top tabs layout */
.public-tree-tab-page {
  display: block !important;
  width: 100% !important;
  max-width: 100% !important;
  padding: 16px !important;
  box-sizing: border-box !important;
  background: transparent !important;
}

.public-tree-tab-sidebar {
  position: sticky !important;
  top: 0 !important;
  z-index: 20 !important;
  display: flex !important;
  align-items: center !important;
  gap: 8px !important;
  width: 100% !important;
  margin-bottom: 14px !important;
  padding: 10px 12px !important;
  border-radius: 16px !important;
  background: rgba(255, 250, 243, 0.96) !important;
  border: 1px solid #ead8c0 !important;
  box-shadow: 0 10px 28px rgba(80, 45, 20, 0.06) !important;
  overflow-x: auto !important;
}

.public-tree-tab-title {
  display: flex !important;
  align-items: center !important;
  gap: 8px !important;
  flex: 0 0 auto !important;
  padding: 0 14px 0 2px !important;
  border-right: 1px solid #ead8c0 !important;
  border-bottom: 0 !important;
  white-space: nowrap !important;
}

.public-tree-tab-title strong {
  color: #3f2418 !important;
  font-size: 15px !important;
  font-weight: 900 !important;
}

.public-tree-tab-title span {
  display: none !important;
}

.public-tree-tab-btn {
  flex: 0 0 auto !important;
  width: auto !important;
  min-width: 92px !important;
  height: 34px !important;
  min-height: 34px !important;
  padding: 0 16px !important;
  border-radius: 999px !important;
  border: 1px solid #ead8c0 !important;
  background: #fff7ed !important;
  color: #7c4a2a !important;
  text-align: center !important;
  cursor: pointer !important;
  box-shadow: none !important;
}

.public-tree-tab-btn strong {
  display: block !important;
  color: inherit !important;
  font-size: 14px !important;
  line-height: 32px !important;
  font-weight: 900 !important;
}

.public-tree-tab-btn span {
  display: none !important;
}

.public-tree-tab-btn.active {
  background: linear-gradient(135deg, #8f1d1d, #b91c1c) !important;
  border-color: #8f1d1d !important;
  color: #fde68a !important;
  box-shadow: 0 8px 18px rgba(127, 29, 29, 0.16) !important;
}

.public-tree-tab-content {
  width: 100% !important;
  min-width: 0 !important;
}

.public-tree-tab-panel {
  display: grid !important;
  gap: 14px !important;
  min-width: 0 !important;
}

.public-tree-tab-panel > * {
  max-width: 100% !important;
}

.public-tree-tab-page .card,
.public-tree-tab-page .panel,
.public-tree-tab-page article {
  border-radius: 18px !important;
}

.public-tree-tab-page .table-wrap,
.public-tree-tab-page .table-wrapper,
.public-tree-tab-page .list-wrap {
  max-height: 460px !important;
  overflow: auto !important;
}

@media (max-width: 768px) {
  .public-tree-tab-page {
    padding: 10px !important;
  }

  .public-tree-tab-sidebar {
    top: 0 !important;
    border-radius: 14px !important;
  }

  .public-tree-tab-title {
    display: none !important;
  }

  .public-tree-tab-btn {
    min-width: 84px !important;
    padding: 0 12px !important;
  }
}


/* fix public tree tabs switch */
.public-tree-tab-panel[style*="display: none"] {
  display: none !important;
}

.public-tree-tab-panel {
  min-width: 0;
}

</style>
