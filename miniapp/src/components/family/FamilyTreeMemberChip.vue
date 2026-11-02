<template>
  <view
    class="tree-member-chip"
    :class="[
      livingClass,
      nodeKindClass,
      {
        self: isSelf,
        compact,
        interactive,
        selectable,
        'in-focus': inFocus === true,
        'out-of-focus': inFocus === false,
        root: isRoot,
        'portrait-imperial': portraitTheme === 'imperial',
        'portrait-regional': portraitTheme === 'regional',
        [`age-${avatarAgeBand}`]: true,
        [`regional-${regionalPortraitStyle}`]: portraitTheme === 'regional',
        'portrait-glasses': node.portraitGlassesStyle === 'ON',
        [`lineage-${lineageRole || 'default'}`]: Boolean(lineageRole)
      }
    ]"
    @tap="handleSelect"
  >
    <text v-if="node.generationCharacter" class="generation-mark">{{ node.generationCharacter }}</text>
    <text v-if="isRoot" class="root-mark">源</text>
    <view v-if="healthHints?.length" class="health-dot"><text>!</text></view>
    <text v-if="spouseRelationLabel" class="spouse-relation-tag">{{ spouseRelationLabel }}</text>
    <view
      class="portrait-frame"
      :class="[
        { 'portrait-deceased': isDeceased },
        `portrait-appearance-${portraitAppearanceVariant}`
      ]"
    >
      <image
        :key="glassesAvatarUrl || displayAvatarUrl"
        class="portrait"
        :src="glassesAvatarUrl || displayAvatarUrl"
        mode="aspectFit"
        @error="handleAvatarError"
      />
    </view>
    <view
      class="member-base"
      :class="{ 'member-bound': showBinding && bindingText === '已绑定' && !isDeceased }"
    >
      <view class="chip-head">
        <view class="name-wrap">
          <text v-if="shouldEmphasizeSurname" class="name-emphasis">
            <text class="surname-seal">{{ surnamePart.slice(0, 1) }}</text>
            <text class="given-name" :class="genderNameClass">{{ surnamePart.slice(1) }}{{ givenPart }}</text>
          </text>
          <text v-else class="plain-name" :class="genderNameClass">{{ displayName }}</text>
          <text v-if="isSelf" class="self-tag">我</text>
        </view>
      </view>
      <text v-if="memberSubtitle" class="member-subtitle">{{ memberSubtitle }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { resolveAvatarAgeBand, type AvatarAgeBand } from '@/features/family-tree/avatarAgeBand'
import {
  resolvePortraitAppearanceVariant,
  resolveFemalePortraitVariant,
  resolveRegionalPortraitAsset,
  resolveRegionalPortraitGlassesAsset,
  resolveSelectedPortraitAsset,
  resolvePortraitAvatarStyle,
  portraitAssetBaseURL,
  resolveRegionalPortraitTheme,
  type RegionalPortraitTheme
} from '@/features/family-tree/portraitCatalog'
import type { TreeNode } from '@/types/api'
import { formatTreeBindingState } from '@/utils/memberFormat'

const props = defineProps<{
  node: TreeNode
  isSelf?: boolean
  showBinding?: boolean
  kinshipTitle?: string
  compact?: boolean
  familySurname?: string
  interactive?: boolean
  selectable?: boolean
  inFocus?: boolean
  isRoot?: boolean
  lineageRole?: string
  spouseRelationLabel?: string
  healthHints?: string[]
  portraitTheme?: 'default' | 'imperial' | 'regional'
  portraitRegion?: string | null
  portraitRegionCode?: string | null
  portraitAppearanceVariant?: number
}>()

const emit = defineEmits<{
  select: [node: TreeNode]
}>()

function handleSelect() {
  if (props.interactive || props.selectable) emit('select', props.node)
}

const displayName = computed(() => props.node.displayName.trim() || '未命名')
const identityText = computed(() => props.node.identityText?.trim() || '')
const memberSubtitle = computed(() => identityText.value || (props.isSelf ? '' : props.kinshipTitle?.trim() || ''))
const surnamePart = computed(() => {
  const surname = (props.node.surname || displayName.value.slice(0, 1)).trim()
  if (!surname || !displayName.value.startsWith(surname)) return ''
  return surname
})
const givenPart = computed(() => displayName.value.slice(surnamePart.value.length))
const shouldEmphasizeSurname = computed(() => {
  const familySurname = props.familySurname?.trim()
  return Boolean(familySurname && surnamePart.value === familySurname)
})
const genderNameClass = computed(() => {
  if (props.node.gender === 'MALE') return 'name-male'
  if (props.node.gender === 'FEMALE') return 'name-female'
  return 'name-unknown'
})
const bindingText = computed(() =>
  props.showBinding ? formatTreeBindingState(props.node.userBindingState) : ''
)
const isDeceased = computed(() => props.node.isLiving === false)
const regionalPortraitStyle = computed<RegionalPortraitTheme>(() =>
  resolveRegionalPortraitTheme(props.portraitRegionCode, props.portraitRegion)
)
const avatarAgeBand = computed<AvatarAgeBand>(() => resolveAvatarAgeBand(
  props.node.birthDate,
  isDeceased.value ? props.node.deathDate : undefined
))
const portraitAppearanceVariant = computed(() =>
  props.portraitAppearanceVariant || resolvePortraitAppearanceVariant(props.node.memberId)
)
const avatarFailed = ref(false)

const avatarUrl = computed(() => {
  const band = avatarAgeBand.value
  if (props.portraitTheme === 'imperial') {
    const imperialBand = band === 'toddler' || band === 'child'
      ? 'child'
      : band === 'teen' || band === 'youth'
        ? 'youth'
        : band === 'adult' || band === 'mature'
          ? 'mature'
          : 'elder'
    if (props.node.gender === 'MALE') return `${portraitAssetBaseURL}/imperial/imperial-male-${imperialBand}.png`
    if (props.node.gender === 'FEMALE') return `${portraitAssetBaseURL}/imperial/imperial-female-${imperialBand}.png`
  }
  if (props.node.gender === 'MALE' || props.node.gender === 'FEMALE') {
    const selected = resolveSelectedPortraitAsset({
      gender: props.node.gender === 'MALE' ? 'male' : 'female',
      ageBand: band,
      style: props.node.portraitAvatarStyle
    })
    if (selected) return selected
  }
  if (props.portraitTheme === 'regional') {
    if (props.node.gender === 'MALE') return resolveRegionalPortraitAsset({
      theme: regionalPortraitStyle.value,
      gender: 'male',
      ageBand: band,
      memberId: props.node.memberId
    })
    if (props.node.gender === 'FEMALE') return resolveRegionalPortraitAsset({
      theme: regionalPortraitStyle.value,
      gender: 'female',
      ageBand: band,
      memberId: props.node.memberId,
      variant: resolveFemalePortraitVariant(props.node.memberId, props.node.portraitHairStyle, band)
    })
  }
  // Public and quality-family trees have no family hometown, so use the north default.
  if (props.node.gender === 'MALE') return resolveRegionalPortraitAsset({
    theme: 'north', gender: 'male', ageBand: band, memberId: props.node.memberId
  })
  if (props.node.gender === 'FEMALE') return resolveRegionalPortraitAsset({
    theme: 'north', gender: 'female', ageBand: band, memberId: props.node.memberId,
    variant: resolveFemalePortraitVariant(props.node.memberId, props.node.portraitHairStyle, band)
  })
  return '/static/tree/avatars/avatar-neutral.png'
})
const displayAvatarUrl = computed(() =>
  avatarFailed.value ? '/static/tree/avatars/avatar-neutral.png' : avatarUrl.value
)
const glassesAvatarUrl = computed(() => {
  // The historical showcase is a curated royal illustration set. It must not
  // inherit personal portrait preferences from ordinary family members.
  if (props.portraitTheme === 'imperial') return ''
  if (resolvePortraitAvatarStyle(props.node.portraitAvatarStyle)) return ''
  if (props.node.portraitGlassesStyle !== 'ON') return ''
  const band = avatarAgeBand.value
  const variant = props.node.gender === 'FEMALE'
    ? resolveFemalePortraitVariant(props.node.memberId, props.node.portraitHairStyle, band)
    : undefined
  const theme = props.portraitTheme === 'regional' ? regionalPortraitStyle.value : 'north'
  if (props.node.gender === 'MALE') return resolveRegionalPortraitGlassesAsset({
    theme, gender: 'male', ageBand: band, memberId: props.node.memberId
  })
  if (props.node.gender === 'FEMALE') return resolveRegionalPortraitGlassesAsset({
    theme, gender: 'female', ageBand: band, memberId: props.node.memberId, variant
  })
  return ''
})

// A node can stay mounted while its family region or its birth information changes.
// Retry the newly resolved portrait instead of preserving an earlier failed fallback.
watch(avatarUrl, () => {
  avatarFailed.value = false
})

function handleAvatarError() {
  avatarFailed.value = true
}
const livingClass = computed(() => {
  if (props.node.isLiving === true) return 'living-active'
  if (isDeceased.value) return 'living-deceased'
  return 'living-unknown'
})
const nodeKindClass = computed(() => {
  if (props.node.memberType === 'SPOUSE') return 'spouse-tag'
  if (props.node.memberType === 'EXTERNAL_MEMBER') return 'external-tag'
  if (props.node.memberType === 'LINEAGE_MEMBER') return 'lineage-tag'
  return 'lineage-tag'
})
</script>

<style scoped>
.tree-member-chip {
  position: relative;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  width: 144rpx;
  height: 144rpx;
  min-width: 144rpx;
  max-width: 144rpx;
  overflow: hidden;
  border: 1rpx solid var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  border-radius: 4rpx;
  background-color: #fff7e8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.028) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
  padding: 10rpx 9rpx 8rpx;
  box-shadow: none;
}
.health-dot {
  position: absolute;
  top: 6rpx;
  right: 6rpx;
  z-index: 3;
  display: grid;
  width: 20rpx;
  height: 20rpx;
  place-items: center;
  border: 2rpx solid #fff7e8;
  border-radius: 50%;
  background: var(--archive-cinnabar, #a83b2d);
  color: #fff;
  font-size: 15rpx;
  font-weight: 800;
  line-height: 1;
}
.spouse-relation-tag {
  position: absolute;
  top: 8rpx;
  left: 2rpx;
  z-index: 4;
  min-width: 28rpx;
  padding: 2rpx 4rpx;
  border: 1rpx solid rgba(168, 59, 45, 0.48);
  background: rgba(255, 247, 232, 0.96);
  color: var(--archive-cinnabar, #a83b2d);
  font-size: 15rpx;
  font-weight: 700;
  line-height: 1.2;
  text-align: center;
}
.tree-member-chip.self .health-dot {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.self {
  z-index: 6;
  border: 2rpx solid var(--archive-blue, #163353);
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.08), transparent),
    var(--archive-blue, #163353);
  box-shadow: 0 8rpx 16rpx rgba(22, 51, 83, 0.18);
}

/* Keep the current viewpoint identifiable without changing the node size. */
.tree-member-chip.self.compact {
  overflow: visible;
  border: 4rpx solid #fff7e8;
  border-radius: 50%;
  background:
    radial-gradient(circle at 50% 34%, rgba(255, 255, 255, 0.14), transparent 42%),
    var(--archive-blue, #163353);
  box-shadow:
    0 0 0 2rpx rgba(200, 164, 93, 0.95),
    0 8rpx 18rpx rgba(22, 51, 83, 0.24);
  transform: none;
}

.tree-member-chip.self.compact::before {
  position: absolute;
  inset: -10rpx;
  border: 1rpx solid rgba(200, 164, 93, 0.72);
  border-radius: 50%;
  content: '';
  pointer-events: none;
}

.tree-member-chip.self.compact::after {
  position: absolute;
  top: -18rpx;
  left: 50%;
  width: 22rpx;
  height: 22rpx;
  border: 2rpx solid #fff7e8;
  border-radius: 50%;
  background: var(--archive-cinnabar, #a83b2d);
  content: '';
  transform: translateX(-50%);
}

.tree-member-chip.in-focus:not(.self) {
  border-color: transparent;
  box-shadow: none;
}

.tree-member-chip.out-of-focus {
  opacity: 1;
  filter: none;
}
.tree-member-chip.interactive {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  box-shadow: inset 0 0 0 2rpx rgba(255, 255, 255, 0.34);
}
.tree-member-chip.selectable {
  border-color: rgba(168, 59, 45, 0.42);
}

.tree-member-chip.in-focus.interactive:not(.self) {
  border-color: transparent;
  box-shadow: none;
}

.tree-member-chip.out-of-focus.interactive {
  border-color: transparent;
  box-shadow: none;
}
.tree-member-chip.compact {
  width: 160rpx;
  height: 178rpx;
  min-width: 160rpx;
  max-width: 160rpx;
  padding: 10rpx 10rpx 8rpx;
  border-radius: 10rpx;
}
.portrait-frame {
  position: relative;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 82rpx;
  height: 82rpx;
  overflow: hidden;
  border: 2rpx solid rgba(181, 122, 41, 0.54);
  border-radius: 50%;
  background: rgba(255, 248, 234, 0.74);
}

.portrait {
  width: 100%;
  height: 100%;
}
.portrait-deceased {
  filter: grayscale(0.7);
  opacity: 0.72;
}
.generation-mark {
  position: absolute;
  top: 50%;
  left: 7rpx;
  color: rgba(22, 51, 83, 0.38);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 18rpx;
  font-weight: 700;
  line-height: 1;
  transform: translateY(-50%);
}
.root-mark {
  position: absolute;
  top: -11rpx;
  right: -10rpx;
  z-index: 4;
  display: grid;
  width: 30rpx;
  height: 30rpx;
  place-items: center;
  border: 1rpx solid rgba(168, 59, 45, 0.58);
  border-radius: 50%;
  background: #fff7e8;
  color: var(--archive-cinnabar, #a83b2d);
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 17rpx;
  font-weight: 800;
  line-height: 1;
}
.tree-member-chip.root .health-dot {
  top: 19rpx;
}
.tree-member-chip.compact .plain-name,
.tree-member-chip.compact .name-emphasis,
.tree-member-chip.compact .given-name {
  font-size: 23rpx;
  line-height: 1.25;
}

.tree-member-chip.self.compact .plain-name,
.tree-member-chip.self.compact .name-emphasis,
.tree-member-chip.self.compact .given-name {
  font-size: 27rpx;
  font-weight: 700;
}

.tree-member-chip.self.compact .portrait-frame {
  border-color: rgba(255, 247, 232, 0.84);
  background: rgba(255, 255, 255, 0.12);
}
.tree-member-chip.living-active {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
}
.tree-member-chip.living-deceased {
  border-color: rgba(101, 112, 128, 0.32);
  background-color: #eee7d8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.026) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
}
.tree-member-chip.living-unknown {
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
}
.tree-member-chip.self.living-active {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.self.living-deceased {
  border-color: var(--archive-blue, #163353);
}
.tree-member-chip.self.living-unknown {
  border-color: var(--archive-blue, #163353);
}

.tree-member-chip.self.compact .self-tag {
  background: var(--archive-cinnabar, #a83b2d);
  font-size: 16rpx;
}
.tree-member-chip.lineage-tag {
  border-style: solid;
  border-color: var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  background-color: #fff7e8;
  background-image: linear-gradient(rgba(92, 74, 48, 0.028) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
}
.tree-member-chip.lineage-tag.lineage-PATRILINEAL {
  background-color: #fff8e9;
}
.tree-member-chip.lineage-tag.lineage-NATAL_FEMALE {
  background-color: #fff5e8;
}
.tree-member-chip.lineage-tag.lineage-EXTERNAL {
  border-style: dashed;
  border-color: rgba(101, 112, 128, 0.36);
  background-color: #f2eadc;
}
.tree-member-chip.spouse-tag {
  border-color: rgba(181, 122, 41, 0.52);
  border-style: dashed;
  background-color: #fcf2de;
  background-image: linear-gradient(rgba(92, 74, 48, 0.022) 1rpx, transparent 1rpx);
  background-size: 100% 24rpx;
  opacity: 0.95;
}
.tree-member-chip.spouse-tag .portrait-frame {
  border-color: rgba(168, 59, 45, 0.72);
  box-shadow: 0 0 0 3rpx rgba(168, 59, 45, 0.12);
}
.tree-member-chip.spouse-tag .portrait-frame::after {
  position: absolute;
  right: -7rpx;
  bottom: -4rpx;
  display: grid;
  width: 25rpx;
  height: 25rpx;
  place-items: center;
  border: 1rpx solid rgba(168, 59, 45, 0.72);
  border-radius: 50%;
  background: #fff7e8;
  color: var(--archive-cinnabar, #a83b2d);
  content: '婚';
  font-family: 'Songti SC', 'STSong', serif;
  font-size: 14rpx;
  font-weight: 700;
}
.tree-member-chip.external-tag {
  background: #f2eadc;
}
.chip-head {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  margin-top: 6rpx;
}
.name-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5rpx;
  width: 100%;
  min-width: 0;
}
.name-emphasis,
.plain-name {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4rpx;
  max-width: 100%;
  min-width: 0;
}
.plain-name {
  display: block;
  overflow: hidden;
  font-size: 26rpx;
  font-weight: 500;
  line-height: 1.35;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.plain-name.name-male,
.given-name.name-male {
  color: var(--archive-blue, #163353);
}
.plain-name.name-female,
.given-name.name-female {
  color: var(--archive-cinnabar, #a83b2d);
}
.plain-name.name-unknown,
.given-name.name-unknown {
  color: var(--archive-ink-soft, #657080);
}
.living-deceased .plain-name,
.living-deceased .given-name {
  opacity: 0.62;
}
.surname-seal {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 28rpx;
  height: 28rpx;
  border: 1rpx solid var(--archive-line-strong, rgba(92, 74, 48, 0.28));
  background: rgba(255, 248, 234, 0.88);
  color: var(--archive-ink, #243244);
  font-size: 18rpx;
  font-weight: 800;
  line-height: 1;
}
.given-name {
  overflow: hidden;
  min-width: 0;
  font-size: 26rpx;
  font-weight: 500;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tree-member-chip.compact .surname-seal {
  width: 24rpx;
  height: 24rpx;
  font-size: 18rpx;
}
.tree-member-chip.compact .given-name {
  font-size: 21rpx;
}
.self-tag {
  flex-shrink: 0;
  border-radius: 999rpx;
  background: var(--archive-cinnabar, #a83b2d);
  color: #fff;
  font-size: 18rpx;
  line-height: 1;
  padding: 3rpx 8rpx;
}
.kinship-title {
  display: block;
  overflow: hidden;
  width: 100%;
  color: var(--archive-blue, #163353);
  font-size: 18rpx;
  font-weight: 500;
  line-height: 1.3;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.identity-tag {
  display: block;
  overflow: hidden;
  max-width: 100%;
  margin-bottom: 2rpx;
  color: var(--archive-cinnabar, #a83b2d);
  font-size: 17rpx;
  font-weight: 600;
  line-height: 1.25;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.living-deceased .kinship-title {
  color: #6b7280;
}
.living-deceased .identity-tag {
  color: #7a6a5f;
}
.chip-detail {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-width: 0;
  margin-top: 3rpx;
  padding-right: 0;
}
.tree-member-chip.self .plain-name,
.tree-member-chip.self .given-name,
.tree-member-chip.self .surname-seal,
.tree-member-chip.self .kinship-title {
  color: rgba(255, 255, 255, 0.92);
}
.tree-member-chip.self .surname-seal {
  border-color: rgba(255, 255, 255, 0.42);
  background: rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.95);
}
.tree-member-chip.self .identity-tag {
  color: rgba(255, 255, 255, 0.74);
}
/* Portrait and nameplate layout: the node is no longer a surrounding card. */
.tree-member-chip,
.tree-member-chip.lineage-tag,
.tree-member-chip.external-tag,
.tree-member-chip.spouse-tag,
.tree-member-chip.living-active,
.tree-member-chip.living-deceased,
.tree-member-chip.living-unknown,
.tree-member-chip.self,
.tree-member-chip.self.living-active,
.tree-member-chip.self.living-deceased,
.tree-member-chip.self.living-unknown {
  overflow: visible;
  border: 0;
  background: transparent;
  box-shadow: none;
}

.tree-member-chip.compact,
.tree-member-chip.self.compact {
  width: 160rpx;
  height: 160rpx;
  min-width: 160rpx;
  max-width: 160rpx;
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.tree-member-chip.self.compact::before,
.tree-member-chip.self.compact::after {
  display: none;
}

.portrait-frame {
  z-index: 2;
  width: 88rpx;
  height: 88rpx;
  border-width: 2rpx;
  background: rgba(255, 248, 234, 0.86);
  box-shadow: 0 4rpx 9rpx rgba(58, 47, 31, 0.1);
}

.member-base {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 152rpx;
  min-height: 48rpx;
  margin-top: -3rpx;
  padding: 5rpx 7rpx 9rpx;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.member-base::before,
.member-base::after {
  position: absolute;
  right: auto;
  left: 50%;
  width: 82rpx;
  height: 1rpx;
  background: rgba(181, 122, 41, 0.54);
  content: '';
  transform: translateX(-50%);
}

.member-base::before { top: 0; }
.member-base::after {
  bottom: 2rpx;
  width: 44rpx;
  background: rgba(92, 74, 48, 0.34);
}

.tree-member-chip.lineage-PATRILINEAL .member-base {
  box-shadow: inset 0 2rpx 0 var(--archive-blue, #163353);
}

.tree-member-chip.lineage-NATAL_FEMALE .member-base {
  box-shadow: inset 0 2rpx 0 rgba(168, 59, 45, 0.72);
}

.tree-member-chip.lineage-EXTERNAL .member-base {
  opacity: 0.72;
}

.tree-member-chip.spouse-tag .member-base {
  box-shadow: inset 0 2rpx 0 rgba(168, 59, 45, 0.48);
}

.tree-member-chip.self .portrait-frame {
  border: 3rpx solid var(--archive-blue, #163353);
  box-shadow: 0 0 0 3rpx rgba(200, 164, 93, 0.7);
}

.tree-member-chip.self .member-base {
  background: transparent;
  box-shadow: inset 0 2rpx 0 var(--archive-blue, #163353);
}

.tree-member-chip.self .member-base::before,
.tree-member-chip.self .member-base::after {
  background: var(--archive-blue, #163353);
}

.tree-member-chip.self .chip-detail { padding-right: 0; }
.tree-member-chip.self .self-tag { background: var(--archive-cinnabar, #a83b2d); }
.tree-member-chip.self .plain-name,
.tree-member-chip.self .given-name,
.tree-member-chip.self .surname-seal,
.tree-member-chip.self .kinship-title { color: var(--archive-blue, #163353); }
.tree-member-chip.self .surname-seal {
  border-color: rgba(22, 51, 83, 0.4);
  background: transparent;
}
.tree-member-chip.self .identity-tag { color: var(--archive-cinnabar, #a83b2d); }

.chip-head { margin-top: 0; }
.chip-detail { margin-top: 1rpx; }
.tree-member-chip.compact .plain-name,
.tree-member-chip.compact .name-emphasis,
.tree-member-chip.compact .given-name { font-size: 22rpx; }
.tree-member-chip.self.compact .plain-name,
.tree-member-chip.self.compact .name-emphasis,
.tree-member-chip.self.compact .given-name { font-size: 22rpx; }

.generation-mark {
  top: 50rpx;
  left: 11rpx;
  z-index: 3;
}

.surname-seal {
  border-radius: 50%;
  background: transparent;
}

.tree-member-chip.living-deceased .member-base {
  background: rgba(238, 231, 216, 0.94);
}

/* A portrait, floating type and a short pedestal: no enclosing card. */
.tree-member-chip.compact,
.tree-member-chip.self.compact {
  height: 156rpx;
}

.portrait-frame {
  width: 94rpx;
  height: 94rpx;
  border: 2rpx solid rgba(126, 88, 42, 0.68);
  background: #efe0bd;
  box-shadow: 0 2rpx 5rpx rgba(58, 47, 31, 0.1);
}

/* Four illustration bases, each with three restrained textile treatments.
 * This keeps a large peer group visually distinct without individualised faces. */
.portrait-appearance-1,
.portrait-appearance-5,
.portrait-appearance-9 { border-color: rgba(126, 88, 42, 0.68); }
.portrait-appearance-2,
.portrait-appearance-6,
.portrait-appearance-10 { border-color: rgba(154, 102, 58, 0.72); }
.portrait-appearance-3,
.portrait-appearance-7,
.portrait-appearance-11 { border-color: rgba(83, 111, 97, 0.7); }
.portrait-appearance-4,
.portrait-appearance-8,
.portrait-appearance-12 { border-color: rgba(143, 80, 67, 0.7); }

.portrait-appearance-5 .portrait,
.portrait-appearance-6 .portrait,
.portrait-appearance-7 .portrait,
.portrait-appearance-8 .portrait {
  filter: sepia(0.14) saturate(0.84) hue-rotate(8deg) brightness(1.04);
}

.portrait-appearance-9 .portrait,
.portrait-appearance-10 .portrait,
.portrait-appearance-11 .portrait,
.portrait-appearance-12 .portrait {
  filter: sepia(0.16) saturate(1.12) hue-rotate(-10deg) brightness(0.95);
}

.member-base {
  position: absolute;
  top: 98rpx;
  left: 50%;
  z-index: 3;
  width: 108rpx;
  min-height: 41rpx;
  margin: 0;
  padding: 1rpx 0 8rpx;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  clip-path: none;
  transform: translateX(-50%);
}

.member-base::before {
  display: none;
}

.member-base::after {
  position: absolute;
  bottom: 1rpx;
  left: 50%;
  display: block;
  width: 76rpx;
  height: 5rpx;
  border-radius: 999rpx;
  background: rgba(126, 88, 42, 0.58);
  content: '';
  transform: translateX(-50%);
}

.member-base.member-bound::after {
  background: #1f9d68;
}

.chip-head {
  min-height: 21rpx;
  margin: 0;
}

.member-subtitle {
  display: block;
  overflow: hidden;
  width: 100%;
  margin-top: 3rpx;
  color: var(--archive-ink-soft, #657080);
  font-size: 17rpx;
  line-height: 1.25;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-member-chip.compact .plain-name,
.tree-member-chip.compact .name-emphasis,
.tree-member-chip.compact .given-name,
.tree-member-chip.self.compact .plain-name,
.tree-member-chip.self.compact .name-emphasis,
.tree-member-chip.self.compact .given-name {
  font-size: 20rpx;
  line-height: 1.2;
}

.tree-member-chip.lineage-PATRILINEAL .member-base,
.tree-member-chip.lineage-NATAL_FEMALE .member-base,
.tree-member-chip.spouse-tag .member-base,
.tree-member-chip.self .member-base {
  box-shadow: none;
}

.tree-member-chip.lineage-PATRILINEAL .portrait-frame {
  border-color: rgba(126, 88, 42, 0.68);
}

.tree-member-chip.lineage-NATAL_FEMALE .portrait-frame {
  border-color: rgba(126, 88, 42, 0.68);
}

.tree-member-chip.spouse-tag .member-base {
  background: transparent;
}

.tree-member-chip.self .member-base {
  background: transparent;
}

.tree-member-chip.self .portrait-frame {
  background: #edf2f5;
}

.tree-member-chip.self .member-subtitle { color: var(--archive-blue, #163353); }

.generation-mark {
  top: 60rpx;
  left: 3rpx;
}

/* Remove every remaining visual trace of the old rectangular node shell. */
.tree-member-chip,
.tree-member-chip.self,
.tree-member-chip.self.compact,
.tree-member-chip.lineage-tag.lineage-PATRILINEAL,
.tree-member-chip.lineage-tag.lineage-NATAL_FEMALE,
.tree-member-chip.lineage-tag.lineage-EXTERNAL,
.tree-member-chip.spouse-tag,
.tree-member-chip.in-focus:not(.self),
.tree-member-chip.interactive,
.tree-member-chip.selectable {
  border: 0;
  border-radius: 0;
  background: transparent;
  background-image: none;
  box-shadow: none;
  overflow: visible;
  padding: 0;
}

.self-tag,
.tree-member-chip.self .self-tag,
.tree-member-chip.self.compact .self-tag {
  padding: 0;
  border-radius: 0;
  background: transparent;
  color: var(--archive-cinnabar, #a83b2d);
  font-size: 15rpx;
  line-height: 1;
}

.tree-member-chip.self.compact::before,
.tree-member-chip.self.compact::after {
  display: none;
}

.tree-member-chip.living-deceased .member-base,
.tree-member-chip.spouse-tag .member-base,
.tree-member-chip.self .member-base {
  background: transparent;
}

.tree-member-chip.living-deceased .member-base::after {
  background: rgba(116, 122, 128, 0.54);
}

.tree-member-chip.self .member-base::after {
  background: var(--archive-blue, #163353);
}

.tree-member-chip.self .member-base.member-bound::after {
  background: #1f9d68;
}

/* Historical examples use a distinct portrait treatment without affecting family trees. */
.tree-member-chip.portrait-imperial .portrait-frame {
  border-color: #9d6727;
  background:
    radial-gradient(circle at 50% 32%, rgba(255, 249, 224, 0.96), rgba(234, 208, 151, 0.94));
  box-shadow:
    0 0 0 3rpx rgba(188, 137, 57, 0.28),
    0 3rpx 8rpx rgba(72, 48, 19, 0.16);
}

.tree-member-chip.portrait-imperial .member-base::after {
  background: #a87531;
}

.tree-member-chip.portrait-imperial .member-base.member-bound::after {
  background: #1f9d68;
}

.tree-member-chip.portrait-imperial .surname-seal {
  border-color: rgba(157, 103, 39, 0.62);
  color: #8c5b24;
}

/* Regional assets include the face, hairstyle and collar; no decorative overlay is added. */
.tree-member-chip.portrait-regional .portrait-frame {
  background: #f4ead7;
}

/* Keep the same regional portrait language while making adult age states legible in a tiny node. */
.tree-member-chip.portrait-regional.age-youth .portrait {
  filter: brightness(1.06) saturate(1.08);
}
.tree-member-chip.portrait-regional.age-adult .portrait-frame {
  border-color: rgba(126, 88, 42, 0.74);
}
.tree-member-chip.portrait-regional.age-mature .portrait {
  filter: saturate(0.82) sepia(0.08);
}
.tree-member-chip.portrait-regional.age-mature .portrait-frame {
  border-color: rgba(111, 84, 53, 0.76);
}
.tree-member-chip.portrait-regional.age-senior .portrait {
  filter: grayscale(0.38) saturate(0.68) brightness(0.98);
}
.tree-member-chip.portrait-regional.age-senior .portrait-frame {
  border-color: rgba(105, 110, 112, 0.72);
}
.tree-member-chip.portrait-regional.age-elder .portrait {
  filter: grayscale(0.66) saturate(0.45) brightness(0.97);
}
.tree-member-chip.portrait-regional.age-centenarian .portrait {
  filter: grayscale(0.78) saturate(0.32) brightness(1.02);
}
.tree-member-chip.portrait-regional.age-elder .portrait-frame {
  border-color: rgba(89, 96, 102, 0.78);
  box-shadow: 0 2rpx 5rpx rgba(58, 47, 31, 0.14);
}
.tree-member-chip.portrait-regional.age-centenarian .portrait-frame {
  border-color: rgba(111, 101, 82, 0.8);
  box-shadow: 0 0 0 2rpx rgba(190, 159, 104, 0.36), 0 2rpx 5rpx rgba(58, 47, 31, 0.14);
}
</style>
