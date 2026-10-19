import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, ref } from 'vue';

import OpsVariableEditor from './OpsVariableEditor.vue';

describe('naive UI variable editor', () => {
  it('keeps the original duplicate-key validation and corrects values through library inputs', async () => {
    const values = ref<Record<string, string>>({ FIRST: 'a', SECOND: 'b' });
    const valid = ref(true);
    const host = defineComponent({
      setup: () => () =>
        h(OpsVariableEditor, {
          id: 'variables',
          modelValue: values.value,
          'onUpdate:modelValue': (value: Record<string, string>) => {
            values.value = value;
          },
          'onUpdate:isValid': (value: boolean) => {
            valid.value = value;
          },
        }),
    });
    const wrapper = mount(host);
    const keys = wrapper.findAll('input[id*="-key-"]');
    expect(keys).toHaveLength(3);
    await keys[1].setValue('FIRST');
    expect(valid.value).toBe(false);
    await keys[1].setValue('OTHER');
    expect(valid.value).toBe(true);
    expect(values.value).toEqual({ FIRST: 'a', OTHER: 'b' });
    wrapper.unmount();
  });
});
