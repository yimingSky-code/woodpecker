<template>
  <fieldset :id="id" :aria-label="label" class="space-y-2">
    <Checkbox
      v-for="option in options"
      :key="option"
      :model-value="selected.includes(option)"
      :label="option"
      @update:model-value="toggle(option, $event)"
    />
  </fieldset>
</template>

<script setup lang="ts">
import { computed } from 'vue';

import Checkbox from './Checkbox.vue';

const props = defineProps<{ id?: string; label: string; modelValue: string; options: string[] }>();
const emit = defineEmits<{ (event: 'update:modelValue', value: string): void }>();
const selected = computed(() => (props.modelValue === '' ? [] : props.modelValue.split(',')));

function toggle(option: string, checked: boolean) {
  const values = new Set(selected.value);
  if (checked) values.add(option);
  else values.delete(option);
  // 采用配置顺序，避免点击顺序改变传给执行器的服务集合表示。
  emit('update:modelValue', props.options.filter((value) => values.has(value)).join(','));
}
</script>
