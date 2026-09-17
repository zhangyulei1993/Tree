import { reactive } from 'vue'

import { isRealApiMode, request } from '@/api/client'

export type CommonToolKey =
  | 'CHOOSE'
  | 'KINSHIP_QUERY'
  | 'TRADITIONAL_FESTIVALS'
  | 'SOLAR_TERMS'
  | 'FAMILY_STORIES'

export interface ToolConfigState {
  displayName: string
  description: string
  visible: boolean
  enabled: boolean
  pinned: boolean
  highlighted: boolean
  sortOrder: number
}

export interface PublicToolConfig {
  toolKey: string
  displayName: string
  description: string
  visible?: boolean
  enabled?: boolean
  pinned?: boolean
  highlighted?: boolean
  sortOrder?: number
  updatedAt?: string
}

export interface SupportedToolDefinition {
  key: string
  path?: string
  icon: string
  displayName: string
  description: string
}

export const supportedToolDefinitions: SupportedToolDefinition[] = [
  { key: 'CHOOSE', path: '/pages/tools/choose', icon: '选', displayName: '该选什么', description: '把选项交给随机选择，帮你快速做决定' },
  { key: 'KINSHIP_QUERY', path: '/pages/tools/kinship', icon: '称', displayName: '称谓查询', description: '按关系、性别和长幼查询称谓' },
  { key: 'TRADITIONAL_FESTIVALS', path: '/pages/tools/festivals', icon: '节', displayName: '传统节日', description: '查看传统节日、法定假期与节日文案' },
  { key: 'SOLAR_TERMS', path: '/pages/tools/solar-terms', icon: '时', displayName: '农历节气', description: '查看二十四节气与时令变化' },
  { key: 'FAMILY_STORIES', path: '/pages/family/stories', icon: '记', displayName: '家庭事迹', description: '查看家庭成员与关系变动记录' }
]

export const toolConfigState = reactive<Record<string, ToolConfigState>>({
  CHOOSE: { displayName: '该选什么', description: '把选项交给随机选择，帮你快速做决定', visible: true, enabled: false, pinned: false, highlighted: false, sortOrder: 0 },
  KINSHIP_QUERY: { displayName: '称谓查询', description: '按关系、性别和长幼查询称谓', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0 },
  TRADITIONAL_FESTIVALS: { displayName: '传统节日', description: '查看传统节日、法定假期与节日文案', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0 },
  SOLAR_TERMS: { displayName: '农历节气', description: '查看二十四节气与时令变化', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0 },
  FAMILY_STORIES: { displayName: '家庭事迹', description: '查看家庭成员与关系变动记录', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0 }
})

export const configuredToolDefinitions = reactive<SupportedToolDefinition[]>(
  supportedToolDefinitions.map((definition) => ({ ...definition }))
)

function fallbackIcon(displayName: string, toolKey: string) {
  return displayName.trim().slice(0, 1) || toolKey.trim().slice(0, 1) || '用'
}

function stateFor(toolKey: string) {
  return toolConfigState[toolKey]
}

function definitionFor(toolKey: string) {
  return configuredToolDefinitions.find((definition) => definition.key === toolKey)
}

let loadedAt = 0
let loading: Promise<void> | null = null
const cacheTtlMs = 30_000

export function isToolEnabled(toolKey: string) {
  return stateFor(toolKey)?.enabled ?? false
}

export function isToolVisible(toolKey: string) {
  return stateFor(toolKey)?.visible ?? false
}

export function isToolHighlighted(toolKey: string) {
  return stateFor(toolKey)?.highlighted ?? false
}

export function getToolDisplayName(toolKey: string) {
  return stateFor(toolKey)?.displayName || definitionFor(toolKey)?.displayName || toolKey
}

export function getToolDescription(toolKey: string) {
  return stateFor(toolKey)?.description || definitionFor(toolKey)?.description || '该工具暂未提供说明'
}

export function getToolDefinition(toolKey: string) {
  return definitionFor(toolKey)
}

export function getVisibleToolDefinitions() {
  return configuredToolDefinitions
    .filter((item) => isToolVisible(item.key))
    .sort((left, right) => {
      const leftState = stateFor(left.key)
      const rightState = stateFor(right.key)
      if (!leftState || !rightState) return 0
      if (leftState.pinned !== rightState.pinned) return leftState.pinned ? -1 : 1
      if (leftState.sortOrder !== rightState.sortOrder) return leftState.sortOrder - rightState.sortOrder
      if (leftState.highlighted !== rightState.highlighted) return leftState.highlighted ? -1 : 1
      return configuredToolDefinitions.indexOf(left) - configuredToolDefinitions.indexOf(right)
    })
}

export async function loadToolConfigs(options: { force?: boolean } = {}) {
  if (!isRealApiMode) return
  const fresh = loadedAt > 0 && Date.now() - loadedAt < cacheTtlMs
  if (!options.force && fresh) return
  if (loading) return loading
  loading = (async () => {
    try {
      const rows = await request<PublicToolConfig[]>('/tool-configs', { public: true })
      const configuredKeys = new Set<string>()
      const nextDefinitions: SupportedToolDefinition[] = rows
        .filter((row) => row.toolKey.trim())
        .map((row) => {
          const key = row.toolKey.trim()
          const fallback = supportedToolDefinitions.find((definition) => definition.key === key)
          const state = toolConfigState[key] || (toolConfigState[key] = {
            displayName: row.displayName?.trim() || key,
            description: row.description?.trim() || '该工具暂未提供说明',
            visible: true,
            enabled: false,
            pinned: false,
            highlighted: false,
            sortOrder: 0
          })
          state.displayName = row.displayName?.trim() || fallback?.displayName || state.displayName
          state.description = row.description?.trim() || fallback?.description || state.description
          state.visible = row.visible ?? true
          state.enabled = row.enabled ?? true
          state.pinned = row.pinned ?? false
          state.highlighted = row.highlighted ?? false
          state.sortOrder = row.sortOrder ?? 0
          configuredKeys.add(key)
          return {
            key,
            path: fallback?.path,
            icon: fallback?.icon || fallbackIcon(state.displayName, key),
            displayName: state.displayName,
            description: state.description
          }
        })
      supportedToolDefinitions.forEach((definition) => {
        if (!configuredKeys.has(definition.key)) nextDefinitions.push({ ...definition })
      })
      configuredToolDefinitions.splice(0, configuredToolDefinitions.length, ...nextDefinitions)
      loadedAt = Date.now()
    } catch {
      // Keep bundled defaults when the endpoint is unavailable.
    } finally {
      loading = null
    }
  })()
  return loading
}

export async function ensureToolEnabled(toolKey: string) {
  await loadToolConfigs()
  if (isToolEnabled(toolKey)) return true
  uni.showToast({ title: '该工具暂未开放', icon: 'none' })
  setTimeout(() => {
    uni.navigateBack({ fail: () => uni.switchTab({ url: '/pages/home/index' }) })
  }, 350)
  return false
}
