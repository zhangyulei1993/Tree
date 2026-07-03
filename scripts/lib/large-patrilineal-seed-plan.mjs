export const LARGE_FAMILY_NAME = 'Tree父系六代真实验收家'
export const LARGE_FAMILY_SURNAME = '张'
export const LARGE_UNLOCATED_NAME = '张待安'

const child = (key, name, gender, birthYear) => ({ key, name, gender, birthYear })
const spouse = (name, gender, birthYear) => ({ name, gender, birthYear })

export function buildLargePatrilinealSeedPlan(founderMemberId) {
  if (!founderMemberId) throw new Error('founderMemberId is required')

  const families = [
    {
      key: 'root',
      baseRef: '$founder',
      baseGender: 'MALE',
      spouse: spouse('李淑珍', 'FEMALE', 1932),
      children: [
        child('jianguo', '张建国', 'MALE', 1952),
        child('jianmin', '张建民', 'MALE', 1955),
        child('jianhua', '张建华', 'MALE', 1958)
      ]
    },
    {
      key: 'jianguo-family',
      baseRef: 'jianguo',
      baseGender: 'MALE',
      spouse: spouse('王秀兰', 'FEMALE', 1954),
      children: [
        child('zhiyuan', '张志远', 'MALE', 1975),
        child('zhiqiang', '张志强', 'MALE', 1978),
        child('zhilan', '张志兰', 'FEMALE', 1981)
      ]
    },
    {
      key: 'jianmin-family',
      baseRef: 'jianmin',
      baseGender: 'MALE',
      spouse: spouse('陈玉梅', 'FEMALE', 1957),
      children: [
        child('zhiping', '张志平', 'MALE', 1977),
        child('zhihai', '张志海', 'MALE', 1980),
        child('zhifang', '张志芳', 'FEMALE', 1983)
      ]
    },
    {
      key: 'jianhua-family',
      baseRef: 'jianhua',
      baseGender: 'MALE',
      spouse: spouse('赵桂芳', 'FEMALE', 1960),
      children: [
        child('zhicheng', '张志成', 'MALE', 1979),
        child('zhian', '张志安', 'MALE', 1982),
        child('zhimin', '张志敏', 'FEMALE', 1985)
      ]
    },
    {
      key: 'zhiyuan-family',
      baseRef: 'zhiyuan',
      baseGender: 'MALE',
      spouse: spouse('陈慧', 'FEMALE', 1977),
      children: [
        child('chengzhi', '张承志', 'MALE', 1998),
        child('chengya', '张承雅', 'FEMALE', 2000),
        child('chengli', '张承礼', 'MALE', 2003)
      ]
    },
    {
      key: 'zhiqiang-family',
      baseRef: 'zhiqiang',
      baseGender: 'MALE',
      spouse: spouse('刘敏', 'FEMALE', 1980),
      children: [
        child('chenghao', '张承浩', 'MALE', 2000),
        child('chengan', '张承安', 'MALE', 2002),
        child('chengning', '张承宁', 'FEMALE', 2005)
      ]
    },
    {
      key: 'zhiping-family',
      baseRef: 'zhiping',
      baseGender: 'MALE',
      spouse: spouse('何静', 'FEMALE', 1979),
      children: [
        child('chengping', '张承平', 'MALE', 2001),
        child('chengyue', '张承月', 'FEMALE', 2004),
        child('chengfeng', '张承峰', 'MALE', 2006)
      ]
    },
    {
      key: 'zhihai-family',
      baseRef: 'zhihai',
      baseGender: 'MALE',
      spouse: spouse('郭宁', 'FEMALE', 1982),
      children: [
        child('chenghai', '张承海', 'MALE', 2003),
        child('chengqing', '张承清', 'FEMALE', 2006),
        child('chengze', '张承泽', 'MALE', 2008)
      ]
    },
    {
      key: 'zhicheng-family',
      baseRef: 'zhicheng',
      baseGender: 'MALE',
      spouse: spouse('郑琳', 'FEMALE', 1981),
      children: [
        child('chengcheng', '张承成', 'MALE', 2004),
        child('chengyue2', '张承悦', 'FEMALE', 2007),
        child('chengkang', '张承康', 'MALE', 2009)
      ]
    },
    {
      key: 'zhian-family',
      baseRef: 'zhian',
      baseGender: 'MALE',
      spouse: spouse('方晴', 'FEMALE', 1984),
      children: [
        child('chengtai', '张承泰', 'MALE', 2006),
        child('chengxin', '张承欣', 'FEMALE', 2009),
        child('chengwen', '张承文', 'MALE', 2011)
      ]
    },
    {
      key: 'chengzhi-family',
      baseRef: 'chengzhi',
      baseGender: 'MALE',
      spouse: spouse('林雪', 'FEMALE', 2000),
      children: [
        child('siyuan', '张思远', 'MALE', 2021),
        child('sihan', '张思涵', 'FEMALE', 2023)
      ]
    },
    {
      key: 'chenghao-family',
      baseRef: 'chenghao',
      baseGender: 'MALE',
      spouse: spouse('孙丽华', 'FEMALE', 2002),
      children: [
        child('siqi', '张思齐', 'MALE', 2022),
        child('simin', '张思敏', 'FEMALE', 2024)
      ]
    },
    {
      key: 'siyuan-family',
      baseRef: 'siyuan',
      baseGender: 'MALE',
      spouse: spouse('苏婉', 'FEMALE', 2022),
      children: [
        child('nianzu', '张念祖', 'MALE', 2045)
      ]
    }
  ]

  const memberNames = [
    '张德厚',
    ...families.flatMap((family) => [
      family.spouse.name,
      ...family.children.map((item) => item.name)
    ]),
    LARGE_UNLOCATED_NAME
  ]
  if (new Set(memberNames).size !== memberNames.length) {
    throw new Error('large seed plan contains duplicate member names')
  }

  const relationshipCount = families.reduce(
    (total, family) => total + 1 + family.children.length * 2,
    0
  )

  return {
    founderMemberId,
    founderPatch: {
      name: '张德厚',
      gender: 'MALE',
      birthYear: 1930,
      isAlive: false,
      deathYear: 2012,
      userBindingPolicy: 'OPTIONAL'
    },
    families,
    unlocated: child('unlocated', LARGE_UNLOCATED_NAME, 'MALE', 1995),
    accountTargets: {
      familyAdmin: '张承志',
      member: '张思远'
    },
    expected: {
      memberCount: memberNames.length,
      relationshipCount,
      coupleUnits: families.length,
      unlocatedCount: 1,
      renderedCount: memberNames.length - 1
    }
  }
}
