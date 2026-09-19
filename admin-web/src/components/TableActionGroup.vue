<template>
  <div class="table-action-group">
    <el-button size="small" type="primary" :loading="loading" :disabled="disabled" @click="emit('primary')">
      {{ primaryText }}
    </el-button>
    <el-dropdown trigger="click" @command="(command: string) => emit('command', command)">
      <el-button size="small" text :disabled="disabled || loading">更多</el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item
            v-for="action in actions"
            :key="action.command"
            :command="action.command"
            :divided="action.divided"
            :disabled="action.disabled"
            :class="{ 'is-danger': action.danger }"
          >
            {{ action.label }}
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
type TableAction = {
  command: string
  label: string
  danger?: boolean
  divided?: boolean
  disabled?: boolean
}

withDefaults(defineProps<{
  primaryText: string
  actions: TableAction[]
  loading?: boolean
  disabled?: boolean
}>(), {
  loading: false,
  disabled: false
})

const emit = defineEmits<{
  primary: []
  command: [command: string]
}>()
</script>

<style scoped>
.table-action-group { display: flex; align-items: center; gap: 6px; min-width: 140px; white-space: nowrap; }
:deep(.el-dropdown-menu__item.is-danger) { color: var(--el-color-danger); }
</style>
