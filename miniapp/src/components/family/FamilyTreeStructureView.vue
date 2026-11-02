<template>
  <FamilyTreeGraph
    :tree="tree"
    :viewer-member-id="viewerMemberId"
    :show-binding="showBinding"
    :interactive="interactive"
    :selectable-all="selectableAll"
    :selectable-deceased="selectableDeceased"
    :health-hints="healthHints"
    :portrait-theme="portraitTheme"
    :portrait-region="portraitRegion"
    :portrait-region-code="portraitRegionCode"
    @select="emit('select', $event)"
  />
</template>

<script setup lang="ts">
import FamilyTreeGraph from '@/components/family/FamilyTreeGraph.vue'
import type { FamilyTreeResult } from '@/types/api'

defineProps<{
  tree: FamilyTreeResult
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

const emit = defineEmits<{
  select: [node: FamilyTreeResult['nodes'][number]]
}>()
</script>
