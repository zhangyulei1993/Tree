export const SEED_UNLOCATED_NAME = '张待安'

export function buildSeedPlan({ founderMemberId, founderDisplayName = '创始人' }) {
  if (!founderMemberId) {
    throw new Error('founderMemberId is required')
  }

  const steps = [
    {
      key: 'father',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_FATHER',
      baseMemberId: founderMemberId,
      baseLabel: founderDisplayName,
      newMember: {
        name: '张志远',
        gender: 'MALE',
        memberType: 'LINEAGE_MEMBER',
        birthYear: 1995
      },
      relationships: ['PARENT_CHILD:张志远 → 创始人']
    },
    {
      key: 'mother',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_MOTHER',
      baseMemberId: founderMemberId,
      baseLabel: founderDisplayName,
      newMember: {
        name: '陈慧',
        gender: 'FEMALE',
        memberType: 'SPOUSE',
        birthYear: 1997
      },
      relationships: ['PARENT_CHILD:陈慧 → 创始人', 'SPOUSE:张志远 ↔ 陈慧']
    },
    {
      key: 'grandfather',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_FATHER',
      baseMemberId: '$father',
      baseLabel: '张志远',
      newMember: {
        name: '张建国',
        gender: 'MALE',
        memberType: 'LINEAGE_MEMBER',
        birthYear: 1968
      },
      relationships: ['PARENT_CHILD:张建国 → 张志远']
    },
    {
      key: 'grandmother',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_MOTHER',
      baseMemberId: '$father',
      baseLabel: '张志远',
      newMember: {
        name: '王秀兰',
        gender: 'FEMALE',
        memberType: 'SPOUSE',
        birthYear: 1970
      },
      relationships: ['PARENT_CHILD:王秀兰 → 张志远', 'SPOUSE:张建国 ↔ 王秀兰']
    },
    {
      key: 'great-grandfather',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_FATHER',
      baseMemberId: '$grandfather',
      baseLabel: '张建国',
      newMember: {
        name: '张德厚',
        gender: 'MALE',
        memberType: 'LINEAGE_MEMBER',
        birthYear: 1940
      },
      relationships: ['PARENT_CHILD:张德厚 → 张建国']
    },
    {
      key: 'great-grandmother',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_MOTHER',
      baseMemberId: '$grandfather',
      baseLabel: '张建国',
      newMember: {
        name: '李淑珍',
        gender: 'FEMALE',
        memberType: 'SPOUSE',
        birthYear: 1942
      },
      relationships: ['PARENT_CHILD:李淑珍 → 张建国', 'SPOUSE:张德厚 ↔ 李淑珍']
    },
    {
      key: 'uncle',
      action: 'CREATE_RELATIONSHIP',
      addType: 'ADD_SIBLING',
      baseMemberId: '$father',
      baseLabel: '张志远',
      newMember: {
        name: '张叔平',
        gender: 'MALE',
        birthYear: 1972
      },
      relationships: ['PARENT_CHILD:张建国 → 张叔平', 'PARENT_CHILD:王秀兰 → 张叔平']
    },
    {
      key: 'unlocated',
      action: 'CREATE_MEMBER',
      newMember: {
        name: SEED_UNLOCATED_NAME,
        gender: 'MALE',
        memberType: 'LINEAGE_MEMBER',
        birthYear: 1999
      },
      relationships: []
    }
  ]

  const plannedNames = steps.map((step) => step.newMember.name)
  if (plannedNames.includes(founderDisplayName)) {
    throw new Error(`seed plan must not recreate founder name: ${founderDisplayName}`)
  }
  if (new Set(plannedNames).size !== plannedNames.length) {
    throw new Error(`seed plan contains duplicate names: ${plannedNames.join(', ')}`)
  }

  return {
    founderMemberId,
    founderDisplayName,
    steps,
    expected: {
      memberCount: 1 + plannedNames.length,
      relationshipCount: 11,
      coupleUnits: 3,
      unlocatedCount: 1,
      createdMemberCount: plannedNames.length
    }
  }
}

export function summarizeSeedPlan(plan) {
  return {
    founderMemberId: plan.founderMemberId,
    founderDisplayName: plan.founderDisplayName,
    nodesToCreate: plan.steps.map((step) => ({
      key: step.key,
      action: step.action,
      name: step.newMember.name,
      memberType: step.newMember.memberType || 'LINEAGE_MEMBER',
      addType: step.addType || null,
      base: step.baseLabel || null
    })),
    expected: plan.expected
  }
}
