import type { Parameter, ParameterEvent } from '~/lib/api/types';

export function parametersForEvent(parameters: Parameter[], event: ParameterEvent): Parameter[] {
  return parameters.filter((parameter) => (parameter.event || 'manual') === event);
}

export function initialParameterValues(
  parameters: Parameter[],
  inherited: Record<string, string> = {},
): Record<string, string> {
  return Object.fromEntries(
    parameters.map((parameter) => [
      parameter.name,
      Object.hasOwn(inherited, parameter.name)
        ? inherited[parameter.name]
        : parameter.default || (parameter.type === 'boolean' ? 'false' : ''),
    ]),
  );
}

export function isParameterValueValid(parameter: Parameter, value: string): boolean {
  if (value === '') return !parameter.required && parameter.type !== 'boolean';
  switch (parameter.type) {
    case 'boolean':
      return value === 'true' || value === 'false';
    case 'number':
      return /^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i.test(value) && Number.isFinite(Number(value));
    case 'choice':
      return parameter.options.includes(value);
    case 'multi_choice': {
      const selected = value.split(',');
      return (
        new Set(selected).size === selected.length && selected.every((option) => parameter.options.includes(option))
      );
    }
    case 'string':
      return true;
  }
}
