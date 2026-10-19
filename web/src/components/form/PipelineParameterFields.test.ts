import { mount } from '@vue/test-utils';
import { NSelect } from 'naive-ui';
import { describe, expect, it } from 'vitest';
import { defineComponent, h, nextTick, ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Parameter } from '~/lib/api/types';
import { initialParameterValues } from '~/lib/pipelineParameters';

import PipelineParameterFields from './PipelineParameterFields.vue';

const definitions: Parameter[] = [
  {
    id: 1,
    repo_id: 1,
    event: 'deployment',
    name: 'SERVICES',
    type: 'multi_choice',
    options: ['server', 'scheduler', 'worker'],
    default: 'server',
    required: true,
    description: '',
    order: 1,
    source: 'repo_config',
  },
  {
    id: 2,
    repo_id: 1,
    event: 'deployment',
    name: 'deploy_to',
    type: 'choice',
    options: ['testing', 'production'],
    default: 'testing',
    required: true,
    description: '',
    order: 2,
    source: 'repo_config',
  },
];

describe('pipeline parameter fields', () => {
  it('submits a stable service set and validates clearing all choices without restoring defaults', async () => {
    const values = ref(initialParameterValues(definitions));
    const valid = ref(false);
    const host = defineComponent({
      setup: () => () =>
        h(PipelineParameterFields, {
          parameters: definitions,
          modelValue: values.value,
          'onUpdate:modelValue': (value: Record<string, string>) => {
            values.value = value;
          },
          'onUpdate:isValid': (value: boolean) => {
            valid.value = value;
          },
        }),
    });
    const wrapper = mount(host, {
      global: { plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })] },
    });
    expect(valid.value).toBe(true);
    const boxes = wrapper.findAll('[role="checkbox"]');
    await boxes[2].trigger('click');
    expect(values.value.SERVICES).toBe('server,worker');
    wrapper.findComponent(NSelect).vm.$emit('update:value', 'production');
    await nextTick();
    expect(values.value.deploy_to).toBe('production');
    await boxes[0].trigger('click');
    await boxes[2].trigger('click');
    expect(values.value.SERVICES).toBe('');
    expect(valid.value).toBe(false);
    await boxes[1].trigger('click');
    expect(values.value.SERVICES).toBe('scheduler');
    expect(valid.value).toBe(true);
  });
});
