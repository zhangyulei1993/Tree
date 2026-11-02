import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { test } from 'node:test'

const page = readFileSync(join(process.cwd(), 'src', 'pages', 'tools', 'kinship.vue'), 'utf8')

test('kinship first step keeps relation selection visible and requests self gender only when needed', () => {
  assert.match(page, /<view v-if="showSelfGenderPanel" class="archive-form-panel kinship-panel self-gender-panel">[\s\S]*补充本人性别/)
  assert.match(page, /<text class="archive-section-title">选择下一层关系<\/text>/)
  assert.match(page, /<view v-if="steps\.length > 0" class="archive-form-panel kinship-panel">[\s\S]*关系路径/)
  assert.match(page, /<view v-if="pendingRelation" class="archive-form-panel kinship-panel">[\s\S]*补充这个人的信息/)
  assert.match(page, /<view v-if="needsRelativeAgeForCurrentRelation" class="relative-age-block">[\s\S]*relativeAgePrompt/)
  assert.match(page, /return '年龄关系（相对本人）'/)
  assert.match(page, /function requiresRelativeAgeForRelation\(\s*relation: KinshipRelation \| null,\s*currentSteps: KinshipStep\[\],\s*pendingGender: Gender\s*\): boolean/)
  assert.match(page, /const relationOptions = computed\(\(\) => getRelationOptions\(context\.value\)\)/)
  assert.match(page, /<view v-if="steps\.length > 0" class="kinship-result"/)
})
