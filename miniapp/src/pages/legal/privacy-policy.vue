<template>
  <LegalDocumentView
    kicker="Privacy Policy"
    seal-letter="隐"
    title="隐私政策"
    :product-name="LEGAL_PRODUCT_NAME"
    :updated-date="LEGAL_UPDATED_DATE"
    :meta-line="LEGAL_META_LINE"
    :sections="PRIVACY_POLICY_SECTIONS"
    :append-sections="COLLECTED_INFO_ITEMS"
  />
</template>

<script setup lang="ts">
import LegalDocumentView from '@/components/legal/LegalDocumentView.vue'
import { PRIVACY_POLICY_SECTIONS } from '@/features/legal/legalContent'
import { LEGAL_META_LINE, LEGAL_PRODUCT_NAME, LEGAL_UPDATED_DATE } from '@/features/legal/legalMeta'
import { COLLECTED_INFO_ITEMS } from '@/features/legal/privacyCatalog'
</script>
