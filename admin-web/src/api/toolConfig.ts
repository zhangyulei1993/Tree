import { apiClient, apiMode, unwrapData } from '@/api/client'
import type { ToolConfig } from '@/types/api'

const mockConfigs: ToolConfig[] = [
  { id: 5, toolKey: 'CHOOSE', displayName: '该选什么', description: '把选项交给随机选择，帮你快速做决定', visible: true, enabled: false, pinned: false, highlighted: false, sortOrder: 0, updatedAt: new Date().toISOString() },
  { id: 1, toolKey: 'KINSHIP_QUERY', displayName: '称谓查询', description: '按关系、性别和长幼查询称谓', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0, updatedAt: new Date().toISOString() },
  { id: 2, toolKey: 'TRADITIONAL_FESTIVALS', displayName: '传统节日', description: '查看传统节日、法定假期与节日文案', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0, updatedAt: new Date().toISOString() },
  { id: 3, toolKey: 'SOLAR_TERMS', displayName: '农历节气', description: '查看二十四节气与时令变化', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0, updatedAt: new Date().toISOString() },
  { id: 4, toolKey: 'FAMILY_STORIES', displayName: '家庭事迹', description: '查看家庭成员与关系变动记录', visible: true, enabled: true, pinned: false, highlighted: false, sortOrder: 0, updatedAt: new Date().toISOString() }
]

export interface CreateToolConfigInput {
  toolKey: string
  displayName: string
  description: string
}

export async function listToolConfigs() {
  if (apiMode === 'mock') return mockConfigs
  const response = await apiClient.get('/admin/tool-configs')
  return unwrapData<ToolConfig[]>(response)
}

export async function createToolConfig(input: CreateToolConfigInput) {
  if (apiMode === 'mock') {
    const item: ToolConfig = {
      id: Math.max(0, ...mockConfigs.map((config) => config.id)) + 1,
      toolKey: input.toolKey.toUpperCase(),
      displayName: input.displayName,
      description: input.description,
      visible: false,
      enabled: false,
      pinned: false,
      highlighted: false,
      sortOrder: 0,
      updatedAt: new Date().toISOString()
    }
    mockConfigs.unshift(item)
    return item
  }
  const response = await apiClient.post('/admin/tool-configs', input)
  return unwrapData<ToolConfig>(response)
}

export type ToolConfigUpdateInput = Partial<Pick<ToolConfig, 'visible' | 'enabled' | 'pinned' | 'highlighted' | 'sortOrder'>>

export async function updateToolConfig(toolKey: string, input: ToolConfigUpdateInput) {
  if (apiMode === 'mock') {
    const item = mockConfigs.find((config) => config.toolKey === toolKey)
    if (item) Object.assign(item, input, { updatedAt: new Date().toISOString() })
    return item
  }
  const response = await apiClient.put(`/admin/tool-configs/${toolKey}`, input)
  return unwrapData<ToolConfig>(response)
}
