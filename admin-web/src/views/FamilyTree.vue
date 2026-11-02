<template>
  <div class="page-stack">
    <PageHeader :title="isShowcase ? '帝王世家树' : '家庭树'" :description="isShowcase ? '独立运营展示内容，与用户家庭数据无关。' : '后台成员管理主入口：选中成员后查看资料、亲属关系与后续操作。'">
      <el-button @click="$router.push(isShowcase ? '/admin/showcase-families' : `/admin/families/${familyId}`)">返回{{ isShowcase ? '帝王世家配置' : '家庭详情' }}</el-button>
      <el-button v-if="!isShowcase" @click="openMemberIndex()">成员索引</el-button>
    </PageHeader>
    <el-alert v-if="error" :title="error" type="error" show-icon />
    <section v-else-if="tree" class="page-stack">
      <el-card v-if="selectedNode && isShowcase" shadow="never" class="detail-card">
        <template #header>
          <div class="node-info-header">
            <div>
              <strong>{{ selectedNode.displayName }}</strong>
              <span>成员 {{ tree.nodes.length }} · 亲属关系 {{ tree.edges.length }} · {{ isShowcase ? '独立展示' : `家谱版本 ${tree.graphVersion}` }}</span>
            </div>
            <div v-if="!isShowcase" class="node-actions">
              <el-button @click="openMemberIndex(selectedNode.memberId)">在成员索引中查看</el-button>
              <el-button type="primary" plain @click="openAdminMemberEdit(selectedNode)">编辑成员</el-button>
              <el-tooltip v-if="deleteBlockReason(selectedNode)" :content="deleteBlockReason(selectedNode)" placement="top">
                <span><el-button type="danger" plain disabled>删除成员</el-button></span>
              </el-tooltip>
              <el-button v-else type="danger" plain :loading="deletingMember" @click="removeAdminMember(selectedNode)">删除成员</el-button>
            </div>
          </div>
        </template>
        <el-descriptions v-if="isShowcase" :column="4" border>
          <el-descriptions-item label="成员 ID">{{ selectedNode.memberId }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ selectedNode.displayName }}</el-descriptions-item>
          <el-descriptions-item label="称号 / 封号">{{ selectedNode.identityText || '历史人物' }}</el-descriptions-item>
          <el-descriptions-item label="性别">{{ genderText(selectedNode.gender) }}</el-descriptions-item>
          <el-descriptions-item label="父母">{{ parentNames(selectedNode.memberId).join('、') || '未记录' }}</el-descriptions-item>
          <el-descriptions-item label="配偶">{{ spouseNames(selectedNode.memberId).join('、') || '未记录' }}</el-descriptions-item>
          <el-descriptions-item label="子女">{{ childNames(selectedNode.memberId).join('、') || '未记录' }}</el-descriptions-item>
        </el-descriptions>
        <template v-else>
          <el-descriptions :column="4" border>
            <el-descriptions-item label="成员 ID">{{ selectedNode.memberId }}</el-descriptions-item>
            <el-descriptions-item label="姓名">{{ selectedNode.displayName }}</el-descriptions-item>
            <el-descriptions-item label="性别">{{ genderText(selectedNode.gender) }}</el-descriptions-item>
            <el-descriptions-item label="状态">{{ lifeText(selectedNode.isLiving) }}</el-descriptions-item>
            <el-descriptions-item label="出生日期">{{ selectedNode.birthDate || '未记录' }}</el-descriptions-item>
            <el-descriptions-item label="节点类型">{{ memberTypeText(selectedNode.memberType) }}</el-descriptions-item>
            <el-descriptions-item label="绑定状态">{{ bindingText(selectedNode.userBindingState) }}</el-descriptions-item>
            <el-descriptions-item label="管理口径">家庭树成员卡片</el-descriptions-item>
          </el-descriptions>
          <div class="relation-grid">
            <div v-for="block in selectedRelationBlocks" :key="block.key" class="relation-block" :class="{ empty: block.members.length === 0 }">
              <span>{{ block.label }}</span>
              <div v-if="block.members.length" class="relation-members">
                <button v-for="member in block.members" :key="member.memberId" type="button" class="relation-chip" @click="selectMember(member.memberId)">
                  {{ member.displayName }}
                </button>
              </div>
              <strong v-else>未记录</strong>
              <small>{{ block.hint }}</small>
            </div>
          </div>
          <p class="operation-note">后台变更会写入操作日志并刷新家谱版本；小程序会按同一套成员与关系数据重新计算亲属称谓。</p>
        </template>
      </el-card>
      <div :class="['tree-workspace', { 'tree-workspace--showcase': isShowcase }]">
      <el-card shadow="never" class="tree-card">
        <template #header><div class="card-header"><div><strong>树结构</strong><span>配偶以同一单元并列；父母与子女用连线表示。</span></div><el-button text @click="loadTree">刷新</el-button></div></template>
        <el-empty v-if="tree.nodes.length === 0" description="该家庭暂未创建成员节点" />
        <el-empty v-else-if="!layout.success" description="暂无可展示的主干关系，可在关系明细中核对" />
        <el-scrollbar v-else class="tree-scroll"><div class="tree-canvas" :style="{ width: `${layout.canvasWidth}px`, height: `${layout.canvasHeight}px` }">
          <div v-for="link in layout.links" :key="link.id" class="links-layer" aria-hidden="true"><i v-for="(segment, index) in link.segments" :key="`${link.id}-${index}`" class="link-segment" :style="{ left: `${segment.x}px`, top: `${segment.y}px`, width: `${segment.width}px`, height: `${segment.height}px` }" /></div>
          <div v-for="unit in layout.units" :key="unit.id" class="couple-unit" :style="{ left: `${unit.x}px`, top: `${unit.y}px`, width: `${unit.width}px` }"><template v-for="(node, index) in unit.parents" :key="node.memberId"><span v-if="index > 0" class="spouse-join" aria-hidden="true"><i /></span><button class="member-card" :class="[isShowcase ? showcaseNodeClass(node) : '', { selected: selectedMemberId === node.memberId, deceased: node.isLiving === false }]" type="button" @click="selectMember(node.memberId)"><strong>{{ node.displayName }}</strong><small>{{ isShowcase ? (node.identityText || '历史人物') : `${genderText(node.gender)} · ${lifeText(node.isLiving)}` }}</small><small v-if="!isShowcase" class="binding">{{ bindingText(node.userBindingState) }}</small><span v-if="isShowcase && hasKeyIdentity(node)" class="identity-seal">名</span><span v-if="index > 0" class="spouse-badge">配偶</span><template v-if="isShowcase"><span class="node-action node-add" aria-label="添加亲属" title="添加配偶或子女" @click.stop="openRelativeCreate(node.memberId)">＋</span><span class="node-action node-remove" aria-label="删除成员" title="删除成员" @click.stop="removeMember(node.memberId)">−</span></template></button></template></div>
        </div></el-scrollbar>
        <p class="tree-footnote">共 {{ tree.nodes.length }} 位成员，树中展示 {{ layout.renderedCount }} 位；可左右、上下滚动查看。</p>
        <section v-if="layout.unlocatedNodes.length" class="unlocated-section"><strong>未定位成员</strong><span>以下成员尚未接入当前主干关系图。</span><div class="unlocated-list"><button v-for="node in layout.unlocatedNodes" :key="node.memberId" class="member-card unlocated-card" :class="{ selected: selectedMemberId === node.memberId, deceased: node.isLiving === false }" type="button" @click="selectMember(node.memberId)"><strong>{{ node.displayName }}</strong><small>{{ isShowcase ? (node.identityText || '历史人物') : `${genderText(node.gender)} · ${lifeText(node.isLiving)}` }}</small><small v-if="!isShowcase" class="binding">{{ bindingText(node.userBindingState) }}</small><template v-if="isShowcase"><span class="node-action node-add" aria-label="添加亲属" title="添加配偶或子女" @click.stop="openRelativeCreate(node.memberId)">＋</span><span class="node-action node-remove" aria-label="删除成员" title="删除成员" @click.stop="removeMember(node.memberId)">−</span></template></button></div></section>
      </el-card>
      <aside v-if="!isShowcase" class="member-inspector">
        <el-card v-if="selectedNode" shadow="never" class="detail-card inspector-card">
          <template #header>
            <div class="node-info-header">
              <div>
                <strong>{{ selectedNode.displayName }}</strong>
                <span>成员 {{ tree.nodes.length }} · 亲属关系 {{ tree.edges.length }} · 家谱版本 {{ tree.graphVersion }}</span>
              </div>
              <div class="node-actions">
                <el-button @click="openMemberIndex(selectedNode.memberId)">成员索引</el-button>
                <el-button type="primary" plain @click="openAdminMemberEdit(selectedNode)">编辑</el-button>
                <el-tooltip v-if="deleteBlockReason(selectedNode)" :content="deleteBlockReason(selectedNode)" placement="top">
                  <span><el-button type="danger" plain disabled>删除</el-button></span>
                </el-tooltip>
                <el-button v-else type="danger" plain :loading="deletingMember" @click="removeAdminMember(selectedNode)">删除</el-button>
              </div>
            </div>
          </template>
          <div class="relative-action-panel">
            <div class="relative-action-title">
              <strong>从该成员添加亲属</strong>
              <span>按小程序同一套规则写入关系</span>
            </div>
            <div class="relative-action-buttons">
              <el-tooltip v-for="action in relativeActionOptions" :key="action.addType" :content="action.disabledReason || action.hint" placement="top">
                <span>
                  <el-button :type="action.addType === 'ADD_CHILD' ? 'primary' : undefined" plain :disabled="Boolean(action.disabledReason)" @click="openAdminRelativeCreate(action.addType)">
                    {{ action.label }}
                  </el-button>
                </span>
              </el-tooltip>
            </div>
          </div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="成员 ID">{{ selectedNode.memberId }}</el-descriptions-item>
            <el-descriptions-item label="性别">{{ genderText(selectedNode.gender) }}</el-descriptions-item>
            <el-descriptions-item label="状态">{{ lifeText(selectedNode.isLiving) }}</el-descriptions-item>
            <el-descriptions-item label="出生日期">{{ selectedNode.birthDate || '未记录' }}</el-descriptions-item>
            <el-descriptions-item label="节点类型">{{ memberTypeText(selectedNode.memberType) }}</el-descriptions-item>
            <el-descriptions-item label="绑定状态">{{ bindingText(selectedNode.userBindingState) }}</el-descriptions-item>
          </el-descriptions>
          <div class="relation-grid">
            <div v-for="block in selectedRelationBlocks" :key="block.key" class="relation-block" :class="{ empty: block.members.length === 0 }">
              <span>{{ block.label }}</span>
              <div v-if="block.members.length" class="relation-members">
                <button v-for="member in block.members" :key="member.memberId" type="button" class="relation-chip" @click="selectMember(member.memberId)">
                  {{ member.displayName }}
                </button>
              </div>
              <strong v-else>未记录</strong>
              <small>{{ block.hint }}</small>
            </div>
          </div>
          <p class="operation-note">后台变更会写入操作日志并刷新家谱版本；小程序会按同一套成员与关系数据重新计算亲属称谓。</p>
        </el-card>
      </aside>
      </div>
      <el-card shadow="never"><template #header><strong>关系明细</strong></template><el-table :data="tree.edges" max-height="360" stripe><el-table-column label="关系" width="120"><template #default="{ row }">{{ relationshipText(row.relationshipType) }}</template></el-table-column><el-table-column label="起点" min-width="180" show-overflow-tooltip><template #default="{ row }">{{ memberName(row.fromMemberId) }}</template></el-table-column><el-table-column label="终点" min-width="180" show-overflow-tooltip><template #default="{ row }">{{ memberName(row.toMemberId) }}</template></el-table-column><el-table-column label="属性" width="132"><template #default="{ row }">{{ row.parentLinkType || '—' }}</template></el-table-column></el-table></el-card>
      <el-card v-if="isShowcase" shadow="never"><template #header><strong>成员配置</strong></template><el-table :data="tree.nodes" max-height="360" stripe><el-table-column prop="displayName" label="姓名" min-width="160" show-overflow-tooltip /><el-table-column prop="identityText" label="称号 / 封号" min-width="180" show-overflow-tooltip /><el-table-column label="性别" width="92"><template #default="{ row }">{{ genderText(row.gender) }}</template></el-table-column><el-table-column label="操作" width="84" fixed="right"><template #default="{ row }"><el-button size="small" @click="openMemberEdit(row)">编辑</el-button></template></el-table-column></el-table></el-card>
    </section>
    <el-dialog v-model="adminMemberDialogVisible" title="编辑成员" width="520px" destroy-on-close>
      <el-form :model="adminMemberForm" label-width="96px">
        <el-form-item label="姓名" required><el-input v-model="adminMemberForm.displayName" maxlength="80" /></el-form-item>
        <el-form-item label="性别" required><el-radio-group v-model="adminMemberForm.gender"><el-radio-button value="MALE">男</el-radio-button><el-radio-button value="FEMALE">女</el-radio-button><el-radio-button value="UNKNOWN">未记录</el-radio-button></el-radio-group></el-form-item>
        <el-form-item label="出生日期"><el-date-picker v-model="adminMemberForm.birthDate" value-format="YYYY-MM-DD" type="date" placeholder="未记录" clearable style="width:100%" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="adminMemberForm.isLiving" style="width:100%"><el-option label="健在" value="ALIVE" /><el-option label="已故" value="DECEASED" /><el-option label="未记录" value="UNKNOWN" /></el-select></el-form-item>
        <el-form-item label="绑定策略"><el-select v-model="adminMemberForm.userBindingPolicy" style="width:100%"><el-option label="可选绑定" value="OPTIONAL" /><el-option label="需要绑定" value="REQUIRED" /><el-option label="无需绑定" value="NOT_REQUIRED" :disabled="editingAdminMember?.userBindingState === 'BOUND'" /></el-select></el-form-item>
        <el-alert v-if="editingAdminMember?.userBindingState === 'BOUND'" title="该成员已绑定账号，不能直接改为无需绑定；如需调整，请先处理账号绑定。" type="info" show-icon :closable="false" />
      </el-form>
      <template #footer><el-button @click="adminMemberDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingAdminMember" @click="submitAdminMember">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="adminRelativeDialogVisible" :title="adminRelativeDialogTitle" width="540px" destroy-on-close>
      <el-form :model="adminRelativeForm" label-width="96px">
        <el-form-item label="姓名" required><el-input v-model="adminRelativeForm.displayName" maxlength="80" /></el-form-item>
        <el-form-item label="性别" required><el-radio-group v-model="adminRelativeForm.gender" :disabled="adminRelativeGenderLocked"><el-radio-button value="MALE">男</el-radio-button><el-radio-button value="FEMALE">女</el-radio-button><el-radio-button v-if="!adminRelativeGenderRequired" value="UNKNOWN">未记录</el-radio-button></el-radio-group></el-form-item>
        <el-form-item v-if="adminRelativeForm.addType === 'ADD_FATHER' || adminRelativeForm.addType === 'ADD_MOTHER'" label="成员身份"><el-radio-group v-model="adminRelativeForm.memberType"><el-radio-button value="LINEAGE_MEMBER">本家成员</el-radio-button><el-radio-button value="SPOUSE" :disabled="!canParentBeSpouse(adminRelativeForm.addType, adminRelativeBase)">本家成员配偶</el-radio-button></el-radio-group></el-form-item>
        <el-form-item v-if="adminRelativeForm.addType !== 'ADD_SPOUSE'" label="关系属性"><el-select v-model="adminRelativeForm.parentLinkType" style="width:100%"><el-option label="主要亲属" value="PRIMARY" /><el-option label="继亲" value="STEP" /><el-option label="收养" value="ADOPTIVE" /><el-option label="承嗣" value="SUCCESSION" /><el-option label="仅备注" value="NOTE_ONLY" /><el-option label="其他" value="OTHER" /></el-select></el-form-item>
        <el-form-item label="出生日期"><el-date-picker v-model="adminRelativeForm.birthDate" value-format="YYYY-MM-DD" type="date" placeholder="未记录" clearable style="width:100%" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="adminRelativeForm.isLiving" style="width:100%"><el-option label="健在" value="ALIVE" /><el-option label="已故" value="DECEASED" /><el-option label="未记录" value="UNKNOWN" /></el-select></el-form-item>
        <el-form-item label="绑定策略"><el-select v-model="adminRelativeForm.userBindingPolicy" style="width:100%"><el-option label="可选绑定" value="OPTIONAL" /><el-option label="需要绑定" value="REQUIRED" /><el-option label="无需绑定" value="NOT_REQUIRED" /></el-select></el-form-item>
        <el-alert :title="adminRelativeHint" type="info" show-icon :closable="false" />
      </el-form>
      <template #footer><el-button @click="adminRelativeDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingAdminRelative" @click="submitAdminRelative">保存亲属</el-button></template>
    </el-dialog>
    <el-dialog v-model="memberDialogVisible" :title="editingMember ? '编辑成员' : '新增成员'" width="460px" destroy-on-close><el-form :model="memberForm" label-width="92px"><el-form-item label="姓名" required><el-input v-model="memberForm.displayName" maxlength="80" /></el-form-item><el-form-item label="称号 / 封号"><el-input v-model="memberForm.identityText" maxlength="120" placeholder="如：汉高祖、汉高后、太子" /></el-form-item><el-form-item label="性别" required><el-radio-group v-model="memberForm.gender"><el-radio-button value="MALE">男</el-radio-button><el-radio-button value="FEMALE">女</el-radio-button><el-radio-button value="UNKNOWN">未记录</el-radio-button></el-radio-group></el-form-item><el-form-item label="出生年份"><el-input-number v-model="memberForm.birthYear" :min="1" :max="9999" controls-position="right" /></el-form-item><el-form-item label="排序"><el-input-number v-model="memberForm.sortOrder" :min="0" :max="9999" controls-position="right" /></el-form-item></el-form><template #footer><el-button @click="memberDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingTree" @click="submitMember">保存</el-button></template></el-dialog>
    <el-dialog v-model="relativeDialogVisible" :title="relativeAnchor ? `向「${relativeAnchor.displayName}」添加亲属` : '添加亲属'" width="520px" destroy-on-close><el-form :model="relativeForm" label-width="94px"><el-form-item label="添加类型"><el-radio-group v-model="relativeForm.relationKind"><el-radio-button value="CHILD">子女</el-radio-button><el-radio-button value="SPOUSE">配偶</el-radio-button></el-radio-group></el-form-item><el-form-item :label="relativeForm.relationKind === 'CHILD' ? '子女姓名' : '配偶姓名'" required><el-input v-model="relativeForm.member.displayName" maxlength="80" /></el-form-item><el-form-item label="称号 / 封号"><el-input v-model="relativeForm.member.identityText" maxlength="120" placeholder="如：汉惠帝、皇后、太子" /></el-form-item><el-form-item label="性别" required><el-radio-group v-model="relativeForm.member.gender"><el-radio-button value="MALE">男</el-radio-button><el-radio-button value="FEMALE">女</el-radio-button><el-radio-button value="UNKNOWN">未记录</el-radio-button></el-radio-group></el-form-item><el-form-item label="出生年份"><el-input-number v-model="relativeForm.member.birthYear" :min="1" :max="9999" controls-position="right" /></el-form-item><template v-if="relativeForm.relationKind === 'CHILD'"><el-divider content-position="left">另一位家长</el-divider><el-form-item label="关系处理"><el-radio-group v-model="relativeForm.spouseMode"><el-radio-button value="NONE">暂不设置</el-radio-button><el-radio-button v-if="anchorSpouses.length" value="EXISTING">选择已有配偶</el-radio-button><el-radio-button value="NEW">同时新增配偶</el-radio-button></el-radio-group></el-form-item><el-form-item v-if="relativeForm.spouseMode === 'EXISTING'" label="已有配偶"><el-select v-model="relativeForm.existingSpouseMemberId" style="width:100%"><el-option v-for="node in anchorSpouses" :key="node.memberId" :label="node.displayName" :value="node.memberId" /></el-select></el-form-item><template v-if="relativeForm.spouseMode === 'NEW'"><el-form-item label="配偶姓名" required><el-input v-model="relativeForm.newSpouse.displayName" maxlength="80" /></el-form-item><el-form-item label="称号 / 封号"><el-input v-model="relativeForm.newSpouse.identityText" maxlength="120" placeholder="如：皇后、夫人、王妃" /></el-form-item><el-form-item label="配偶性别" required><el-radio-group v-model="relativeForm.newSpouse.gender"><el-radio-button value="MALE">男</el-radio-button><el-radio-button value="FEMALE">女</el-radio-button><el-radio-button value="UNKNOWN">未记录</el-radio-button></el-radio-group></el-form-item><el-form-item label="出生年份"><el-input-number v-model="relativeForm.newSpouse.birthYear" :min="1" :max="9999" controls-position="right" /></el-form-item></template></template></el-form><template #footer><el-button @click="relativeDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingTree" @click="submitRelative">保存关系</el-button></template></el-dialog>
    <el-dialog v-model="relationshipDialogVisible" title="新增关系" width="500px" destroy-on-close><el-form :model="relationshipForm" label-width="84px"><el-form-item label="关系类型" required><el-radio-group v-model="relationshipForm.relationshipType"><el-radio-button value="PARENT_CHILD">父母子女</el-radio-button><el-radio-button value="SPOUSE">配偶</el-radio-button></el-radio-group></el-form-item><el-form-item :label="relationshipForm.relationshipType === 'PARENT_CHILD' ? '父母' : '成员一'" required><el-select v-model="relationshipForm.fromMemberId" filterable style="width:100%"><el-option v-for="node in tree?.nodes || []" :key="node.memberId" :label="node.displayName" :value="node.memberId" /></el-select></el-form-item><el-form-item :label="relationshipForm.relationshipType === 'PARENT_CHILD' ? '子女' : '成员二'" required><el-select v-model="relationshipForm.toMemberId" filterable style="width:100%"><el-option v-for="node in tree?.nodes || []" :key="node.memberId" :label="node.displayName" :value="node.memberId" /></el-select></el-form-item></el-form><template #footer><el-button @click="relationshipDialogVisible = false">取消</el-button><el-button type="primary" :loading="savingTree" @click="submitRelationship">保存</el-button></template></el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { getApiErrorMessage } from '@/api/client'
import { createFamilyRelative, deleteFamilyMember, getFamilyTree, updateFamilyMember, type ManagedRelativeAddType } from '@/api/management'
import { addShowcaseRelative, createShowcaseFamilyMember, createShowcaseFamilyRelationship, deleteShowcaseFamilyMember, getShowcaseFamilyTree, updateShowcaseFamilyMember, type ShowcaseFamilyMemberInput } from '@/api/showcaseFamily'
import PageHeader from '@/components/PageHeader.vue'
import type { AdminFamilyTree, AdminFamilyTreeNode } from '@/types/api'

type Branch = { id: string; parents: AdminFamilyTreeNode[]; children: Branch[]; width: number; x: number; y: number }
type Segment = { x: number; y: number; width: number; height: number }
type Unit = Pick<Branch, 'id' | 'parents' | 'width' | 'x' | 'y'>
const CARD_WIDTH = 102, CARD_HEIGHT = 66, CARD_GAP = 4, UNIT_GAP = 12, LEVEL_GAP = 82, PADDING = 16, STROKE = 2
const route = useRoute(), router = useRouter(), familyId = String(route.params.familyId || route.params.showcaseFamilyId), isShowcase = route.meta.showcase === true, tree = ref<AdminFamilyTree | null>(null), error = ref(''), selectedMemberId = ref<number | null>(null)
const memberDialogVisible = ref(false), relationshipDialogVisible = ref(false), relativeDialogVisible = ref(false), savingTree = ref(false), editingMember = ref<AdminFamilyTreeNode | null>(null), relativeAnchor = ref<AdminFamilyTreeNode | null>(null)
const adminMemberDialogVisible = ref(false), savingAdminMember = ref(false), deletingMember = ref(false), editingAdminMember = ref<AdminFamilyTreeNode | null>(null)
const adminRelativeDialogVisible = ref(false), savingAdminRelative = ref(false), adminRelativeBase = ref<AdminFamilyTreeNode | null>(null)
const memberForm = reactive<ShowcaseFamilyMemberInput>({ displayName: '', identityText: '', gender: 'UNKNOWN', birthYear: null, sortOrder: 0 })
const adminMemberForm = reactive({ displayName: '', gender: 'UNKNOWN', birthDate: '', isLiving: 'UNKNOWN' as 'ALIVE' | 'DECEASED' | 'UNKNOWN', userBindingPolicy: 'OPTIONAL' })
const adminRelativeForm = reactive({ addType: 'ADD_CHILD' as ManagedRelativeAddType, displayName: '', gender: 'UNKNOWN', birthDate: '', isLiving: 'ALIVE' as 'ALIVE' | 'DECEASED' | 'UNKNOWN', userBindingPolicy: 'OPTIONAL', memberType: 'LINEAGE_MEMBER', parentLinkType: 'PRIMARY' })
const relationshipForm = reactive({ fromMemberId: 0, toMemberId: 0, relationshipType: 'PARENT_CHILD' as 'PARENT_CHILD' | 'SPOUSE' })
const relativeForm = reactive({ relationKind: 'CHILD' as 'CHILD' | 'SPOUSE', member: { displayName: '', identityText: '', gender: 'UNKNOWN' as ShowcaseFamilyMemberInput['gender'], birthYear: null as number | null, sortOrder: 0 }, spouseMode: 'NONE' as 'NONE' | 'EXISTING' | 'NEW', existingSpouseMemberId: 0, newSpouse: { displayName: '', identityText: '', gender: 'UNKNOWN' as ShowcaseFamilyMemberInput['gender'], birthYear: null as number | null, sortOrder: 0 } })
const nodeById = computed(() => new Map(tree.value?.nodes.map((node) => [node.memberId, node]) || []))
const parentEdges = computed(() => tree.value?.edges.filter((edge) => edge.relationshipType === 'PARENT_CHILD') || [])
const anchorSpouses = computed(() => {
  const anchorID = relativeAnchor.value?.memberId
  if (!anchorID) return []
  return (tree.value?.edges || []).filter((edge) => edge.relationshipType === 'SPOUSE' && (edge.fromMemberId === anchorID || edge.toMemberId === anchorID)).map((edge) => nodeById.value.get(edge.fromMemberId === anchorID ? edge.toMemberId : edge.fromMemberId)).filter((node): node is AdminFamilyTreeNode => Boolean(node))
})
const selectedNode = computed(() => selectedMemberId.value ? nodeById.value.get(selectedMemberId.value) || null : null)
const adminRelativeDialogTitle = computed(() => adminRelativeBase.value ? `向「${adminRelativeBase.value.displayName}」添加${relativeAddTypeText(adminRelativeForm.addType)}` : '添加亲属')
const adminRelativeGenderLocked = computed(() => ['ADD_FATHER', 'ADD_MOTHER', 'ADD_SPOUSE'].includes(adminRelativeForm.addType))
const adminRelativeGenderRequired = computed(() => ['ADD_CHILD', 'ADD_SIBLING'].includes(adminRelativeForm.addType))
const adminRelativeHint = computed(() => {
  const map: Record<ManagedRelativeAddType, string> = {
    ADD_FATHER: '父亲会作为当前成员的上级节点；主要父亲已存在时会阻断。',
    ADD_MOTHER: '母亲会作为当前成员的上级节点；主要母亲已存在时会阻断。',
    ADD_CHILD: '子女会挂在当前成员下方，并参与小程序亲属称谓计算。',
    ADD_SPOUSE: '配偶会与当前成员并列展示；配偶节点不能继续作为主干扩展。',
    ADD_SIBLING: '同辈不会保存为兄弟姐妹关系，而是复用当前成员已有父母生成共同父母关系。',
  }
  return map[adminRelativeForm.addType]
})
const relativeActionOptions = computed(() => {
  const node = selectedNode.value
  const actions: { addType: ManagedRelativeAddType; label: string; hint: string; disabledReason: string }[] = [
    { addType: 'ADD_FATHER', label: '父亲', hint: '添加父亲节点', disabledReason: '' },
    { addType: 'ADD_MOTHER', label: '母亲', hint: '添加母亲节点', disabledReason: '' },
    { addType: 'ADD_SPOUSE', label: '配偶', hint: '添加配偶节点', disabledReason: '' },
    { addType: 'ADD_CHILD', label: '子女', hint: '添加子女节点', disabledReason: '' },
    { addType: 'ADD_SIBLING', label: '同辈', hint: '通过共同父母添加兄弟姐妹', disabledReason: '' },
  ]
  return actions.map((action) => ({
    ...action,
    disabledReason: node ? relativeActionDisabledReason(node, action.addType) : '请先选择成员',
  }))
})
const selectedRelationBlocks = computed(() => {
  const memberId = selectedMemberId.value
  if (!memberId) return []
  const siblings = siblingMembers(memberId)
  return [
    { key: 'parents', label: '父母', members: relationMembers(parentIds(memberId)), hint: '决定上代结构与同辈联动' },
    { key: 'spouses', label: '配偶', members: relationMembers(spouseIds(memberId)), hint: '影响配偶并列和共同子女' },
    { key: 'children', label: '子女', members: relationMembers(childIds(memberId)), hint: '影响下代展开和删除限制' },
    { key: 'siblings', label: '同辈', members: siblings, hint: '由共同父母推导，不单独入库' },
  ]
})
const isLineage = (node?: AdminFamilyTreeNode) => String(node?.memberType || '').toUpperCase() === 'LINEAGE_MEMBER'
const birthValue = (node: AdminFamilyTreeNode) => { const value = String(node.birthDate || '').slice(0, 10).replace(/-/g, ''); return /^\d{8}$/.test(value) ? Number(value) : Number.MAX_SAFE_INTEGER }
const compareBirth = (left: AdminFamilyTreeNode, right: AdminFamilyTreeNode) => birthValue(left) - birthValue(right) || left.memberId - right.memberId

function sortChildren(ids: number[], nodes: Map<number, AdminFamilyTreeNode>) {
  const male: AdminFamilyTreeNode[] = [], female: AdminFamilyTreeNode[] = [], other: AdminFamilyTreeNode[] = []
  ids.forEach((id) => { const node = nodes.get(id); if (!node) return; if (node.gender === 'MALE') male.push(node); else if (node.gender === 'FEMALE') female.push(node); else other.push(node) })
  female.sort((left, right) => (left.surname || left.displayName.slice(0, 1)).localeCompare(right.surname || right.displayName.slice(0, 1), 'zh-CN') || left.displayName.localeCompare(right.displayName, 'zh-CN') || left.memberId - right.memberId)
  return [...male.sort(compareBirth), ...female, ...other.sort(compareBirth)].map((node) => node.memberId)
}

const layout = computed(() => {
  const nodes = tree.value?.nodes || [], edges = tree.value?.edges || [], nodeMap = new Map(nodes.map((node) => [node.memberId, node]))
  const parents = new Map<number, number[]>(), children = new Map<number, number[]>(), spouses = new Map<number, number[]>()
  const append = (map: Map<number, number[]>, key: number, value: number) => { const values = map.get(key) || []; if (key !== value && !values.includes(value)) values.push(value); map.set(key, values) }
  edges.forEach((edge) => { if (edge.relationshipType === 'PARENT_CHILD') { append(parents, edge.toMemberId, edge.fromMemberId); append(children, edge.fromMemberId, edge.toMemberId) } else if (edge.relationshipType === 'SPOUSE') { append(spouses, edge.fromMemberId, edge.toMemberId); append(spouses, edge.toMemberId, edge.fromMemberId) } })
  const lineageIds = nodes.filter(isLineage).map((node) => node.memberId), lineageSet = new Set(lineageIds), adjacency = new Map<number, Set<number>>(lineageIds.map((id) => [id, new Set()]))
  parentEdges.value.forEach((edge) => { if (lineageSet.has(edge.fromMemberId) && lineageSet.has(edge.toMemberId)) { adjacency.get(edge.fromMemberId)?.add(edge.toMemberId); adjacency.get(edge.toMemberId)?.add(edge.fromMemberId) } })
  const visited = new Set<number>(), components: number[][] = []
  lineageIds.forEach((start) => { if (visited.has(start)) return; const component: number[] = [], queue = [start]; visited.add(start); while (queue.length) { const current = queue.shift()!; component.push(current); adjacency.get(current)?.forEach((next) => { if (!visited.has(next)) { visited.add(next); queue.push(next) } }) }; components.push(component) })
  const main = components.sort((left, right) => right.length - left.length)[0] || [], mainSet = new Set(main)
  const rootId = main.filter((id) => !(parents.get(id) || []).some((parentId) => mainSet.has(parentId))).map((id) => nodeMap.get(id)).filter((node): node is AdminFamilyTreeNode => Boolean(node)).sort((left, right) => (left.gender === 'MALE' ? -1 : 0) - (right.gender === 'MALE' ? -1 : 0) || compareBirth(left, right) || left.displayName.localeCompare(right.displayName, 'zh-CN'))[0]?.memberId
  if (!rootId) return { success: false, units: [] as Unit[], links: [] as { id: string; segments: Segment[] }[], canvasWidth: 0, canvasHeight: 0, renderedCount: 0, unlocatedNodes: nodes.filter(isLineage) }
  const rendered = new Set<number>()
  const buildBranch = (anchorId: number, path: Set<number>): Branch | null => {
    if (path.has(anchorId) || rendered.has(anchorId)) return null
    const anchor = nodeMap.get(anchorId); if (!anchor || !isLineage(anchor)) return null
    const nextPath = new Set(path); nextPath.add(anchorId); rendered.add(anchorId)
    const pair = (spouses.get(anchorId) || []).map((id) => nodeMap.get(id)).filter((node): node is AdminFamilyTreeNode => node !== undefined && !rendered.has(node.memberId)).sort(compareBirth)
    pair.forEach((node) => { rendered.add(node.memberId); nextPath.add(node.memberId) })
    const family = [anchor, ...pair], childIds = new Set<number>(), branch: Branch = { id: family.map((node) => node.memberId).sort((a, b) => a - b).join('-'), parents: family, children: [], width: 0, x: 0, y: 0 }
    family.forEach((parent) => { ;[...(children.get(parent.memberId) || []), ...((spouses.get(parent.memberId) || []).flatMap((id) => children.get(id) || []))].filter((id) => isLineage(nodeMap.get(id))).forEach((id) => childIds.add(id)) })
    sortChildren([...childIds], nodeMap).forEach((childId) => { const child = buildBranch(childId, nextPath); if (child) branch.children.push(child) })
    return branch
  }
  const root = buildBranch(rootId, new Set()); if (!root) return { success: false, units: [] as Unit[], links: [] as { id: string; segments: Segment[] }[], canvasWidth: 0, canvasHeight: 0, renderedCount: 0, unlocatedNodes: nodes.filter(isLineage) }
  const unitWidth = (branch: Branch) => branch.parents.length * CARD_WIDTH + Math.max(0, branch.parents.length - 1) * CARD_GAP
  const measure = (branch: Branch): number => { const childWidth = branch.children.reduce((sum, child) => sum + measure(child), 0) + Math.max(0, branch.children.length - 1) * UNIT_GAP; return (branch.width = Math.max(unitWidth(branch), childWidth)) }
  const width = measure(root), units: Unit[] = [], links: { id: string; segments: Segment[] }[] = []; let bottom = 0
  const place = (branch: Branch, left: number, level: number) => { branch.x = left + (branch.width - unitWidth(branch)) / 2; branch.y = PADDING + level * (CARD_HEIGHT + LEVEL_GAP); bottom = Math.max(bottom, branch.y + CARD_HEIGHT); units.push({ id: branch.id, parents: branch.parents, width: unitWidth(branch), x: branch.x, y: branch.y }); let childLeft = left; branch.children.forEach((child) => { place(child, childLeft, level + 1); const sourceX = branch.x + unitWidth(branch) / 2, sourceY = branch.y + CARD_HEIGHT, targetX = child.x + unitWidth(child) / 2, targetY = child.y, middleY = sourceY + (targetY - sourceY) / 2; const segments: Segment[] = [{ x: sourceX - STROKE / 2, y: sourceY, width: STROKE, height: Math.max(middleY - sourceY, STROKE) }]; if (Math.abs(targetX - sourceX) > STROKE) segments.push({ x: Math.min(sourceX, targetX), y: middleY - STROKE / 2, width: Math.abs(targetX - sourceX), height: STROKE }); segments.push({ x: targetX - STROKE / 2, y: middleY, width: STROKE, height: Math.max(targetY - middleY, STROKE) }); links.push({ id: `${branch.id}-${child.id}`, segments }); childLeft += child.width + UNIT_GAP }) }
  place(root, PADDING, 0)
  return { success: true, units, links, canvasWidth: Math.max(width + PADDING * 2, 420), canvasHeight: bottom + PADDING, renderedCount: rendered.size, unlocatedNodes: nodes.filter((node) => isLineage(node) && !rendered.has(node.memberId)) }
})
function memberName(memberId: number) { return nodeById.value.get(memberId)?.displayName || `成员 #${memberId}` }
function parentIds(memberId: number) { return parentEdges.value.filter((edge) => edge.toMemberId === memberId).map((edge) => edge.fromMemberId) }
function childIds(memberId: number) { return parentEdges.value.filter((edge) => edge.fromMemberId === memberId).map((edge) => edge.toMemberId) }
function spouseIds(memberId: number) { return (tree.value?.edges || []).filter((edge) => edge.relationshipType === 'SPOUSE' && (edge.fromMemberId === memberId || edge.toMemberId === memberId)).map((edge) => edge.fromMemberId === memberId ? edge.toMemberId : edge.fromMemberId) }
function relationMembers(memberIds: number[]) { return memberIds.map((memberId) => nodeById.value.get(memberId)).filter((node): node is AdminFamilyTreeNode => Boolean(node)) }
function siblingMembers(memberId: number) {
  const parents = parentIds(memberId)
  const siblingSet = new Set<number>()
  parents.forEach((parentId) => childIds(parentId).forEach((childId) => { if (childId !== memberId) siblingSet.add(childId) }))
  return relationMembers([...siblingSet])
}
function primaryParentByGender(memberId: number, gender: string, options: { lineageOnly?: boolean } = {}) {
  const parentSet = new Set(parentEdges.value.filter((edge) => edge.toMemberId === memberId && (edge.parentLinkType || 'PRIMARY') === 'PRIMARY').map((edge) => edge.fromMemberId))
  return (tree.value?.nodes || []).find((node) => parentSet.has(node.memberId) && node.gender === gender && (!options.lineageOnly || node.memberType === 'LINEAGE_MEMBER')) || null
}
function primaryUnknownGenderParent(memberId: number) {
  const parentSet = new Set(parentEdges.value.filter((edge) => edge.toMemberId === memberId && (edge.parentLinkType || 'PRIMARY') === 'PRIMARY').map((edge) => edge.fromMemberId))
  return (tree.value?.nodes || []).find((node) => parentSet.has(node.memberId) && node.gender !== 'MALE' && node.gender !== 'FEMALE') || null
}
function canParentBeSpouse(addType: ManagedRelativeAddType, base?: AdminFamilyTreeNode | null) {
  if (!base) return false
  if (addType === 'ADD_MOTHER') return Boolean(primaryParentByGender(base.memberId, 'MALE', { lineageOnly: true }))
  if (addType === 'ADD_FATHER') return Boolean(primaryParentByGender(base.memberId, 'FEMALE', { lineageOnly: true }))
  return false
}
function defaultParentMemberType(addType: ManagedRelativeAddType, base?: AdminFamilyTreeNode | null) {
  return canParentBeSpouse(addType, base) ? 'SPOUSE' : 'LINEAGE_MEMBER'
}
function relativeActionDisabledReason(node: AdminFamilyTreeNode, addType: ManagedRelativeAddType) {
  if (node.memberType === 'SPOUSE') return '配偶节点不能作为家谱主干扩展，请从本家成员节点操作'
  if ((addType === 'ADD_FATHER' || addType === 'ADD_MOTHER') && primaryUnknownGenderParent(node.memberId)) return '该成员已有性别未确定的父母节点，请先完善父母性别后再添加父亲或母亲'
  if (addType === 'ADD_FATHER' && primaryParentByGender(node.memberId, 'MALE')) return '该成员已有父亲节点，不能重复添加父亲'
  if (addType === 'ADD_MOTHER' && primaryParentByGender(node.memberId, 'FEMALE')) return '该成员已有母亲节点，不能重复添加母亲'
  if ((addType === 'ADD_CHILD' || addType === 'ADD_SPOUSE') && node.gender !== 'MALE' && node.gender !== 'FEMALE') return '请先完善该成员性别，再添加配偶或子女'
  if (addType === 'ADD_SIBLING' && parentIds(node.memberId).length === 0) return '该成员还没有父亲或母亲节点，请先创建父亲或母亲节点'
  return ''
}
function parentNames(memberId: number) { return parentIds(memberId).map(memberName) }
function childNames(memberId: number) { return childIds(memberId).map(memberName) }
function spouseNames(memberId: number) { return spouseIds(memberId).map(memberName) }
function genderText(gender: string) { return gender === 'MALE' ? '男' : gender === 'FEMALE' ? '女' : '未记录' }
function lifeText(isLiving?: boolean | null) { return isLiving === true ? '健在' : isLiving === false ? '已故' : '状态未录' }
function bindingText(state: string) { return state === 'BOUND' ? '已绑定' : state === 'NOT_REQUIRED' ? '无需绑定' : '未绑定' }
function memberTypeText(type: string) { return ({ LINEAGE_MEMBER: '本家成员', SPOUSE: '外姓配偶', SPOUSE_MEMBER: '配偶成员' } as Record<string, string>)[type] || type || '未标注' }
function hasKeyIdentity(node: AdminFamilyTreeNode) { return /皇帝|皇后|太子|宣祖|[高太文景武昭惠中睿真仁英神徽钦孝光宁]宗/.test(node.identityText || '') }
function showcaseNodeClass(node: AdminFamilyTreeNode) { return node.memberType === 'SPOUSE' ? 'showcase-spouse' : hasKeyIdentity(node) ? 'showcase-key-identity' : 'showcase-lineage' }
function relationshipText(type: string) { return type === 'PARENT_CHILD' ? '父母子女' : type === 'SPOUSE' ? '配偶' : type }
function relativeAddTypeText(type: ManagedRelativeAddType) { return ({ ADD_FATHER: '父亲', ADD_MOTHER: '母亲', ADD_CHILD: '子女', ADD_SPOUSE: '配偶', ADD_SIBLING: '同辈' } as Record<ManagedRelativeAddType, string>)[type] }
function requestedMemberId() { const value = Number(route.query.memberId); return Number.isFinite(value) && value > 0 ? value : null }
function selectMember(memberId: number) {
  selectedMemberId.value = memberId
  if (!isShowcase) router.replace({ query: { ...route.query, memberId: String(memberId) } })
}
function openMemberIndex(memberId = selectedMemberId.value) {
  router.push({ path: `/admin/families/${familyId}/members`, query: memberId ? { memberId: String(memberId) } : undefined })
}
function normalizeDate(value?: string | null) {
  return String(value || '').slice(0, 10)
}
function deleteBlockReason(node: AdminFamilyTreeNode) {
  const childCount = childIds(node.memberId).length
  return childCount > 0 ? `该成员已有 ${childCount} 位子女，不能直接删除` : ''
}
function openAdminMemberEdit(node: AdminFamilyTreeNode) {
  editingAdminMember.value = node
  Object.assign(adminMemberForm, {
    displayName: node.displayName,
    gender: node.gender || 'UNKNOWN',
    birthDate: normalizeDate(node.birthDate),
    isLiving: node.isLiving === true ? 'ALIVE' : node.isLiving === false ? 'DECEASED' : 'UNKNOWN',
    userBindingPolicy: node.userBindingPolicy || (node.userBindingState === 'NOT_REQUIRED' ? 'NOT_REQUIRED' : 'OPTIONAL'),
  })
  adminMemberDialogVisible.value = true
}
async function submitAdminMember() {
  if (!editingAdminMember.value) return
  const displayName = adminMemberForm.displayName.trim()
  if (!displayName) {
    error.value = '请填写成员姓名'
    return
  }
  savingAdminMember.value = true
  error.value = ''
  try {
    await updateFamilyMember(familyId, editingAdminMember.value.memberId, {
      displayName,
      gender: adminMemberForm.gender,
      birthDate: adminMemberForm.birthDate || null,
      isLiving: adminMemberForm.isLiving === 'UNKNOWN' ? null : adminMemberForm.isLiving === 'ALIVE',
      userBindingPolicy: adminMemberForm.userBindingPolicy,
    })
    adminMemberDialogVisible.value = false
    const memberId = editingAdminMember.value.memberId
    await loadTree()
    if (nodeById.value.has(memberId)) selectMember(memberId)
    ElMessage.success('成员资料已更新')
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  } finally {
    savingAdminMember.value = false
  }
}
async function removeAdminMember(node: AdminFamilyTreeNode) {
  const reasonText = deleteBlockReason(node)
  if (reasonText) {
    error.value = reasonText
    return
  }
  let reason = ''
  try {
    const result = await ElMessageBox.prompt(`确认删除「${node.displayName}」？删除后会同步解除该成员相关关系。`, '删除成员', {
      confirmButtonText: '确认删除',
      cancelButtonText: '取消',
      inputPlaceholder: '可填写删除原因',
      type: 'warning',
    })
    reason = result.value || ''
  } catch {
    return
  }
  deletingMember.value = true
  error.value = ''
  try {
    await deleteFamilyMember(familyId, node.memberId, reason)
    await loadTree()
    ElMessage.success('成员已删除')
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  } finally {
    deletingMember.value = false
  }
}
function defaultRelativeGender(addType: ManagedRelativeAddType, base?: AdminFamilyTreeNode | null) {
  if (addType === 'ADD_FATHER') return 'MALE'
  if (addType === 'ADD_MOTHER') return 'FEMALE'
  if (addType === 'ADD_SPOUSE') return base?.gender === 'MALE' ? 'FEMALE' : base?.gender === 'FEMALE' ? 'MALE' : 'UNKNOWN'
  return 'UNKNOWN'
}
function openAdminRelativeCreate(addType: ManagedRelativeAddType) {
  if (!selectedNode.value) return
  const disabledReason = relativeActionDisabledReason(selectedNode.value, addType)
  if (disabledReason) {
    error.value = disabledReason
    return
  }
  adminRelativeBase.value = selectedNode.value
  Object.assign(adminRelativeForm, {
    addType,
    displayName: '',
    gender: defaultRelativeGender(addType, selectedNode.value),
    birthDate: '',
    isLiving: 'ALIVE',
    userBindingPolicy: 'OPTIONAL',
    memberType: addType === 'ADD_SPOUSE' ? 'SPOUSE' : defaultParentMemberType(addType, selectedNode.value),
    parentLinkType: 'PRIMARY',
  })
  adminRelativeDialogVisible.value = true
}
async function submitAdminRelative() {
  const base = adminRelativeBase.value
  if (!base) return
  const displayName = adminRelativeForm.displayName.trim()
  if (!displayName) {
    error.value = '请填写成员姓名'
    return
  }
  if (adminRelativeGenderRequired.value && adminRelativeForm.gender === 'UNKNOWN') {
    error.value = '子女或同辈必须选择男或女'
    return
  }
  savingAdminRelative.value = true
  error.value = ''
  try {
    const result = await createFamilyRelative(familyId, base.memberId, {
      addType: adminRelativeForm.addType,
      newMember: {
        displayName,
        gender: adminRelativeForm.gender,
        birthDate: adminRelativeForm.birthDate || null,
        isLiving: adminRelativeForm.isLiving === 'UNKNOWN' ? null : adminRelativeForm.isLiving === 'ALIVE',
        userBindingPolicy: adminRelativeForm.userBindingPolicy,
        memberType: adminRelativeForm.memberType,
      },
      relationship: adminRelativeForm.addType === 'ADD_SPOUSE' ? undefined : { parentLinkType: adminRelativeForm.parentLinkType },
    })
    adminRelativeDialogVisible.value = false
    await loadTree()
    if (result.memberId && nodeById.value.has(result.memberId)) selectMember(result.memberId)
    ElMessage.success('亲属已添加')
  } catch (requestError) {
    error.value = getApiErrorMessage(requestError)
  } finally {
    savingAdminRelative.value = false
  }
}
function openMemberCreate() { editingMember.value = null; Object.assign(memberForm, { displayName: '', identityText: '', gender: 'UNKNOWN', birthYear: null, sortOrder: (tree.value?.nodes.length || 0) * 10 }); memberDialogVisible.value = true }
function openRelativeCreate(memberId: number) { const anchor = nodeById.value.get(memberId); if (!anchor) return; relativeAnchor.value = anchor; const order = (tree.value?.nodes.length || 0) * 10; Object.assign(relativeForm, { relationKind: 'CHILD', spouseMode: anchorSpouses.value.length ? 'EXISTING' : 'NONE', existingSpouseMemberId: anchorSpouses.value[0]?.memberId || 0 }); Object.assign(relativeForm.member, { displayName: '', identityText: '', gender: 'UNKNOWN', birthYear: null, sortOrder: order }); Object.assign(relativeForm.newSpouse, { displayName: '', identityText: '', gender: 'UNKNOWN', birthYear: null, sortOrder: order + 1 }); relativeDialogVisible.value = true }
function openMemberEdit(node: AdminFamilyTreeNode) { editingMember.value = node; Object.assign(memberForm, { displayName: node.displayName, identityText: node.identityText || '', gender: node.gender as ShowcaseFamilyMemberInput['gender'], birthYear: node.birthDate ? Number(node.birthDate.slice(0, 4)) : null, sortOrder: node.sortOrder || 0 }); memberDialogVisible.value = true }
function openRelationshipCreate() { Object.assign(relationshipForm, { fromMemberId: 0, toMemberId: 0, relationshipType: 'PARENT_CHILD' }); relationshipDialogVisible.value = true }
async function submitMember() { if (!memberForm.displayName.trim()) { error.value = '请填写成员姓名'; return } savingTree.value = true; error.value = ''; try { if (editingMember.value) await updateShowcaseFamilyMember(Number(familyId), editingMember.value.memberId, memberForm); else await createShowcaseFamilyMember(Number(familyId), memberForm); memberDialogVisible.value = false; await loadTree() } catch (requestError) { error.value = getApiErrorMessage(requestError) } finally { savingTree.value = false } }
async function submitRelative() { if (!relativeAnchor.value || !relativeForm.member.displayName.trim()) { error.value = '请填写成员姓名'; return } if (relativeForm.relationKind === 'CHILD' && relativeForm.spouseMode === 'EXISTING' && !relativeForm.existingSpouseMemberId) { error.value = '请选择已有配偶'; return } if (relativeForm.relationKind === 'CHILD' && relativeForm.spouseMode === 'NEW' && !relativeForm.newSpouse.displayName.trim()) { error.value = '请填写配偶姓名'; return } savingTree.value = true; error.value = ''; try { await addShowcaseRelative(Number(familyId), relativeAnchor.value.memberId, { relationKind: relativeForm.relationKind, member: relativeForm.member, existingSpouseMemberId: relativeForm.relationKind === 'CHILD' && relativeForm.spouseMode === 'EXISTING' ? relativeForm.existingSpouseMemberId : undefined, newSpouse: relativeForm.relationKind === 'CHILD' && relativeForm.spouseMode === 'NEW' ? relativeForm.newSpouse : undefined }); relativeDialogVisible.value = false; relativeAnchor.value = null; await loadTree() } catch (requestError) { error.value = getApiErrorMessage(requestError) } finally { savingTree.value = false } }
async function submitRelationship() { if (!relationshipForm.fromMemberId || !relationshipForm.toMemberId || relationshipForm.fromMemberId === relationshipForm.toMemberId) { error.value = '请选择两个不同成员'; return } savingTree.value = true; error.value = ''; try { await createShowcaseFamilyRelationship(Number(familyId), relationshipForm); relationshipDialogVisible.value = false; await loadTree() } catch (requestError) { error.value = getApiErrorMessage(requestError) } finally { savingTree.value = false } }
function deletionBlockReason(memberId: number) { const descendants = (tree.value?.edges || []).filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.fromMemberId === memberId).length; return descendants > 0 ? `该成员有 ${descendants} 位子女，不能删除；如资料有误请直接编辑` : '' }
async function removeMember(memberId: number) { const reason = deletionBlockReason(memberId); if (reason) { error.value = reason; return } const relatedEdges = (tree.value?.edges || []).filter((edge) => edge.fromMemberId === memberId || edge.toMemberId === memberId); const parentLinks = relatedEdges.filter((edge) => edge.relationshipType === 'PARENT_CHILD' && edge.toMemberId === memberId).length; const spouseLinks = relatedEdges.filter((edge) => edge.relationshipType === 'SPOUSE').length; const cleanupText = [`${parentLinks} 条上级关系`, `${spouseLinks} 条配偶关系`].filter((_, index) => index === 0 ? parentLinks > 0 : spouseLinks > 0).join('、'); try { await ElMessageBox.confirm(cleanupText ? `删除后将同步解除 ${cleanupText}，确认继续？` : '确认删除这个末端节点？', '确认删除', { type: 'warning' }) } catch { return } savingTree.value = true; error.value = ''; try { await deleteShowcaseFamilyMember(Number(familyId), memberId); await loadTree() } catch (requestError) { error.value = getApiErrorMessage(requestError) } finally { savingTree.value = false } }
async function loadTree() { error.value = ''; try { if (isShowcase) { const result = await getShowcaseFamilyTree(Number(familyId)); tree.value = { familyId: result.showcaseFamilyId, graphVersion: 0, nodes: result.nodes.map((node) => ({ memberId: node.memberId, displayName: node.displayName, identityText: node.identityText, gender: node.gender, memberType: node.memberType, birthDate: node.birthYear ? `${node.birthYear}-01-01` : null, isLiving: null, userBindingState: '', sortOrder: node.sortOrder })), edges: result.edges.map((edge) => ({ relationshipId: edge.relationshipId, fromMemberId: edge.fromMemberId, toMemberId: edge.toMemberId, relationshipType: edge.relationshipType })) } } else { tree.value = await getFamilyTree(familyId) }; const preferred = requestedMemberId(); selectedMemberId.value = preferred && nodeById.value.has(preferred) ? preferred : tree.value.nodes[0]?.memberId || null } catch (requestError) { tree.value = null; error.value = getApiErrorMessage(requestError) } }
onMounted(loadTree)
watch(() => route.query.memberId, () => {
  const memberId = requestedMemberId()
  if (memberId && nodeById.value.has(memberId)) selectedMemberId.value = memberId
})
</script>

<style scoped>
.node-info-header{display:flex;align-items:baseline;gap:10px}.node-info-header span{color:var(--color-text-secondary);font-size:12px}
.summary-card { display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:24px }.tree-card :deep(.el-card__body){padding:0}.card-header{display:flex;align-items:center;justify-content:space-between;gap:16px}.card-header span{display:block;margin-top:4px;color:var(--color-text-secondary);font-size:13px}.tree-scroll{height:min(720px,calc(100vh - 300px));min-height:360px}.tree-scroll :deep(.el-scrollbar__wrap){overflow:auto}.tree-canvas{position:relative;min-width:100%;background-color:rgba(255,249,236,.36);background-image:linear-gradient(rgba(92,74,48,.05) 1px,transparent 1px);background-size:100% 28px}.links-layer{position:absolute;inset:0;pointer-events:none;z-index:1}.link-segment{position:absolute;display:flex;width:6px;align-items:center;justify-content:center}.spouse-join i{display:block;width:4px;border-top:2px dashed rgba(92,74,48,.55)}.member-card{position:relative;width:102px;height:66px;display:grid;flex:0 0 102px;align-content:center;gap:2px;padding:5px;color:var(--color-text-primary);text-align:center;border:1px solid var(--color-border);border-radius:8px;background:var(--color-bg-card);cursor:pointer}.member-card:hover,.member-card.selected{border-color:var(--color-primary);box-shadow:var(--shadow-light)}.member-card.deceased{background:#fafafa;opacity:.82}.member-card strong{overflow:hidden;font-size:13px;text-overflow:ellipsis;white-space:nowrap}.member-card small{overflow:hidden;color:var(--color-text-secondary);font-size:10px;text-overflow:ellipsis;white-space:nowrap}.member-card .binding{color:var(--color-heritage-green)}.showcase-lineage{border-color:#b99c68;background:#fffdf7}.showcase-spouse{border-style:dashed;border-color:#c9a88a;background:#fff9f2}.showcase-key-identity{border:2px solid #b6473e;background:#fffaf2;box-shadow:inset 0 3px 0 rgba(182,71,62,.14)}.identity-seal{position:absolute;right:4px;top:4px;display:grid;width:15px;height:15px;place-items:center;border:1px solid #b6473e;border-radius:50%;color:#b6473e;font-size:9px;font-weight:700;line-height:1}.spouse-badge{position:absolute;right:3px;bottom:3px;padding:0 3px;border-radius:3px;background:#f4e8cf;color:#8b6b35;font-size:9px;line-height:12px}.node-action{position:absolute;top:3px;display:grid;width:15px;height:15px;place-items:center;border:1px solid rgba(35,70,108,.45);border-radius:50%;background:#fff;color:var(--color-primary);font-size:12px;font-weight:700;line-height:1}.node-action:hover{background:var(--color-primary);color:#fff}.node-add{right:3px}.node-remove{left:3px;border-color:rgba(190,70,62,.48);color:#b6473e}.node-remove:hover{background:#b6473e;color:#fff}.tree-footnote{margin:0;padding:12px 20px;color:var(--color-text-secondary);font-size:13px}.unlocated-section{padding:18px 20px 20px;border-top:1px dashed var(--color-border)}.unlocated-section>span{display:block;margin-top:5px;color:var(--color-text-secondary);font-size:13px}.unlocated-list{display:flex;flex-wrap:wrap;gap:10px;margin-top:14px}.unlocated-card{height:66px}@media (max-width:760px){.summary-card{grid-template-columns:1fr}.tree-scroll{height:520px;min-height:320px}}
.link-segment {
  display: block;
  background: rgba(92, 74, 48, .55);
  border-radius: 2px;
}

.couple-unit {
  position: absolute;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 6px;
}

.spouse-join {
  display: flex;
  width: 6px;
  align-items: center;
  justify-content: center;
}

.tree-workspace {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(340px, 390px);
  align-items: start;
  gap: 16px;
}

.tree-workspace--showcase {
  grid-template-columns: minmax(0, 1fr);
}

.member-inspector {
  min-width: 0;
}

.inspector-card {
  position: sticky;
  top: 16px;
}

.member-inspector .node-info-header {
  align-items: flex-start;
  flex-direction: column;
}

.member-inspector .node-info-header > div:first-child {
  align-items: flex-start;
  flex-direction: column;
  gap: 4px;
}

.member-inspector .node-actions {
  justify-content: flex-start;
}

.member-inspector .relation-grid {
  grid-template-columns: 1fr;
}

.member-inspector .relation-block {
  min-height: auto;
}

.member-inspector .operation-note {
  font-size: 12px;
}

.node-info-header {
  align-items: center;
  justify-content: space-between;
}

.node-info-header > div:first-child {
  display: flex;
  min-width: 0;
  align-items: baseline;
  gap: 10px;
}

.node-info-header strong {
  font-size: 18px;
}

.node-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.relative-action-panel {
  margin-bottom: 14px;
  padding: 14px;
  border: 1px solid rgba(35, 70, 108, .16);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(247, 250, 252, .96), rgba(255, 252, 244, .82));
}

.relative-action-title {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}

.relative-action-title strong {
  color: var(--color-text-primary);
  font-size: 15px;
}

.relative-action-title span {
  color: var(--color-text-secondary);
  font-size: 12px;
}

.relative-action-buttons {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.relative-action-buttons .el-button {
  width: 100%;
  margin-left: 0;
}

.relation-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.relation-block {
  min-height: 92px;
  padding: 14px;
  text-align: left;
  border: 1px solid rgba(35, 70, 108, .14);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(255, 252, 244, .95), rgba(250, 246, 236, .88));
  transition: border-color .2s ease, box-shadow .2s ease, transform .2s ease;
}

.relation-block:not(.empty):hover {
  border-color: rgba(35, 70, 108, .36);
  box-shadow: 0 8px 20px rgba(35, 70, 108, .08);
}

.member-card:active {
  transform: translateY(1px);
}

.relation-block.empty {
  opacity: .78;
}

.relation-block.empty:hover {
  border-color: rgba(35, 70, 108, .14);
  box-shadow: none;
}

.relation-block span {
  display: block;
  color: #b6473e;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: .08em;
}

.relation-block strong {
  display: block;
  margin-top: 8px;
  overflow: hidden;
  color: var(--color-text-primary);
  font-size: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.relation-members {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 9px;
}

.relation-chip {
  max-width: 100%;
  padding: 4px 8px;
  overflow: hidden;
  color: var(--color-primary);
  font-size: 13px;
  font-weight: 600;
  text-align: left;
  text-overflow: ellipsis;
  white-space: nowrap;
  border: 1px solid rgba(35, 70, 108, .22);
  border-radius: 6px;
  background: rgba(255, 255, 255, .72);
  cursor: pointer;
  transition: border-color .2s ease, background .2s ease, transform .2s ease;
}

.relation-chip:hover {
  border-color: rgba(35, 70, 108, .44);
  background: #fff;
}

.relation-chip:active {
  transform: translateY(1px);
}

.relation-block small {
  display: block;
  margin-top: 8px;
  color: var(--color-text-secondary);
  font-size: 12px;
  line-height: 1.45;
}

.operation-note {
  margin: 14px 0 0;
  padding: 10px 12px;
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.6;
  border-left: 3px solid rgba(182, 71, 62, .72);
  background: rgba(182, 71, 62, .06);
}

.member-card {
  transition: border-color .2s ease, box-shadow .2s ease, transform .2s ease;
}

@media (max-width: 960px) {
  .tree-workspace {
    grid-template-columns: 1fr;
  }

  .inspector-card {
    position: static;
  }

  .node-info-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .node-actions {
    justify-content: flex-start;
  }

  .relation-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .relative-action-buttons {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .relation-grid {
    grid-template-columns: 1fr;
  }
}
</style>
