<template>
  <OpsUiProvider>
    <Settings
      :title="$t('repo.settings.parameters.parameters')"
      :description="$t('repo.settings.parameters.desc')"
      docs-url="docs/usage/parameters"
    >
      <template #headerActions>
        <NButton v-if="selectedParameter" @click="selectedParameter = undefined">
          {{ $t('repo.settings.parameters.show') }}
        </NButton>
        <NButton v-else type="primary" @click="selectedParameter = { type: 'string', event: 'manual' }">
          {{ $t('repo.settings.parameters.add') }}
        </NButton>
      </template>
      <div v-if="!selectedParameter">
        <div v-for="parameter in parameters" :key="parameter.id" class="ops-parameter-row">
          <div class="min-w-0">
            <div class="ops-parameter-name">{{ parameter.name }}</div>
            <div class="ops-parameter-meta">
              <NTag :bordered="false" size="small" :type="parameter.event === 'deployment' ? 'success' : 'info'">
                {{
                  parameter.event === 'deployment'
                    ? $t('repo.settings.parameters.event.deployment')
                    : $t('repo.settings.parameters.event.manual')
                }}
              </NTag>
              <NTag :bordered="false" size="small">{{ parameterTypeLabel(parameter.type) }}</NTag>
            </div>
            <p v-if="parameter.description" class="ops-field-help">{{ parameter.description }}</p>
          </div>
          <div class="flex shrink-0 gap-2">
            <NButton
              quaternary
              circle
              :title="$t('repo.settings.parameters.edit')"
              :aria-label="$t('repo.settings.parameters.edit')"
              @click="selectedParameter = { ...parameter, options: [...(parameter.options ?? [])] }"
            >
              <template #icon><Icon name="edit" class="h-4! w-4!" /></template>
            </NButton>
            <NButton
              quaternary
              circle
              type="error"
              :loading="isDeleting"
              :title="$t('repo.settings.parameters.delete')"
              :aria-label="$t('repo.settings.parameters.delete')"
              @click="deleteParameter(parameter)"
            >
              <template #icon><Icon name="trash" class="h-4! w-4!" /></template>
            </NButton>
          </div>
        </div>
        <div v-if="loading" class="flex justify-center py-8"><NSpin /></div>
        <NEmpty v-else-if="parameters?.length === 0" :description="$t('repo.settings.parameters.none')" class="py-8" />
      </div>
      <NForm v-else label-placement="top" @submit.prevent="saveParameter">
        <div class="ops-form-grid">
          <NFormItem :label="$t('repo.settings.parameters.name.name')" required>
            <NInput
              v-model:value="selectedParameter.name"
              :placeholder="$t('repo.settings.parameters.name.placeholder')"
              :input-props="{ required: true, 'aria-label': $t('repo.settings.parameters.name.name') }"
            />
          </NFormItem>
          <NFormItem :label="$t('repo.settings.parameters.event.title')">
            <NSelect
              v-model:value="selectedParameterEvent"
              :options="eventOptions"
              :input-props="{ 'aria-label': $t('repo.settings.parameters.event.title') }"
            />
          </NFormItem>
          <NFormItem :label="$t('repo.settings.parameters.type.type')">
            <NSelect
              v-model:value="selectedParameterType"
              :options="parameterTypeOptions"
              :input-props="{ 'aria-label': $t('repo.settings.parameters.type.type') }"
            />
          </NFormItem>
          <NFormItem :label="$t('repo.settings.parameters.order.order')">
            <NInputNumber
              v-model:value="selectedParameterOrder"
              :input-props="{ 'aria-label': $t('repo.settings.parameters.order.order') }"
              class="w-full"
            />
          </NFormItem>
        </div>
        <NFormItem :label="$t('repo.settings.parameters.description.description')">
          <NInput
            v-model:value="selectedParameter.description"
            :placeholder="$t('repo.settings.parameters.description.placeholder')"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 4 }"
            :input-props="{ 'aria-label': $t('repo.settings.parameters.description.description') }"
          />
        </NFormItem>
        <NFormItem
          v-if="selectedParameter.type === 'choice' || selectedParameter.type === 'multi_choice'"
          :label="$t('repo.settings.parameters.options.options')"
          required
        >
          <div class="w-full">
            <NInput
              v-model:value="selectedParameterOptionsText"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              :input-props="{ required: true, 'aria-label': $t('repo.settings.parameters.options.options') }"
            />
            <p class="ops-field-help">{{ $t('repo.settings.parameters.options.desc') }}</p>
          </div>
        </NFormItem>
        <NFormItem :label="$t('repo.settings.parameters.default')">
          <NSelect
            v-if="selectedParameter.type === 'choice'"
            v-model:value="selectedParameterDefault"
            :options="defaultChoiceOptions"
            :input-props="{ 'aria-label': $t('repo.settings.parameters.default') }"
          />
          <MultiSelectField
            v-else-if="selectedParameter.type === 'multi_choice'"
            v-model="selectedParameterDefault"
            :label="$t('repo.settings.parameters.default')"
            :options="selectedParameter.options ?? []"
            class="w-full"
          />
          <NCheckbox v-else-if="selectedParameter.type === 'boolean'" v-model:checked="selectedParameterDefaultBoolean">
            {{ $t('repo.settings.parameters.default') }}
          </NCheckbox>
          <NInput
            v-else
            v-model:value="selectedParameterDefault"
            :input-props="{ 'aria-label': $t('repo.settings.parameters.default') }"
          />
        </NFormItem>
        <NCheckbox v-if="selectedParameter.type !== 'boolean'" v-model:checked="selectedParameterRequired">
          {{ $t('repo.settings.parameters.required') }}
        </NCheckbox>
        <div class="ops-actions ops-section">
          <NButton @click="selectedParameter = undefined">{{ $t('cancel') }}</NButton>
          <NButton attr-type="submit" type="primary" :loading="isSaving">
            {{ isEditingParameter ? $t('repo.settings.parameters.save') : $t('repo.settings.parameters.add') }}
          </NButton>
        </div>
      </NForm>
    </Settings>
  </OpsUiProvider>
</template>

<script lang="ts" setup>
import { NButton, NCheckbox, NEmpty, NForm, NFormItem, NInput, NInputNumber, NSelect, NSpin, NTag } from 'naive-ui';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import MultiSelectField from '~/components/form/MultiSelectField.vue';
import Settings from '~/components/layout/Settings.vue';
import OpsUiProvider from '~/components/ops/OpsUiProvider.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Parameter, ParameterType, ParameterEvent } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = requiredInject('repo');
const selectedParameter = ref<Partial<Parameter>>();
const isEditingParameter = computed(() => !!selectedParameter.value?.id);

const parameterTypeLabels: Record<ParameterType, string> = {
  string: i18n.t('repo.settings.parameters.type.string'),
  number: i18n.t('repo.settings.parameters.type.number'),
  boolean: i18n.t('repo.settings.parameters.type.boolean'),
  choice: i18n.t('repo.settings.parameters.type.choice'),
  multi_choice: i18n.t('repo.settings.parameters.type.multi_choice'),
};

function parameterTypeLabel(type: ParameterType): string {
  return parameterTypeLabels[type];
}

const parameterTypeOptions = Object.entries(parameterTypeLabels).map(([type, label]) => ({
  value: type,
  label,
}));

const eventOptions = [
  { value: 'manual', label: i18n.t('repo.settings.parameters.event.manual') },
  { value: 'deployment', label: i18n.t('repo.settings.parameters.event.deployment') },
];
const selectedParameterEvent = computed<string>({
  get: () => selectedParameter.value!.event || 'manual',
  set: (value) => {
    selectedParameter.value!.event = value as ParameterEvent;
  },
});

const selectedParameterType = computed<string>({
  set(type) {
    selectedParameter.value!.type = type as ParameterType;
    // a default rarely stays valid across type changes (e.g. free text -> choice)
    selectedParameter.value!.default = '';
    if (type === 'boolean') {
      selectedParameter.value!.required = false;
    }
  },
  get() {
    return selectedParameter.value!.type ?? 'string';
  },
});

const selectedParameterOptions = computed(() =>
  (selectedParameter.value?.options ?? []).map((option) => ({ value: option, label: option })),
);

// leading empty entry so a previously picked default can be unset again
const defaultChoiceOptions = computed(() => [
  { value: '', label: i18n.t('repo.settings.parameters.no_default') },
  ...selectedParameterOptions.value,
]);

const selectedParameterOptionsText = computed<string>({
  set(text) {
    selectedParameter.value!.options = text
      .split('\n')
      .map((option) => option.trim())
      .filter((option) => option !== '');
  },
  get() {
    return (selectedParameter.value!.options ?? []).join('\n');
  },
});

const selectedParameterDefault = computed<string>({
  set(value) {
    // Vue's v-model on <input type="number"> emits numbers, but Default is a string
    selectedParameter.value!.default = String(value);
  },
  get() {
    return selectedParameter.value!.default ?? '';
  },
});

const selectedParameterDefaultBoolean = computed<boolean>({
  set(value) {
    selectedParameter.value!.default = value ? 'true' : 'false';
  },
  get() {
    return selectedParameter.value!.default === 'true';
  },
});

const selectedParameterRequired = computed<boolean>({
  set(required) {
    selectedParameter.value!.required = required;
  },
  get() {
    return selectedParameter.value!.required ?? false;
  },
});

const selectedParameterOrder = computed<number | null>({
  set(order) {
    selectedParameter.value!.order = order === null || Number.isNaN(order) ? 0 : order;
  },
  get() {
    return selectedParameter.value!.order ?? 0;
  },
});

async function loadParameters(page: number): Promise<Parameter[] | null> {
  return apiClient.getParameterList(repo.value.id, { page });
}

const { resetPage, data: parameters, loading } = usePagination(loadParameters, () => !selectedParameter.value);

const { doSubmit: saveParameter, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedParameter.value) {
    throw new Error("Unexpected: Can't get parameter");
  }

  if (isEditingParameter.value) {
    await apiClient.updateParameter(repo.value.id, selectedParameter.value);
  } else {
    await apiClient.createParameter(repo.value.id, selectedParameter.value);
  }
  notifications.notify({
    title: isEditingParameter.value
      ? i18n.t('repo.settings.parameters.saved')
      : i18n.t('repo.settings.parameters.created'),
    type: 'success',
  });
  selectedParameter.value = undefined;
  await resetPage();
});

const { doSubmit: deleteParameter, isLoading: isDeleting } = useAsyncAction(async (_parameter: Parameter) => {
  await apiClient.deleteParameter(repo.value.id, _parameter.id);
  notifications.notify({ title: i18n.t('repo.settings.parameters.deleted'), type: 'success' });
  await resetPage();
});

useWPTitle(computed(() => [i18n.t('repo.settings.parameters.parameters'), repo.value.full_name]));
</script>
