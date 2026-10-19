import { describe, expect, it } from 'vitest';

import { resolveLocale } from './locale';

const supported = ['en', 'fr', 'zh-Hans', 'zh-Hant', 'pt-BR'];

describe('browser and stored locale resolution', () => {
  it.each([
    ['zh-CN', 'zh-Hans'],
    ['zh-SG', 'zh-Hans'],
    ['zh', 'zh-Hans'],
    ['zh_CN', 'zh-Hans'],
    ['zh-TW', 'zh-Hant'],
    ['zh-HK', 'zh-Hant'],
    ['zh-Hans', 'zh-Hans'],
    ['en', 'en'],
    ['fr-CA', 'fr'],
    ['pt-br', 'pt-BR'],
    ['unsupported', 'en'],
  ])('resolves %s to a supported locale without overriding explicit choices', (input, expected) => {
    expect(resolveLocale(input, supported)).toBe(expected);
  });
});
