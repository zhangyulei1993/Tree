<template>
  <el-dialog v-model="visible" :title="title" width="420px">
    <p class="confirm-text">{{ message }}</p>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button :type="danger ? 'danger' : 'primary'" @click="confirm">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
const visible = defineModel<boolean>({ required: true })

defineProps<{
  title: string
  message: string
  danger?: boolean
}>()

const emit = defineEmits<{ confirm: [] }>()

function confirm() {
  visible.value = false
  emit('confirm')
}
</script>

<style scoped>
.confirm-text {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
}
</style>
