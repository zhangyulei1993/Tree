import type { TreeEdge, TreeNode } from '@/types/api'

function lineageNode(
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
    memberType: 'LINEAGE_MEMBER',
    userBindingState: 'NOT_REQUIRED',
    canExpand: false
  }
}

function spouseNode(
  memberId: number,
  displayName: string,
  gender: 'MALE' | 'FEMALE',
  birthDate: string
): TreeNode {
  return { ...lineageNode(memberId, displayName, gender, birthDate), memberType: 'SPOUSE' }
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

/**
 * 43 个族内成员 + 若干配偶节点的综合家谱，用于称谓矩阵测试。
 */
export function buildIntraLineage43Fixture(): {
  nodes: TreeNode[]
  edges: TreeEdge[]
  lineageMemberIds: number[]
  spouseMemberIds: number[]
  ids: Record<string, number>
} {
  const ids = {
    zengfu: 1,
    zengmu: 2,
    gong: 3,
    gongSpouse: 4,
    gongBro1: 5,
    gongBro2: 6,
    gongSis: 7,
    fu: 8,
    mu: 9,
    shufu: 10,
    gufu: 11,
    muBro: 12,
    muSis: 13,
    me: 14,
    spouse: 15,
    tangGe: 16,
    tangDi: 17,
    biaoGe: 18,
    biaoMei: 19,
    nephew: 20,
    niece: 21,
    waiSheng: 22,
    waiShengNv: 23,
    son: 24,
    daughter: 25,
    grandson: 26,
    granddaughter: 27,
    greatGrandson: 28,
    cousin2: 29,
    cousin3: 30,
    cousin4: 31,
    cousin5: 32,
    extraUncleChild1: 33,
    extraUncleChild2: 34,
    secondSon: 35,
    secondSonWife: 36,
    secondGrandson: 37,
    secondGranddaughter: 38,
    branchNephew: 39,
    branchNiece: 40,
    waiGong: 41,
    waiPo: 42,
    elderBro: 43,
    jieJie: 44,
    muCousin: 45,
    tangJie: 46
  }

  let rid = 1000
  const nodes: TreeNode[] = [
    lineageNode(ids.zengfu, '曾祖父', 'MALE', '1900-01-01'),
    lineageNode(ids.zengmu, '曾祖母', 'FEMALE', '1902-02-02'),
    lineageNode(ids.gong, '祖父', 'MALE', '1928-03-12'),
    spouseNode(ids.gongSpouse, '祖母', 'FEMALE', '1931-09-20'),
    lineageNode(ids.gongBro1, '伯祖父', 'MALE', '1925-01-01'),
    lineageNode(ids.gongBro2, '叔祖父', 'MALE', '1935-06-06'),
    lineageNode(ids.gongSis, '姑祖母', 'FEMALE', '1930-03-03'),
    lineageNode(ids.fu, '父亲', 'MALE', '1952-05-04'),
    lineageNode(ids.shufu, '叔父', 'MALE', '1958-02-09'),
    lineageNode(ids.gufu, '姑母', 'FEMALE', '1960-04-04'),
    lineageNode(ids.mu, '母亲', 'FEMALE', '1954-08-16'),
    lineageNode(ids.waiGong, '外祖父', 'MALE', '1950-02-02'),
    lineageNode(ids.waiPo, '外祖母', 'FEMALE', '1953-03-03'),
    lineageNode(ids.muBro, '舅父', 'MALE', '1956-06-06'),
    lineageNode(ids.muSis, '姨母', 'FEMALE', '1959-07-07'),
    lineageNode(ids.elderBro, '兄长', 'MALE', '1975-03-03'),
    lineageNode(ids.jieJie, '姐姐', 'FEMALE', '1977-04-04'),
    lineageNode(ids.me, '我', 'MALE', '1980-01-01'),
    spouseNode(ids.spouse, '妻子', 'FEMALE', '1981-02-02'),
    lineageNode(ids.tangGe, '堂兄', 'MALE', '1978-05-05'),
    lineageNode(ids.tangDi, '堂弟', 'MALE', '1985-08-08'),
    lineageNode(ids.biaoGe, '表兄', 'MALE', '1979-09-09'),
    lineageNode(ids.biaoMei, '表妹', 'FEMALE', '1988-10-10'),
    lineageNode(ids.nephew, '侄子', 'MALE', '2005-03-03'),
    lineageNode(ids.niece, '侄女', 'FEMALE', '2008-04-04'),
    lineageNode(ids.waiSheng, '外甥', 'MALE', '2006-05-05'),
    lineageNode(ids.waiShengNv, '外甥女', 'FEMALE', '2009-06-06'),
    lineageNode(ids.son, '儿子', 'MALE', '2004-03-16'),
    lineageNode(ids.daughter, '女儿', 'FEMALE', '2007-09-01'),
    lineageNode(ids.grandson, '孙子', 'MALE', '2028-02-14'),
    lineageNode(ids.granddaughter, '孙女', 'FEMALE', '2030-08-08'),
    lineageNode(ids.greatGrandson, '曾孙', 'MALE', '2048-01-01'),
    lineageNode(ids.cousin2, '堂弟二', 'MALE', '1986-01-01'),
    lineageNode(ids.cousin3, '堂姐', 'FEMALE', '1983-02-02'),
    lineageNode(ids.cousin4, '表弟', 'MALE', '1990-03-03'),
    lineageNode(ids.cousin5, '表姐', 'FEMALE', '1987-04-04'),
    lineageNode(ids.extraUncleChild1, '叔父长子', 'MALE', '1984-05-05'),
    lineageNode(ids.extraUncleChild2, '叔父次女', 'FEMALE', '1989-06-06'),
    lineageNode(ids.secondSon, '次子', 'MALE', '2010-07-07'),
    spouseNode(ids.secondSonWife, '次媳', 'FEMALE', '2011-08-08'),
    lineageNode(ids.secondGrandson, '次孙', 'MALE', '2032-09-09'),
    lineageNode(ids.secondGranddaughter, '次孙女', 'FEMALE', '2034-10-10'),
    lineageNode(ids.branchNephew, '侄孙', 'MALE', '2035-11-11'),
    lineageNode(ids.branchNiece, '侄孙女', 'FEMALE', '2036-12-12'),
    lineageNode(ids.muCousin, '表兄二', 'MALE', '1984-11-11'),
    lineageNode(ids.tangJie, '堂姐', 'FEMALE', '1982-12-12')
  ]

  const edges: TreeEdge[] = [
    sp(ids.zengfu, ids.zengmu, rid++),
    sp(ids.gong, ids.gongSpouse, rid++),
    sp(ids.fu, ids.mu, rid++),
    sp(ids.me, ids.spouse, rid++),
    sp(ids.secondSon, ids.secondSonWife, rid++),
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
    pc(ids.gongBro1, ids.cousin3, rid++),
    pc(ids.shufu, ids.tangGe, rid++),
    pc(ids.shufu, ids.tangDi, rid++),
    pc(ids.gongBro2, ids.cousin2, rid++),
    pc(ids.shufu, ids.extraUncleChild1, rid++),
    pc(ids.shufu, ids.extraUncleChild2, rid++),
    pc(ids.shufu, ids.tangJie, rid++),
    pc(ids.fu, ids.me, rid++),
    pc(ids.mu, ids.me, rid++),
    pc(ids.fu, ids.elderBro, rid++),
    pc(ids.mu, ids.elderBro, rid++),
    pc(ids.fu, ids.jieJie, rid++),
    pc(ids.mu, ids.jieJie, rid++),
    pc(ids.elderBro, ids.nephew, rid++),
    pc(ids.elderBro, ids.niece, rid++),
    pc(ids.jieJie, ids.waiSheng, rid++),
    pc(ids.jieJie, ids.waiShengNv, rid++),
    pc(ids.waiGong, ids.mu, rid++),
    pc(ids.waiPo, ids.mu, rid++),
    pc(ids.waiGong, ids.muBro, rid++),
    pc(ids.waiPo, ids.muBro, rid++),
    pc(ids.waiGong, ids.muSis, rid++),
    pc(ids.waiPo, ids.muSis, rid++),
    pc(ids.muBro, ids.biaoGe, rid++),
    pc(ids.muBro, ids.biaoMei, rid++),
    pc(ids.muBro, ids.cousin4, rid++),
    pc(ids.muBro, ids.muCousin, rid++),
    pc(ids.muSis, ids.cousin5, rid++),
    pc(ids.me, ids.son, rid++),
    pc(ids.spouse, ids.son, rid++),
    pc(ids.me, ids.daughter, rid++),
    pc(ids.spouse, ids.daughter, rid++),
    pc(ids.me, ids.secondSon, rid++),
    pc(ids.spouse, ids.secondSon, rid++),
    pc(ids.son, ids.grandson, rid++),
    pc(ids.son, ids.granddaughter, rid++),
    pc(ids.grandson, ids.greatGrandson, rid++),
    pc(ids.secondSon, ids.secondGrandson, rid++),
    pc(ids.secondSon, ids.secondGranddaughter, rid++),
    pc(ids.nephew, ids.branchNephew, rid++),
    pc(ids.nephew, ids.branchNiece, rid++)
  ]

  const lineageMemberIds = nodes
    .filter((node) => node.memberType === 'LINEAGE_MEMBER')
    .map((node) => node.memberId)

  const spouseMemberIds = nodes
    .filter((node) => node.memberType === 'SPOUSE')
    .map((node) => node.memberId)

  return { nodes, edges, lineageMemberIds, spouseMemberIds, ids }
}
