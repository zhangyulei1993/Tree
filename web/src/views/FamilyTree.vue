<template>
  <section class="section tree-page">
    <div class="container">
      <div class="page-head">
        <p class="eyebrow">&#x5BB6;&#x65CF;&#x8109;&#x7EDC;</p>
        <h1 class="section-title">&#x4EB2;&#x5C5E;&#x5173;&#x7CFB;&#x6811;</h1>
        <p class="section-subtitle">
          &#x6309;&#x4EE3;&#x9645;&#x4ECE;&#x5DE6;&#x5230;&#x53F3;&#x6A2A;&#x5411;&#x5C55;&#x5F00;&#xFF0C;&#x53EF;&#x6E05;&#x6670;&#x67E5;&#x770B;&#x5BB6;&#x65CF;&#x6210;&#x5458;&#x3001;&#x914D;&#x5076;&#x5173;&#x7CFB;&#x548C;&#x7236;&#x6BCD;&#x5B50;&#x5973;&#x5173;&#x7CFB;&#x3002;
        </p>
      </div>

      <div v-if="loading" class="empty">&#x52A0;&#x8F7D;&#x4E2D;...</div>

      <div v-else-if="error" class="empty">
        {{ error }}
      </div>

      <div v-else-if="!tree.members.length" class="empty">
        &#x6682;&#x65E0;&#x5BB6;&#x65CF;&#x6210;&#x5458;&#x6570;&#x636E;&#xFF0C;&#x8BF7;&#x5148;&#x5230;&#x540E;&#x53F0;&#x5F55;&#x5165;&#x5BB6;&#x65CF;&#x3001;&#x6210;&#x5458;&#x548C;&#x5173;&#x7CFB;&#x3002;
      </div>

      <div v-else>
        <div class="family-card card">
          <img v-if="tree.family.cover" :src="tree.family.cover" alt="family cover" />
          <div>
            <h2>{{ tree.family.name }}</h2>
            <p v-if="tree.family.origin">&#x7956;&#x7C4D;&#xFF1A;{{ tree.family.origin }}</p>
            <p v-if="tree.family.description">{{ tree.family.description }}</p>
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
                  v-for="member in listByGeneration(generation)"
                  :key="member.id"
                  class="member-card"
                >
                  <div class="avatar">
                    <img v-if="member.avatar" :src="member.avatar" alt="avatar" />
                    <span v-else>{{ firstChar(member.name) }}</span>
                  </div>

                  <div class="member-main">
                    <h3>{{ member.name }}</h3>
                    <p>{{ genderLabel(member.gender) }}</p>

                    <p v-if="member.birthday">
                      &#x751F;&#x65E5;&#xFF1A;{{ member.birthday }}
                    </p>

                    <p v-if="parentNames(member.id).length">
                      &#x7236;&#x6BCD;&#xFF1A;{{ formatNames(parentNames(member.id)) }}
                    </p>

                    <p v-if="spouseNames(member.id).length">
                      &#x914D;&#x5076;&#xFF1A;{{ formatNames(spouseNames(member.id)) }}
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
              <strong>{{ memberName(item.from_member_id) }}</strong>
              <span>{{ relationLabel(item.relation_type) }}</span>
              <strong>{{ memberName(item.to_member_id) }}</strong>
            </div>
          </div>
        </section>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type FamilyTreeResult } from '../api'

const loading = ref(false)
const error = ref('')

const tree = reactive<FamilyTreeResult>({
  family: {
    id: 0,
    name: '',
    surname: '',
    origin: '',
    cover: '',
    description: '',
    status: 1
  },
  members: [],
  relationships: [],
  generation_keys: [],
  generations: {}
})

onMounted(load)

async function load() {
  loading.value = true
  error.value = ''

  try {
    const res = await api.familyTree()
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

function memberName(id: number) {
  return tree.members.find((item) => item.id === id)?.name || '-'
}

function firstChar(name: string) {
  return name ? name.slice(0, 1) : '?'
}

function genderLabel(gender: string) {
  if (gender === 'male') return '\u7537'
  if (gender === 'female') return '\u5973'
  return '\u672A\u77E5'
}

function relationLabel(type: string) {
  if (type === 'spouse') return '\u914D\u5076'
  if (type === 'parent_child') return '\u7236\u6BCD\u5B50\u5973'
  return type
}

function formatNames(names: string[]) {
  return names.join('\u3001')
}

function parentNames(memberId: number) {
  return tree.relationships
    .filter((item) => item.relation_type === 'parent_child' && item.to_member_id === memberId)
    .map((item) => memberName(item.from_member_id))
    .filter((name) => name !== '-')
}

function spouseNames(memberId: number) {
  return tree.relationships
    .filter((item) => {
      return item.relation_type === 'spouse' &&
        (item.from_member_id === memberId || item.to_member_id === memberId)
    })
    .map((item) => {
      const spouseId = item.from_member_id === memberId ? item.to_member_id : item.from_member_id
      return memberName(spouseId)
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
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 24px;
  padding: 22px;
  margin-bottom: 28px;
}

.family-card img {
  width: 100%;
  height: 160px;
  object-fit: cover;
  border-radius: 16px;
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

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
  .family-card {
    grid-template-columns: 1fr;
  }

  .relation-grid {
    grid-template-columns: 1fr;
  }
}
</style>