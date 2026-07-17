<template>
  <FamilyTreeMemberChip
    :node="node"
    :is-self="isSelf"
    :show-binding="showBinding"
    :kinship-title="kinshipTitle"
    :family-surname="familySurname"
    :interactive="interactive"
    :selectable="selectable"
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
}>()

const emit = defineEmits<{
  select: [node: TreeNode]
}>()
</script>
