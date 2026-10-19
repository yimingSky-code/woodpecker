<template>
  <Panel v-if="!loading && !loadError">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-wp-text-100 text-xl">{{ $t('repo.manual_pipeline.title') }}</span>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.message.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.message.desc') }}</span>
        <TextField :id="id" v-model="payload.message" :placeholder="$t('repo.manual_pipeline.message.title')" />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required />
      </InputField>

      <PipelineParameterFields
        v-model="parameterValues"
        :parameters="parameters"
        @update:is-valid="parametersValid = $event"
      />

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
        <KeyValueEditor
          :id="id"
          v-model="payload.variables"
          :key-placeholder="$t('repo.manual_pipeline.variables.name')"
          :value-placeholder="$t('repo.manual_pipeline.variables.value')"
          :delete-title="$t('repo.manual_pipeline.variables.delete')"
          @update:is-valid="isVariablesValid = $event"
        />
      </InputField>

      <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" :disabled="!isFormValid" />
    </form>
  </Panel>
  <p v-else-if="loadError" role="alert">{{ $t('repo.parameters.load_failed') }}</p>
  <div v-else class="text-wp-text-100 flex justify-center">
    <Icon name="spinner" />
  </div>
</template>

<script lang="ts" setup>
import { useNotification } from '@kyvg/vue3-notification';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineParameterFields from '~/components/form/PipelineParameterFields.vue';
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
