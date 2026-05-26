<template>
  <section class="section tree-page">
    <div class="container">
      <div class="page-head">
        <p class="eyebrow">&#x516C;&#x5F00;&#x5C55;&#x793A;</p>
        <h1 class="section-title">{{ pageTitle }}</h1>
        <p class="section-subtitle">
          {{ pageSubtitle }}
        </p>
      </div>

      <div v-if="loading" class="empty">&#x52A0;&#x8F7D;&#x4E2D;...</div>

      <div v-else-if="error" class="empty">
        {{ error }}
      </div>

      <div v-else-if="!tree.nodes.length" class="empty">
        &#x6682;&#x65E0;&#x516C;&#x5F00;&#x6811;&#x8282;&#x70B9;&#x6570;&#x636E;&#xFF0C;&#x8BF7;&#x5148;&#x5230;&#x540E;&#x53F0;&#x7EF4;&#x62A4;&#x516C;&#x5F00;&#x6811;&#x5C55;&#x793A;&#x3002;
      </div>

      <div v-else>
        <div class="family-card card">
          <div>
            <h2>{{ tree.config.title }}</h2>
            <p v-if="tree.config.description">{{ tree.config.description }}</p>
          </div>
        </div>

        <div class="tree-scroll">
          <div class="tree-lane">
            <section
              v-for="(generation, index) in tree.generation_keys"
              :key="generation"
              class="generation-column"
            >
              <div class="generation-title">
                &#x7B2C; {{ generation }} &#x4EE3;
              </div>

              <div class="member-list">
                <article
                  v-for="node in listByGeneration(generation)"
                  :key="node.id"
                  class="member-card"
                >
                  <div class="avatar">
                    <span>{{ firstChar(node.name) }}</span>
                  </div>

                  <div class="member-main">
                    <h3>{{ node.name }}</h3>
                    <p v-if="node.role">{{ node.role }}</p>
                    <p v-if="node.description">{{ node.description }}</p>

                    <p v-if="parentNames(node.id).length">
                      &#x4E0A;&#x7EA7;&#x5173;&#x7CFB;&#xFF1A;{{ formatNames(parentNames(node.id)) }}
                    </p>

                    <p v-if="spouseNames(node.id).length">
                      &#x914D;&#x5076;&#xFF1A;{{ formatNames(spouseNames(node.id)) }}
                    </p>
                  </div>
                </article>
              </div>

              <div v-if="index < tree.generation_keys.length - 1" class="column-arrow">
                &#8594;
              </div>
            </section>
          </div>
        </div>

        <section class="relations card">
          <h2>&#x5173;&#x7CFB;&#x660E;&#x7EC6;</h2>

          <div v-if="!tree.relationships.length" class="relation-empty">
            &#x6682;&#x65E0;&#x5173;&#x7CFB;&#x6570;&#x636E;
          </div>

          <div v-else class="relation-grid">
            <div
              v-for="item in tree.relationships"
              :key="item.id"
              class="relation-item"
            >
              <strong>{{ nodeName(item.from_node_id) }}</strong>
              <span>{{ relationLabel(item) }}</span>
              <strong>{{ nodeName(item.to_node_id) }}</strong>
            </div>
          </div>
        </section>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type PublicTreeRelationship, type PublicTreeResult } from '../api'

const loading = ref(false)
const error = ref('')

const tree = reactive<PublicTreeResult>({
  config: {
    id: 0,
    title: '',
    subtitle: '',
    description: '',
    status: 1
  },
  nodes: [],
  relationships: [],
  generation_keys: [],
  generations: {}
})

const pageTitle = computed(() => tree.config.title || '\u4EB2\u5C5E\u5173\u7CFB\u6811')
const pageSubtitle = computed(() => {
  return tree.config.subtitle || '\u6309\u4EE3\u9645\u4ECE\u5DE6\u5230\u53F3\u6A2A\u5411\u5C55\u5F00\uFF0C\u5C55\u793A\u516C\u5F00\u5173\u7CFB\u8282\u70B9\u548C\u5173\u7CFB\u3002'
})

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''

  try {
    const res = await api.publicTree()
    Object.assign(tree, res)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '\u52A0\u8F7D\u5931\u8D25'
  } finally {
    loading.value = false
  }
}

function listByGeneration(generation: number) {
  return tree.generations[String(generation)] || []
}

function nodeName(id: number) {
  return tree.nodes.find((item) => item.id === id)?.name || '-'
}

function firstChar(name: string) {
  return name ? name.slice(0, 1) : '?'
}

function relationLabel(item: PublicTreeRelationship) {
  return item.relation_name || item.relation_type
}

function formatNames(names: string[]) {
  return names.join('\u3001')
}

function parentNames(nodeId: number) {
  return tree.relationships
    .filter((item) => item.relation_type !== 'spouse' && item.to_node_id === nodeId)
    .map((item) => nodeName(item.from_node_id))
    .filter((name) => name !== '-')
}

function spouseNames(nodeId: number) {
  return tree.relationships
    .filter((item) => {
      return item.relation_type.startsWith('spouse') &&
        (item.from_node_id === nodeId || item.to_node_id === nodeId)
    })
    .map((item) => {
      const spouseId = item.from_node_id === nodeId ? item.to_node_id : item.from_node_id
      return nodeName(spouseId)
    })
    .filter((name) => name !== '-')
}
</script>

<style scoped>
.tree-page {
  background:
    radial-gradient(circle at 10% 10%, rgba(185, 28, 28, 0.12), transparent 30%),
    linear-gradient(180deg, #fff7ed, #f8f1e8);
}

.page-head {
  max-width: 820px;
}

.eyebrow {
  margin: 0 0 10px;
  color: #b91c1c;
  font-weight: 900;
  letter-spacing: 0.16em;
}

.family-card {
  padding: 22px;
  margin-bottom: 28px;
}

.family-card h2 {
  margin: 0 0 12px;
  font-size: 28px;
}

.family-card p {
  color: #7c4a2a;
  line-height: 1.8;
}

.tree-scroll {
  overflow-x: auto;
  padding: 8px 0 24px;
}

.tree-lane {
  display: flex;
  align-items: flex-start;
  gap: 34px;
  min-width: max-content;
}

.generation-column {
  position: relative;
  width: 300px;
  flex: 0 0 300px;
}

.generation-title {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 42px;
  padding: 0 18px;
  margin-bottom: 18px;
  border-radius: 999px;
  background: #7f1d1d;
  color: #fde68a;
  font-weight: 900;
  box-shadow: 0 12px 26px rgba(127, 29, 29, 0.18);
}

.member-list {
  display: grid;
  gap: 16px;
}

.member-card {
  position: relative;
  display: flex;
  gap: 14px;
  padding: 16px;
  border-radius: 20px;
  background: #fffaf3;
  border: 1px solid #ead8c0;
  box-shadow: 0 14px 38px rgba(80, 45, 20, 0.08);
}

.member-card::before {
  content: "";
  position: absolute;
  left: -10px;
  top: 28px;
  width: 10px;
  border-top: 2px solid #d97706;
  opacity: 0.55;
}

.avatar {
  width: 54px;
  height: 54px;
  flex: 0 0 54px;
  border-radius: 18px;
  background: linear-gradient(135deg, #b91c1c, #7f1d1d);
  color: #fde68a;
  display: grid;
  place-items: center;
  font-size: 22px;
  font-weight: 900;
  overflow: hidden;
}

.member-main h3 {
  margin: 0 0 6px;
  font-size: 18px;
}

.member-main p {
  margin: 5px 0;
  color: #7c4a2a;
  font-size: 14px;
  line-height: 1.55;
}

.column-arrow {
  position: absolute;
  right: -30px;
  top: 54px;
  color: #b91c1c;
  font-size: 28px;
  font-weight: 900;
}

.relations {
  padding: 22px;
  margin-top: 12px;
}

.relations h2 {
  margin: 0 0 18px;
}

.relation-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.relation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #fff4e5;
  border-radius: 14px;
  color: #5b3924;
}

.relation-item span {
  color: #b91c1c;
  font-weight: 900;
}

.relation-empty {
  color: #7c4a2a;
}

@media (max-width: 768px) {
  .relation-grid {
    grid-template-columns: 1fr;
  }
}
</style>
