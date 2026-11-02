<template>
  <FamilyTreeMemberChip
    :node="node"
    :is-self="isSelf"
    :show-binding="showBinding"
    :kinship-title="kinshipTitle"
    :family-surname="familySurname"
    :interactive="interactive"
    :selectable="selectable"
    :in-focus="inFocus"
    :is-root="isRoot"
    :lineage-role="lineageRole"
    :spouse-relation-label="spouseRelationLabel"
    :health-hints="healthHints"
    :portrait-theme="portraitTheme"
    :portrait-region="portraitRegion"
    :portrait-region-code="portraitRegionCode"
    :portrait-appearance-variant="portraitAppearanceVariant"
    compact
    @select="emit('select', $event)"
  />
</template>

<script setup lang="ts">
import FamilyTreeMemberChip from '@/components/family/FamilyTreeMemberChip.vue'
import type { TreeNode } from '@/types/api'

defineProps<{
  node: TreeNode
  isSelf?: boolean
  showBinding?: boolean
  kinshipTitle?: string
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
</script>
