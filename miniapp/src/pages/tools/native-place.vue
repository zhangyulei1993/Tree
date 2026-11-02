<template>
  <view class="archive-page native-place-page">
    <MiniBackHome />

    <view class="native-place-head archive-page-head">
      <view>
        <text class="archive-kicker">Tree Tools</text>
        <text class="archive-title">籍贯查询</text>
        <text class="archive-subtitle">快速核对籍贯的行政层级与地理分区</text>
      </view>
      <view class="archive-seal">籍</view>
    </view>

    <view class="archive-form-panel native-place-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">输入地名</text>
        <text class="archive-section-subtitle">支持省市区县，也支持“县 + 乡镇/街道”</text>
      </view>
      <view class="native-place-search">
        <input v-model.trim="keyword" class="tree-input" maxlength="80" placeholder="如：嘉鱼县鱼岳镇、浙江省杭州市西湖区" confirm-type="search" @confirm="queryPlace" />
        <MiniButton size="sm" :disabled="!keyword" @click="queryPlace">查询</MiniButton>
      </view>
      <view class="native-place-examples">
        <text v-for="item in examples" :key="item" class="native-place-example" @click="useExample(item)">{{ item }}</text>
      </view>
    </view>

    <view v-if="resolution" class="native-place-result archive-form-panel">
      <view class="native-place-result-head">
        <text class="native-place-result-kicker">查询结果</text>
        <text class="native-place-kind">{{ nativePlaceKindLabel(resolution.kind) }}</text>
      </view>
      <text class="native-place-path">{{ nativePlaceQueryText(resolution) }}</text>
      <view class="native-place-row">
        <text>籍贯保存值</text>
        <text>{{ resolution.nativePlace }}</text>
      </view>
      <view class="native-place-row">
        <text>头像风格分区</text>
        <text>{{ resolution.zone.label }}</text>
      </view>
      <MiniNotice tone="info" class="native-place-notice">
        {{ resultNote }}
      </MiniNotice>
      <MiniButton variant="secondary" size="sm" @click="copyResult">复制结果</MiniButton>
    </view>

    <view v-else-if="candidates.length" class="native-place-candidates archive-form-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">找到 {{ candidates.length }} 个可能地点</text>
        <text class="archive-section-subtitle">请选择准确的行政区划</text>
      </view>
      <view
        v-for="item in candidates"
        :key="`${item.regionCode}-${item.township}`"
        class="native-place-candidate"
        @click="selectCandidate(item)"
      >
        <text class="native-place-candidate-path">{{ nativePlaceQueryText(item) }}</text>
        <text class="native-place-candidate-action">选择</text>
      </view>
    </view>

    <MiniNotice v-else-if="queried" tone="warm" title="未找到可匹配地点">
      请检查地名用字是否准确，或补充省、市、县、镇/乡/街道等行政层级。例如“嘉鱼县鱼岳镇”。同名地名会展示候选供你选择。
    </MiniNotice>

    <view class="archive-form-panel native-place-rule-panel">
      <view class="archive-section-head">
        <text class="archive-section-title">保存规则</text>
      </view>
      <text class="native-place-rule">直辖市或特别行政区：地理分区 · 省/市 · 市辖区</text>
      <text class="native-place-rule">其他地区：地理分区 · 省 · 市 · 县/区</text>
      <text class="native-place-rule-muted">家庭资料以系统三级行政区选择结果为准，本工具仅用于快速查询与核对。</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'

import { ensureToolEnabled } from '@/api/toolConfigs'
import MiniBackHome from '@/components/base/MiniBackHome.vue'
import MiniButton from '@/components/base/MiniButton.vue'
import MiniNotice from '@/components/base/MiniNotice.vue'
import { nativePlaceKindLabel, nativePlaceQueryText, resolveNativePlaceKeyword, searchNativePlaceKeywords, type NativePlaceResolution } from '@/features/native-place/geoHierarchy'

const keyword = ref('')
const queried = ref(false)
const resolution = ref<NativePlaceResolution | null>(null)
const candidates = ref<NativePlaceResolution[]>([])
const examples = ['嘉鱼县鱼岳镇', '北京市朝阳区', '浙江省杭州市西湖区', '重庆市渝中区']

const resultNote = computed(() => {
  if (!resolution.value) return ''
  return resolution.value.kind === 'MUNICIPALITY_DISTRICT'
    ? '该地属于直辖市或特别行政区，保存时不重复写入同名城市层级。'
    : '该地保留省、市与区县层级，便于家庭资料展示和后续核对。'
})

onLoad(() => {
  void ensureToolEnabled('REGIONAL_DIVISION')
})

function queryPlace() {
  queried.value = true
  resolution.value = resolveNativePlaceKeyword(keyword.value)
  candidates.value = resolution.value ? [] : searchNativePlaceKeywords(keyword.value)
}

function useExample(value: string) {
  keyword.value = value
  queryPlace()
}

function selectCandidate(value: NativePlaceResolution) {
  resolution.value = value
  candidates.value = []
}

function copyResult() {
  if (!resolution.value) return
  uni.setClipboardData({
    data: `${nativePlaceQueryText(resolution.value)}\n籍贯：${resolution.value.nativePlace}`,
    success: () => uni.showToast({ title: '已复制', icon: 'success' })
  })
}
</script>

<style scoped>
.native-place-page { padding-top: 28rpx; }
.native-place-panel, .native-place-result, .native-place-rule-panel { margin-top: 24rpx; }
.native-place-search { display:flex; align-items:center; gap:14rpx; }
.native-place-search .tree-input { min-width:0; flex:1; }
.native-place-examples { display:flex; flex-wrap:wrap; gap:10rpx; margin-top:16rpx; }
.native-place-example { padding:7rpx 12rpx; border:1rpx solid var(--archive-line); border-radius:999rpx; color:var(--archive-blue); font-size:21rpx; line-height:1.35; }
.native-place-result-head { display:flex; align-items:center; justify-content:space-between; gap:16rpx; }
.native-place-candidate { display:flex; align-items:center; gap:16rpx; padding:18rpx 0; border-top:1rpx solid var(--archive-line); }
.native-place-candidate-path { flex:1; color:var(--archive-ink); font-size:24rpx; line-height:1.5; }
.native-place-candidate-action { flex:0 0 auto; color:var(--archive-cinnabar); font-size:23rpx; font-weight:700; }
.native-place-result-kicker { color:var(--archive-cinnabar); font-size:22rpx; font-weight:700; }
.native-place-kind { color:var(--archive-ink-soft); font-size:21rpx; }
.native-place-path { display:block; margin:18rpx 0; color:var(--archive-ink); font-size:31rpx; font-weight:700; line-height:1.45; }
.native-place-row { display:flex; justify-content:space-between; gap:20rpx; padding:13rpx 0; border-top:1rpx solid var(--archive-line); color:var(--archive-ink-soft); font-size:23rpx; }
.native-place-row text:last-child { color:var(--archive-ink); text-align:right; }
.native-place-notice { margin:18rpx 0; }
.native-place-rule { display:block; margin-top:12rpx; color:var(--archive-ink); font-size:24rpx; line-height:1.6; }
.native-place-rule-muted { display:block; margin-top:16rpx; color:var(--archive-ink-soft); font-size:21rpx; line-height:1.55; }
</style>
