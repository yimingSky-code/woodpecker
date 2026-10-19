import { describe, expect, it } from 'vitest';

import type { Parameter } from '~/lib/api/types';

import { initialParameterValues, isParameterValueValid, parametersForEvent } from './pipelineParameters';

const parameter: Parameter = {
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
};

describe('pipeline parameter values', () => {
  it('distinguishes missing values from explicitly cleared choices', () => {
    expect(initialParameterValues([parameter])).toEqual({ SERVICES: 'server' });
    expect(initialParameterValues([parameter], { SERVICES: '' })).toEqual({ SERVICES: '' });
    expect(isParameterValueValid(parameter, '')).toBe(false);
    expect(isParameterValueValid(parameter, 'server,worker')).toBe(true);
    expect(isParameterValueValid(parameter, 'server,server')).toBe(false);
    expect(isParameterValueValid(parameter, 'server,unknown')).toBe(false);
  });
  it('keeps manual and Deploy definitions independent', () => {
    expect(parametersForEvent([parameter], 'manual')).toEqual([]);
    expect(parametersForEvent([parameter], 'deployment')).toEqual([parameter]);
  });
});
