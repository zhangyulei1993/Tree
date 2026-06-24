<template>
  <view class="family-tree-graph">
    <view v-if="!layout.success" class="tree-fallback">
      <MiniNotice tone="warm" title="结构生成提示">
        {{ layout.message || '家谱结构暂时无法生成，可查看关系明细' }}
      </MiniNotice>
    </view>

    <template v-else>
      <view class="scroll-hint">
        <text class="hint-text">可左右、上下滑动查看完整家谱</text>
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
                  :is-self="parent.memberId === viewerMemberId"
                  :show-binding="showBinding"
                  :kinship-title="kinshipTitles[parent.memberId]"
                  :family-surname="familySurname"
                />
              </template>
            </view>
          </view>
        </view>
      </scroll-view>

      <view class="tree-footnote">
        <text class="footnote">
          共 {{ layout.memberCount }} 位成员，树中展示 {{ layout.renderedCount }} 位
          <text v-if="graphVersion"> · 家谱版本 {{ graphVersion }}</text>
        </text>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import MiniNotice from '@/components/base/MiniNotice.vue'
import FamilyTreePersonCard from '@/components/family/FamilyTreePersonCard.vue'
import { layoutFamilyTree } from '@/features/family-tree/layoutFamilyTree'
import { buildKinshipTitleMap } from '@/features/family-tree/relativeTitle'
import type { FamilyTreeResult } from '@/types/api'

const props = defineProps<{
  tree?: FamilyTreeResult | null
  viewerMemberId?: number | null
  showBinding?: boolean
}>()

const layout = computed(() =>
  layoutFamilyTree({
    nodes: props.tree?.nodes || [],
    edges: props.tree?.edges || [],
    tree: props.tree?.tree || [],
    viewerMemberId: props.viewerMemberId
  })
)

/** 视角相关称谓：viewerMemberId 或 graphVersion 变化时全量重算 */
const kinshipTitles = computed(() =>
  buildKinshipTitleMap({
    viewerMemberId: props.viewerMemberId,
    nodes: props.tree?.nodes ?? [],
    edges: props.tree?.edges ?? [],
    graphVersion: props.tree?.graphVersion ?? 0
  })
)

const graphVersion = computed(() => props.tree?.graphVersion || null)

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
  color: var(--tree-text-weak);
  font-size: 22rpx;
}
.tree-scroll {
  width: 100%;
}
.tree-canvas {
  position: relative;
}
.links-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}
.link-segment {
  position: absolute;
  background: #285c4a;
  opacity: 0.62;
  border-radius: 1rpx;
}
.link-segment.highlighted {
  opacity: 0.96;
  background: var(--tree-green-dark, #1f4f40);
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
  background: #285c4a;
  opacity: 0.72;
}
.tree-footnote {
  padding: 8rpx 8rpx 4rpx;
}
.footnote {
  color: var(--tree-text-weak);
  font-size: 22rpx;
  line-height: 1.5;
}
</style>
