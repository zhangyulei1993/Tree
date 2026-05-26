<template>
  <div class="public-tree-page">
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
            <input v-model="relationshipForm.relation_type" class="input" placeholder="father_child / spouse / uncle_nephew" />
          </div>
          <div class="form-row">
            <label>关系名称</label>
            <input v-model="relationshipForm.relation_name" class="input" placeholder="父子 / 配偶 / 叔侄" />
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
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type PublicTreeConfig, type PublicTreeNode, type PublicTreeRelationship } from '../api'

const nodes = ref<PublicTreeNode[]>([])
const relationships = ref<PublicTreeRelationship[]>([])
const showNodeModal = ref(false)
const showRelationshipModal = ref(false)

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

const nodeForm = reactive<PublicTreeNode>(emptyNode())
const relationshipForm = reactive<PublicTreeRelationship>(emptyRelationship())

onMounted(load)

async function load() {
  const [configRes, nodeRes, relationshipRes] = await Promise.all([
    api.publicTreeConfig(),
    api.publicTreeNodeList(),
    api.publicTreeRelationshipList()
  ])

  Object.assign(config, configRes)
  nodes.value = nodeRes
  relationships.value = relationshipRes
}

function statusText(status: number) {
  return status === 1 ? '\u663E\u793A' : '\u9690\u85CF'
}

function nodeName(id: number) {
  return nodes.value.find((item) => item.id === id)?.name || '-'
}

async function saveConfig() {
  await api.publicTreeConfigUpdate(config)
  alert('\u516C\u5F00\u6811\u914D\u7F6E\u5DF2\u4FDD\u5B58')
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
    alert('\u8BF7\u8F93\u5165\u8282\u70B9\u540D\u79F0')
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
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u4E2A\u8282\u70B9\u5417\uFF1F\u76F8\u5173\u5173\u7CFB\u4E5F\u4F1A\u4E00\u8D77\u5220\u9664\u3002')) return

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

async function saveRelationship() {
  if (!relationshipForm.from_node_id || !relationshipForm.to_node_id) {
    alert('\u8BF7\u9009\u62E9\u5173\u7CFB\u8282\u70B9')
    return
  }

  if (!relationshipForm.relation_type) {
    alert('\u8BF7\u586B\u5199\u5173\u7CFB\u7C7B\u578B')
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
  if (!confirm('\u786E\u5B9A\u5220\u9664\u8FD9\u6761\u516C\u5F00\u6811\u5173\u7CFB\u5417\uFF1F')) return

  await api.publicTreeRelationshipDelete(id)
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
</style>
