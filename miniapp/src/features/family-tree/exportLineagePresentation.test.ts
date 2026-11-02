import assert from 'node:assert/strict'
import test from 'node:test'

import { getExportLineageDisplay } from './exportLineagePresentation'
import type { ExportLineageNode } from './exportLineageTree'

const item: ExportLineageNode = {
  id: 'member-1', generation: 2, role: 'PATRILINEAL', children: [],
  parentLabels: ['父：张团'], spouseLabels: ['妻：李兰'],
  member: { memberId: 1, displayName: '张达望', gender: 'MALE', memberType: 'LINEAGE_MEMBER', birthDate: '1938-01-01', deathDate: '2020-01-01', userBindingState: 'NOT_REQUIRED', canExpand: false }
}

test('export scopes expose intentionally different information density', () => {
  assert.deepEqual(getExportLineageDisplay(item, 'INTERNAL'), { name: '张达望', yearText: '1938-2020', parentLabels: ['父：张团'], spouseLabels: ['妻：李兰'] })
  assert.deepEqual(getExportLineageDisplay(item, 'SHARE'), { name: '张达望', yearText: '1938-2020', parentLabels: ['父：张团'], spouseLabels: ['妻：李兰'] })
  assert.deepEqual(getExportLineageDisplay(item, 'PUBLIC'), { name: '张*', yearText: '', parentLabels: ['父：张*'], spouseLabels: ['妻：李*'] })
})

test('living members do not show a birth year in any non-public export', () => {
  const livingItem: ExportLineageNode = {
    ...item,
    member: { ...item.member, deathDate: undefined }
  }

  assert.equal(getExportLineageDisplay(livingItem, 'INTERNAL').yearText, '')
  assert.equal(getExportLineageDisplay(livingItem, 'SHARE').yearText, '')
})

test('public export masks a spouse death year together with the spouse identity', () => {
  const deceasedSpouseItem: ExportLineageNode = { ...item, spouseLabels: ['妻：李兰（卒：2020）'] }

  assert.deepEqual(getExportLineageDisplay(deceasedSpouseItem, 'PUBLIC').spouseLabels, ['妻：李*'])
})

test('share and public exports hide former marriages while the family archive retains them', () => {
  const remarriedItem: ExportLineageNode = {
    ...item,
    spouseLabels: ['前妻：李兰', '现任妻：王梅']
  }

  assert.deepEqual(getExportLineageDisplay(remarriedItem, 'INTERNAL').spouseLabels, ['前妻：李兰', '现任妻：王梅'])
  assert.deepEqual(getExportLineageDisplay(remarriedItem, 'SHARE').spouseLabels, ['现任妻：王梅'])
  assert.deepEqual(getExportLineageDisplay(remarriedItem, 'PUBLIC').spouseLabels, ['现任妻：王*'])
})
