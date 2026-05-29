import type { Gender, KinshipContext, KinshipStep } from './types'
import { allowedRelations } from './allowedRelations'
import { resolveKinship } from './resolveKinship'

function p(relation: KinshipStep['relation'], gender: Gender, age: number | string = ''): KinshipStep {
  return { relation, gender, age }
}

const maleSelf = { name: '我', gender: 'male' as Gender, age: 30 }
const femaleSelf = { name: '我', gender: 'female' as Gender, age: 30 }

export interface KinshipSelfTestCase {
  id: string
  label: string
  context: KinshipContext
  expectedTitle?: string
  expectedAllowed?: KinshipStep['relation'][]
  expectedTerminal?: boolean
}

export const KINSHIP_SELF_TEST_CASES: KinshipSelfTestCase[] = [
  { id: 'T001', label: '父亲', context: { self: maleSelf, steps: [p('parent', 'male', 60)] }, expectedTitle: '父亲' },
  { id: 'T002', label: '爷爷', context: { self: maleSelf, steps: [p('parent', 'male', 60), p('parent', 'male', 85)] }, expectedTitle: '爷爷' },
  { id: 'T003', label: '岳父', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('parent', 'male', 60)] }, expectedTitle: '岳父', expectedAllowed: ['parent', 'sibling'] },
  { id: 'T004', label: '公公', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('parent', 'male', 60)] }, expectedTitle: '公公', expectedAllowed: ['parent', 'sibling'] },
  { id: 'T005', label: '大舅子', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'male', 35)] }, expectedTitle: '大舅子' },
  { id: 'T006', label: '小舅子', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'male', 20)] }, expectedTitle: '小舅子' },
  { id: 'T007', label: '大姨子', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'female', 35)] }, expectedTitle: '大姨子' },
  { id: 'T008', label: '小姨子', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'female', 20)] }, expectedTitle: '小姨子' },
  { id: 'T009', label: '大伯哥', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'male', 38)] }, expectedTitle: '大伯哥' },
  { id: 'T010', label: '小叔子', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'male', 25)] }, expectedTitle: '小叔子' },
  { id: 'T011', label: '大姑姐', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'female', 38)] }, expectedTitle: '大姑姐' },
  { id: 'T012', label: '小姑妹', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'female', 25)] }, expectedTitle: '小姑妹' },
  { id: 'T013', label: '内兄嫂', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'male', 35), p('spouse', 'female', 33)] }, expectedTitle: '内兄嫂', expectedTerminal: true },
  { id: 'T014', label: '内弟媳', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'male', 20), p('spouse', 'female', 20)] }, expectedTitle: '内弟媳', expectedTerminal: true },
  { id: 'T015', label: '连襟', context: { self: maleSelf, steps: [p('spouse', 'female', 28), p('sibling', 'female', 24), p('spouse', 'male', 26)] }, expectedTitle: '连襟', expectedTerminal: true, expectedAllowed: [] },
  { id: 'T016', label: '妯娌', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'male', 29), p('spouse', 'female', 28)] }, expectedTitle: '妯娌', expectedTerminal: true },
  { id: 'T017', label: '姑姐夫', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'female', 40), p('spouse', 'male', 41)] }, expectedTitle: '姑姐夫', expectedTerminal: true },
  { id: 'T018', label: '姑妹夫', context: { self: femaleSelf, steps: [p('spouse', 'male', 32), p('sibling', 'female', 25), p('spouse', 'male', 26)] }, expectedTitle: '姑妹夫', expectedTerminal: true },
  { id: 'T019', label: '伯父', context: { self: maleSelf, steps: [p('parent', 'male', 60), p('sibling', 'male', 66)] }, expectedTitle: '伯父' },
  { id: 'T020', label: '叔叔', context: { self: maleSelf, steps: [p('parent', 'male', 60), p('sibling', 'male', 50)] }, expectedTitle: '叔叔' },
  { id: 'T021', label: '堂哥', context: { self: maleSelf, steps: [p('parent', 'male', 60), p('sibling', 'male', 66), p('child', 'male', 35)] }, expectedTitle: '堂哥' },
  { id: 'T022', label: '表妹', context: { self: maleSelf, steps: [p('parent', 'female', 60), p('sibling', 'female', 55), p('child', 'female', 20)] }, expectedTitle: '表妹' },
  { id: 'T023', label: '侄女婿', context: { self: maleSelf, steps: [p('sibling', 'male', 35), p('child', 'female', 8), p('spouse', 'male', 8)] }, expectedTitle: '侄女婿', expectedTerminal: true },
  { id: 'T024', label: '孙女婿', context: { self: maleSelf, steps: [p('child', 'male', 8), p('child', 'female', 1), p('spouse', 'male', 1)] }, expectedTitle: '孙女婿', expectedTerminal: true }
]

export function runKinshipSelfTest() {
  return KINSHIP_SELF_TEST_CASES.map((item) => {
    const actual = resolveKinship(item.context)
    const allowed = allowedRelations(item.context)
    const titleOk = item.expectedTitle ? actual.title === item.expectedTitle : true
    const terminalOk = item.expectedTerminal === undefined ? true : actual.terminal === item.expectedTerminal
    const allowedOk = item.expectedAllowed === undefined
      ? true
      : JSON.stringify(allowed) === JSON.stringify(item.expectedAllowed)

    return {
      id: item.id,
      label: item.label,
      expectedTitle: item.expectedTitle,
      actualTitle: actual.title,
      matched: actual.matched,
      terminal: actual.terminal,
      allowed,
      pass: titleOk && terminalOk && allowedOk
    }
  })
}
