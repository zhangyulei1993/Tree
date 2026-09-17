import assert from 'node:assert/strict'
import test from 'node:test'

import { parseChoiceOptions, pickRandomChoice } from './choiceOptions'

test('parseChoiceOptions supports line and comma separated options without duplicates', () => {
  assert.deepEqual(parseChoiceOptions('火锅\n烧烤，火锅\n 日料 '), ['火锅', '烧烤', '日料'])
})

test('pickRandomChoice returns a random option and avoids immediate repeats', () => {
  const options = ['火锅', '烧烤', '日料']
  assert.equal(pickRandomChoice(options, () => 0), '火锅')
  assert.equal(pickRandomChoice(options, () => 0, '火锅'), '烧烤')
  assert.equal(pickRandomChoice(options, () => 0.99, '日料'), '烧烤')
})
