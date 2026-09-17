import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { test } from 'node:test'

const page = readFileSync(join(process.cwd(), 'src', 'pages', 'tools', 'kinship.vue'), 'utf8')

test('kinship first step keeps gender and relation selection visible', () => {
  assert.match(page, /<text class="archive-section-title">本人信息（可选）<\/text>/)
  assert.match(page, /<text class="archive-section-title">选择下一层关系<\/text>/)
  assert.match(page, /<view v-if="steps\.length > 0" class="archive-form-panel kinship-panel">[\s\S]*关系路径/)
  assert.match(page, /<view v-if="pendingRelation" class="archive-form-panel kinship-panel">[\s\S]*补充这个人的信息/)
  assert.match(page, /<view v-if="needsRelativeAgeForCurrentRelation" class="relative-age-block">[\s\S]*relativeAgePrompt/)
  assert.match(page, /return '年龄关系（相对本人）'/)
  assert.match(page, /function requiresRelativeAgeForRelation\(\s*relation: KinshipRelation \| null,\s*currentSteps: KinshipStep\[\],\s*pendingGender: Gender\s*\): boolean/)
  assert.match(page, /const relationOptions = computed\(\(\) => getRelationOptions\(context\.value\)\)/)
  assert.match(page, /<view v-if="steps\.length > 0" class="kinship-result"/)
})
