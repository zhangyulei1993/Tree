<template>
  <PageShell>
    <section class="container page-section">
      <div class="page-heading">
        <div>
          <span class="eyebrow">家庭核心流程</span>
          <h1>创建家庭</h1>
          <p class="muted">创建成功后，当前账号会成为家庭创始人。</p>
        </div>
        <RouterLink class="button secondary" to="/me/families">返回我的家庭</RouterLink>
      </div>

      <form class="card form-card" @submit.prevent="submit">
        <label>
          <span>姓氏 <strong>*</strong></span>
          <input v-model.trim="form.surname" class="field" maxlength="50" placeholder="例如：张" />
        </label>
        <div class="form-grid">
          <label>
            <span>本人姓名 <strong>*</strong></span>
            <input v-model.trim="form.founderName" class="field" maxlength="80" placeholder="用于家庭树中的本人节点" />
          </label>
          <label>
            <span>性别 <strong>*</strong></span>
            <select v-model="form.founderGender" class="field">
              <option value="MALE">男</option>
              <option value="FEMALE">女</option>
            </select>
          </label>
        </div>
        <label>
          <span>出生年份 <strong>*</strong></span>
          <input v-model.number="form.founderBirthYear" class="field" type="number" min="1" :max="currentYear" placeholder="例如：1990" />
        </label>
        <label>
          <span>家庭名称</span>
          <input v-model.trim="form.familyName" class="field" maxlength="100" placeholder="例如：张氏家族" />
        </label>
        <div class="form-grid">
          <label>
            <span>籍贯</span>
            <input v-model.trim="form.nativePlace" class="field" maxlength="100" placeholder="例如：山东济南" />
          </label>
          <label>
            <span>地区</span>
            <input v-model.trim="form.regionText" class="field" maxlength="100" placeholder="例如：山东省济南市" />
          </label>
        </div>
        <label>
          <span>家庭简介</span>
          <textarea v-model.trim="form.description" class="field textarea" maxlength="500" placeholder="简要说明家庭来源或记录范围" />
        </label>
        <p v-if="error" class="feedback error" role="alert">{{ error }}</p>
        <div class="actions">
          <button class="button" :disabled="submitting">
            {{ submitting ? '创建中...' : '创建家庭' }}
          </button>
        </div>
      </form>
    </section>
  </PageShell>
</template>

<script setup lang="ts">
import axios from 'axios'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import { apiErrorMessage } from '@/api/client'
import { createFamily } from '@/api/families'
import PageShell from '@/components/PageShell.vue'
import type { ApiResponse, CreateFamilyInput, Gender } from '@/types/api'

const router = useRouter()
const submitting = ref(false)
const error = ref('')
const currentYear = new Date().getFullYear()
const form = reactive({
  surname: '',
  founderName: '',
  founderGender: 'MALE' as Gender,
  founderBirthYear: undefined as number | undefined,
  familyName: '',
  nativePlace: '',
  regionText: '',
  description: ''
})

function errorText(requestError: unknown) {
  if (axios.isAxiosError<ApiResponse<unknown>>(requestError) && requestError.response?.data?.code) {
    return `${requestError.response.data.code}：${requestError.response.data.message}`
  }
  return apiErrorMessage(requestError, '家庭创建失败。')
}

async function submit() {
  error.value = ''
  if (!form.surname || !form.founderName) {
    error.value = '请填写姓氏和本人姓名。'
    return
  }
  if (!Number.isInteger(form.founderBirthYear) || !form.founderBirthYear || form.founderBirthYear < 1 || form.founderBirthYear > currentYear) {
    error.value = `请输入 1 到 ${currentYear} 之间的出生年份。`
    return
  }

  const input: CreateFamilyInput = {
    surname: form.surname,
    founderName: form.founderName,
    founderGender: form.founderGender,
    founderBirthYear: form.founderBirthYear,
    familyName: form.familyName || undefined,
    nativePlace: form.nativePlace || undefined,
    regionText: form.regionText || undefined,
    description: form.description || undefined
  }

  submitting.value = true
  try {
    const family = await createFamily(input)
    await router.push(`/families/${family.id}`)
  } catch (requestError) {
    error.value = errorText(requestError)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page-section {
  padding: 32px 0;
}

.page-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

h1 {
  margin: 8px 0;
}

.eyebrow {
  color: var(--color-heritage-green);
  font-size: 13px;
  font-weight: 700;
}

.form-card {
  display: grid;
  gap: 18px;
  max-width: 760px;
  padding: 24px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

label {
  display: grid;
  gap: 7px;
  font-size: 14px;
  font-weight: 600;
}

label strong,
.error {
  color: var(--color-danger);
}

.feedback {
  margin: 0;
  font-weight: 600;
}

.actions {
  display: flex;
  justify-content: flex-end;
}

.button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

@media (max-width: 640px) {
  .page-heading,
  .actions {
    align-items: stretch;
    flex-direction: column;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
