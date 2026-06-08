<template>
  <el-dialog v-model="visible" :title="title" width="520px">
    <el-form label-width="84px">
      <el-form-item label="审核结论">
        <el-tag :type="action === 'approve' ? 'success' : 'danger'">
          {{ action === 'approve' ? '审核通过' : '审核拒绝' }}
        </el-tag>
      </el-form-item>
      <el-form-item label="审核备注">
        <el-input v-model="comment" type="textarea" :rows="4" placeholder="填写 mock 审核备注" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button :type="action === 'approve' ? 'primary' : 'danger'" @click="submit">提交</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const visible = defineModel<boolean>({ required: true })

defineProps<{
  title: string
  action: 'approve' | 'reject'
}>()

const emit = defineEmits<{ submit: [comment: string] }>()
const comment = ref('')

function submit() {
  emit('submit', comment.value)
  comment.value = ''
  visible.value = false
}
</script>
