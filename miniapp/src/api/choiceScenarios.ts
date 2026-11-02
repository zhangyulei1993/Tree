import { isRealApiMode, request } from '@/api/client'

export type ChoiceScenarioPeriod = 'today' | '7d' | 'month'

export interface SharedChoiceScenario {
  id: number | string
  title: string
  options: string[]
  useCount: number
  createdAt: string
}

const mockScenarios: SharedChoiceScenario[] = []

export async function listSharedChoiceScenarios(period: ChoiceScenarioPeriod) {
  if (!isRealApiMode) return mockScenarios.map((item) => ({ ...item, options: [...item.options] }))
  return request<SharedChoiceScenario[]>(`/choice-scenarios?period=${period}`, { public: true })
}

export async function publishSharedChoiceScenario(title: string, options: string[]) {
  if (!isRealApiMode) {
    const item: SharedChoiceScenario = {
      id: `mock_shared_${Date.now()}`,
      title,
      options: [...options],
      useCount: 0,
      createdAt: new Date().toISOString()
    }
    mockScenarios.unshift(item)
    return item
  }
  return request<SharedChoiceScenario, { title: string; options: string[] }>('/choice-scenarios', {
    method: 'POST',
    data: { title, options }
  })
}

export async function recordSharedChoiceScenarioUse(scenarioId: number | string) {
  if (!isRealApiMode) return
  await request<{ used: boolean }>(`/choice-scenarios/${scenarioId}/use`, { method: 'POST', data: {} })
}
