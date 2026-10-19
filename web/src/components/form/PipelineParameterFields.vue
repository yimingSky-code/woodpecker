<template>
  <div>
    <NFormItem
      v-for="parameter in parameters"
      :key="parameter.name"
      :label="label(parameter)"
      :show-feedback="false"
      :show-label="parameter.type !== 'boolean'"
      :required="parameter.required"
      :validation-status="
        touched.has(parameter.name) && !isParameterValueValid(parameter, value(parameter)) ? 'error' : undefined
      "
      class="mb-6"
    >
      <div class="w-full">
        <NSelect
          v-if="parameter.type === 'choice'"
          :value="value(parameter) || null"
          :options="parameter.options.map((option) => ({ value: option, label: optionLabel(parameter, option) }))"
          :placeholder="$t('repo.parameters.select_placeholder')"
          :input-props="{ 'aria-label': label(parameter) }"
          clearable
          @update:value="update(parameter, $event ?? '')"
        />
        <MultiSelectField
          v-else-if="parameter.type === 'multi_choice'"
          :label="label(parameter)"
          :model-value="value(parameter)"
          :options="parameter.options"
          @update:model-value="update(parameter, $event)"
        />
        <NCheckbox
          v-else-if="parameter.type === 'boolean'"
          :checked="value(parameter) === 'true'"
          @update:checked="update(parameter, String($event))"
        >
          {{ label(parameter) }}
        </NCheckbox>
        <NInputNumber
          v-else-if="parameter.type === 'number'"
          :value="value(parameter) === '' ? null : Number(value(parameter))"
          :show-button="false"
          :input-props="{ 'aria-label': label(parameter) }"
          class="w-full"
          @update:value="update(parameter, $event === null ? '' : String($event))"
        />
        <NInput
          v-else
          :value="value(parameter)"
          :input-props="{ 'aria-label': label(parameter) }"
          @update:value="update(parameter, $event)"
        />
        <div v-if="parameter.description" class="ops-field-help">{{ parameter.description }}</div>
      </div>
    </NFormItem>
    <p v-if="touched.size > 0 && !valid" role="status" class="ops-description">{{ $t('repo.parameters.invalid') }}</p>
  </div>
</template>

<script setup lang="ts">
import { NCheckbox, NFormItem, NInput, NInputNumber, NSelect } from 'naive-ui';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { Parameter } from '~/lib/api/types';
import { isParameterValueValid } from '~/lib/pipelineParameters';

import MultiSelectField from './MultiSelectField.vue';

const props = defineProps<{ parameters: Parameter[]; modelValue: Record<string, string> }>();
const emit = defineEmits<{
  (event: 'update:modelValue', value: Record<string, string>): void;
  (event: 'update:isValid', value: boolean): void;
}>();
const i18n = useI18n();
const touched = ref(new Set<string>());
const label = (parameter: Parameter) => {
  if (parameter.name === 'deploy_to') return i18n.t('repo.deploy_pipeline.enter_target');
  if (parameter.name === 'deploy_task') return i18n.t('repo.deploy_pipeline.enter_task');
  if (parameter.name === 'BUILD_SERVICES') return i18n.t('repo.parameters.services');
  if (parameter.name === 'AUTO_DEPLOY') return i18n.t('repo.parameters.auto_deploy');
  if (parameter.name === 'AUTO_DEPLOY_TARGET') return i18n.t('repo.parameters.auto_deploy_target');
  const text = parameter.name.toLowerCase().replaceAll('_', ' ');
  return text.charAt(0).toUpperCase() + text.slice(1);
};
function optionLabel(parameter: Parameter, option: string): string {
  if (parameter.name === 'deploy_to' || parameter.name === 'AUTO_DEPLOY_TARGET') {
    if (option === 'test' || option === 'testing') return `${i18n.t('repo.parameters.testing')} · ${option}`;
    if (option === 'prod' || option === 'production') return `${i18n.t('repo.parameters.production')} · ${option}`;
  }
  if (parameter.name === 'deploy_task') {
    if (option === 'hot') return `${i18n.t('repo.parameters.hot')} · ${option}`;
    if (option === 'cold') return `${i18n.t('repo.parameters.cold')} · ${option}`;
  }
  return option;
}
const value = (parameter: Parameter) =>
  Object.hasOwn(props.modelValue, parameter.name) ? props.modelValue[parameter.name] : '';
const valid = computed(() => props.parameters.every((parameter) => isParameterValueValid(parameter, value(parameter))));
watch(valid, (isValid) => emit('update:isValid', isValid), { immediate: true });
function update(parameter: Parameter, selected: string) {
  touched.value.add(parameter.name);
  emit('update:modelValue', { ...props.modelValue, [parameter.name]: selected });
}
</script>
