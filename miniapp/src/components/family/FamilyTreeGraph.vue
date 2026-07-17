<template>
  <view class="family-tree-graph">
    <view v-if="!layout.success" class="tree-fallback">
      <MiniNotice tone="warm" title="结构生成提示">
        {{ layout.message || '家庭树暂时无法生成，可查看关系明细' }}
      </MiniNotice>
    </view>

    <template v-else>
      <view class="scroll-hint">
        <text class="hint-text">按父母、子女、配偶关系自然展开，可左右、上下滑动查看</text>
      </view>

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
              :class="[segment.orientation, { highlighted: link.highlighted }]"
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
                <FamilyTreePersonCard
                  :node="parent"
                  :is-self="parent.memberId === effectiveViewerMemberId"
                  :show-binding="showBinding"
                  :kinship-title="kinshipTitles[parent.memberId]"
                  :family-surname="familySurname"
                  :interactive="interactive"
                  :selectable="isNodeSelectable(parent)"
                  @select="emit('select', $event)"
                />
              </template>
            </view>
          </view>
        </view>
      </scroll-view>

      <view class="tree-footnote">
        <text class="footnote">
          共 {{ layout.memberCount }} 位成员，树中展示 {{ layout.renderedCount }} 位
          <text v-if="effectiveViewerMemberId"> · 深蓝纸签为当前中心视角</text>
        </text>
      </view>

      <view v-if="layout.unlocatedMembers.length > 0" class="unlocated-section">
        <view class="unlocated-head">
          <text class="unlocated-title">未定位成员</text>
          <text class="unlocated-desc">以下成员暂未接入当前关系图，可在「定位成员」中挂接。</text>
        </view>
        <view class="unlocated-list">
          <view
            v-for="node in layout.unlocatedMembers"
            :key="node.memberId"
            class="unlocated-card-hit"
            :class="{ interactive, selectable: isNodeSelectable(node) }"
            @tap.stop="handleUnlocatedSelect(node)"
          >
            <FamilyTreePersonCard
              :node="node"
              :is-self="node.memberId === effectiveViewerMemberId"
              :show-binding="showBinding"
              :kinship-title="kinshipTitles[node.memberId]"
              :family-surname="familySurname"
              :interactive="false"
              :selectable="isNodeSelectable(node)"
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
import { buildRelGraph, resolveLineageViewerId } from '@/features/family-tree/relativeGraph'
import { buildKinshipTitleMap } from '@/features/family-tree/relativeTitle'
import type { FamilyTreeResult } from '@/types/api'

const props = defineProps<{
  tree?: FamilyTreeResult | null
  viewerMemberId?: number | null
  showBinding?: boolean
  interactive?: boolean
  selectableDeceased?: boolean
}>()

const emit = defineEmits<{
  select: [node: FamilyTreeResult['nodes'][number]]
}>()

function handleUnlocatedSelect(node: FamilyTreeResult['nodes'][number]) {
  if (!isNodeSelectable(node)) return
  emit('select', node)
}

function isNodeSelectable(node: FamilyTreeResult['nodes'][number]) {
  return Boolean(props.interactive || (props.selectableDeceased && node.isLiving === false))
}

const effectiveViewerMemberId = computed(() => {
  if (props.viewerMemberId == null) return props.viewerMemberId
  const graph = buildRelGraph(props.tree?.nodes ?? [], props.tree?.edges ?? [])
  return resolveLineageViewerId(graph, props.viewerMemberId)
})

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
.scroll-hint {
  padding: 0 8rpx 10rpx;
}
.hint-text {
  color: var(--archive-ink-soft, #657080);
  font-size: 22rpx;
}
.tree-scroll {
  width: 100%;
}
.tree-canvas {
  position: relative;
  background-color: rgba(255, 249, 236, 0.28);
  background-image: linear-gradient(rgba(92, 74, 48, 0.024) 1rpx, transparent 1rpx);
  background-size: 100% 34rpx;
}
.links-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}
.link-segment {
  position: absolute;
  background: rgba(92, 74, 48, 0.58);
  opacity: 0.82;
  border-radius: 1rpx;
}
.link-segment.highlighted {
  opacity: 0.96;
  background: var(--archive-blue, #163353);
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
.spouse-join {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  padding: 0 4rpx;
}
.spouse-line {
  width: 14rpx;
  height: 2rpx;
  border-top: 2rpx dashed var(--archive-line-strong, rgba(92, 74, 48, 0.28));
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
.unlocated-card-hit.interactive::after {
  content: '管';
  position: absolute;
  top: 2rpx;
  right: 6rpx;
  z-index: 4;
  color: #fff;
  font-size: 16rpx;
  font-weight: 700;
  line-height: 1;
}
</style>
