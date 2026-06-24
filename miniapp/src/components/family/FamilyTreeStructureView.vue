<template>
  <FamilyTreeGraph
    :tree="tree"
    :viewer-member-id="viewerMemberId"
    :show-binding="showBinding"
  />
</template>

<script setup lang="ts">
import FamilyTreeGraph from '@/components/family/FamilyTreeGraph.vue'
import type { FamilyTreeResult } from '@/types/api'

defineProps<{
  tree: FamilyTreeResult
  viewerMemberId?: number | null
  showBinding?: boolean
}>()
</script>
