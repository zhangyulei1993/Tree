<template>
  <view class="tree-branch">
    <view class="couple-block">
      <view class="couple-row">
        <template v-for="(parent, index) in branch.parents" :key="parent.memberId">
          <view v-if="index > 0" class="spouse-join">
            <view class="spouse-line" />
            <text class="spouse-mark">♥</text>
            <view class="spouse-line" />
          </view>
          <FamilyTreePersonCard
            :node="parent"
            :is-self="parent.memberId === currentMemberId"
            :show-binding="showBinding"
          />
        </template>
      </view>

      <view v-if="branch.childBranches.length > 0" class="descendant-zone">
        <view class="stem-down" />

        <view class="children-rail-wrap">
          <view
            v-if="branch.childBranches.length > 1"
            class="children-rail"
            :style="railWidthStyle"
          />
          <view class="children-row">
            <view
              v-for="child in branch.childBranches"
              :key="child.id"
              class="child-column"
            >
              <view class="stem-up" />
              <FamilyTreeBranch
                :branch="child"
                :current-member-id="currentMemberId"
                :show-binding="showBinding"
              />
            </view>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import FamilyTreePersonCard from '@/components/family/FamilyTreePersonCard.vue'
import type { FamilyTreeBranchNode } from '@/features/family-tree/types'

import FamilyTreeBranch from './FamilyTreeBranch.vue'

export default defineComponent({
  name: 'FamilyTreeBranch',
  components: {
    FamilyTreePersonCard,
    FamilyTreeBranch
  },
  props: {
    branch: {
      type: Object as PropType<FamilyTreeBranchNode>,
      required: true
    },
    currentMemberId: {
      type: Number as PropType<number | null>,
      default: null
    },
    showBinding: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    railWidthStyle(): Record<string, string> {
      const count = this.branch.childBranches.length
      if (count <= 1) return {}
      const unit = 220
      const gap = 24
      const width = count * unit + (count - 1) * gap
      return {
        width: `${width}rpx`
      }
    }
  }
})
</script>

<style scoped>
.tree-branch {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
}
.couple-block {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.couple-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  justify-content: center;
  gap: 8rpx;
}
.spouse-join {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  gap: 4rpx;
  padding: 0 2rpx;
}
.spouse-line {
  width: 16rpx;
  height: 2rpx;
  background: var(--tree-border-subtle, #d8dee6);
}
.spouse-mark {
  color: var(--tree-text-weak);
  font-size: 18rpx;
  line-height: 1;
}
.descendant-zone {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}
.stem-down {
  width: 2rpx;
  height: 28rpx;
  background: var(--tree-green, #2f6b57);
  opacity: 0.45;
}
.children-rail-wrap {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.children-rail {
  position: absolute;
  top: 0;
  left: 50%;
  height: 2rpx;
  background: var(--tree-green, #2f6b57);
  opacity: 0.45;
  transform: translateX(-50%);
}
.children-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: flex-start;
  gap: 24rpx;
  padding-top: 0;
}
.child-column {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}
.stem-up {
  width: 2rpx;
  height: 24rpx;
  background: var(--tree-green, #2f6b57);
  opacity: 0.45;
}
</style>
