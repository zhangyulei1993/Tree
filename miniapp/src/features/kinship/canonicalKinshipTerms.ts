import type { CanonicalKinshipTerm, CanonicalTermKey } from './canonicalTypes'

/** 规范称谓注册表：业务代码仅通过 key 引用，禁止散落硬编码称谓字符串。 */
export const CANONICAL_KINSHIP_TERMS: Record<CanonicalTermKey, CanonicalKinshipTerm> = {
  SELF: { key: 'SELF', canonicalTitle: '我' },

  FATHER: { key: 'FATHER', canonicalTitle: '父亲', pathKeys: ['parent:male'], targetGender: 'male' },
  MOTHER: { key: 'MOTHER', canonicalTitle: '母亲', pathKeys: ['parent:female'], targetGender: 'female' },
  SON: { key: 'SON', canonicalTitle: '儿子', pathKeys: ['child:male'], targetGender: 'male' },
  DAUGHTER: { key: 'DAUGHTER', canonicalTitle: '女儿', pathKeys: ['child:female'], targetGender: 'female' },
  HUSBAND: { key: 'HUSBAND', canonicalTitle: '丈夫', pathKeys: ['spouse:male'], targetGender: 'male' },
  WIFE: { key: 'WIFE', canonicalTitle: '妻子', pathKeys: ['spouse:female'], targetGender: 'female' },

  ELDER_BROTHER: {
    key: 'ELDER_BROTHER',
    canonicalTitle: '兄长',
    pathKeys: ['sibling:male:older'],
    targetGender: 'male',
    seniority: 'older'
  },
  YOUNGER_BROTHER: {
    key: 'YOUNGER_BROTHER',
    canonicalTitle: '弟弟',
    pathKeys: ['sibling:male:younger'],
    targetGender: 'male',
    seniority: 'younger'
  },
  ELDER_SISTER: {
    key: 'ELDER_SISTER',
    canonicalTitle: '姐姐',
    pathKeys: ['sibling:female:older'],
    targetGender: 'female',
    seniority: 'older'
  },
  YOUNGER_SISTER: {
    key: 'YOUNGER_SISTER',
    canonicalTitle: '妹妹',
    pathKeys: ['sibling:female:younger'],
    targetGender: 'female',
    seniority: 'younger'
  },

  PATERNAL_GRANDFATHER: {
    key: 'PATERNAL_GRANDFATHER',
    canonicalTitle: '祖父',
    pathKeys: ['parent:male>parent:male'],
    targetGender: 'male',
    lineageSide: 'paternal'
  },
  PATERNAL_GRANDMOTHER: {
    key: 'PATERNAL_GRANDMOTHER',
    canonicalTitle: '祖母',
    pathKeys: ['parent:male>parent:female'],
    targetGender: 'female',
    lineageSide: 'paternal'
  },
  MATERNAL_GRANDFATHER: {
    key: 'MATERNAL_GRANDFATHER',
    canonicalTitle: '外祖父',
    pathKeys: ['parent:female>parent:male'],
    targetGender: 'male',
    lineageSide: 'maternal'
  },
  MATERNAL_GRANDMOTHER: {
    key: 'MATERNAL_GRANDMOTHER',
    canonicalTitle: '外祖母',
    pathKeys: ['parent:female>parent:female'],
    targetGender: 'female',
    lineageSide: 'maternal'
  },
  GREAT_PATERNAL_GRANDFATHER: {
    key: 'GREAT_PATERNAL_GRANDFATHER',
    canonicalTitle: '曾祖父',
    pathKeys: ['parent:male>parent:male>parent:male'],
    targetGender: 'male'
  },
  GREAT_PATERNAL_GRANDMOTHER: {
    key: 'GREAT_PATERNAL_GRANDMOTHER',
    canonicalTitle: '曾祖母',
    pathKeys: ['parent:male>parent:male>parent:female'],
    targetGender: 'female'
  },
  GREAT_GREAT_PATERNAL_GRANDFATHER: {
    key: 'GREAT_GREAT_PATERNAL_GRANDFATHER',
    canonicalTitle: '高祖父',
    pathKeys: ['parent:male>parent:male>parent:male>parent:male'],
    targetGender: 'male'
  },
  GREAT_GREAT_PATERNAL_GRANDMOTHER: {
    key: 'GREAT_GREAT_PATERNAL_GRANDMOTHER',
    canonicalTitle: '高祖母',
    pathKeys: ['parent:male>parent:male>parent:male>parent:female'],
    targetGender: 'female'
  },
  FIFTH_PATERNAL_GRANDFATHER: {
    key: 'FIFTH_PATERNAL_GRANDFATHER',
    canonicalTitle: '五世祖父',
    pathKeys: ['parent:male>parent:male>parent:male>parent:male>parent:male'],
    targetGender: 'male'
  },
  FIFTH_PATERNAL_GRANDMOTHER: {
    key: 'FIFTH_PATERNAL_GRANDMOTHER',
    canonicalTitle: '五世祖母',
    pathKeys: ['parent:male>parent:male>parent:male>parent:male>parent:female'],
    targetGender: 'female'
  },
  MATERNAL_GREAT_GREAT_GRANDFATHER: {
    key: 'MATERNAL_GREAT_GREAT_GRANDFATHER',
    canonicalTitle: '外曾外祖父',
    pathKeys: ['parent:female>parent:female>parent:male'],
    targetGender: 'male'
  },
  MATERNAL_GREAT_GREAT_GRANDMOTHER: {
    key: 'MATERNAL_GREAT_GREAT_GRANDMOTHER',
    canonicalTitle: '外曾外祖母',
    pathKeys: ['parent:female>parent:female>parent:female'],
    targetGender: 'female'
  },
  MATERNAL_HIGH_GRANDMOTHER: {
    key: 'MATERNAL_HIGH_GRANDMOTHER',
    canonicalTitle: '外高祖母',
    pathKeys: [
      'parent:female>parent:female>parent:female>parent:female'
    ],
    targetGender: 'female'
  },
  MATERNAL_FIFTH_GRANDMOTHER: {
    key: 'MATERNAL_FIFTH_GRANDMOTHER',
    canonicalTitle: '外五世祖母',
    pathKeys: [
      'parent:female>parent:female>parent:female>parent:female>parent:female'
    ],
    targetGender: 'female'
  },
  GREAT_MATERNAL_GRANDFATHER: {
    key: 'GREAT_MATERNAL_GRANDFATHER',
    canonicalTitle: '外曾祖父',
    pathKeys: ['parent:female>parent:male>parent:male'],
    targetGender: 'male'
  },
  GREAT_MATERNAL_GRANDMOTHER: {
    key: 'GREAT_MATERNAL_GRANDMOTHER',
    canonicalTitle: '外曾祖母',
    pathKeys: ['parent:female>parent:male>parent:female'],
    targetGender: 'female'
  },

  PATERNAL_UNCLE_ELDER: {
    key: 'PATERNAL_UNCLE_ELDER',
    canonicalTitle: '伯父',
    pathKeys: ['parent:male>sibling:male:older'],
    targetGender: 'male',
    seniority: 'older',
    lineageSide: 'paternal'
  },
  PATERNAL_UNCLE_YOUNGER: {
    key: 'PATERNAL_UNCLE_YOUNGER',
    canonicalTitle: '叔父',
    pathKeys: ['parent:male>sibling:male:younger'],
    targetGender: 'male',
    seniority: 'younger',
    lineageSide: 'paternal'
  },
  PATERNAL_AUNT: {
    key: 'PATERNAL_AUNT',
    canonicalTitle: '姑母',
    pathKeys: ['parent:male>sibling:female:older', 'parent:male>sibling:female:younger'],
    targetGender: 'female',
    lineageSide: 'paternal'
  },
  MATERNAL_UNCLE: {
    key: 'MATERNAL_UNCLE',
    canonicalTitle: '舅父',
    pathKeys: ['parent:female>sibling:male:older', 'parent:female>sibling:male:younger'],
    targetGender: 'male',
    lineageSide: 'maternal'
  },
  MATERNAL_AUNT: {
    key: 'MATERNAL_AUNT',
    canonicalTitle: '姨母',
    pathKeys: ['parent:female>sibling:female:older', 'parent:female>sibling:female:younger'],
    targetGender: 'female',
    lineageSide: 'maternal'
  },

  NEPHEW: {
    key: 'NEPHEW',
    canonicalTitle: '侄子',
    pathKeys: [
      'sibling:male:older>child:male',
      'sibling:male:younger>child:male',
      'sibling:male:unknown>child:male'
    ],
    targetGender: 'male'
  },
  NIECE: {
    key: 'NIECE',
    canonicalTitle: '侄女',
    pathKeys: [
      'sibling:male:older>child:female',
      'sibling:male:younger>child:female',
      'sibling:male:unknown>child:female'
    ],
    targetGender: 'female'
  },
  MATERNAL_NEPHEW: {
    key: 'MATERNAL_NEPHEW',
    canonicalTitle: '外甥',
    pathKeys: [
      'sibling:female:older>child:male',
      'sibling:female:younger>child:male',
      'sibling:female:unknown>child:male'
    ],
    targetGender: 'male'
  },
  MATERNAL_NIECE: {
    key: 'MATERNAL_NIECE',
    canonicalTitle: '外甥女',
    pathKeys: [
      'sibling:female:older>child:female',
      'sibling:female:younger>child:female',
      'sibling:female:unknown>child:female'
    ],
    targetGender: 'female'
  },

  NEPHEW_GRANDSON: {
    key: 'NEPHEW_GRANDSON',
    canonicalTitle: '侄孙',
    pathKeys: [
      'sibling:male:older>child:male>child:male',
      'sibling:male:younger>child:male>child:male',
      'sibling:male:unknown>child:male>child:male'
    ],
    targetGender: 'male'
  },
  NEPHEW_GRANDDAUGHTER: {
    key: 'NEPHEW_GRANDDAUGHTER',
    canonicalTitle: '侄孙女',
    pathKeys: [
      'sibling:male:older>child:male>child:female',
      'sibling:male:younger>child:male>child:female',
      'sibling:male:unknown>child:male>child:female'
    ],
    targetGender: 'female'
  },
  MATERNAL_NEPHEW_GRANDSON: {
    key: 'MATERNAL_NEPHEW_GRANDSON',
    canonicalTitle: '外甥孙',
    pathKeys: [
      'sibling:female:older>child:male>child:male',
      'sibling:female:younger>child:male>child:male',
      'sibling:female:unknown>child:male>child:male'
    ],
    targetGender: 'male'
  },
  MATERNAL_NEPHEW_GRANDDAUGHTER: {
    key: 'MATERNAL_NEPHEW_GRANDDAUGHTER',
    canonicalTitle: '外甥孙女',
    pathKeys: [
      'sibling:female:older>child:male>child:female',
      'sibling:female:younger>child:male>child:female',
      'sibling:female:unknown>child:male>child:female'
    ],
    targetGender: 'female'
  },
  TANG_NEPHEW: {
    key: 'TANG_NEPHEW',
    canonicalTitle: '堂侄',
    pathKeys: [
      'parent:male>sibling:male:older>child:male>child:male',
      'parent:male>sibling:male:younger>child:male>child:male'
    ],
    targetGender: 'male'
  },
  MATERNAL_GRANDUNCLE_ELDER: {
    key: 'MATERNAL_GRANDUNCLE_ELDER',
    canonicalTitle: '外伯祖父',
    pathKeys: ['parent:female>parent:male>sibling:male:older'],
    targetGender: 'male',
    seniority: 'older'
  },
  MATERNAL_GRANDUNCLE_YOUNGER: {
    key: 'MATERNAL_GRANDUNCLE_YOUNGER',
    canonicalTitle: '外叔祖父',
    pathKeys: ['parent:female>parent:male>sibling:male:younger'],
    targetGender: 'male',
    seniority: 'younger'
  },
  MATERNAL_GRANDAUNT_MATERNAL: {
    key: 'MATERNAL_GRANDAUNT_MATERNAL',
    canonicalTitle: '外姨祖母',
    pathKeys: [
      'parent:female>parent:female>sibling:female:older',
      'parent:female>parent:female>sibling:female:younger'
    ],
    targetGender: 'female'
  },
  GREAT_GRANDUNCLE_PATERNAL: {
    key: 'GREAT_GRANDUNCLE_PATERNAL',
    canonicalTitle: '曾伯祖父',
    pathKeys: ['parent:male>parent:male>parent:male>sibling:male:older'],
    targetGender: 'male',
    seniority: 'older'
  },
  GREAT_GRANDUNCLE_PATERNAL_YOUNGER: {
    key: 'GREAT_GRANDUNCLE_PATERNAL_YOUNGER',
    canonicalTitle: '曾叔祖父',
    pathKeys: ['parent:male>parent:male>parent:male>sibling:male:younger'],
    targetGender: 'male',
    seniority: 'younger'
  },
  GREAT_GRANDAUNT_PATERNAL: {
    key: 'GREAT_GRANDAUNT_PATERNAL',
    canonicalTitle: '曾姑祖母',
    pathKeys: [
      'parent:male>parent:male>parent:male>sibling:female:older',
      'parent:male>parent:male>parent:male>sibling:female:younger'
    ],
    targetGender: 'female'
  },
  HIGH_GRANDUNCLE_PATERNAL: {
    key: 'HIGH_GRANDUNCLE_PATERNAL',
    canonicalTitle: '高伯祖父',
    pathKeys: ['parent:male>parent:male>parent:male>parent:male>sibling:male:older'],
    targetGender: 'male',
    seniority: 'older'
  },
  HIGH_GRANDUNCLE_PATERNAL_YOUNGER: {
    key: 'HIGH_GRANDUNCLE_PATERNAL_YOUNGER',
    canonicalTitle: '高叔祖父',
    pathKeys: ['parent:male>parent:male>parent:male>parent:male>sibling:male:younger'],
    targetGender: 'male',
    seniority: 'younger'
  },
  HIGH_GRANDAUNT_PATERNAL: {
    key: 'HIGH_GRANDAUNT_PATERNAL',
    canonicalTitle: '高姑祖母',
    pathKeys: [
      'parent:male>parent:male>parent:male>parent:male>sibling:female:older',
      'parent:male>parent:male>parent:male>parent:male>sibling:female:younger'
    ],
    targetGender: 'female'
  },
  GREAT_GRANDFATHER_PATERNAL_MIXED: {
    key: 'GREAT_GRANDFATHER_PATERNAL_MIXED',
    canonicalTitle: '曾外祖父',
    pathKeys: ['parent:male>parent:female>parent:male'],
    targetGender: 'male'
  },
  GREAT_GRANDNEPHEW: {
    key: 'GREAT_GRANDNEPHEW',
    canonicalTitle: '侄重孙',
    pathKeys: [
      'sibling:male:older>child:male>child:male>child:male',
      'sibling:male:younger>child:male>child:male>child:male'
    ],
    targetGender: 'male'
  },
  MATERNAL_GREAT_GRANDNIECE: {
    key: 'MATERNAL_GREAT_GRANDNIECE',
    canonicalTitle: '外甥重孙女',
    pathKeys: [
      'sibling:female:older>child:male>child:male>child:female',
      'sibling:female:younger>child:male>child:male>child:female'
    ],
    targetGender: 'female'
  },
  BIAO_NIECE: {
    key: 'BIAO_NIECE',
    canonicalTitle: '表侄女',
    pathKeys: ['parent:female>sibling:female:older>child:female>child:female'],
    targetGender: 'female'
  },
  TANG_GRANDNEPHEW: {
    key: 'TANG_GRANDNEPHEW',
    canonicalTitle: '堂侄孙',
    pathKeys: [
      'parent:male>sibling:male:older>child:male>child:male>child:male',
      'parent:male>sibling:male:younger>child:male>child:male>child:male'
    ],
    targetGender: 'male'
  },
  BIAO_GRANDNIECE: {
    key: 'BIAO_GRANDNIECE',
    canonicalTitle: '表侄孙女',
    pathKeys: ['parent:female>sibling:female:older>child:female>child:female>child:female'],
    targetGender: 'female'
  },

  TANG_ELDER_BROTHER: {
    key: 'TANG_ELDER_BROTHER',
    canonicalTitle: '堂兄',
    pathKeys: ['parent:male>sibling:male:older>child:male'],
    targetGender: 'male',
    seniority: 'older'
  },
  TANG_YOUNGER_BROTHER: {
    key: 'TANG_YOUNGER_BROTHER',
    canonicalTitle: '堂弟',
    pathKeys: ['parent:male>sibling:male:younger>child:male'],
    targetGender: 'male',
    seniority: 'younger'
  },
  TANG_ELDER_SISTER: {
    key: 'TANG_ELDER_SISTER',
    canonicalTitle: '堂姐',
    pathKeys: ['parent:male>sibling:male:older>child:female'],
    targetGender: 'female',
    seniority: 'older'
  },
  TANG_YOUNGER_SISTER: {
    key: 'TANG_YOUNGER_SISTER',
    canonicalTitle: '堂妹',
    pathKeys: ['parent:male>sibling:male:younger>child:female'],
    targetGender: 'female',
    seniority: 'younger'
  },
  BIAO_ELDER_BROTHER: {
    key: 'BIAO_ELDER_BROTHER',
    canonicalTitle: '表兄',
    pathKeys: [
      'parent:female>sibling:male:older>child:male',
      'parent:female>sibling:female:older>child:male'
    ],
    targetGender: 'male',
    seniority: 'older'
  },
  BIAO_YOUNGER_BROTHER: {
    key: 'BIAO_YOUNGER_BROTHER',
    canonicalTitle: '表弟',
    pathKeys: [
      'parent:female>sibling:male:younger>child:male',
      'parent:female>sibling:female:younger>child:male'
    ],
    targetGender: 'male',
    seniority: 'younger'
  },
  BIAO_ELDER_SISTER: {
    key: 'BIAO_ELDER_SISTER',
    canonicalTitle: '表姐',
    pathKeys: [
      'parent:female>sibling:male:older>child:female',
      'parent:female>sibling:female:older>child:female'
    ],
    targetGender: 'female',
    seniority: 'older'
  },
  BIAO_YOUNGER_SISTER: {
    key: 'BIAO_YOUNGER_SISTER',
    canonicalTitle: '表妹',
    pathKeys: [
      'parent:female>sibling:male:younger>child:female',
      'parent:female>sibling:female:younger>child:female'
    ],
    targetGender: 'female',
    seniority: 'younger'
  },

  PATERNAL_UNCLE_WIFE_ELDER: {
    key: 'PATERNAL_UNCLE_WIFE_ELDER',
    canonicalTitle: '伯母',
    targetGender: 'female'
  },
  PATERNAL_UNCLE_WIFE_YOUNGER: {
    key: 'PATERNAL_UNCLE_WIFE_YOUNGER',
    canonicalTitle: '婶母',
    targetGender: 'female'
  },
  PATERNAL_AUNT_HUSBAND: { key: 'PATERNAL_AUNT_HUSBAND', canonicalTitle: '姑父', targetGender: 'male' },
  MATERNAL_UNCLE_WIFE: {
    key: 'MATERNAL_UNCLE_WIFE',
    canonicalTitle: '舅母',
    pathKeys: [
      'parent:female>sibling:male:older>spouse:female',
      'parent:female>sibling:male:younger>spouse:female'
    ],
    targetGender: 'female'
  },
  MATERNAL_AUNT_HUSBAND: { key: 'MATERNAL_AUNT_HUSBAND', canonicalTitle: '姨父', targetGender: 'male' },
  DAUGHTER_IN_LAW: {
    key: 'DAUGHTER_IN_LAW',
    canonicalTitle: '儿媳',
    pathKeys: ['child:male>spouse:female'],
    targetGender: 'female'
  },
  SON_IN_LAW: {
    key: 'SON_IN_LAW',
    canonicalTitle: '女婿',
    pathKeys: ['child:female>spouse:male'],
    targetGender: 'male'
  },
  ELDER_BROTHER_WIFE: {
    key: 'ELDER_BROTHER_WIFE',
    canonicalTitle: '嫂子',
    pathKeys: ['sibling:male:older>spouse:female'],
    targetGender: 'female'
  },
  YOUNGER_BROTHER_WIFE: {
    key: 'YOUNGER_BROTHER_WIFE',
    canonicalTitle: '弟媳',
    pathKeys: ['sibling:male:younger>spouse:female'],
    targetGender: 'female'
  },
  ELDER_SISTER_HUSBAND: {
    key: 'ELDER_SISTER_HUSBAND',
    canonicalTitle: '姐夫',
    pathKeys: ['sibling:female:older>spouse:male'],
    targetGender: 'male'
  },
  YOUNGER_SISTER_HUSBAND: {
    key: 'YOUNGER_SISTER_HUSBAND',
    canonicalTitle: '妹夫',
    pathKeys: ['sibling:female:younger>spouse:male'],
    targetGender: 'male'
  },
  GRANDSON: {
    key: 'GRANDSON',
    canonicalTitle: '孙子',
    pathKeys: ['child:male>child:male'],
    targetGender: 'male'
  },
  GRANDDAUGHTER: {
    key: 'GRANDDAUGHTER',
    canonicalTitle: '孙女',
    pathKeys: ['child:male>child:female'],
    targetGender: 'female'
  },
  MATERNAL_GRANDSON: { key: 'MATERNAL_GRANDSON', canonicalTitle: '外孙', targetGender: 'male' },
  MATERNAL_GRANDDAUGHTER: { key: 'MATERNAL_GRANDDAUGHTER', canonicalTitle: '外孙女', targetGender: 'female' },
  GREAT_GRANDSON: { key: 'GREAT_GRANDSON', canonicalTitle: '曾孙', targetGender: 'male' },
  GREAT_GRANDDAUGHTER: { key: 'GREAT_GRANDDAUGHTER', canonicalTitle: '曾孙女', targetGender: 'female' },
  MATERNAL_GREAT_GRANDSON: { key: 'MATERNAL_GREAT_GRANDSON', canonicalTitle: '外曾孙', targetGender: 'male' },
  MATERNAL_GREAT_GRANDDAUGHTER: {
    key: 'MATERNAL_GREAT_GRANDDAUGHTER',
    canonicalTitle: '外曾孙女',
    targetGender: 'female'
  },

  PATERNAL_GRANDUNCLE_ELDER: { key: 'PATERNAL_GRANDUNCLE_ELDER', canonicalTitle: '伯祖父', targetGender: 'male', seniority: 'older' },
  PATERNAL_GRANDUNCLE_YOUNGER: { key: 'PATERNAL_GRANDUNCLE_YOUNGER', canonicalTitle: '叔祖父', targetGender: 'male', seniority: 'younger' },
  PATERNAL_GRANDAUNT: { key: 'PATERNAL_GRANDAUNT', canonicalTitle: '姑祖母', targetGender: 'female' },
  MATERNAL_GRANDUNCLE: { key: 'MATERNAL_GRANDUNCLE', canonicalTitle: '舅祖父', targetGender: 'male' },
  MATERNAL_GRANDAUNT: { key: 'MATERNAL_GRANDAUNT', canonicalTitle: '姨祖母', targetGender: 'female' },

  NEPHEW_WIFE: { key: 'NEPHEW_WIFE', canonicalTitle: '侄媳', targetGender: 'female' },
  NEPHEW_HUSBAND: { key: 'NEPHEW_HUSBAND', canonicalTitle: '侄女婿', targetGender: 'male' },
  MATERNAL_NEPHEW_WIFE: { key: 'MATERNAL_NEPHEW_WIFE', canonicalTitle: '甥媳', targetGender: 'female' },
  MATERNAL_NEPHEW_HUSBAND: { key: 'MATERNAL_NEPHEW_HUSBAND', canonicalTitle: '甥女婿', targetGender: 'male' },
  GRANDSON_WIFE: { key: 'GRANDSON_WIFE', canonicalTitle: '孙媳', targetGender: 'female' },
  GRANDDAUGHTER_HUSBAND: { key: 'GRANDDAUGHTER_HUSBAND', canonicalTitle: '孙女婿', targetGender: 'male' },
  TANG_NEPHEW_WIFE: { key: 'TANG_NEPHEW_WIFE', canonicalTitle: '堂侄媳', targetGender: 'female' },
  TANG_NEPHEW_HUSBAND: { key: 'TANG_NEPHEW_HUSBAND', canonicalTitle: '堂侄女婿', targetGender: 'male' },
  BIAO_NEPHEW_WIFE: { key: 'BIAO_NEPHEW_WIFE', canonicalTitle: '表侄媳', targetGender: 'female' },
  BIAO_NEPHEW_HUSBAND: { key: 'BIAO_NEPHEW_HUSBAND', canonicalTitle: '表侄女婿', targetGender: 'male' },

  TANG_NIECE: { key: 'TANG_NIECE', canonicalTitle: '堂侄女', targetGender: 'female' },
  TANG_AUNT: { key: 'TANG_AUNT', canonicalTitle: '堂姑', targetGender: 'female' },
  TANG_UNCLE_ELDER: { key: 'TANG_UNCLE_ELDER', canonicalTitle: '堂伯', targetGender: 'male', seniority: 'older' },
  TANG_UNCLE_YOUNGER: { key: 'TANG_UNCLE_YOUNGER', canonicalTitle: '堂叔', targetGender: 'male', seniority: 'younger' },
  TANG_GRANDNIECE: { key: 'TANG_GRANDNIECE', canonicalTitle: '堂侄孙女', targetGender: 'female' },
  BIAO_NEPHEW: { key: 'BIAO_NEPHEW', canonicalTitle: '表侄', targetGender: 'male' },
  BIAO_UNCLE_ELDER: { key: 'BIAO_UNCLE_ELDER', canonicalTitle: '表伯', targetGender: 'male', seniority: 'older' },
  BIAO_UNCLE_YOUNGER: { key: 'BIAO_UNCLE_YOUNGER', canonicalTitle: '表叔', targetGender: 'male', seniority: 'younger' },
  BIAO_AUNT: { key: 'BIAO_AUNT', canonicalTitle: '表姑', targetGender: 'female' },
  BIAO_MATERNAL_AUNT: { key: 'BIAO_MATERNAL_AUNT', canonicalTitle: '表姨', targetGender: 'female' },
  BIAO_MATERNAL_UNCLE: { key: 'BIAO_MATERNAL_UNCLE', canonicalTitle: '表舅', targetGender: 'male' },
  BIAO_NEPHEW_GRANDSON: { key: 'BIAO_NEPHEW_GRANDSON', canonicalTitle: '表侄孙', targetGender: 'male' },
  XUAN_GRANDSON: { key: 'XUAN_GRANDSON', canonicalTitle: '玄孙', targetGender: 'male' },
  XUAN_GRANDDAUGHTER: { key: 'XUAN_GRANDDAUGHTER', canonicalTitle: '玄孙女', targetGender: 'female' },
  MATERNAL_XUAN_GRANDSON: { key: 'MATERNAL_XUAN_GRANDSON', canonicalTitle: '外玄孙', targetGender: 'male' },
  MATERNAL_XUAN_GRANDDAUGHTER: {
    key: 'MATERNAL_XUAN_GRANDDAUGHTER',
    canonicalTitle: '外玄孙女',
    targetGender: 'female'
  },
  MATERNAL_HIGH_GRANDFATHER: {
    key: 'MATERNAL_HIGH_GRANDFATHER',
    canonicalTitle: '外高祖父',
    targetGender: 'male'
  },
  GREAT_GRAND_MATERNAL_UNCLE: {
    key: 'GREAT_GRAND_MATERNAL_UNCLE',
    canonicalTitle: '曾舅祖父',
    targetGender: 'male'
  },
  GREAT_GRAND_MATERNAL_AUNT: {
    key: 'GREAT_GRAND_MATERNAL_AUNT',
    canonicalTitle: '曾姨祖母',
    targetGender: 'female'
  },
  GREAT_GRAND_MATERNAL_GRANDMOTHER: {
    key: 'GREAT_GRAND_MATERNAL_GRANDMOTHER',
    canonicalTitle: '曾外祖母',
    targetGender: 'female'
  },
  NEPHEW_MATERNAL_GRANDSON: { key: 'NEPHEW_MATERNAL_GRANDSON', canonicalTitle: '侄外孙', targetGender: 'male' },
  NEPHEW_MATERNAL_GRANDDAUGHTER: {
    key: 'NEPHEW_MATERNAL_GRANDDAUGHTER',
    canonicalTitle: '侄外孙女',
    targetGender: 'female'
  },
  MATERNAL_NEPHEW_MATERNAL_GRANDSON: {
    key: 'MATERNAL_NEPHEW_MATERNAL_GRANDSON',
    canonicalTitle: '外甥外孙',
    targetGender: 'male'
  },
  MATERNAL_NEPHEW_MATERNAL_GRANDDAUGHTER: {
    key: 'MATERNAL_NEPHEW_MATERNAL_GRANDDAUGHTER',
    canonicalTitle: '外甥外孙女',
    targetGender: 'female'
  },
  MATERNAL_GREAT_GRANDUNCLE_ELDER: {
    key: 'MATERNAL_GREAT_GRANDUNCLE_ELDER',
    canonicalTitle: '外曾伯祖父',
    targetGender: 'male',
    seniority: 'older'
  },
  MATERNAL_GREAT_GRANDUNCLE_YOUNGER: {
    key: 'MATERNAL_GREAT_GRANDUNCLE_YOUNGER',
    canonicalTitle: '外曾叔祖父',
    targetGender: 'male',
    seniority: 'younger'
  },
  MATERNAL_GREAT_GRANDAUNT: {
    key: 'MATERNAL_GREAT_GRANDAUNT',
    canonicalTitle: '外曾姑祖母',
    targetGender: 'female'
  },
  MATERNAL_GRANDAUNT_PATERNAL: {
    key: 'MATERNAL_GRANDAUNT_PATERNAL',
    canonicalTitle: '外姑祖母',
    targetGender: 'female'
  },
  MATERNAL_GRANDUNCLE_MATERNAL: {
    key: 'MATERNAL_GRANDUNCLE_MATERNAL',
    canonicalTitle: '外舅祖父',
    targetGender: 'male'
  },
  TANG_ELDER_COUSIN_WIFE: {
    key: 'TANG_ELDER_COUSIN_WIFE',
    canonicalTitle: '堂嫂',
    pathKeys: ['parent:male>sibling:male:older>child:male>spouse:female'],
    targetGender: 'female'
  },
  TANG_YOUNGER_COUSIN_WIFE: {
    key: 'TANG_YOUNGER_COUSIN_WIFE',
    canonicalTitle: '堂弟媳',
    pathKeys: ['parent:male>sibling:male:younger>child:male>spouse:female'],
    targetGender: 'female'
  },
  TANG_ELDER_COUSIN_HUSBAND: {
    key: 'TANG_ELDER_COUSIN_HUSBAND',
    canonicalTitle: '堂姐夫',
    pathKeys: ['parent:male>sibling:male:older>child:female>spouse:male'],
    targetGender: 'male'
  },
  TANG_YOUNGER_COUSIN_HUSBAND: {
    key: 'TANG_YOUNGER_COUSIN_HUSBAND',
    canonicalTitle: '堂妹夫',
    pathKeys: ['parent:male>sibling:male:younger>child:female>spouse:male'],
    targetGender: 'male'
  },
  BIAO_ELDER_COUSIN_WIFE: {
    key: 'BIAO_ELDER_COUSIN_WIFE',
    canonicalTitle: '表嫂',
    pathKeys: [
      'parent:female>sibling:male:older>child:male>spouse:female',
      'parent:male>sibling:female:older>child:male>spouse:female'
    ],
    targetGender: 'female'
  },
  BIAO_YOUNGER_COUSIN_WIFE: {
    key: 'BIAO_YOUNGER_COUSIN_WIFE',
    canonicalTitle: '表弟媳',
    pathKeys: [
      'parent:female>sibling:male:younger>child:male>spouse:female',
      'parent:male>sibling:female:younger>child:male>spouse:female'
    ],
    targetGender: 'female'
  },
  BIAO_ELDER_COUSIN_HUSBAND: {
    key: 'BIAO_ELDER_COUSIN_HUSBAND',
    canonicalTitle: '表姐夫',
    pathKeys: [
      'parent:female>sibling:female:older>child:female>spouse:male',
      'parent:male>sibling:female:older>child:female>spouse:male'
    ],
    targetGender: 'male'
  },
  BIAO_YOUNGER_COUSIN_HUSBAND: {
    key: 'BIAO_YOUNGER_COUSIN_HUSBAND',
    canonicalTitle: '表妹夫',
    pathKeys: [
      'parent:female>sibling:female:younger>child:female>spouse:male',
      'parent:male>sibling:female:younger>child:female>spouse:male'
    ],
    targetGender: 'male'
  },
  FIFTH_GRANDUNCLE_ELDER: {
    key: 'FIFTH_GRANDUNCLE_ELDER',
    canonicalTitle: '五世伯祖父',
    targetGender: 'male',
    seniority: 'older'
  },
  PATERNAL_GRANDUNCLE_WIFE_ELDER: { key: 'PATERNAL_GRANDUNCLE_WIFE_ELDER', canonicalTitle: '伯祖母', targetGender: 'female' },
  PATERNAL_GRANDUNCLE_WIFE_YOUNGER: {
    key: 'PATERNAL_GRANDUNCLE_WIFE_YOUNGER',
    canonicalTitle: '叔祖母',
    targetGender: 'female'
  },
  PATERNAL_GRANDAUNT_HUSBAND: { key: 'PATERNAL_GRANDAUNT_HUSBAND', canonicalTitle: '姑祖父', targetGender: 'male' },
  MATERNAL_GRANDUNCLE_WIFE_COLLATERAL: {
    key: 'MATERNAL_GRANDUNCLE_WIFE_COLLATERAL',
    canonicalTitle: '舅祖母',
    targetGender: 'female'
  },
  MATERNAL_GRANDAUNT_HUSBAND: { key: 'MATERNAL_GRANDAUNT_HUSBAND', canonicalTitle: '姨祖父', targetGender: 'male' },
  GREAT_GRANDUNCLE_WIFE_ELDER: { key: 'GREAT_GRANDUNCLE_WIFE_ELDER', canonicalTitle: '曾伯祖母', targetGender: 'female' },
  GREAT_GRANDUNCLE_WIFE_YOUNGER: {
    key: 'GREAT_GRANDUNCLE_WIFE_YOUNGER',
    canonicalTitle: '曾叔祖母',
    targetGender: 'female'
  },
  MATERNAL_FIFTH_GRANDFATHER: {
    key: 'MATERNAL_FIFTH_GRANDFATHER',
    canonicalTitle: '外五世祖父',
    pathKeys: [
      'parent:female>parent:female>parent:female>parent:female>parent:male'
    ],
    targetGender: 'male'
  }
}

const PATH_KEY_TO_CANONICAL_KEY: Record<string, CanonicalTermKey> = {}

for (const term of Object.values(CANONICAL_KINSHIP_TERMS)) {
  for (const pathKey of term.pathKeys || []) {
    PATH_KEY_TO_CANONICAL_KEY[pathKey] = term.key
  }
}

export function getCanonicalTerm(key: CanonicalTermKey): CanonicalKinshipTerm | undefined {
  return CANONICAL_KINSHIP_TERMS[key]
}

export function getCanonicalTitle(key: CanonicalTermKey): string | undefined {
  return CANONICAL_KINSHIP_TERMS[key]?.canonicalTitle
}

export function lookupCanonicalKeyByPathKey(pathKey: string): CanonicalTermKey | undefined {
  return PATH_KEY_TO_CANONICAL_KEY[pathKey]
}

export function buildCanonicalResult(
  key: CanonicalTermKey,
  pathDescription: string,
  rankLabel?: string
): import('./canonicalTypes').CanonicalKinshipResult {
  const term = CANONICAL_KINSHIP_TERMS[key]
  return {
    canonicalTitle: term.canonicalTitle,
    rankLabel,
    unsupportedCanonicalTitle: false,
    incompleteInfo: false,
    pathDescription
  }
}

export function buildUnsupportedCanonicalResult(
  pathDescription: string,
  options: { incompleteInfo?: boolean; explanation?: string } = {}
): import('./canonicalTypes').CanonicalKinshipResult {
  return {
    unsupportedCanonicalTitle: true,
    incompleteInfo: Boolean(options.incompleteInfo),
    pathDescription,
    explanation: options.explanation
  }
}
