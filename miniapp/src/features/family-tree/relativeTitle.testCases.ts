import type { TreeEdge, TreeNode } from '@/types/api'

import { buildRelGraph } from './relativeGraph'
import { buildTitleCacheKey, clearRelativeTitleCache, setCachedRelativeTitle } from './relativeTitleCache'
import { buildKinshipTitleMap, getRelativeTitle } from './relativeTitle'

export interface RelativeTitleTestCase {
  name: string
  viewerMemberId: number
  targetMemberId: number
  nodes: TreeNode[]
  edges: TreeEdge[]
  graphVersion?: number
  expected: string
}

function n(
  memberId: number,
  displayName: string,
  gender: 'MALE' | 'FEMALE',
  birthDate: string
): TreeNode {
  return {
    memberId,
    displayName,
    gender,
    birthDate,
    memberType: 'NORMAL',
    userBindingState: 'NOT_REQUIRED',
    canExpand: false
  }
}

function pc(from: number, to: number, id: number): TreeEdge {
  return {
    relationshipId: id,
    fromMemberId: from,
    toMemberId: to,
    relationshipType: 'PARENT_CHILD'
  }
}

function sp(a: number, b: number, id: number): TreeEdge {
  return {
    relationshipId: id,
    fromMemberId: a,
    toMemberId: b,
    relationshipType: 'SPOUSE'
  }
}

/** 标准三代张氏家族（用于多数用例） */
export function buildZhangFamilyFixture(): {
  nodes: TreeNode[]
  edges: TreeEdge[]
  ids: Record<string, number>
} {
  const ids = {
    zengfu: 1,
    zengmu: 2,
    gong: 3,
    gongSpouse: 4,
    fu: 5,
    mu: 6,
    waiGong: 25,
    waiPo: 26,
    shufu: 7,
    gufu: 8,
    me: 10,
    spouse: 11,
    son: 12,
    daughter: 13,
    gongBro1: 20,
    gongBro2: 21,
    gongSis: 22,
    muBro: 23,
    muSis: 24,
    tangGe: 30,
    tangDi: 31,
    biaoGe: 32,
    nephew: 40,
    niece: 41,
    waiSheng: 42,
    grandson: 50,
    granddaughter: 51,
    greatGrandson: 52
  }

  const nodes: TreeNode[] = [
    n(ids.zengfu, '曾祖父', 'MALE', '1900-01-01'),
    n(ids.zengmu, '曾祖母', 'FEMALE', '1902-02-02'),
    n(ids.gong, '祖父', 'MALE', '1928-03-12'),
    n(ids.gongSpouse, '祖母', 'FEMALE', '1931-09-20'),
    n(ids.gongBro1, '大伯祖父', 'MALE', '1925-01-01'),
    n(ids.gongBro2, '三叔祖父', 'MALE', '1935-06-06'),
    n(ids.gongSis, '姑祖母', 'FEMALE', '1930-03-03'),
    n(ids.fu, '父亲', 'MALE', '1952-05-04'),
    n(ids.shufu, '叔父', 'MALE', '1958-02-09'),
    n(ids.gufu, '姑母', 'FEMALE', '1960-04-04'),
    n(ids.mu, '母亲', 'FEMALE', '1954-08-16'),
    n(ids.waiGong, '外祖父', 'MALE', '1925-01-01'),
    n(ids.waiPo, '外祖母', 'FEMALE', '1927-03-03'),
    n(ids.muBro, '舅父', 'MALE', '1953-03-03'),
    n(ids.muSis, '姨母', 'FEMALE', '1956-05-05'),
    n(ids.me, '我', 'MALE', '2004-03-16'),
    n(ids.spouse, '妻子', 'FEMALE', '2005-05-29'),
    n(ids.son, '儿子', 'MALE', '2028-02-14'),
    n(ids.daughter, '女儿', 'FEMALE', '2030-08-08'),
    n(ids.tangGe, '堂哥', 'MALE', '2000-01-01'),
    n(ids.tangDi, '堂弟', 'MALE', '2010-01-01'),
    n(ids.biaoGe, '表哥', 'MALE', '2002-02-02'),
    n(ids.nephew, '侄子', 'MALE', '2032-06-06'),
    n(ids.niece, '侄女', 'FEMALE', '2033-11-11'),
    n(ids.waiSheng, '外甥', 'MALE', '2034-01-01'),
    n(ids.grandson, '孙子', 'MALE', '2050-01-01'),
    n(ids.granddaughter, '孙女', 'FEMALE', '2051-02-02'),
    n(ids.greatGrandson, '曾孙', 'MALE', '2070-01-01')
  ]

  let rid = 1
  const edges: TreeEdge[] = [
    pc(ids.zengfu, ids.gong, rid++),
    pc(ids.zengmu, ids.gong, rid++),
    pc(ids.zengfu, ids.gongBro1, rid++),
    pc(ids.zengmu, ids.gongBro1, rid++),
    pc(ids.zengfu, ids.gongBro2, rid++),
    pc(ids.zengmu, ids.gongBro2, rid++),
    pc(ids.zengfu, ids.gongSis, rid++),
    pc(ids.zengmu, ids.gongSis, rid++),
    pc(ids.gong, ids.fu, rid++),
    pc(ids.gongSpouse, ids.fu, rid++),
    pc(ids.gong, ids.shufu, rid++),
    pc(ids.gongSpouse, ids.shufu, rid++),
    pc(ids.gong, ids.gufu, rid++),
    pc(ids.gongSpouse, ids.gufu, rid++),
    pc(ids.fu, ids.me, rid++),
    pc(ids.mu, ids.me, rid++),
    pc(ids.waiGong, ids.mu, rid++),
    pc(ids.waiPo, ids.mu, rid++),
    pc(ids.waiGong, ids.muBro, rid++),
    pc(ids.waiPo, ids.muBro, rid++),
    pc(ids.waiGong, ids.muSis, rid++),
    pc(ids.waiPo, ids.muSis, rid++),
    pc(ids.muBro, ids.biaoGe, rid++),
    sp(ids.gong, ids.gongSpouse, rid++),
    sp(ids.fu, ids.mu, rid++),
    sp(ids.me, ids.spouse, rid++),
    pc(ids.me, ids.son, rid++),
    pc(ids.spouse, ids.son, rid++),
    pc(ids.me, ids.daughter, rid++),
    pc(ids.spouse, ids.daughter, rid++),
    pc(ids.shufu, ids.tangGe, rid++),
    pc(ids.shufu, ids.tangDi, rid++),
    pc(ids.muBro, ids.biaoGe, rid++),
    pc(ids.son, ids.grandson, rid++),
    pc(ids.son, ids.granddaughter, rid++),
    pc(ids.grandson, ids.greatGrandson, rid++)
  ]

  return { nodes, edges, ids }
}

/** staging familyId=12 结构近似（memberId 为测试用虚拟 ID） */
export function buildFamily12Fixture(): {
  nodes: TreeNode[]
  edges: TreeEdge[]
  ids: Record<string, number>
} {
  const ids = {
    zhangDehou: 101,
    liSuqin: 102,
    zhangGuoan: 103,
    wangXiulan: 104,
    zhangGuoping: 105,
    zhangGuohua: 106,
    zhaoMeihua: 107,
    zhangJianguo: 108,
    zhangJianjun: 109,
    zhangJianfang: 110,
    zhangJianmin: 111,
    liuYaqin: 112,
    zhangMingyuan: 113,
    chenYuwei: 114,
    zhangMingcheng: 115,
    zhangMinghui: 116,
    zhangSiyuan: 117,
    zhangSining: 118
  }

  const nodes: TreeNode[] = [
    n(ids.zhangDehou, '张德厚', 'MALE', '1928-03-12'),
    n(ids.liSuqin, '李素琴', 'FEMALE', '1931-09-20'),
    n(ids.zhangGuoan, '张国安', 'MALE', '1952-05-04'),
    n(ids.wangXiulan, '王秀兰', 'FEMALE', '1954-08-16'),
    n(ids.zhangGuoping, '张国平', 'MALE', '1955-11-18'),
    n(ids.zhangGuohua, '张国华', 'FEMALE', '1958-02-09'),
    n(ids.zhaoMeihua, '赵美华', 'FEMALE', '1957-12-10'),
    n(ids.zhangJianguo, '张建国', 'MALE', '1978-01-23'),
    n(ids.zhangJianjun, '张建军', 'MALE', '1981-06-03'),
    n(ids.zhangJianfang, '张建芳', 'FEMALE', '1984-04-28'),
    n(ids.zhangJianmin, '张建民', 'MALE', '1982-07-15'),
    n(ids.liuYaqin, '刘雅琴', 'FEMALE', '1980-10-11'),
    n(ids.zhangMingyuan, '张明远', 'MALE', '2004-03-16'),
    n(ids.chenYuwei, '陈雨薇', 'FEMALE', '2005-05-29'),
    n(ids.zhangMingcheng, '张明诚', 'MALE', '2007-09-01'),
    n(ids.zhangMinghui, '张明慧', 'FEMALE', '2010-12-22'),
    n(ids.zhangSiyuan, '张思源', 'MALE', '2028-02-14'),
    n(ids.zhangSining, '张思宁', 'FEMALE', '2030-08-08')
  ]

  let rid = 1000
  const edges: TreeEdge[] = [
    sp(ids.zhangDehou, ids.liSuqin, rid++),
    sp(ids.zhangGuoan, ids.wangXiulan, rid++),
    sp(ids.zhangGuoping, ids.zhaoMeihua, rid++),
    sp(ids.zhangJianguo, ids.liuYaqin, rid++),
    sp(ids.zhangMingyuan, ids.chenYuwei, rid++),
    pc(ids.zhangDehou, ids.zhangGuoan, rid++),
    pc(ids.liSuqin, ids.zhangGuoan, rid++),
    pc(ids.zhangDehou, ids.zhangGuoping, rid++),
    pc(ids.liSuqin, ids.zhangGuoping, rid++),
    pc(ids.zhangDehou, ids.zhangGuohua, rid++),
    pc(ids.liSuqin, ids.zhangGuohua, rid++),
    pc(ids.zhangGuoan, ids.zhangJianguo, rid++),
    pc(ids.wangXiulan, ids.zhangJianguo, rid++),
    pc(ids.zhangGuoan, ids.zhangJianjun, rid++),
    pc(ids.wangXiulan, ids.zhangJianjun, rid++),
    pc(ids.zhangGuoan, ids.zhangJianfang, rid++),
    pc(ids.wangXiulan, ids.zhangJianfang, rid++),
    pc(ids.zhangGuoping, ids.zhangJianmin, rid++),
    pc(ids.zhaoMeihua, ids.zhangJianmin, rid++),
    pc(ids.zhangJianguo, ids.zhangMingyuan, rid++),
    pc(ids.liuYaqin, ids.zhangMingyuan, rid++),
    pc(ids.zhangJianguo, ids.zhangMingcheng, rid++),
    pc(ids.liuYaqin, ids.zhangMingcheng, rid++),
    pc(ids.zhangJianguo, ids.zhangMinghui, rid++),
    pc(ids.liuYaqin, ids.zhangMinghui, rid++),
    pc(ids.zhangMingyuan, ids.zhangSiyuan, rid++),
    pc(ids.chenYuwei, ids.zhangSiyuan, rid++),
    pc(ids.zhangMingyuan, ids.zhangSining, rid++),
    pc(ids.chenYuwei, ids.zhangSining, rid++)
  ]

  return { nodes, edges, ids }
}

function zhangCases(): RelativeTitleTestCase[] {
  const { nodes, edges, ids } = buildZhangFamilyFixture()
  const base = { nodes, edges }
  return [
    { name: 'self', viewerMemberId: ids.me, targetMemberId: ids.me, expected: '我', ...base },
    { name: 'father', viewerMemberId: ids.me, targetMemberId: ids.fu, expected: '父亲', ...base },
    { name: 'mother', viewerMemberId: ids.me, targetMemberId: ids.mu, expected: '母亲', ...base },
    { name: 'grandfather', viewerMemberId: ids.me, targetMemberId: ids.gong, expected: '祖父', ...base },
    { name: 'grandmother', viewerMemberId: ids.me, targetMemberId: ids.gongSpouse, expected: '祖母', ...base },
    { name: 'great-grandfather', viewerMemberId: ids.me, targetMemberId: ids.zengfu, expected: '曾祖父', ...base },
    { name: 'great-grandmother', viewerMemberId: ids.me, targetMemberId: ids.zengmu, expected: '曾祖母', ...base },
    { name: 'wife', viewerMemberId: ids.me, targetMemberId: ids.spouse, expected: '妻子', ...base },
    { name: 'son', viewerMemberId: ids.me, targetMemberId: ids.son, expected: '儿子', ...base },
    { name: 'daughter', viewerMemberId: ids.me, targetMemberId: ids.daughter, expected: '女儿', ...base },
    { name: 'grandson', viewerMemberId: ids.me, targetMemberId: ids.grandson, expected: '孙子', ...base },
    { name: 'granddaughter', viewerMemberId: ids.me, targetMemberId: ids.granddaughter, expected: '孙女', ...base },
    { name: 'great-grandson', viewerMemberId: ids.me, targetMemberId: ids.greatGrandson, expected: '曾孙', ...base },
    { name: 'uncle-older', viewerMemberId: ids.me, targetMemberId: ids.shufu, expected: '叔父', ...base },
    { name: 'aunt-paternal', viewerMemberId: ids.me, targetMemberId: ids.gufu, expected: '姑母', ...base },
    { name: 'uncle-maternal', viewerMemberId: ids.me, targetMemberId: ids.muBro, expected: '舅父', ...base },
    { name: 'aunt-maternal', viewerMemberId: ids.me, targetMemberId: ids.muSis, expected: '姨母', ...base },
    { name: 'granduncle-older', viewerMemberId: ids.me, targetMemberId: ids.gongBro1, expected: '伯祖父', ...base },
    { name: 'granduncle-younger', viewerMemberId: ids.me, targetMemberId: ids.gongBro2, expected: '叔祖父', ...base },
    { name: 'grandaunt', viewerMemberId: ids.me, targetMemberId: ids.gongSis, expected: '姑祖母', ...base },
    { name: 'cousin-tang-older', viewerMemberId: ids.me, targetMemberId: ids.tangGe, expected: '堂兄', ...base },
    { name: 'cousin-tang-younger', viewerMemberId: ids.me, targetMemberId: ids.tangDi, expected: '堂弟', ...base },
    { name: 'cousin-biao', viewerMemberId: ids.me, targetMemberId: ids.biaoGe, expected: '表兄', ...base }
  ]
}

function siblingCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '父', 'MALE', '1970-01-01'),
    n(2, '母', 'FEMALE', '1972-01-01'),
    n(3, '大哥', 'MALE', '1995-01-01'),
    n(4, '我', 'MALE', '2000-01-01'),
    n(5, '二弟', 'MALE', '2005-01-01'),
    n(6, '大姐', 'FEMALE', '1996-01-01'),
    n(7, '小妹', 'FEMALE', '2008-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 4, 3),
    pc(2, 4, 4),
    pc(1, 5, 5),
    pc(2, 5, 6),
    pc(1, 6, 7),
    pc(2, 6, 8),
    pc(1, 7, 9),
    pc(2, 7, 10)
  ]
  return [
    { name: 'sibling-elder-brother', viewerMemberId: 4, targetMemberId: 3, nodes, edges, expected: '兄长' },
    { name: 'sibling-younger-brother', viewerMemberId: 4, targetMemberId: 5, nodes, edges, expected: '弟弟' },
    { name: 'sibling-elder-sister', viewerMemberId: 4, targetMemberId: 6, nodes, edges, expected: '姐姐' },
    { name: 'sibling-younger-sister', viewerMemberId: 4, targetMemberId: 7, nodes, edges, expected: '妹妹' }
  ]
}

function inlawCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '父', 'MALE', '1970-01-01'),
    n(2, '母', 'FEMALE', '1972-01-01'),
    n(3, '我', 'MALE', '2000-01-01'),
    n(4, '兄', 'MALE', '1998-01-01'),
    n(5, '弟', 'MALE', '2002-01-01'),
    n(6, '姐', 'FEMALE', '1997-01-01'),
    n(7, '妹', 'FEMALE', '2004-01-01'),
    n(8, '妻', 'FEMALE', '2001-01-01'),
    n(9, '嫂', 'FEMALE', '1999-01-01'),
    n(10, '弟媳', 'FEMALE', '2003-01-01'),
    n(11, '姐夫', 'MALE', '1996-01-01'),
    n(12, '妹夫', 'MALE', '2005-01-01'),
    n(13, '儿', 'MALE', '2025-01-01'),
    n(14, '媳', 'FEMALE', '2026-01-01'),
    n(15, '女', 'FEMALE', '2027-01-01'),
    n(16, '婿', 'MALE', '2028-01-01'),
    n(17, '孙', 'MALE', '2050-01-01'),
    n(18, '孙媳', 'FEMALE', '2051-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 4, 3),
    pc(2, 4, 4),
    pc(1, 5, 5),
    pc(2, 5, 6),
    pc(1, 6, 7),
    pc(2, 6, 8),
    pc(1, 7, 9),
    pc(2, 7, 10),
    sp(3, 8, 11),
    sp(4, 9, 12),
    sp(5, 10, 13),
    sp(6, 11, 14),
    sp(7, 12, 15),
    pc(3, 13, 16),
    pc(8, 13, 17),
    sp(13, 14, 18),
    pc(3, 15, 19),
    pc(8, 15, 20),
    sp(15, 16, 21),
    pc(13, 17, 22),
    sp(17, 18, 23)
  ]
  return [
    { name: 'wife', viewerMemberId: 3, targetMemberId: 8, nodes, edges, expected: '妻子' },
    { name: 'sister-in-law-elder', viewerMemberId: 3, targetMemberId: 9, nodes, edges, expected: '嫂子' },
    { name: 'sister-in-law-younger', viewerMemberId: 3, targetMemberId: 10, nodes, edges, expected: '弟媳' },
    { name: 'brother-in-law-elder', viewerMemberId: 3, targetMemberId: 11, nodes, edges, expected: '姐夫' },
    { name: 'brother-in-law-younger', viewerMemberId: 3, targetMemberId: 12, nodes, edges, expected: '妹夫' },
    { name: 'daughter-in-law', viewerMemberId: 3, targetMemberId: 14, nodes, edges, expected: '儿媳' },
    { name: 'son-in-law', viewerMemberId: 3, targetMemberId: 16, nodes, edges, expected: '女婿' },
    { name: 'granddaughter-in-law', viewerMemberId: 3, targetMemberId: 18, nodes, edges, expected: '孙媳' }
  ]
}

function nephewCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '父', 'MALE', '1970-01-01'),
    n(2, '母', 'FEMALE', '1972-01-01'),
    n(3, '我', 'MALE', '2000-01-01'),
    n(4, '兄', 'MALE', '1998-01-01'),
    n(5, '妹', 'FEMALE', '2002-01-01'),
    n(6, '侄', 'MALE', '2030-01-01'),
    n(7, '侄', 'FEMALE', '2031-01-01'),
    n(8, '甥', 'MALE', '2032-01-01'),
    n(9, '甥', 'FEMALE', '2033-01-01'),
    n(10, '侄孙', 'MALE', '2055-01-01'),
    n(11, '甥孙', 'MALE', '2056-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 4, 3),
    pc(2, 4, 4),
    pc(1, 5, 5),
    pc(2, 5, 6),
    pc(4, 6, 7),
    pc(4, 7, 8),
    pc(5, 8, 9),
    pc(5, 9, 10),
    pc(6, 10, 11),
    pc(8, 11, 12)
  ]
  return [
    { name: 'nephew', viewerMemberId: 3, targetMemberId: 6, nodes, edges, expected: '侄子' },
    { name: 'niece', viewerMemberId: 3, targetMemberId: 7, nodes, edges, expected: '侄女' },
    { name: 'sister-son', viewerMemberId: 3, targetMemberId: 8, nodes, edges, expected: '外甥' },
    { name: 'sister-daughter', viewerMemberId: 3, targetMemberId: 9, nodes, edges, expected: '外甥女' },
    { name: 'nephew-grandson', viewerMemberId: 3, targetMemberId: 10, nodes, edges, expected: '侄孙' },
    { name: 'sister-grandson', viewerMemberId: 3, targetMemberId: 11, nodes, edges, expected: '外甥孙' }
  ]
}

function cousinChildCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '祖父', 'MALE', '1930-01-01'),
    n(2, '祖母', 'FEMALE', '1932-01-01'),
    n(3, '父', 'MALE', '1960-01-01'),
    n(4, '母', 'FEMALE', '1962-01-01'),
    n(5, '外祖父', 'MALE', '1931-01-01'),
    n(6, '外祖母', 'FEMALE', '1933-01-01'),
    n(7, '我', 'MALE', '2000-01-01'),
    n(8, '伯父', 'MALE', '1955-01-01'),
    n(9, '舅父', 'MALE', '1959-01-01'),
    n(10, '堂兄', 'MALE', '1995-01-01'),
    n(11, '表哥', 'MALE', '1996-01-01'),
    n(12, '堂侄', 'MALE', '2025-01-01'),
    n(13, '堂侄女', 'FEMALE', '2026-01-01'),
    n(14, '表侄', 'MALE', '2027-01-01'),
    n(15, '堂嫂', 'FEMALE', '1997-01-01'),
    n(16, '堂侄孙', 'MALE', '2050-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(5, 4, 3),
    pc(6, 4, 4),
    pc(1, 8, 5),
    pc(2, 8, 6),
    pc(5, 9, 7),
    pc(6, 9, 8),
    pc(3, 7, 9),
    pc(4, 7, 10),
    pc(8, 10, 11),
    pc(9, 11, 12),
    sp(10, 15, 13),
    pc(10, 12, 14),
    pc(10, 13, 15),
    pc(11, 14, 16),
    pc(12, 16, 17)
  ]
  return [
    { name: 'cousin-child-tang-male', viewerMemberId: 7, targetMemberId: 12, nodes, edges, expected: '堂侄' },
    { name: 'cousin-child-tang-female', viewerMemberId: 7, targetMemberId: 13, nodes, edges, expected: '堂侄女' },
    { name: 'cousin-child-biao-male', viewerMemberId: 7, targetMemberId: 14, nodes, edges, expected: '表侄' },
    { name: 'cousin-spouse-tang', viewerMemberId: 7, targetMemberId: 15, nodes, edges, expected: '堂嫂' },
    { name: 'cousin-grandchild-tang', viewerMemberId: 7, targetMemberId: 16, nodes, edges, expected: '堂侄孙' }
  ]
}

function maternalLineCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '外曾祖父', 'MALE', '1900-01-01'),
    n(2, '外曾祖母', 'FEMALE', '1902-01-01'),
    n(3, '外祖父', 'MALE', '1928-01-01'),
    n(4, '外祖母', 'FEMALE', '1930-01-01'),
    n(5, '母', 'FEMALE', '1960-01-01'),
    n(6, '父', 'MALE', '1962-01-01'),
    n(7, '我', 'MALE', '2000-01-01'),
    n(8, '外伯祖父', 'MALE', '1925-01-01'),
    n(9, '外叔祖父', 'MALE', '1935-01-01'),
    n(10, '外姑祖母', 'FEMALE', '1927-01-01'),
    n(11, '外曾外祖父', 'MALE', '1898-01-01'),
    n(12, '外曾外祖母', 'FEMALE', '1899-01-01'),
    n(13, '外舅祖父', 'MALE', '1926-01-01'),
    n(14, '外姨祖母', 'FEMALE', '1929-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 8, 3),
    pc(2, 8, 4),
    pc(1, 9, 5),
    pc(2, 9, 6),
    pc(1, 10, 7),
    pc(2, 10, 8),
    pc(11, 4, 9),
    pc(12, 4, 10),
    pc(11, 13, 11),
    pc(12, 13, 12),
    pc(11, 14, 13),
    pc(12, 14, 14),
    pc(3, 5, 15),
    pc(4, 5, 16),
    pc(6, 7, 17),
    pc(5, 7, 18)
  ]
  return [
    { name: 'maternal-grandfather', viewerMemberId: 7, targetMemberId: 3, nodes, edges, expected: '外祖父' },
    { name: 'maternal-grandmother', viewerMemberId: 7, targetMemberId: 4, nodes, edges, expected: '外祖母' },
    { name: 'maternal-granduncle-older', viewerMemberId: 7, targetMemberId: 8, nodes, edges, expected: '外伯祖父' },
    { name: 'maternal-granduncle-younger', viewerMemberId: 7, targetMemberId: 9, nodes, edges, expected: '外叔祖父' },
    { name: 'maternal-grandaunt', viewerMemberId: 7, targetMemberId: 10, nodes, edges, expected: '外姑祖母' },
    { name: 'maternal-grand-uncle-maternal', viewerMemberId: 7, targetMemberId: 13, nodes, edges, expected: '外舅祖父' },
    { name: 'maternal-grand-aunt-maternal', viewerMemberId: 7, targetMemberId: 14, nodes, edges, expected: '外姨祖母' }
  ]
}

function parentSiblingSpouseCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '祖父', 'MALE', '1930-01-01'),
    n(2, '祖母', 'FEMALE', '1932-01-01'),
    n(3, '外祖父', 'MALE', '1931-01-01'),
    n(4, '外祖母', 'FEMALE', '1933-01-01'),
    n(5, '父', 'MALE', '1960-01-01'),
    n(6, '母', 'FEMALE', '1962-01-01'),
    n(7, '我', 'MALE', '2000-01-01'),
    n(8, '伯父', 'MALE', '1955-01-01'),
    n(9, '叔父', 'MALE', '1965-01-01'),
    n(10, '姑母', 'FEMALE', '1958-01-01'),
    n(11, '伯母', 'FEMALE', '1956-01-01'),
    n(12, '婶婶', 'FEMALE', '1966-01-01'),
    n(13, '姑父', 'MALE', '1957-01-01'),
    n(14, '舅父', 'MALE', '1959-01-01'),
    n(15, '舅母', 'FEMALE', '1960-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 5, 1),
    pc(2, 5, 2),
    pc(1, 8, 3),
    pc(2, 8, 4),
    pc(1, 9, 5),
    pc(2, 9, 6),
    pc(1, 10, 7),
    pc(2, 10, 8),
    pc(3, 6, 9),
    pc(4, 6, 10),
    pc(3, 14, 11),
    pc(4, 14, 12),
    pc(5, 7, 13),
    pc(6, 7, 14),
    sp(8, 11, 15),
    sp(9, 12, 16),
    sp(10, 13, 17),
    sp(14, 15, 18)
  ]
  return [
    { name: 'uncle-wife-elder', viewerMemberId: 7, targetMemberId: 11, nodes, edges, expected: '伯母' },
    { name: 'uncle-wife-younger', viewerMemberId: 7, targetMemberId: 12, nodes, edges, expected: '婶母' },
    { name: 'aunt-husband', viewerMemberId: 7, targetMemberId: 13, nodes, edges, expected: '姑父' },
    { name: 'maternal-uncle-wife', viewerMemberId: 7, targetMemberId: 15, nodes, edges, expected: '舅母' }
  ]
}

function family12SampleCases(): RelativeTitleTestCase[] {
  const { nodes, edges, ids } = buildFamily12Fixture()
  const base = { nodes, edges }
  const me = ids.zhangMingyuan
  return [
    { name: 'f12-self', viewerMemberId: me, targetMemberId: me, expected: '我', ...base },
    { name: 'f12-father', viewerMemberId: me, targetMemberId: ids.zhangJianguo, expected: '父亲', ...base },
    { name: 'f12-mother', viewerMemberId: me, targetMemberId: ids.liuYaqin, expected: '母亲', ...base },
    { name: 'f12-grandfather', viewerMemberId: me, targetMemberId: ids.zhangGuoan, expected: '祖父', ...base },
    { name: 'f12-grandmother', viewerMemberId: me, targetMemberId: ids.wangXiulan, expected: '祖母', ...base },
    { name: 'f12-great-grandfather', viewerMemberId: me, targetMemberId: ids.zhangDehou, expected: '曾祖父', ...base },
    { name: 'f12-great-grandmother', viewerMemberId: me, targetMemberId: ids.liSuqin, expected: '曾祖母', ...base },
    { name: 'f12-wife', viewerMemberId: me, targetMemberId: ids.chenYuwei, expected: '妻子', ...base },
    { name: 'f12-son', viewerMemberId: me, targetMemberId: ids.zhangSiyuan, expected: '儿子', ...base },
    { name: 'f12-daughter', viewerMemberId: me, targetMemberId: ids.zhangSining, expected: '女儿', ...base },
    { name: 'f12-elder-brother', viewerMemberId: me, targetMemberId: ids.zhangMingcheng, expected: '弟弟', ...base },
    { name: 'f12-younger-sister', viewerMemberId: me, targetMemberId: ids.zhangMinghui, expected: '妹妹', ...base },
    { name: 'f12-uncle-paternal', viewerMemberId: me, targetMemberId: ids.zhangJianjun, expected: '叔父', ...base },
    { name: 'f12-aunt-paternal', viewerMemberId: me, targetMemberId: ids.zhangJianfang, expected: '姑母', ...base },
    { name: 'f12-cousin-tang', viewerMemberId: me, targetMemberId: ids.zhangJianmin, expected: '堂叔', ...base },
    { name: 'f12-granduncle', viewerMemberId: me, targetMemberId: ids.zhangGuoping, expected: '叔祖父', ...base },
    { name: 'f12-granduncle-wife', viewerMemberId: me, targetMemberId: ids.zhaoMeihua, expected: '叔祖母', ...base },
    { name: 'f12-grandaunt', viewerMemberId: me, targetMemberId: ids.zhangGuohua, expected: '姑祖母', ...base }
  ]
}

function extraBloodCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '父', 'MALE', '1970-01-01'),
    n(2, '母', 'FEMALE', '1972-01-01'),
    n(3, '我', 'FEMALE', '2000-01-01'),
    n(4, '夫', 'MALE', '1999-01-01'),
    n(5, '儿', 'MALE', '2025-01-01'),
    n(6, '女', 'FEMALE', '2027-01-01'),
    n(7, '孙', 'MALE', '2050-01-01'),
    n(8, '外孙', 'MALE', '2051-01-01'),
    n(9, '外孙女', 'FEMALE', '2052-01-01'),
    n(10, '外曾孙', 'MALE', '2070-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    sp(3, 4, 3),
    pc(3, 5, 4),
    pc(4, 5, 5),
    pc(3, 6, 6),
    pc(4, 6, 7),
    pc(5, 7, 8),
    pc(6, 8, 9),
    pc(6, 9, 10),
    pc(8, 10, 11)
  ]
  return [
    { name: 'husband', viewerMemberId: 3, targetMemberId: 4, nodes, edges, expected: '丈夫' },
    { name: 'daughter-as-mother', viewerMemberId: 3, targetMemberId: 6, nodes, edges, expected: '女儿' },
    { name: 'maternal-grandson', viewerMemberId: 3, targetMemberId: 8, nodes, edges, expected: '外孙' },
    { name: 'maternal-granddaughter', viewerMemberId: 3, targetMemberId: 9, nodes, edges, expected: '外孙女' },
    { name: 'maternal-great-grandson', viewerMemberId: 3, targetMemberId: 10, nodes, edges, expected: '外曾孙' }
  ]
}

function rankedGrandparentCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '曾祖父', 'MALE', '1900-01-01'),
    n(2, '曾祖母', 'FEMALE', '1902-01-01'),
    n(3, '大伯祖父', 'MALE', '1924-01-01'),
    n(4, '二伯祖父', 'MALE', '1926-01-01'),
    n(5, '祖父', 'MALE', '1928-01-01'),
    n(6, '叔祖父', 'MALE', '1932-01-01'),
    n(7, '小姑祖母', 'FEMALE', '1935-01-01'),
    n(8, '父', 'MALE', '1960-01-01'),
    n(9, '母', 'FEMALE', '1962-01-01'),
    n(10, '我', 'MALE', '2000-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 4, 3),
    pc(2, 4, 4),
    pc(1, 5, 5),
    pc(2, 5, 6),
    pc(1, 6, 7),
    pc(2, 6, 8),
    pc(1, 7, 9),
    pc(2, 7, 10),
    pc(5, 8, 11),
    pc(9, 8, 12),
    pc(8, 10, 13),
    pc(9, 10, 14)
  ]
  return [
    { name: 'ranked-da-bo-grandfather', viewerMemberId: 10, targetMemberId: 3, nodes, edges, expected: '伯祖父' },
    { name: 'ranked-er-bo-grandfather', viewerMemberId: 10, targetMemberId: 4, nodes, edges, expected: '伯祖父' },
    { name: 'ranked-shu-grandfather', viewerMemberId: 10, targetMemberId: 6, nodes, edges, expected: '叔祖父' },
    { name: 'ranked-xiao-gu-grandmother', viewerMemberId: 10, targetMemberId: 7, nodes, edges, expected: '姑祖母' }
  ]
}

/** 祖辈旁系后代 = 父母辈堂/表亲（非我同辈） */
function grandparentCollateralDescendantCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '曾祖父', 'MALE', '1900-01-01'),
    n(2, '曾祖母', 'FEMALE', '1902-01-01'),
    n(3, '祖父', 'MALE', '1928-01-01'),
    n(4, '祖母', 'FEMALE', '1930-01-01'),
    n(20, '祖母之父', 'MALE', '1898-01-01'),
    n(21, '祖母之母', 'FEMALE', '1899-01-01'),
    n(5, '祖母兄', 'MALE', '1926-01-01'),
    n(6, '祖母妹', 'FEMALE', '1933-01-01'),
    n(7, '伯祖父', 'MALE', '1924-01-01'),
    n(8, '叔祖父', 'MALE', '1935-01-01'),
    n(9, '父', 'MALE', '1978-01-01'),
    n(10, '母', 'FEMALE', '1980-01-01'),
    n(11, '我', 'MALE', '2004-01-01'),
    n(12, '堂伯', 'MALE', '1970-01-01'),
    n(13, '堂叔', 'MALE', '1985-01-01'),
    n(14, '堂姑', 'FEMALE', '1975-01-01'),
    n(15, '表伯', 'MALE', '1972-01-01'),
    n(16, '表姑', 'FEMALE', '1976-01-01'),
    n(18, '叔父', 'MALE', '1980-01-01'),
    n(17, '同辈堂哥', 'MALE', '2000-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(1, 3, 1),
    pc(2, 3, 2),
    pc(1, 7, 3),
    pc(2, 7, 4),
    pc(1, 8, 5),
    pc(2, 8, 6),
    pc(3, 9, 7),
    pc(4, 9, 8),
    pc(20, 4, 9),
    pc(21, 4, 10),
    pc(20, 5, 11),
    pc(21, 5, 12),
    pc(20, 6, 13),
    pc(21, 6, 14),
    pc(7, 12, 15),
    pc(8, 13, 16),
    pc(8, 14, 17),
    pc(5, 15, 18),
    pc(6, 16, 19),
    pc(3, 18, 20),
    pc(4, 18, 21),
    pc(18, 17, 22),
    pc(9, 11, 23),
    pc(10, 11, 24)
  ]
  return [
    { name: 'gp-paternal-uncle-son-older', viewerMemberId: 11, targetMemberId: 12, nodes, edges, expected: '堂伯' },
    { name: 'gp-paternal-uncle-son-younger', viewerMemberId: 11, targetMemberId: 13, nodes, edges, expected: '堂叔' },
    { name: 'gp-paternal-uncle-daughter', viewerMemberId: 11, targetMemberId: 14, nodes, edges, expected: '堂姑' },
    { name: 'gp-maternal-uncle-son', viewerMemberId: 11, targetMemberId: 15, nodes, edges, expected: '表伯' },
    { name: 'gp-maternal-aunt-daughter', viewerMemberId: 11, targetMemberId: 16, nodes, edges, expected: '表姑' },
    { name: 'parent-sibling-child-peer', viewerMemberId: 11, targetMemberId: 17, nodes, edges, expected: '堂兄' }
  ]
}

/** 父母辈堂亲（祖辈旁系）的子女 = 我同辈，不得误标为堂侄/堂侄女 */
function parentCousinPeerCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(99, '曾祖父', 'MALE', '1900-01-01'),
    n(100, '曾祖母', 'FEMALE', '1902-01-01'),
    n(1, '祖父', 'MALE', '1930-01-01'),
    n(2, '祖母', 'FEMALE', '1932-01-01'),
    n(3, '伯祖父', 'MALE', '1925-01-01'),
    n(4, '父', 'MALE', '1960-01-01'),
    n(5, '父的堂兄', 'MALE', '1958-01-01'),
    n(6, '我', 'MALE', '1990-01-01'),
    n(7, '堂兄', 'MALE', '1988-01-01'),
    n(8, '堂妹', 'FEMALE', '1992-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(99, 1, 1),
    pc(100, 1, 2),
    pc(99, 3, 3),
    pc(100, 3, 4),
    pc(1, 4, 5),
    pc(2, 4, 6),
    pc(3, 5, 7),
    pc(4, 6, 8),
    pc(5, 7, 9),
    pc(5, 8, 10)
  ]
  return [
    { name: 'parent-cousin-son-elder', viewerMemberId: 6, targetMemberId: 7, nodes, edges, expected: '堂兄' },
    { name: 'parent-cousin-daughter-younger', viewerMemberId: 6, targetMemberId: 8, nodes, edges, expected: '堂妹' },
    { name: 'parent-cousin-peer-reverse', viewerMemberId: 7, targetMemberId: 6, nodes, edges, expected: '堂弟' }
  ]
}

function deepAncestorAndCollateralCases(): RelativeTitleTestCase[] {
  const nodes: TreeNode[] = [
    n(1, '我', 'MALE', '2000-01-01'),
    n(2, '父', 'MALE', '1970-01-01'),
    n(3, '父亲的母亲', 'FEMALE', '1940-01-01'),
    n(4, '父亲的母亲的父亲', 'MALE', '1910-01-01'),
    n(5, '父亲的母亲的父亲的母亲', 'FEMALE', '1880-01-01'),
    n(6, '父亲的母亲的父亲的母亲的父亲', 'MALE', '1850-01-01'),
    n(7, '祖父', 'MALE', '1942-01-01'),
    n(8, '曾祖父', 'MALE', '1912-01-01'),
    n(9, '高祖母', 'FEMALE', '1882-01-01'),
    n(10, '五世祖母', 'FEMALE', '1852-01-01'),
    n(20, '高祖父', 'MALE', '1880-01-01'),
    n(21, '高祖母', 'FEMALE', '1881-01-01'),
    n(22, '曾伯祖父', 'MALE', '1908-01-01'),
    n(23, '曾叔祖父', 'MALE', '1918-01-01'),
    n(24, '曾姑祖母', 'FEMALE', '1915-01-01'),
    n(25, '曾叔祖母', 'FEMALE', '1919-01-01'),
    n(26, '曾祖母', 'FEMALE', '1914-01-01'),
    n(27, '曾舅祖父', 'MALE', '1910-01-01'),
    n(28, '曾姨祖母', 'FEMALE', '1916-01-01'),
    n(30, '天祖父', 'MALE', '1850-01-01'),
    n(31, '天祖母', 'FEMALE', '1852-01-01'),
    n(32, '高伯祖父', 'MALE', '1876-01-01'),
    n(33, '高姑祖母', 'FEMALE', '1885-01-01'),
    n(40, '曾祖母之父', 'MALE', '1888-01-01'),
    n(41, '曾祖母之母', 'FEMALE', '1889-01-01')
  ]
  const edges: TreeEdge[] = [
    pc(2, 1, 1),
    pc(3, 2, 2),
    pc(4, 3, 3),
    pc(5, 4, 4),
    pc(6, 5, 5),
    pc(7, 2, 6),
    pc(8, 7, 7),
    pc(9, 8, 8),
    pc(10, 9, 9),
    pc(20, 8, 10),
    pc(21, 8, 11),
    pc(20, 22, 12),
    pc(21, 22, 13),
    pc(20, 23, 14),
    pc(21, 23, 15),
    pc(20, 24, 16),
    pc(21, 24, 17),
    sp(23, 25, 18),
    pc(26, 7, 19),
    pc(40, 26, 20),
    pc(41, 26, 21),
    pc(30, 20, 31),
    pc(31, 20, 32),
    pc(30, 32, 33),
    pc(31, 32, 34),
    pc(30, 33, 35),
    pc(31, 33, 36),
    pc(40, 27, 37),
    pc(41, 27, 38),
    pc(40, 28, 39),
    pc(41, 28, 40)
  ]

  return [
    { name: 'deep-chain-father-mother-father-mother-father', viewerMemberId: 1, targetMemberId: 6, nodes, edges, expected: '外五世祖父' },
    { name: 'deep-chain-father-father-father-mother-mother', viewerMemberId: 1, targetMemberId: 10, nodes, edges, expected: '外五世祖母' },
    { name: 'great-grandfather-older-brother', viewerMemberId: 1, targetMemberId: 22, nodes, edges, expected: '曾伯祖父' },
    { name: 'great-grandfather-younger-brother', viewerMemberId: 1, targetMemberId: 23, nodes, edges, expected: '曾叔祖父' },
    { name: 'great-grandfather-sister', viewerMemberId: 1, targetMemberId: 24, nodes, edges, expected: '曾姑祖母' },
    { name: 'great-grandfather-younger-brother-wife', viewerMemberId: 1, targetMemberId: 25, nodes, edges, expected: '曾叔祖母' },
    { name: 'great-grandmother-brother', viewerMemberId: 1, targetMemberId: 27, nodes, edges, expected: '曾舅祖父' },
    { name: 'great-grandmother-sister', viewerMemberId: 1, targetMemberId: 28, nodes, edges, expected: '曾姨祖母' },
    { name: 'great-great-grandfather-older-brother', viewerMemberId: 1, targetMemberId: 32, nodes, edges, expected: '高伯祖父' },
    { name: 'great-great-grandfather-sister', viewerMemberId: 1, targetMemberId: 33, nodes, edges, expected: '高姑祖母' }
  ]
}

export function getFamily12SampleTitles(): Array<{ name: string; title: string }> {
  const { nodes, edges, ids } = buildFamily12Fixture()
  const viewerMemberId = ids.zhangMingyuan
  const graph = buildRelGraph(nodes, edges)
  return nodes.map((node) => ({
    name: node.displayName,
    title: getRelativeTitle(viewerMemberId, node.memberId, graph) ?? ''
  }))
}

function cousinPeerSeniorityCounterCases(): RelativeTitleTestCase[] {
  /** 叔父（父辈年幼）之子，出生早于「我」→ 堂兄，不得因叔父支判为堂弟 */
  const tangNodes: TreeNode[] = [
    n(10, '祖父', 'MALE', '1940-01-01'),
    n(11, '祖母', 'FEMALE', '1942-01-01'),
    n(1, '父亲', 'MALE', '1970-01-01'),
    n(4, '叔父', 'MALE', '1975-01-01'),
    n(2, '母亲', 'FEMALE', '1972-01-01'),
    n(3, '我', 'MALE', '2000-01-01'),
    n(5, '堂亲', 'MALE', '1998-01-01')
  ]
  const tangEdges: TreeEdge[] = [
    pc(10, 1, 1),
    pc(11, 1, 2),
    pc(10, 4, 3),
    pc(11, 4, 4),
    pc(1, 3, 5),
    pc(2, 3, 6),
    pc(4, 5, 7)
  ]

  /** 舅父（父辈年长）之子，出生晚于「我」→ 表弟，不得因舅父支判为表兄 */
  const biaoNodes: TreeNode[] = [
    n(20, '外祖父', 'MALE', '1945-01-01'),
    n(21, '外祖母', 'FEMALE', '1947-01-01'),
    n(1, '母亲', 'FEMALE', '1972-01-01'),
    n(4, '舅父', 'MALE', '1968-01-01'),
    n(2, '父亲', 'MALE', '1970-01-01'),
    n(3, '我', 'MALE', '2000-01-01'),
    n(5, '表亲', 'MALE', '2005-01-01')
  ]
  const biaoEdges: TreeEdge[] = [
    pc(20, 1, 1),
    pc(21, 1, 2),
    pc(20, 4, 3),
    pc(21, 4, 4),
    pc(1, 3, 5),
    pc(2, 3, 6),
    pc(4, 5, 7)
  ]

  return [
    {
      name: 'cousin-tang-peer-older-not-paternal-branch',
      viewerMemberId: 3,
      targetMemberId: 5,
      nodes: tangNodes,
      edges: tangEdges,
      expected: '堂兄'
    },
    {
      name: 'cousin-biao-peer-younger-not-paternal-branch',
      viewerMemberId: 3,
      targetMemberId: 5,
      nodes: biaoNodes,
      edges: biaoEdges,
      expected: '表弟'
    }
  ]
}

export function allRelativeTitleTestCases(): RelativeTitleTestCase[] {
  return [
    ...zhangCases(),
    ...siblingCases(),
    ...inlawCases(),
    ...nephewCases(),
    ...cousinChildCases(),
    ...cousinPeerSeniorityCounterCases(),
    ...maternalLineCases(),
    ...parentSiblingSpouseCases(),
    ...family12SampleCases(),
    ...extraBloodCases(),
    ...rankedGrandparentCases(),
    ...grandparentCollateralDescendantCases(),
    ...parentCousinPeerCases(),
    ...deepAncestorAndCollateralCases()
  ]
}

export function runRelativeTitleSelfCheck(): {
  total: number
  passed: number
  failed: number
  failures: Array<{ name: string; expected: string; actual: string }>
  cacheOk: boolean
} {
  clearRelativeTitleCache()
  const cases = allRelativeTitleTestCases()
  const failures: Array<{ name: string; expected: string; actual: string }> = []

  for (const testCase of cases) {
    const graph = buildRelGraph(testCase.nodes, testCase.edges)
    const actual =
      getRelativeTitle(testCase.viewerMemberId, testCase.targetMemberId, graph) ?? ''
    if (actual !== testCase.expected) {
      failures.push({ name: testCase.name, expected: testCase.expected, actual })
    }
  }

  const cacheOk = runCacheKeySelfCheck()

  return {
    total: cases.length,
    passed: cases.length - failures.length,
    failed: failures.length,
    failures,
    cacheOk
  }
}

function runCacheKeySelfCheck(): boolean {
  const { nodes, edges, ids } = buildFamily12Fixture()
  const viewerMemberId = ids.zhangMingyuan
  const targetMemberId = ids.zhangJianmin
  const graphVersion = 3
  const poisonedTitle = '错误称谓'

  clearRelativeTitleCache()
  const graph = buildRelGraph(nodes, edges)
  const key = buildTitleCacheKey(viewerMemberId, targetMemberId, graphVersion)
  setCachedRelativeTitle(key, poisonedTitle)

  const map1 = buildKinshipTitleMap({
    viewerMemberId,
    nodes,
    edges,
    graphVersion
  })
  if (map1[targetMemberId] !== poisonedTitle) return false

  const expected = getRelativeTitle(viewerMemberId, targetMemberId, graph) ?? ''
  const map2 = buildKinshipTitleMap({
    viewerMemberId,
    nodes,
    edges,
    graphVersion: graphVersion + 1
  })
  return map2[targetMemberId] === expected && map2[targetMemberId] !== poisonedTitle
}

const isDirectRun = typeof process !== 'undefined' && process.argv[1]?.includes('relativeTitle.testCases')
if (isDirectRun) {
  const result = runRelativeTitleSelfCheck()
  console.log(`relativeTitle self-check: ${result.passed}/${result.total} passed`)
  console.log(`cache key self-check: ${result.cacheOk ? 'OK' : 'FAIL'}`)
  if (result.failures.length > 0 || !result.cacheOk) {
    for (const failure of result.failures) {
      console.log(`  FAIL ${failure.name}: expected "${failure.expected}", got "${failure.actual}"`)
    }
    if (!result.cacheOk) {
      console.log('  FAIL cache: graphVersion must invalidate cached titles')
    }
    process.exitCode = 1
  } else {
    console.log('\nfamilyId=12 sample titles (张明远视角):')
    for (const item of getFamily12SampleTitles()) {
      console.log(`  ${item.name}: ${item.title}`)
    }
  }
}
