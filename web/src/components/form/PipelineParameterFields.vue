<template>
  <div>
    <InputField v-for="parameter in parameters" :key="parameter.name" :label="label(parameter)">
      <template #description>{{ parameter.description }}</template>
      <template #default="{ id }">
        <SelectField
          v-if="parameter.type === 'choice'"
          :id="id"
          :model-value="value(parameter)"
          :options="[{ value: '', text: '—' }, ...parameter.options.map((option) => ({ value: option, text: option }))]"
          :required="parameter.required"
          @update:model-value="update(parameter, $event)"
        />
        <MultiSelectField
          v-else-if="parameter.type === 'multi_choice'"
          :id="id"
          :label="label(parameter)"
          :model-value="value(parameter)"
          :options="parameter.options"
          @update:model-value="update(parameter, $event)"
        />
        <Checkbox
          v-else-if="parameter.type === 'boolean'"
          :model-value="value(parameter) === 'true'"
          :label="label(parameter)"
          @update:model-value="update(parameter, String($event))"
        />
        <TextField
          v-else
          :id="id"
          :model-value="value(parameter)"
          :type="parameter.type === 'number' ? 'number' : 'text'"
          :required="parameter.required"
          :step="parameter.type === 'number' ? 'any' : undefined"
          @update:model-value="update(parameter, String($event))"
        />
      </template>
    </InputField>
    <p v-if="!valid" role="status" class="text-wp-text-alt-100 mb-4 text-sm">{{ $t('repo.parameters.invalid') }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import type { Parameter } from '~/lib/api/types';
import { isParameterValueValid } from '~/lib/pipelineParameters';

import Checkbox from './Checkbox.vue';
import InputField from './InputField.vue';
import MultiSelectField from './MultiSelectField.vue';
import SelectField from './SelectField.vue';
import TextField from './TextField.vue';

const props = defineProps<{ parameters: Parameter[]; modelValue: Record<string, string> }>();
const emit = defineEmits<{
  (event: 'update:modelValue', value: Record<string, string>): void;
  (event: 'update:isValid', value: boolean): void;
}>();
const i18n = useI18n();
const label = (parameter: Parameter) => {
  if (parameter.name === 'deploy_to') return i18n.t('repo.deploy_pipeline.enter_target');
  if (parameter.name === 'deploy_task') return i18n.t('repo.deploy_pipeline.enter_task');
  return parameter.name;
};
const value = (parameter: Parameter) =>
  Object.hasOwn(props.modelValue, parameter.name) ? props.modelValue[parameter.name] : '';
const valid = computed(() => props.parameters.every((parameter) => isParameterValueValid(parameter, value(parameter))));
watch(valid, (isValid) => emit('update:isValid', isValid), { immediate: true });
function update(parameter: Parameter, selected: string) {
  emit('update:modelValue', { ...props.modelValue, [parameter.name]: selected });
}
</script>
