<template>
  <LegalDocumentView
    title="用户协议"
    :meta-line="LEGAL_META_LINE"
    :sections="USER_AGREEMENT_SECTIONS"
  />
</template>

<script setup lang="ts">
import LegalDocumentView from '@/components/legal/LegalDocumentView.vue'
import { USER_AGREEMENT_SECTIONS } from '@/features/legal/legalContent'
import { LEGAL_META_LINE } from '@/features/legal/legalMeta'
</script>
