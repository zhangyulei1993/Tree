<template>
  <LegalDocumentView
    title="隐私政策"
    :meta-line="LEGAL_META_LINE"
    :sections="PRIVACY_POLICY_SECTIONS"
    :append-sections="COLLECTED_INFO_ITEMS"
  />
</template>

<script setup lang="ts">
import LegalDocumentView from '@/components/legal/LegalDocumentView.vue'
import { PRIVACY_POLICY_SECTIONS } from '@/features/legal/legalContent'
import { LEGAL_META_LINE } from '@/features/legal/legalMeta'
import { COLLECTED_INFO_ITEMS } from '@/features/legal/privacyCatalog'
</script>
