<template>
  <el-dialog v-model="visible" :title="title" width="420px">
    <p class="confirm-text">{{ message }}</p>
    <el-input
      v-if="reasonLabel"
      v-model="reason"
      type="textarea"
      :rows="3"
      :disabled="submitting"
      :placeholder="reasonLabel"
    />
    <template #footer>
      <el-button :disabled="submitting" @click="visible = false">取消</el-button>
      <el-button :type="danger ? 'danger' : 'primary'" :loading="submitting" @click="confirm">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const visible = defineModel<boolean>({ required: true })

const props = withDefaults(defineProps<{
  title: string
  message: string
  danger?: boolean
  reasonLabel?: string
  submitting?: boolean
  closeOnConfirm?: boolean
}>(), {
  danger: false,
  reasonLabel: '',
  submitting: false,
  closeOnConfirm: true
})

const emit = defineEmits<{ confirm: [reason: string] }>()
const reason = ref('')

watch(visible, (value) => {
  if (!value && !props.submitting) {
    reason.value = ''
  }
})

function confirm() {
  emit('confirm', reason.value)
  if (props.closeOnConfirm) {
    reason.value = ''
    visible.value = false
  }
}
</script>

<style scoped>
.confirm-text {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
}
</style>
