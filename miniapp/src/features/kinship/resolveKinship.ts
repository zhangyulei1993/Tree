import type { Gender, KinshipContext, KinshipResolution, KinshipStep, PersonFacts, RelativeAge } from './types'
import { formatPathDescription, inferRelativeAgeFromFacts } from './helpers'
import { validateContextStructure } from './terminalRules'

function resolved(
  context: KinshipContext,
  primaryTitle: string,
  aliases: string[] = [],
  explanation?: string
): KinshipResolution {
  return {
    status: 'resolved',
    primaryTitle,
    aliases,
    pathDescription: formatPathDescription(context),
    explanation
  }
}

function ambiguous(
  context: KinshipContext,
  candidates: string[],
  aliases: string[] = [],
  explanation?: string
): KinshipResolution {
  return {
    status: 'ambiguous',
    primaryTitle: candidates[0],
    candidates,
    aliases,
    pathDescription: formatPathDescription(context),
    explanation: explanation || '信息不足，可能存在多种称谓，请补充性别或长幼信息。'
  }
}

function unsupported(context: KinshipContext, explanation?: string): KinshipResolution {
  return {
    status: 'unsupported',
    aliases: [],
    pathDescription: formatPathDescription(context),
    explanation: explanation || '暂未收录该关系的常用称谓，你仍然可以保留完整关系路径。'
  }
}

function stepKey(steps: KinshipStep[]): string {
  return steps.map((step) => step.relation).join('>')
}

function genderChildLabel(gender: Gender, male: string, female: string, neutral: string): string {
  if (gender === 'male') return male
  if (gender === 'female') return female
  return neutral
}

function resolveParentParent(
  context: KinshipContext,
  first: KinshipStep,
  second: KinshipStep
): KinshipResolution {
  const firstGender = first.person.gender
  const secondGender = second.person.gender

  if (firstGender === 'male' && secondGender === 'male') {
    return resolved(context, '祖父', ['爷爷'])
  }
  if (firstGender === 'male' && secondGender === 'female') {
    return resolved(context, '祖母', ['奶奶'])
  }
  if (firstGender === 'female' && secondGender === 'male') {
    return resolved(context, '外祖父', ['外公'])
  }
  if (firstGender === 'female' && secondGender === 'female') {
    return resolved(context, '外祖母', ['外婆'])
  }

  if (firstGender === 'unknown') {
    if (secondGender === 'male') {
      return ambiguous(context, ['祖父', '外祖父'], ['爷爷', '外公'], '需要补充第一层父母性别以区分父系或母系祖辈。')
    }
    if (secondGender === 'female') {
      return ambiguous(context, ['祖母', '外祖母'], ['奶奶', '外婆'], '需要补充第一层父母性别以区分父系或母系祖辈。')
    }
    return ambiguous(context, ['祖父母', '外祖父母'], [], '需要补充父母性别以区分父系或母系祖辈。')
  }

  return ambiguous(context, ['祖父母', '外祖父母'], [], '需要补充父母性别以区分父系或母系祖辈。')
}

function getSiblingRelativeAge(reference: PersonFacts | undefined, step: KinshipStep): RelativeAge {
  return inferRelativeAgeFromFacts(reference, step.person, step.relativeAge)
}

function getParentChainGenders(steps: KinshipStep[]): Gender[] {
  return steps.map((step) => step.person.gender)
}

function hasUnknownGenderInParentChain(genders: Gender[]): boolean {
  return genders.some((gender) => gender === 'unknown')
}

function countLeadingFemales(genders: Gender[]): number {
  let count = 0
  for (const gender of genders) {
    if (gender === 'female') count += 1
    else break
  }
  return count
}

function isPurePaternalChain(genders: Gender[]): boolean {
  if (genders[0] !== 'male' || hasUnknownGenderInParentChain(genders)) return false
  if (genders.every((gender) => gender === 'male')) return true
  return (
    genders[genders.length - 1] === 'female' &&
    genders.slice(0, -1).every((gender) => gender === 'male')
  )
}

const PATERNAL_ANCESTOR_TITLES: Record<number, { male: string; female: string; aliases?: string[] }> = {
  3: { male: '曾祖父', female: '曾祖母' },
  4: { male: '高祖父', female: '高祖母' },
  5: { male: '五世祖父', female: '五世祖母', aliases: ['天祖父'] }
}

const MATERNAL_ANCESTOR_TITLES: Record<
  number,
  { male: string; female: string; maleAliases?: string[]; femaleAliases?: string[] }
> = {
  3: {
    male: '外曾祖父',
    female: '外曾祖母',
    maleAliases: ['母系曾祖父'],
    femaleAliases: ['母系曾祖母']
  },
  4: {
    male: '外高祖父',
    female: '外高祖母',
    maleAliases: ['母系高祖父'],
    femaleAliases: ['母系高祖母']
  },
  5: {
    male: '外五世祖父',
    female: '外五世祖母',
    maleAliases: ['母系五世祖父'],
    femaleAliases: ['母系五世祖母']
  }
}

const MATERNAL_DOUBLE_FEMALE_TITLES: Record<
  number,
  { male: string; female: string; maleAliases?: string[]; femaleAliases?: string[] }
> = {
  3: {
    male: '外曾外祖父',
    female: '外曾外祖母',
    maleAliases: ['母系曾外祖父'],
    femaleAliases: ['母系曾外祖母']
  }
}

function getGenderedAliases(
  gender: Gender,
  config: { maleAliases?: string[]; femaleAliases?: string[] }
): string[] {
  if (gender === 'male') return config.maleAliases || []
  if (gender === 'female') return config.femaleAliases || []
  return []
}

function resolveAmbiguousMaternalAncestor(
  context: KinshipContext,
  lastGender: Gender,
  maleCandidates: string[],
  femaleCandidates: string[],
  explanation: string
): KinshipResolution {
  if (lastGender === 'male') {
    return ambiguous(context, ['祖辈亲属', ...maleCandidates], [], explanation)
  }
  if (lastGender === 'female') {
    return ambiguous(context, ['祖辈亲属', ...femaleCandidates], [], explanation)
  }
  return ambiguous(
    context,
    ['祖辈亲属', ...maleCandidates, ...femaleCandidates],
    [],
    `${explanation}还需要补充终点性别。`
  )
}

function resolveUnknownAncestorChain(
  context: KinshipContext,
  genders: Gender[]
): KinshipResolution {
  const lastGender = genders[genders.length - 1]
  const specificCandidates =
    lastGender === 'male'
      ? ['曾祖父', '外曾祖父', '其他祖辈称谓']
      : lastGender === 'female'
        ? ['曾祖母', '外曾祖母', '其他祖辈称谓']
        : ['曾祖父母', '外曾祖父母', '其他祖辈称谓']
  return ambiguous(
    context,
    ['祖辈亲属', ...specificCandidates],
    [],
    '需要补充父系/母系方向才能确定祖辈称谓。'
  )
}

function resolveMaternalAncestorChain(
  context: KinshipContext,
  genders: Gender[],
  depth: number
): KinshipResolution {
  const leadingFemales = countLeadingFemales(genders)
  const lastGender = genders[genders.length - 1]

  if (leadingFemales === 1) {
    const rest = genders.slice(1)
    const restIsPaternalMale =
      rest.every((gender) => gender === 'male') ||
      (rest[rest.length - 1] === 'female' &&
        rest.slice(0, -1).every((gender) => gender === 'male'))

    if (restIsPaternalMale) {
      const config = MATERNAL_ANCESTOR_TITLES[depth]
      if (config) {
        return resolved(
          context,
          genderChildLabel(lastGender, config.male, config.female, '外系祖辈'),
          getGenderedAliases(lastGender, config)
        )
      }
    }
  }

  if (leadingFemales >= 2 && depth === 3) {
    const config = MATERNAL_DOUBLE_FEMALE_TITLES[depth]
    return resolved(
      context,
      genderChildLabel(lastGender, config.male, config.female, '外系祖辈'),
      getGenderedAliases(lastGender, config)
    )
  }

  if (leadingFemales >= 2 && depth >= 4) {
    return resolveAmbiguousMaternalAncestor(
      context,
      lastGender,
      depth === 5
        ? ['外五世祖父', '外五世外祖父', '母系男性五世祖辈']
        : ['外高祖父', '外高外祖父', '母系男性高祖辈'],
      depth === 5
        ? ['外五世祖母', '外五世外祖母', '母系女性五世祖辈']
        : ['外高祖母', '外高外祖母', '母系女性高祖辈'],
      '该祖辈路径涉及母系多代交叉，称谓存在地区差异，当前版本给出保守结果。'
    )
  }

  return resolveAmbiguousMaternalAncestor(
    context,
    lastGender,
    ['外曾祖父', '外曾外祖父', '母系男性祖辈'],
    ['外曾祖母', '外曾外祖母', '母系女性祖辈'],
    '该祖辈路径涉及母系方向，称谓存在地区差异，当前版本给出保守结果。'
  )
}

function resolveAncestorChain(context: KinshipContext, steps: KinshipStep[]): KinshipResolution {
  const genders = getParentChainGenders(steps)
  const depth = genders.length

  if (depth === 2) {
    return resolveParentParent(context, steps[0], steps[1])
  }

  if (depth < 3) {
    return unsupported(context)
  }

  if (hasUnknownGenderInParentChain(genders)) {
    return resolveUnknownAncestorChain(context, genders)
  }

  if (isPurePaternalChain(genders)) {
    const config = PATERNAL_ANCESTOR_TITLES[depth]
    if (!config) {
      return unsupported(context)
    }
    const lastGender = genders[genders.length - 1]
    return resolved(
      context,
      genderChildLabel(lastGender, config.male, config.female, `${depth}世祖辈`),
      config.aliases || []
    )
  }

  if (genders[0] === 'female') {
    return resolveMaternalAncestorChain(context, genders, depth)
  }

  return ambiguous(
    context,
    ['祖辈亲属', '曾祖父', '外曾祖父', '曾祖母', '外曾祖母'],
    [],
    '该祖辈路径涉及父系/母系交叉，称谓存在地区差异，当前版本给出保守结果。'
  )
}

function resolveParentSibling(
  context: KinshipContext,
  first: KinshipStep,
  second: KinshipStep
): KinshipResolution {
  const lineGender = first.person.gender
  const siblingGender = second.person.gender
  const relativeAge = getSiblingRelativeAge(first.person, second)

  if (lineGender === 'male') {
    if (siblingGender === 'male' && relativeAge === 'older') {
      return resolved(context, '伯父', ['大伯', '伯伯'])
    }
    if (siblingGender === 'male' && relativeAge === 'younger') {
      return resolved(context, '叔父', ['叔叔'])
    }
    if (siblingGender === 'female') {
      return resolved(context, '姑母', ['姑姑', '姑妈'])
    }
    if (siblingGender === 'male') {
      return ambiguous(context, ['伯父', '叔父'], [], '父系兄弟姐妹需要补充长幼信息。')
    }
    return ambiguous(context, ['伯父', '叔父', '姑母'], [], '父系兄弟姐妹需要补充性别或长幼信息。')
  }

  if (lineGender === 'female') {
    if (siblingGender === 'male') {
      return resolved(context, '舅父', ['舅舅'])
    }
    if (siblingGender === 'female') {
      return resolved(context, '姨母', ['姨妈', '阿姨'])
    }
    return ambiguous(context, ['舅父', '姨母'], [], '母系兄弟姐妹需要补充性别信息。')
  }

  if (lineGender === 'unknown') {
    if (siblingGender === 'male') {
      return ambiguous(context, ['伯父', '叔父', '舅父'], [], '需要补充第一层父母性别以区分父系或母系旁系。')
    }
    if (siblingGender === 'female') {
      return ambiguous(context, ['姑母', '姨母'], [], '需要补充第一层父母性别以区分父系或母系旁系。')
    }
    return ambiguous(context, ['父母的兄弟姐妹'], [], '父母的兄弟姐妹需要补充性别或长幼信息。')
  }

  return ambiguous(context, ['父母的兄弟姐妹'], [], '父母的兄弟姐妹需要补充性别或长幼信息。')
}

function resolveChildChild(
  context: KinshipContext,
  first: KinshipStep,
  second: KinshipStep
): KinshipResolution {
  const lineGender = first.person.gender
  const childGender = second.person.gender

  if (lineGender === 'male') {
    return resolved(
      context,
      genderChildLabel(childGender, '孙子', '孙女', '孙子女'),
      ['孙儿']
    )
  }

  if (lineGender === 'female') {
    return resolved(
      context,
      genderChildLabel(childGender, '外孙', '外孙女', '外孙子女')
    )
  }

  if (lineGender === 'unknown') {
    if (childGender === 'male') {
      return ambiguous(context, ['孙子', '外孙'], [], '需要补充第一层子女性别以区分儿系或女系孙辈。')
    }
    if (childGender === 'female') {
      return ambiguous(context, ['孙女', '外孙女'], [], '需要补充第一层子女性别以区分儿系或女系孙辈。')
    }
    return ambiguous(context, ['孙辈', '外孙辈'], [], '需要补充子女性别以区分儿系或女系孙辈。')
  }

  return ambiguous(context, ['孙辈', '外孙辈'], [], '需要补充子女性别以区分儿系或女系孙辈。')
}

function resolveSpouseSibling(
  context: KinshipContext,
  first: KinshipStep,
  second: KinshipStep
): KinshipResolution {
  const spouseGender = first.person.gender
  const siblingGender = second.person.gender
  const relativeAge = getSiblingRelativeAge(first.person, second)

  if (spouseGender === 'female') {
    if (siblingGender === 'male' && relativeAge === 'older') {
      return resolved(context, '大舅子', ['内兄'])
    }
    if (siblingGender === 'male' && relativeAge === 'younger') {
      return resolved(context, '小舅子', ['内弟'])
    }
    if (siblingGender === 'female' && relativeAge === 'older') {
      return resolved(context, '大姨子', ['姨姐'])
    }
    if (siblingGender === 'female' && relativeAge === 'younger') {
      return resolved(context, '小姨子', ['姨妹'])
    }
    if (siblingGender === 'male') {
      return ambiguous(context, ['大舅子', '小舅子'], ['内兄', '内弟'], '妻子的兄弟需要补充长幼信息。')
    }
    if (siblingGender === 'female') {
      return ambiguous(context, ['大姨子', '小姨子'], ['姨姐', '姨妹'], '妻子的姐妹需要补充长幼信息。')
    }
    return ambiguous(context, ['妻子的兄弟姐妹'], [], '妻子的兄弟姐妹需要补充性别或长幼信息。')
  }

  if (spouseGender === 'male') {
    if (siblingGender === 'male' && relativeAge === 'older') {
      return resolved(context, '大伯子', ['大伯哥'])
    }
    if (siblingGender === 'male' && relativeAge === 'younger') {
      return resolved(context, '小叔子')
    }
    if (siblingGender === 'female' && relativeAge === 'older') {
      return resolved(context, '大姑子')
    }
    if (siblingGender === 'female' && relativeAge === 'younger') {
      return resolved(context, '小姑子')
    }
    if (siblingGender === 'male') {
      return ambiguous(context, ['大伯子', '小叔子'], ['大伯哥'], '丈夫的兄弟需要补充长幼信息。')
    }
    if (siblingGender === 'female') {
      return ambiguous(context, ['大姑子', '小姑子'], [], '丈夫的姐妹需要补充长幼信息。')
    }
    return ambiguous(context, ['丈夫的兄弟姐妹'], [], '丈夫的兄弟姐妹需要补充性别或长幼信息。')
  }

  if (siblingGender === 'male' && relativeAge === 'older') {
    return ambiguous(context, ['配偶的哥哥', '大舅子', '大伯子'], [], '需要补充配偶性别以区分称谓。')
  }
  if (siblingGender === 'male' && relativeAge === 'younger') {
    return ambiguous(context, ['配偶的弟弟', '小舅子', '小叔子'], [], '需要补充配偶性别以区分称谓。')
  }
  if (siblingGender === 'female' && relativeAge === 'older') {
    return ambiguous(context, ['配偶的姐姐', '大姨子', '大姑子'], [], '需要补充配偶性别以区分称谓。')
  }
  if (siblingGender === 'female' && relativeAge === 'younger') {
    return ambiguous(context, ['配偶的妹妹', '小姨子', '小姑子'], [], '需要补充配偶性别以区分称谓。')
  }

  return ambiguous(
    context,
    ['配偶的哥哥', '配偶的弟弟', '配偶的姐姐', '配偶的妹妹'],
    [],
    '配偶兄弟姐妹的称谓因配偶性别、长幼而异，请补充信息。'
  )
}

function resolveSingleStep(context: KinshipContext, step: KinshipStep): KinshipResolution {
  const { relation, person } = step
  switch (relation) {
    case 'parent':
      return resolved(
        context,
        genderChildLabel(person.gender, '父亲', '母亲', '父母'),
        person.gender === 'male' ? ['爸爸', '爹'] : person.gender === 'female' ? ['妈妈', '娘'] : []
      )
    case 'child':
      return resolved(
        context,
        genderChildLabel(person.gender, '儿子', '女儿', '子女')
      )
    case 'spouse':
      return resolved(
        context,
        genderChildLabel(person.gender, '丈夫', '妻子', '配偶'),
        person.gender === 'male' ? ['老公'] : person.gender === 'female' ? ['老婆'] : []
      )
    case 'sibling': {
      const relativeAge = getSiblingRelativeAge(context.self, step)
      if (person.gender === 'male' && relativeAge === 'older') {
        return resolved(context, '哥哥', ['兄长'])
      }
      if (person.gender === 'male' && relativeAge === 'younger') {
        return resolved(context, '弟弟')
      }
      if (person.gender === 'female' && relativeAge === 'older') {
        return resolved(context, '姐姐', ['姊姊'])
      }
      if (person.gender === 'female' && relativeAge === 'younger') {
        return resolved(context, '妹妹')
      }
      return ambiguous(context, ['哥哥', '弟弟', '姐姐', '妹妹', '兄弟姐妹'], [], '兄弟姐妹关系需要补充性别或长幼信息。')
    }
    default:
      return unsupported(context)
  }
}

function resolveParentChain(context: KinshipContext, steps: KinshipStep[]): KinshipResolution | null {
  if (!steps.every((step) => step.relation === 'parent')) return null
  return resolveAncestorChain(context, steps)
}

function resolveTwoLayers(context: KinshipContext, steps: KinshipStep[]): KinshipResolution | null {
  const [first, second] = steps
  const key = stepKey(steps)

  if (key === 'parent>sibling') {
    return resolveParentSibling(context, first, second)
  }

  if (key === 'sibling>child') {
    const siblingGender = first.person.gender
    const childGender = second.person.gender
    if (siblingGender === 'male') {
      return resolved(
        context,
        genderChildLabel(childGender, '侄子', '侄女', '侄子女')
      )
    }
    if (siblingGender === 'female') {
      return resolved(
        context,
        genderChildLabel(childGender, '外甥', '外甥女', '外甥子女')
      )
    }
    return ambiguous(context, ['侄子', '侄女', '外甥', '外甥女'], [], '需要知道兄弟姐妹的性别才能判断称谓。')
  }

  if (key === 'sibling>spouse') {
    return resolved(context, '兄弟姐妹的配偶')
  }

  if (key === 'spouse>parent') {
    const selfGender = context.self.gender
    const parentGender = second.person.gender
    if (selfGender === 'male' && parentGender === 'male') {
      return resolved(context, '岳父', ['丈人'])
    }
    if (selfGender === 'male' && parentGender === 'female') {
      return resolved(context, '岳母', ['丈母娘'])
    }
    if (selfGender === 'female' && parentGender === 'male') {
      return resolved(context, '公公')
    }
    if (selfGender === 'female' && parentGender === 'female') {
      return resolved(context, '婆婆')
    }
    return ambiguous(context, ['公公', '婆婆', '岳父', '岳母', '配偶的父母'], [], '需要补充本人性别和对方父母性别。')
  }

  if (key === 'spouse>sibling') {
    return resolveSpouseSibling(context, first, second)
  }

  if (key === 'spouse>child') {
    return resolved(
      context,
      genderChildLabel(second.person.gender, '继子', '继女', '配偶的子女'),
      ['继子女']
    )
  }

  if (key === 'child>child') {
    return resolveChildChild(context, first, second)
  }

  return null
}

function resolveMultiLayer(context: KinshipContext, steps: KinshipStep[]): KinshipResolution | null {
  const parentChain = resolveParentChain(context, steps)
  if (parentChain) return parentChain

  const twoLayer = resolveTwoLayers(context, steps)
  if (twoLayer) return twoLayer

  return null
}

export function resolveKinship(context: KinshipContext): KinshipResolution {
  const structure = validateContextStructure(context)
  if (!structure.valid) {
    return unsupported(context, structure.reason)
  }

  if (context.steps.length === 0) {
    return resolved(context, '我', [], '这是关系路径的起点。')
  }

  if (context.steps.length === 1) {
    return resolveSingleStep(context, context.steps[0])
  }

  const multi = resolveMultiLayer(context, context.steps)
  if (multi) return multi

  return unsupported(context)
}
