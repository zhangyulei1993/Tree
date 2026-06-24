import type { FamilyTreeResult, TreeNode } from '@/types/api'

export type FamilyTreeViewMode = 'structure' | 'detail'

export interface FamilyTreeBuildInput {
  nodes: FamilyTreeResult['nodes']
  edges: FamilyTreeResult['edges']
  tree?: FamilyTreeResult['tree']
  /** 称谓视角成员；仅影响布局中的「我」标记与直系高亮，不嵌入固定称谓 */
  viewerMemberId?: number | null
}

export interface FamilyTreeUnit {
  memberIds: number[]
  members: TreeNode[]
}

export interface FamilyTreeGeneration {
  generation: number
  label: string
  units: FamilyTreeUnit[]
}

export interface BuiltFamilyTree {
  success: boolean
  layers: FamilyTreeGeneration[]
  orphanUnits: FamilyTreeUnit[]
  memberCount: number
  visibleMemberCount: number
  message?: string
}

export interface MemberGraph {
  parentIds: Map<number, number[]>
  childrenIds: Map<number, number[]>
  spouseIds: Map<number, number[]>
  nodeMap: Map<number, TreeNode>
}

/** 递归家谱分支：夫妻单元 + 子女子树 */
export interface FamilyTreeBranchNode {
  id: string
  parents: TreeNode[]
  childBranches: FamilyTreeBranchNode[]
}

export interface BuiltFamilyGraph {
  success: boolean
  roots: FamilyTreeBranchNode[]
  orphanBranches: FamilyTreeBranchNode[]
  memberCount: number
  renderedCount: number
  message?: string
}

export interface CoupleHierarchyDatum {
  id: string
  virtual?: boolean
  parents: TreeNode[]
  coupleWidth: number
  children?: CoupleHierarchyDatum[]
}

export interface RenderCoupleNode {
  id: string
  parents: TreeNode[]
  x: number
  y: number
  width: number
  height: number
  scrollAnchorId?: string
}

export interface RenderLinkSegment {
  x: number
  y: number
  width: number
  height: number
  orientation: 'h' | 'v'
}

export interface RenderTreeLink {
  id: string
  segments: RenderLinkSegment[]
  highlighted?: boolean
}

export interface FamilyTreeLayoutResult {
  success: boolean
  renderNodes: RenderCoupleNode[]
  renderLinks: RenderTreeLink[]
  canvasWidth: number
  canvasHeight: number
  memberCount: number
  renderedCount: number
  scrollIntoViewId?: string
  message?: string
}
