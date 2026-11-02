<template>
  <pre class="json-viewer">{{ formatted }}</pre>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ value: unknown }>()
const formatted = computed(() => JSON.stringify(props.value, null, 2))
</script>

<style scoped>
.json-viewer {
  max-height: 420px;
  margin: 0;
  overflow: auto;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  background: #0f172a;
  color: #dbeafe;
  font-size: 12px;
  line-height: 1.6;
  padding: 14px;
}
</style>
