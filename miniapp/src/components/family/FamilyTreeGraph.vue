<template>
  <view class="family-tree-graph">
    <view v-if="!layout.success" class="tree-fallback">
      <MiniNotice tone="warm" title="查看提示">
        {{ layout.message || '家庭树暂时无法生成，可查看关系明细' }}
      </MiniNotice>
    </view>

    <template v-else>
      <view class="tree-stage-head">
        <view>
          <text class="stage-kicker">Genealogy Scroll</text>
          <text class="stage-title">世系脉络</text>
        </view>
        <text class="stage-legend">墨线血缘 · 金线姻亲</text>
      </view>

      <view class="tree-stage">
      <scroll-view
        class="tree-scroll"
        scroll-x
        scroll-y
        :style="{ height: scrollViewHeight }"
        :scroll-into-view="scrollIntoViewId"
        scroll-with-animation
        show-scrollbar
      >
        <view
          class="tree-canvas"
          :style="{
            width: `${layout.canvasWidth}rpx`,
            height: `${layout.canvasHeight}rpx`
          }"
        >
          <view
            v-for="link in layout.renderLinks"
            :key="link.id"
            class="links-layer"
          >
            <view
              v-for="(segment, index) in link.segments"
              :key="`${link.id}-${index}`"
              class="link-segment"
              :class="[segment.orientation, linkToneClass(link)]"
              :style="{
                left: `${segment.x}rpx`,
                top: `${segment.y}rpx`,
                width: `${segment.width}rpx`,
                height: `${segment.height}rpx`
              }"
            />
          </view>

          <view
            v-for="node in layout.renderNodes"
            :id="node.scrollAnchorId"
            :key="node.id"
            class="couple-node"
            :class="{ 'root-node': isPatrilinealRoot(node) }"
            :style="{
              left: `${node.x}rpx`,
              top: `${node.y}rpx`,
              width: `${node.width}rpx`,
              height: `${node.height}rpx`
            }"
          >
            <view class="couple-row">
              <template v-for="(parent, index) in node.parents" :key="parent.memberId">
                <view v-if="index > 0" class="spouse-join" aria-hidden="true">
                  <view class="spouse-line" />
                </view>
                <view class="tree-person-hit" @tap.stop="handleNodeSelect(parent)">
                  <FamilyTreePersonCard
                    :node="parent"
                    :is-self="parent.memberId === effectiveViewerMemberId"
                    :show-binding="showBinding"
                    :kinship-title="kinshipTitles[parent.memberId]"
                    :family-surname="familySurname"
                    :interactive="interactive"
                    :selectable="isNodeSelectable(parent)"
                    :in-focus="isNodeInFocus(parent.memberId)"
                    :is-root="isPatrilinealRoot(node) && nodeLineageRole(parent.memberId) === 'PATRILINEAL'"
                    :lineage-role="nodeLineageRole(parent.memberId)"
                    :spouse-relation-label="spouseRelationByMemberId[parent.memberId]"
                    :health-hints="healthHints[parent.memberId]"
                    :portrait-theme="portraitTheme"
                    :portrait-region="portraitRegion"
                    :portrait-region-code="portraitRegionCode"
                    :portrait-appearance-variant="portraitAppearanceByMemberId[parent.memberId]"
                    @select="handleNodeSelect"
                  />
                </view>
              </template>
            </view>
          </view>
        </view>
      </scroll-view>
      </view>

      <view class="tree-footnote">
        <text class="footnote">
          共 {{ layout.memberCount }} 位成员，树中展示 {{ layout.renderedCount }} 位
          <text v-if="effectiveViewerMemberId"> · 深蓝边饰为当前视角</text>
        </text>
      </view>

      <view v-if="layout.unlocatedMembers.length > 0" class="unlocated-section">
        <view class="unlocated-head">
          <text class="unlocated-title">待补充关系成员</text>
          <text class="unlocated-desc">尚未建立亲属关系；可点击成员卡片补充关系或编辑资料。</text>
        </view>
        <view class="unlocated-list">
          <view
            v-for="node in layout.unlocatedMembers"
            :key="node.memberId"
            class="unlocated-card-hit"
            :class="{ interactive, selectable: isUnlocatedNodeSelectable(node) }"
            @tap.stop="handleUnlocatedSelect(node)"
          >
            <FamilyTreePersonCard
              :node="node"
              :is-self="node.memberId === effectiveViewerMemberId"
              :show-binding="showBinding"
              :kinship-title="kinshipTitles[node.memberId]"
              :family-surname="familySurname"
              :interactive="false"
              :selectable="isUnlocatedNodeSelectable(node)"
              :in-focus="isNodeInFocus(node.memberId)"
              :is-root="false"
              :lineage-role="nodeLineageRole(node.memberId)"
              :spouse-relation-label="spouseRelationByMemberId[node.memberId]"
              :health-hints="healthHints[node.memberId]"
              :portrait-theme="portraitTheme"
              :portrait-region="portraitRegion"
              :portrait-region-code="portraitRegionCode"
              :portrait-appearance-variant="portraitAppearanceByMemberId[node.memberId]"
              @select="emit('select', $event)"
            />
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreePersonCard from '@/components/family/FamilyTreePersonCard.vue'
import { layoutFamilyTree } from '@/features/family-tree/layoutFamilyTree'
import { collectDirectLineMemberIds } from '@/features/family-tree/directLine'
import { buildMemberGraph } from '@/features/family-tree/graph'
import { buildPatrilinealDisplayRoles } from '@/features/family-tree/patrilinealDisplay'
import { buildPortraitAppearanceMap } from '@/features/family-tree/portraitCatalog'
import { buildRelGraph, resolveLineageViewerId } from '@/features/family-tree/relativeGraph'
import { buildKinshipTitleMap } from '@/features/family-tree/relativeTitle'
import { buildSpouseRelationLabelMap } from '@/features/family-tree/spouseRelationLabel'
import type { FamilyTreeResult } from '@/types/api'

const props = defineProps<{
  tree?: FamilyTreeResult | null
  viewerMemberId?: number | null
  showBinding?: boolean
  interactive?: boolean
  selectableAll?: boolean
  selectableDeceased?: boolean
  healthHints?: Record<number, string[]>
  portraitTheme?: 'default' | 'imperial' | 'regional'
  portraitRegion?: string | null
  portraitRegionCode?: string | null
}>()

const healthHints = computed(() => props.healthHints || {})
const portraitAppearanceByMemberId = computed(() => buildPortraitAppearanceMap(props.tree?.nodes || []))
const spouseRelationByMemberId = computed(() => buildSpouseRelationLabelMap(props.tree))

const emit = defineEmits<{
  select: [node: FamilyTreeResult['nodes'][number]]
}>()

function handleUnlocatedSelect(node: FamilyTreeResult['nodes'][number]) {
  if (!isUnlocatedNodeSelectable(node)) return
  emit('select', node)
}

function handleNodeSelect(node: FamilyTreeResult['nodes'][number]) {
  if (!isNodeSelectable(node)) return
  if (selectingMemberId.value === node.memberId) return
  selectingMemberId.value = node.memberId
  emit('select', node)
  // Both the card and its hit area can emit within one tap. Release the
  // guard after that event completes, rather than depending on a timer that
  // may be paused while a Mini Program page is hidden.
  Promise.resolve().then(() => {
    selectingMemberId.value = null
  })
}

function isNodeSelectable(node: FamilyTreeResult['nodes'][number]) {
  return Boolean(props.selectableAll || props.interactive || (props.selectableDeceased && node.isLiving === false))
}

function isUnlocatedNodeSelectable(node: FamilyTreeResult['nodes'][number]) {
	return Boolean(props.selectableAll || props.interactive || (props.selectableDeceased && node.isLiving === false))
}

const effectiveViewerMemberId = computed(() => {
  if (props.viewerMemberId == null) return props.viewerMemberId
  const graph = buildRelGraph(props.tree?.nodes ?? [], props.tree?.edges ?? [])
  return resolveLineageViewerId(graph, props.viewerMemberId)
})

// A card's component event and the surrounding full-footprint hit area can both
// fire on different Mini Program renderers. Keep both paths, but only act once.
const selectingMemberId = ref<number | null>(null)

// A return from member details reloads the tree. Make a stale tap guard
// impossible even if the previous page was suspended before its microtask.
watch(() => props.tree, () => {
  selectingMemberId.value = null
})

const focusMemberIds = computed(() => {
  if (effectiveViewerMemberId.value == null) return new Set<number>()
  return collectDirectLineMemberIds(
    effectiveViewerMemberId.value,
    buildMemberGraph({
      nodes: props.tree?.nodes ?? [],
      edges: props.tree?.edges ?? []
    })
  )
})

function isNodeInFocus(memberId: number) {
  return focusMemberIds.value.size === 0 || focusMemberIds.value.has(memberId)
}

const patrilinealRoles = computed(() => buildPatrilinealDisplayRoles({
  nodes: props.tree?.nodes ?? [],
  edges: props.tree?.edges ?? [],
  tree: props.tree?.tree ?? []
}))

function nodeLineageRole(memberId: number) {
  return patrilinealRoles.value[memberId] || 'EXTERNAL'
}

const rootNodeY = computed(() => Math.min(...layout.value.renderNodes.map((node) => node.y)))

function isRootNode(y: number) {
  return Number.isFinite(rootNodeY.value) && y === rootNodeY.value
}

function isPatrilinealRoot(node: typeof layout.value.renderNodes[number]) {
  return isRootNode(node.y) && node.parents.some((parent) => nodeLineageRole(parent.memberId) === 'PATRILINEAL')
}

function linkToneClass(link: typeof layout.value.renderLinks[number]) {
  const sourceIsPatrilineal = link.sourceMemberIds.some((memberId) => nodeLineageRole(memberId) === 'PATRILINEAL')
  const targetIsPatrilineal = link.targetMemberIds.some((memberId) => nodeLineageRole(memberId) === 'PATRILINEAL')
  const targetIsNatalFemale = link.targetMemberIds.some((memberId) => nodeLineageRole(memberId) === 'NATAL_FEMALE')
  return {
    highlighted: link.highlighted,
    'patrilineal-link': sourceIsPatrilineal && targetIsPatrilineal,
    'natal-female-link': sourceIsPatrilineal && targetIsNatalFemale
  }
}

const layout = computed(() =>
  layoutFamilyTree({
    nodes: props.tree?.nodes || [],
    edges: props.tree?.edges || [],
    tree: props.tree?.tree || [],
    viewerMemberId: effectiveViewerMemberId.value
  })
)

/** 视角相关称谓：viewerMemberId 或 graphVersion 变化时全量重算 */
const kinshipTitles = computed(() =>
  buildKinshipTitleMap({
    viewerMemberId: effectiveViewerMemberId.value,
    nodes: props.tree?.nodes ?? [],
    edges: props.tree?.edges ?? [],
    graphVersion: props.tree?.graphVersion ?? 0
  })
)

const familySurname = computed(() => {
  const counts = new Map<string, number>()
  for (const node of props.tree?.nodes || []) {
    const surname = (node.surname || node.displayName.trim().slice(0, 1)).trim()
    if (!surname) continue
    counts.set(surname, (counts.get(surname) || 0) + 1)
  }

  let best = ''
  let bestCount = 0
  for (const [surname, count] of counts.entries()) {
    if (count > bestCount) {
      best = surname
      bestCount = count
    }
  }
  return best
})

const scrollViewHeight = computed(() => {
  const canvasH = layout.value.canvasHeight || 600
  const capped = Math.min(Math.max(canvasH + 48, 480), 1200)
  return `${capped}rpx`
})

const scrollIntoViewId = ref('')

watch(
  () => layout.value.scrollIntoViewId,
  (nextId) => {
    if (!nextId) {
      scrollIntoViewId.value = ''
      return
    }
    scrollIntoViewId.value = ''
    setTimeout(() => {
      scrollIntoViewId.value = nextId
    }, 120)
  },
  { immediate: true }
)

</script>

<style scoped>
.family-tree-graph {
  width: 100%;
}
.tree-fallback {
  padding: 8rpx 0 16rpx;
}
.tree-stage-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18rpx;
  margin: 4rpx 0 10rpx;
  padding: 0 8rpx;
}
.stage-kicker,
.stage-title,
.stage-legend {
  display: block;
}
.stage-kicker {
  color: var(--archive-cinnabar, #a83b2d);
  font-size: 17rpx;
  font-weight: 700;
  letter-spacing: 3rpx;
  line-height: 1.2;
}
.stage-title {
  margin-top: 3rpx;
  color: var(--archive-ink, #243244);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 29rpx;
  font-weight: 700;
  line-height: 1.2;
}
.stage-legend {
  flex-shrink: 0;
  color: var(--archive-ink-soft, #657080);
  font-size: 19rpx;
  line-height: 1.3;
}
.tree-stage {
  position: relative;
  overflow: hidden;
  border-top: 1rpx solid rgba(92, 74, 48, 0.3);
  border-bottom: 1rpx solid rgba(92, 74, 48, 0.2);
  background: rgba(255, 251, 242, 0.42);
}
.tree-scroll {
  width: 100%;
}
.tree-canvas {
  position: relative;
  background-color: rgba(255, 249, 236, 0.36);
  background-image:
    linear-gradient(rgba(92, 74, 48, 0.03) 1rpx, transparent 1rpx),
    linear-gradient(90deg, rgba(168, 59, 45, 0.04) 1rpx, transparent 1rpx);
  background-size: 100% 34rpx, 34rpx 100%;
}
.tree-canvas::before {
  content: '谱';
  position: absolute;
  top: 26rpx;
  left: 18rpx;
  color: rgba(168, 59, 45, 0.1);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 52rpx;
  font-weight: 700;
  line-height: 1;
}
.links-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}
.link-segment {
  position: absolute;
  background: rgba(91, 77, 55, 0.34);
  opacity: 0.72;
  border-radius: 999rpx;
}
.link-segment.highlighted {
  opacity: 0.72;
  background: rgba(91, 77, 55, 0.34);
  box-shadow: none;
}
.link-segment.patrilineal-link {
  background: rgba(91, 77, 55, 0.34);
  opacity: 0.72;
}
.link-segment.natal-female-link {
  background: rgba(91, 77, 55, 0.34);
  opacity: 0.72;
}
.couple-node {
  position: absolute;
  z-index: 2;
  display: flex;
  align-items: flex-start;
  justify-content: center;
}
.couple-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: center;
  gap: 0;
}
.tree-person-hit {
  position: relative;
  z-index: 3;
  display: block;
  flex: 0 0 160rpx;
  width: 160rpx;
  height: 178rpx;
}
.spouse-join {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  padding: 0 6rpx;
}
.spouse-line {
  width: 12rpx;
  height: 4rpx;
  border-top: 2rpx dashed rgba(181, 122, 41, 0.72);
  background: transparent;
  opacity: 1;
}
.tree-footnote {
  padding: 8rpx 8rpx 4rpx;
}
.footnote {
  color: var(--archive-ink-soft, #657080);
  font-size: 22rpx;
  line-height: 1.5;
}
.unlocated-section {
  margin-top: 20rpx;
  padding: 18rpx 12rpx 8rpx;
  border-top: 1rpx dashed var(--archive-line-strong, rgba(92, 74, 48, 0.28));
}
.unlocated-head {
  margin-bottom: 12rpx;
}
.unlocated-title {
  display: block;
  color: var(--archive-ink, #243244);
  font-size: 26rpx;
  font-weight: 700;
}
.unlocated-desc {
  display: block;
  margin-top: 6rpx;
  color: var(--archive-ink-soft, #657080);
  font-size: 22rpx;
  line-height: 1.45;
}
.unlocated-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}
.unlocated-card-hit {
  position: relative;
  display: block;
}
.unlocated-card-hit.interactive::before {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  z-index: 3;
  width: 34rpx;
  height: 34rpx;
  border-radius: 0 0 0 16rpx;
  background: var(--archive-cinnabar, #a83b2d);
}
</style>
