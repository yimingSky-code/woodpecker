<template>
  <Popup :open="open" @close="$emit('close')">
    <Panel
      v-if="!loading && !loadError"
      class="bg-wp-background-100 dark:bg-wp-background-300 max-h-[calc(100vh-2rem)] overflow-y-auto"
    >
      <form @submit.prevent="triggerDeployPipeline">
        <span class="text-wp-text-100 text-xl">{{
          $t('repo.deploy_pipeline.title', { pipelineId: pipelineNumber })
        }}</span>
        <PipelineParameterFields
          v-model="parameterValues"
          :parameters="parameters"
          @update:is-valid="parametersValid = $event"
        />
        <InputField
          v-if="!parameters.some((p) => p.name === 'deploy_to')"
          v-slot="{ id }"
          :label="$t('repo.deploy_pipeline.enter_target')"
        >
          <TextField :id="id" v-model="payload.environment" required />
        </InputField>
        <InputField
          v-if="!parameters.some((p) => p.name === 'deploy_task')"
          v-slot="{ id }"
          :label="$t('repo.deploy_pipeline.enter_task')"
        >
          <TextField :id="id" v-model="payload.task" />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('repo.deploy_pipeline.variables.title')">
          <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.deploy_pipeline.variables.desc') }}</span>
          <KeyValueEditor
            :id="id"
            v-model="payload.variables"
            :key-placeholder="$t('repo.deploy_pipeline.variables.name')"
            :value-placeholder="$t('repo.deploy_pipeline.variables.value')"
            :delete-title="$t('repo.deploy_pipeline.variables.delete')"
            @update:is-valid="isVariablesValid = $event"
          />
        </InputField>
        <Button type="submit" :text="$t('repo.deploy_pipeline.trigger')" :disabled="!isFormValid" />
      </form>
    </Panel>
    <p v-if="loadError" role="alert">{{ $t('repo.parameters.load_failed') }}</p>
  </Popup>
</template>

<script lang="ts" setup>
import { computed, ref, toRef, watch } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineParameterFields from '~/components/form/PipelineParameterFields.vue';
import { usePaginate } from '~/compositions/usePaginate';
import type { Parameter } from '~/lib/api/types';
import { initialParameterValues, parametersForEvent } from '~/lib/pipelineParameters';
import Popup from '~/components/layout/Popup.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';

const props = defineProps<{
  open: boolean;
  pipelineNumber: string;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

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
      parameters.value = parametersForEvent(definitions, 'deployment');
      parameterValues.value = initialParameterValues(parameters.value, {
        ...source.variables,
        ...(source.deploy_to ? { deploy_to: source.deploy_to } : {}),
        ...(source.deploy_task ? { deploy_task: source.deploy_task } : {}),
      });
      payload.value = { id: '', environment: source.deploy_to || '', task: source.deploy_task || '', variables: {} };
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
