<template>
  <view class="family-region-selector">
    <text class="tree-field-label">地理分区</text>
    <picker :range="zoneLabels" :value="zoneIndex" @change="onZoneChange">
      <view class="region-picker-field">
        <text :class="{ placeholder: !selectedZone }">{{ selectedZone?.label || '请选择七大地理分区' }}</text>
        <text class="region-picker-arrow">›</text>
      </view>
    </picker>

    <text class="tree-field-label">省份</text>
    <picker :range="provinceOptions" :disabled="!selectedZone" :value="provinceIndex" @change="onProvinceChange">
      <view class="region-picker-field" :class="{ disabled: !selectedZone }">
        <text :class="{ placeholder: !selectedProvince }">{{ selectedProvince || '请从当前分区选择省份' }}</text>
        <text class="region-picker-arrow">›</text>
      </view>
    </picker>

    <text class="tree-field-label">市与区县</text>
    <picker mode="region" level="region" :disabled="!selectedProvince" :value="pickerValue" @change="onRegionChange">
      <view class="region-picker-field" :class="{ disabled: !selectedProvince }">
        <text :class="{ placeholder: !selectedProvince }">{{ selectedProvince ? placeLabel : '请先选择省份，再选择市与区县' }}</text>
        <text class="region-picker-arrow">›</text>
      </view>
    </picker>

    <text class="region-picker-hint">直辖市保存为“省 · 市辖区”；其他地区保存为“省 · 市 · 县/区”。地理分区将决定家庭树头像风格。</text>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { geographicZones, resolveNativePlace, type NativePlaceResolution } from '@/features/native-place/geoHierarchy'

export interface FamilyRegionValue {
  nativePlace: string
  regionCode: string
  regionText: string
}

const zones = geographicZones

const props = defineProps<{
  modelValue?: Partial<FamilyRegionValue>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: FamilyRegionValue]
}>()

const selectedZoneCode = ref('')
const selectedProvince = ref('')
const selectedCity = ref('')
const selectedDistrict = ref('')
const pickerValue = ref<string[]>([])
const zoneLabels = zones.map((item) => item.label)
const selectedZone = computed(() => zones.find((item) => item.code === selectedZoneCode.value))
const zoneIndex = computed(() => Math.max(0, zones.findIndex((item) => item.code === selectedZoneCode.value)))
const provinceOptions = computed(() => selectedZone.value?.provinces || [])
const provinceIndex = computed(() => Math.max(0, provinceOptions.value.indexOf(selectedProvince.value)))
const currentResolution = computed<NativePlaceResolution | null>(() => resolveNativePlace({
  province: selectedProvince.value,
  city: selectedCity.value,
  district: selectedDistrict.value
}))
const placeLabel = computed(() => currentResolution.value?.regionText.replace(/^.+? · /, '') || selectedProvince.value)

function findZoneByProvince(province: string) {
  return zones.find((item) => item.provinces.includes(province))
}

function parseExistingValue(value?: Partial<FamilyRegionValue>) {
  const code = value?.regionCode || ''
  const text = value?.regionText || ''
  const nativePlace = value?.nativePlace || ''
  const zoneFromCode = zones.find((item) => code.startsWith(`CN-${item.code}-`))
  const zone = zoneFromCode || zones.find((item) => text.includes(item.label))
  const codeParts = zoneFromCode ? code.slice(`CN-${zoneFromCode.code}-`.length).split('-') : []
  const province = codeParts[0] || zones.flatMap((item) => item.provinces).find((item) => text.includes(item) || nativePlace.includes(item)) || ''
  const city = codeParts[1] || (province ? nativePlace.replace(province, '').trim() : '')
  const district = codeParts[2] || ''

  selectedZoneCode.value = zone?.code || findZoneByProvince(province)?.code || ''
  selectedProvince.value = province
  selectedCity.value = city
  selectedDistrict.value = district
  pickerValue.value = province ? [province, city, district] : []
}

function emitValue() {
  const zone = selectedZone.value
  const resolution = currentResolution.value
  if (!zone || !resolution) return
  emit('update:modelValue', {
    nativePlace: resolution.nativePlace,
    regionCode: resolution.regionCode,
    regionText: resolution.regionText
  })
}

function clearValue() {
  emit('update:modelValue', {
    nativePlace: '',
    regionCode: '',
    regionText: ''
  })
}

function onZoneChange(event: { detail: { value: string | number } }) {
  const zone = zones[Number(event.detail.value)]
  if (!zone) return
  selectedZoneCode.value = zone.code
  selectedProvince.value = ''
  selectedCity.value = ''
  selectedDistrict.value = ''
  pickerValue.value = []
  clearValue()
}

function onProvinceChange(event: { detail: { value: string | number } }) {
  const province = provinceOptions.value[Number(event.detail.value)]
  if (!province) return
  selectedProvince.value = province
  selectedCity.value = ''
  selectedDistrict.value = ''
  pickerValue.value = [province, '', '']
  clearValue()
}

function onRegionChange(event: { detail: { value: string[] } }) {
  const [province = '', city = '', district = ''] = event.detail.value || []
  const provinceZone = findZoneByProvince(province)
  if (!provinceZone || provinceZone.code !== selectedZoneCode.value || province !== selectedProvince.value) {
    uni.showToast({ title: `请在${selectedProvince.value || selectedZone.value?.label || '当前分区'}内选择市区县`, icon: 'none' })
    return
  }
  selectedProvince.value = province
  selectedCity.value = city
  selectedDistrict.value = district
  pickerValue.value = [province, city, district]
  emitValue()
}

watch(() => props.modelValue, parseExistingValue, { immediate: true, deep: true })
</script>

<style scoped>
.family-region-selector { margin-top: 18rpx; }
.region-picker-field { display:flex; align-items:center; justify-content:space-between; min-height:76rpx; box-sizing:border-box; padding:0 18rpx; border:1rpx solid var(--archive-line); border-radius:8rpx; color:var(--archive-ink); background:rgba(255,252,244,.7); font-size:25rpx; }
.region-picker-field:active { background:var(--archive-paper-deep); }
.region-picker-field.disabled { opacity:.64; }
.region-picker-field .placeholder { color:var(--archive-ink-soft); }
.region-picker-arrow { color:var(--archive-cinnabar); font-size:34rpx; line-height:1; }
.region-picker-hint { display:block; margin-top:8rpx; color:var(--archive-ink-soft); font-size:20rpx; line-height:1.45; }
</style>
