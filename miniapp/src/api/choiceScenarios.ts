import { isRealApiMode, request } from '@/api/client'

export type ChoiceScenarioPeriod = 'today' | '7d' | 'month'

export interface SharedChoiceScenario {
  id: number | string
  title: string
  options: string[]
  useCount: number
  createdAt: string
}

const mockScenarios: SharedChoiceScenario[] = [
  {
    id: 'mock_shared_1',
    title: '周末去哪儿',
    options: ['公园散步', '看电影', '逛街', '周边短途游', '在家休息'],
    useCount: 12,
    createdAt: new Date().toISOString()
  },
  {
    id: 'mock_shared_2',
    title: '今晚吃什么',
    options: ['家常菜', '火锅', '烧烤', '面食', '外卖'],
    useCount: 8,
    createdAt: new Date().toISOString()
  }
]

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
