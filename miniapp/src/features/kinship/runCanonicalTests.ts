import { runCanonicalMatrixTest, runFullMatrix43x42Test, runReverseKinshipTest } from '../family-tree/canonicalKinshipMatrix'
import { runParentSiblingSeniorityChecks } from '../family-tree/parentSiblingSeniorityCases'
import { runRelativeTitleSelfCheck } from '../family-tree/relativeTitle.testCases'
import { runDeepDirectDescendantKinshipTests } from './deepDirectDescendantKinship'
import { resolveCanonicalKinship } from './resolveCanonicalKinship'
import { runKinshipSelfChecks } from './testCases'
import { MAX_KINSHIP_DEPTH } from './types'

const FORBIDDEN_BARE = ['亲属', '姻亲', '祖辈亲属', '父母辈亲属', '晚辈亲属', '同辈亲属']

function assertGlobalCanonicalRules(): string[] {
  const failures: string[] = []
  const samples = [
    { steps: [{ relation: 'parent' as const, person: { gender: 'male' as const } }] },
    { steps: [{ relation: 'sibling' as const, person: { gender: 'male' as const }, relativeAge: 'older' as const }] }
  ]
  for (const sample of samples) {
    const result = resolveCanonicalKinship({
      self: { gender: 'male' },
      steps: sample.steps,
      maxDepth: MAX_KINSHIP_DEPTH
    })
    if (result.canonicalTitle && FORBIDDEN_BARE.includes(result.canonicalTitle)) {
      failures.push(`规范结果不得为裸称谓「${result.canonicalTitle}」`)
    }
  }
  return failures
}

export function runAllCanonicalKinshipTests(): {
  pathChecks: ReturnType<typeof runKinshipSelfChecks>
  relativeTitle: ReturnType<typeof runRelativeTitleSelfCheck>
  matrix: ReturnType<typeof runCanonicalMatrixTest>
  fullMatrix: ReturnType<typeof runFullMatrix43x42Test>
  reverse: ReturnType<typeof runReverseKinshipTest>
  parentSiblingSeniority: ReturnType<typeof runParentSiblingSeniorityChecks>
  deepDirectDescendant: ReturnType<typeof runDeepDirectDescendantKinshipTests>
  globalRuleFailures: string[]
  ok: boolean
} {
  const pathChecks = runKinshipSelfChecks()
  const relativeTitle = runRelativeTitleSelfCheck()
  const parentSiblingSeniority = runParentSiblingSeniorityChecks()
  const matrix = runCanonicalMatrixTest()
  const fullMatrix = runFullMatrix43x42Test()
  const reverse = runReverseKinshipTest()
  const deepDirectDescendant = runDeepDirectDescendantKinshipTests()
  const globalRuleFailures = assertGlobalCanonicalRules()

  const ok =
    pathChecks.failures.length === 0 &&
    relativeTitle.failed === 0 &&
    relativeTitle.cacheOk &&
    matrix.failures.length === 0 &&
    fullMatrix.registryViolations === 0 &&
    fullMatrix.forbiddenBareViolations === 0 &&
    fullMatrix.invalidPathViolations === 0 &&
    fullMatrix.generationMismatchViolations === 0 &&
    reverse.failures.length === 0 &&
    parentSiblingSeniority.failures.length === 0 &&
    deepDirectDescendant.failures.length === 0 &&
    globalRuleFailures.length === 0

  return {
    pathChecks,
    relativeTitle,
    matrix,
    fullMatrix,
    reverse,
    parentSiblingSeniority,
    deepDirectDescendant,
    globalRuleFailures,
    ok
  }
}

const isDirectRun = typeof process !== 'undefined' && process.argv[1]?.includes('runCanonicalTests')
if (isDirectRun) {
  const result = runAllCanonicalKinshipTests()
  console.log(
    `path kinship: ${result.pathChecks.passed}/${result.pathChecks.total} passed (${result.pathChecks.failures.length} failures)`
  )
  console.log(
    `relative title: ${result.relativeTitle.passed}/${result.relativeTitle.total} passed, cache=${result.relativeTitle.cacheOk ? 'OK' : 'FAIL'}`
  )
  console.log(
    `matrix: centers=${result.matrix.lineageCenterCount}, golden=${result.matrix.goldenPassed}/${result.matrix.goldenTotal}, spouseRejected=${result.matrix.spouseRejected}`
  )
  console.log(
    `full-matrix: pairs=${result.fullMatrix.totalPairs}, supported=${result.fullMatrix.supported}, unsupported=${result.fullMatrix.unsupported}, registryViolations=${result.fullMatrix.registryViolations}, invalidPath=${result.fullMatrix.invalidPathViolations}`
  )
  console.log(`reverse: ${result.reverse.passed}/${result.reverse.total} passed`)
  console.log(
    `parent-sibling-seniority: ${result.parentSiblingSeniority.passed}/${result.parentSiblingSeniority.total} passed`
  )
  console.log(
    `deep-direct-descendant: ${result.deepDirectDescendant.passed}/${result.deepDirectDescendant.total} passed`
  )

  const allFailures = [
    ...result.pathChecks.failures.slice(0, 20),
    ...result.relativeTitle.failures.map((f) => `relative: ${f.name} expected ${f.expected} got ${f.actual}`),
    ...result.matrix.failures,
    ...result.fullMatrix.failures,
    ...result.reverse.failures,
    ...result.parentSiblingSeniority.failures,
    ...result.deepDirectDescendant.failures,
    ...result.globalRuleFailures
  ]

  if (!result.ok) {
    for (const line of allFailures) {
      console.log(`  FAIL ${line}`)
    }
    process.exitCode = 1
  } else {
    console.log('ALL CANONICAL KINSHIP TESTS PASSED')
  }
}
