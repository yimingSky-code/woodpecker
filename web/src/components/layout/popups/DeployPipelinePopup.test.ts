import { flushPromises, mount } from '@vue/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import { createI18n } from 'vue-i18n';

import en from '~/assets/locales/en.json';
import type { Parameter } from '~/lib/api/types';

import DeployPipelinePopup from './DeployPipelinePopup.vue';

const { api, push } = vi.hoisted(() => ({
  api: { getParameterList: vi.fn(), getPipeline: vi.fn(), deployPipeline: vi.fn() },
  push: vi.fn(),
}));
vi.mock('~/compositions/useApiClient', () => ({ default: () => api }));
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }));

const parameters: Parameter[] = [
  {
    id: 1,
    repo_id: 1,
    event: 'deployment',
    name: 'SERVICES',
    type: 'multi_choice',
    options: ['server', 'worker'],
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
  {
    id: 3,
    repo_id: 1,
    event: 'deployment',
    name: 'deploy_task',
    type: 'choice',
    options: ['hot', 'cold'],
    default: 'hot',
    required: true,
    description: '',
    order: 3,
    source: 'repo_config',
  },
  {
    id: 4,
    repo_id: 1,
    event: 'manual',
    name: 'BUILD_ONLY',
    type: 'string',
    options: [],
    default: '',
    required: true,
    description: '',
    order: 4,
    source: 'repo_config',
  },
];

const mountPopup = (open = true) =>
  mount(DeployPipelinePopup, {
    props: { open, pipelineNumber: '42' },
    global: {
      config: { errorHandler: () => {} },
      plugins: [createI18n({ legacy: false, locale: 'en', messages: { en } })],
      provide: { repo: ref({ id: 1 }) },
    },
  });

describe('deploy typed parameters', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    api.getParameterList.mockResolvedValue(parameters);
    api.getPipeline.mockResolvedValue({ number: 42, commit: 'original-sha', variables: { SERVICES: 'server,worker' } });
    api.deployPipeline.mockResolvedValue({ number: 43 });
  });

  it('loads on open, inherits services and deploys the selected build with native target/task fields', async () => {
    const wrapper = mountPopup(false);
    expect(api.getPipeline).not.toHaveBeenCalled();
    await wrapper.setProps({ open: true });
    await flushPromises();
    expect(api.getPipeline).toHaveBeenCalledWith(1, 42);
    expect(wrapper.text()).not.toContain('BUILD_ONLY');
    expect(wrapper.findAll('input[type="checkbox"]').every((box) => (box.element as HTMLInputElement).checked)).toBe(
      true,
    );
    const selects = wrapper.findAll('select');
    await selects[0].setValue('production');
    await selects[1].setValue('cold');
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(api.deployPipeline).toHaveBeenCalledExactlyOnceWith(1, '42', {
      id: '',
      environment: 'production',
      task: 'cold',
      variables: { SERVICES: 'server,worker' },
    });
    expect(push).toHaveBeenCalledWith({ name: 'repo-pipeline', params: { pipelineId: 43 } });
  });

  it('keeps choices and permits correction after a rejected submission', async () => {
    api.deployPipeline.mockRejectedValueOnce(new Error('invalid input'));
    const wrapper = mountPopup();
    await flushPromises();
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(wrapper.find('form').exists()).toBe(true);
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeUndefined();
    await wrapper.get('form').trigger('submit');
    await flushPromises();
    expect(api.deployPipeline).toHaveBeenCalledTimes(2);
  });

  it('blocks submission when definitions cannot load', async () => {
    api.getParameterList.mockRejectedValue(new Error('unavailable'));
    const wrapper = mountPopup();
    await flushPromises();
    expect(wrapper.find('[role="alert"]').exists()).toBe(true);
    expect(wrapper.find('form').exists()).toBe(false);
    expect(api.deployPipeline).not.toHaveBeenCalled();
  });
});
