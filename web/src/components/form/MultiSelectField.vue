<template>
  <NCheckboxGroup :id="id" :aria-label="label" :value="selected" @update:value="update">
    <div class="ops-choice-grid">
      <NCheckbox
        v-for="option in options"
        :key="option"
        :value="option"
        :label="option"
        class="ops-choice-tile"
        :class="{ 'is-selected': selected.includes(option) }"
      />
    </div>
  </NCheckboxGroup>
</template>

<script setup lang="ts">
import { NCheckbox, NCheckboxGroup } from 'naive-ui';
import { computed } from 'vue';

const props = defineProps<{ id?: string; label: string; modelValue: string; options: string[] }>();
const emit = defineEmits<{ (event: 'update:modelValue', value: string): void }>();
const selected = computed(() => (props.modelValue === '' ? [] : props.modelValue.split(',')));
function update(selectedValues: Array<string | number>) {
  // 采用配置顺序，避免点击顺序改变传给执行器的服务集合表示。
  emit('update:modelValue', props.options.filter((value) => selectedValues.includes(value)).join(','));
}
</script>
