<template>
  <LegalDocumentView
    kicker="User Agreement"
    seal-letter="协"
    title="用户协议"
    :product-name="LEGAL_PRODUCT_NAME"
    :updated-date="LEGAL_UPDATED_DATE"
    :meta-line="LEGAL_META_LINE"
    :sections="USER_AGREEMENT_SECTIONS"
  />
</template>

<script setup lang="ts">
import LegalDocumentView from '@/components/legal/LegalDocumentView.vue'
import { USER_AGREEMENT_SECTIONS } from '@/features/legal/legalContent'
import { LEGAL_META_LINE, LEGAL_PRODUCT_NAME, LEGAL_UPDATED_DATE } from '@/features/legal/legalMeta'
</script>
