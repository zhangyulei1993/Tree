<template>
  <view class="archive-page retired-join-page">
    <MiniBackHome />
    <view class="retired-join-panel archive-form-panel">
      <MiniEmptyState
        symbol="家"
        title="家庭成员通过邀请加入"
        description="优质家庭用于浏览与学习。需要加入某个家庭时，请联系家庭创建者或管理员为对应节点发出邀请。"
        action-text="查看优质家庭"
        @action="openShowcase"
      />
    </view>
  </view>
</template>

<script setup lang="ts">
import { onLoad } from '@dcloudio/uni-app'

import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniEmptyState from '@/components/base/MiniEmptyState.vue'

function openShowcase() {
  uni.redirectTo({ url: '/pages/family/showcase' })
}

onLoad(() => {
  uni.showModal({
    title: '当前不支持公开申请加入家庭',
    content: '家庭成员需由家庭创建者或管理员发出节点邀请后绑定。优质家庭仅供浏览与学习，不开放陌生人申请加入。',
    showCancel: false
  })
})
</script>

<style scoped>
.retired-join-page {
  padding-top: 28rpx;
}

.retired-join-page :deep(.mini-back-home) {
  margin-bottom: 22rpx;
}

.retired-join-panel {
  min-height: 430rpx;
  display: flex;
  align-items: center;
}
</style>
