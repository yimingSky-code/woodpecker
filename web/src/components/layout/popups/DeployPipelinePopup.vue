<template>
  <OpsUiProvider>
    <NModal :show="open" @update:show="!$event && $emit('close')">
      <NCard
        role="dialog"
        aria-modal="true"
        :aria-labelledby="`${formId}-title`"
        :bordered="false"
        content-style="padding: 0"
        header-style="padding: 24px var(--ops-dialog-gutter) 20px"
        footer-style="padding: 16px var(--ops-dialog-gutter); border-top: 1px solid var(--wp-control-neutral-200)"
        class="ops-dialog"
      >
        <template #header>
          <div class="ops-heading">
            <div>
              <div class="ops-eyebrow">{{ $t('repo.parameters.deployment_eyebrow') }}</div>
              <h2 :id="`${formId}-title`" class="ops-title">
                {{ $t('repo.parameters.deploy_heading', { number: pipelineNumber }) }}
              </h2>
            </div>
            <NTag :bordered="false" type="success" round>
              {{ $t('repo.parameters.source_number', { number: pipelineNumber }) }}
            </NTag>
          </div>
        </template>
        <template #header-extra>
          <NButton quaternary circle size="small" :aria-label="$t('repo.parameters.close')" @click="$emit('close')">
            <template #icon><Icon name="close" /></template>
          </NButton>
        </template>
        <div v-if="loading" class="flex justify-center py-12"><NSpin size="large" /></div>
        <NAlert v-else-if="loadError" role="alert" type="error">{{ $t('repo.parameters.load_failed') }}</NAlert>
        <NForm
          v-else
          :id="formId"
          label-placement="top"
          class="ops-dialog-content"
          @submit.prevent="triggerDeployPipeline"
        >
          <p class="ops-description">{{ $t('repo.parameters.deploy_description') }}</p>
          <PipelineParameterFields
            v-model="parameterValues"
            :parameters="parameters"
            @update:is-valid="parametersValid = $event"
          />
          <NFormItem
            v-if="!parameters.some((p) => p.name === 'deploy_to')"
            :label="$t('repo.deploy_pipeline.enter_target')"
            required
          >
            <NInput
              v-model:value="payload.environment"
              :input-props="{ 'aria-label': $t('repo.deploy_pipeline.enter_target') }"
            />
          </NFormItem>
          <NFormItem
            v-if="!parameters.some((p) => p.name === 'deploy_task')"
            :label="$t('repo.deploy_pipeline.enter_task')"
          >
            <NInput
              v-model:value="payload.task"
              :input-props="{ 'aria-label': $t('repo.deploy_pipeline.enter_task') }"
            />
          </NFormItem>
          <NCollapse class="ops-section">
            <NCollapseItem
              name="variables"
              display-directive="show"
              :title="$t('repo.deploy_pipeline.variables.title')"
            >
              <p class="ops-description">{{ $t('repo.deploy_pipeline.variables.desc') }}</p>
              <OpsVariableEditor
                v-model="payload.variables"
                :key-placeholder="$t('repo.deploy_pipeline.variables.name')"
                :value-placeholder="$t('repo.deploy_pipeline.variables.value')"
                :delete-title="$t('repo.deploy_pipeline.variables.delete')"
                @update:is-valid="isVariablesValid = $event"
              />
            </NCollapseItem>
          </NCollapse>
        </NForm>
        <template #footer>
          <div class="ops-actions">
            <div v-if="sourceCommit" class="ops-source">
              {{ $t('repo.parameters.commit') }} <code>{{ sourceCommit.slice(0, 10) }}</code>
            </div>
            <NButton @click="$emit('close')">{{ $t('cancel') }}</NButton>
            <NButton
              type="primary"
              attr-type="submit"
              :form="formId"
              :loading="submitting"
              :disabled="loading || loadError || !isFormValid"
            >
              {{ $t('repo.deploy_pipeline.trigger') }}
            </NButton>
          </div>
        </template>
      </NCard>
    </NModal>
  </OpsUiProvider>
</template>

<script lang="ts" setup>
import {
  NAlert,
  NButton,
  NCard,
  NCollapse,
  NCollapseItem,
  NForm,
  NFormItem,
  NInput,
  NModal,
  NSpin,
  NTag,
} from 'naive-ui';
import { computed, ref, toRef, useId, watch } from 'vue';
import { useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import PipelineParameterFields from '~/components/form/PipelineParameterFields.vue';
import OpsUiProvider from '~/components/ops/OpsUiProvider.vue';
import OpsVariableEditor from '~/components/ops/OpsVariableEditor.vue';
import { usePaginate } from '~/compositions/usePaginate';
import type { Parameter } from '~/lib/api/types';
import { initialParameterValues, parametersForEvent } from '~/lib/pipelineParameters';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';

const props = defineProps<{
  open: boolean;
  pipelineNumber: string;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const formId = `deploy-${useId()}`;
const sourceCommit = ref('');
const apiClient = useApiClient();
const repo = requiredInject('repo');
const router = useRouter();

const payload = ref<{
  id: string;
  environment: string;
  task: string;
  variables: Record<string, string>;
}>({
  id: '',
  environment: '',
  task: '',
  variables: {},
});

const parameters = ref<Parameter[]>([]);
const parameterValues = ref<Record<string, string>>({});
const parametersValid = ref(true);
const isVariablesValid = ref(true);
const submitting = ref(false);
const loadError = ref(false);

const pipelineOptions = computed(() => ({
  ...payload.value,
  environment: parameterValues.value.deploy_to ?? payload.value.environment,
  task: parameterValues.value.deploy_task ?? payload.value.task,
  variables: Object.fromEntries(
    Object.entries({ ...payload.value.variables, ...parameterValues.value }).filter(
      ([key]) => key !== 'deploy_to' && key !== 'deploy_task',
    ),
  ),
}));

const isFormValid = computed(() => {
  return (
    pipelineOptions.value.environment !== '' && isVariablesValid.value && parametersValid.value && !submitting.value
  );
});

const loading = ref(true);
let loadGeneration = 0;
watch(
  () => [props.open, props.pipelineNumber] as const,
  async ([open, number]) => {
    const generation = ++loadGeneration;
    if (!open) return;
    loading.value = true;
    loadError.value = false;
    try {
      const [definitions, source] = await Promise.all([
        usePaginate(async (page) => (await apiClient.getParameterList(repo.value.id, { page })) ?? []),
        apiClient.getPipeline(repo.value.id, Number(number)),
      ]);
      if (generation !== loadGeneration) return;
      sourceCommit.value = source.commit;
      parameters.value = parametersForEvent(definitions, 'deployment');
      parameterValues.value = initialParameterValues(parameters.value, {
        ...source.variables,
        ...(source.deploy_to ? { deploy_to: source.deploy_to } : {}),
        ...(source.deploy_task ? { deploy_task: source.deploy_task } : {}),
      });
      payload.value = { id: '', environment: source.deploy_to || '', task: source.deploy_task || '', variables: {} };
      isVariablesValid.value = true;
    } catch {
      if (generation === loadGeneration) loadError.value = true;
    } finally {
      if (generation === loadGeneration) loading.value = false;
    }
  },
  { immediate: true },
);

const pipelineNumber = toRef(props, 'pipelineNumber');
async function triggerDeployPipeline() {
  if (!isFormValid.value) return;
  submitting.value = true;
  try {
    const newPipeline = await apiClient.deployPipeline(repo.value.id, pipelineNumber.value, pipelineOptions.value);

    emit('close');

    await router.push({
      name: 'repo-pipeline',
      params: {
        pipelineId: newPipeline.number,
      },
    });
  } finally {
    submitting.value = false;
  }
}
</script>
