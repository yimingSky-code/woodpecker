<template>
  <OpsUiProvider>
    <NCard v-if="!loading && !loadError" :bordered="false" size="large" class="ops-run-card">
      <div class="ops-eyebrow">{{ $t('repo.parameters.manual_eyebrow') }}</div>
      <h1 class="ops-title">{{ $t('repo.manual_pipeline.title') }}</h1>
      <p class="ops-description">{{ $t('repo.parameters.manual_description') }}</p>
      <NForm label-placement="top" @submit.prevent="triggerManualPipeline">
        <div class="ops-form-grid">
          <NFormItem :label="$t('repo.manual_pipeline.select_branch')" required>
            <NSelect
              v-model:value="payload.branch"
              :options="branches.map((branch) => ({ label: branch.text, value: branch.value }))"
              :input-props="{ 'aria-label': $t('repo.manual_pipeline.select_branch') }"
              filterable
            />
          </NFormItem>
          <NFormItem :label="$t('repo.manual_pipeline.message.title')">
            <NInput
              v-model:value="payload.message"
              :placeholder="$t('repo.parameters.message_placeholder')"
              :input-props="{ 'aria-label': $t('repo.manual_pipeline.message.title') }"
            />
          </NFormItem>
        </div>
        <div v-if="parameters.length" class="ops-section">
          <PipelineParameterFields
            v-model="parameterValues"
            :parameters="parameters"
            @update:is-valid="parametersValid = $event"
          />
        </div>
        <NCollapse class="ops-section">
          <NCollapseItem name="variables" display-directive="show" :title="$t('repo.manual_pipeline.variables.title')">
            <p class="ops-description">{{ $t('repo.manual_pipeline.variables.desc') }}</p>
            <OpsVariableEditor
              v-model="payload.variables"
              :key-placeholder="$t('repo.manual_pipeline.variables.name')"
              :value-placeholder="$t('repo.manual_pipeline.variables.value')"
              :delete-title="$t('repo.manual_pipeline.variables.delete')"
              @update:is-valid="isVariablesValid = $event"
            />
          </NCollapseItem>
        </NCollapse>
        <div class="ops-actions ops-section">
          <NButton attr-type="submit" type="primary" size="large" :loading="submitting" :disabled="!isFormValid">
            {{ $t('repo.manual_pipeline.trigger') }}
          </NButton>
        </div>
      </NForm>
    </NCard>
    <NAlert v-else-if="loadError" role="alert" type="error">{{ $t('repo.parameters.load_failed') }}</NAlert>
    <div v-else class="flex justify-center py-12"><NSpin size="large" /></div>
  </OpsUiProvider>
</template>

<script lang="ts" setup>
import { useNotification } from '@kyvg/vue3-notification';
import { NAlert, NButton, NCard, NCollapse, NCollapseItem, NForm, NFormItem, NInput, NSelect, NSpin } from 'naive-ui';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import PipelineParameterFields from '~/components/form/PipelineParameterFields.vue';
import OpsUiProvider from '~/components/ops/OpsUiProvider.vue';
import OpsVariableEditor from '~/components/ops/OpsVariableEditor.vue';
import type { Parameter } from '~/lib/api/types';
import { initialParameterValues, parametersForEvent } from '~/lib/pipelineParameters';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const notifications = useNotification();
const i18n = useI18n();

const repo = requiredInject('repo');
const repoPermissions = requiredInject('repo-permissions');

const router = useRouter();
const branches = ref<{ text: string; value: string }[]>([]);
const payload = ref<{ message: string; branch: string; variables: Record<string, string> }>({
  message: '',
  branch: 'main',
  variables: {},
});

const parameters = ref<Parameter[]>([]);
const parameterValues = ref<Record<string, string>>({});
const parametersValid = ref(true);
const isVariablesValid = ref(true);
const submitting = ref(false);
const loadError = ref(false);

const isFormValid = computed(() => {
  return payload.value.branch !== '' && isVariablesValid.value && parametersValid.value && !submitting.value;
});

const pipelineOptions = computed(() => ({
  ...payload.value,
  variables: { ...payload.value.variables, ...parameterValues.value },
}));

const loading = ref(true);
onMounted(async () => {
  if (!repoPermissions.value.push) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
    return;
  }

  try {
    const [data, definitions] = await Promise.all([
      usePaginate((page) => apiClient.getRepoBranches(repo.value.id, { page })),
      usePaginate(async (page) => (await apiClient.getParameterList(repo.value.id, { page })) ?? []),
    ]);
    branches.value = data.map((e) => ({ text: e, value: e }));
    if (!data.includes(payload.value.branch)) payload.value.branch = repo.value.default_branch || data[0] || '';
    parameters.value = parametersForEvent(definitions, 'manual');
    parameterValues.value = initialParameterValues(parameters.value);
  } catch {
    loadError.value = true;
  } finally {
    loading.value = false;
  }
});

async function triggerManualPipeline() {
  if (!isFormValid.value) return;
  submitting.value = true;
  try {
    const pipeline = await apiClient.createPipeline(repo.value.id, pipelineOptions.value);

    emit('close');

    if (typeof pipeline == 'string') {
      // if this is a string (http 204) there is no workflow to run with the 'manual' event

      await router.push({
        name: 'repo',
      });

      notifications.notify({ type: 'warn', title: i18n.t('repo.manual_pipeline.no_manual_workflows') });
    } else {
      await router.push({
        name: 'repo-pipeline',
        params: {
          pipelineId: pipeline.number,
        },
      });
    }
  } finally {
    submitting.value = false;
  }
}

useWPTitle(computed(() => [i18n.t('repo.manual_pipeline.trigger'), repo.value.full_name]));
</script>
