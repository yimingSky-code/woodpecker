<template>
  <KeyValueEditor
    v-bind="$props"
    @update:model-value="$emit('update:modelValue', $event)"
    @update:is-valid="$emit('update:isValid', $event)"
  >
    <template #key-input="{ id: inputId, value, invalid, update }">
      <NInput
        :value="value"
        :placeholder="keyPlaceholder"
        :input-props="{ id: inputId, 'aria-label': keyPlaceholder }"
        :status="invalid ? 'error' : undefined"
        @update:value="update"
      />
    </template>
    <template #value-input="{ id: inputId, value, update }">
      <NInput
        :value="value"
        :placeholder="valuePlaceholder"
        :input-props="{ id: inputId, 'aria-label': valuePlaceholder }"
        @update:value="update"
      />
    </template>
    <template #remove-button="{ remove }">
      <NButton quaternary circle :title="deleteTitle" :aria-label="deleteTitle" @click="remove">
        <template #icon><Icon name="close" class="h-4! w-4!" /></template>
      </NButton>
    </template>
  </KeyValueEditor>
</template>

<script setup lang="ts">
import { NButton, NInput } from 'naive-ui';

import Icon from '~/components/atomic/Icon.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';

defineProps<{
  modelValue: Record<string, string>;
  id?: string;
  keyPlaceholder?: string;
  valuePlaceholder?: string;
  deleteTitle?: string;
}>();
defineEmits<{
  (event: 'update:modelValue', value: Record<string, string>): void;
  (event: 'update:isValid', value: boolean): void;
}>();
</script>
