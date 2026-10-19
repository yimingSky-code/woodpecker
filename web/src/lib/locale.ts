export function resolveLocale(language: string, supported: readonly string[]): string {
  const normalized = language.replaceAll('_', '-').toLowerCase();
  const exact = supported.find((locale) => locale.toLowerCase() === normalized);
  if (exact !== undefined) return exact;

  const aliases: Record<string, string> = {
    zh: 'zh-Hans',
    'zh-cn': 'zh-Hans',
    'zh-sg': 'zh-Hans',
    'zh-tw': 'zh-Hant',
    'zh-hk': 'zh-Hant',
    'zh-mo': 'zh-Hant',
  };
  const alias = aliases[normalized];
  if (alias && supported.includes(alias)) return alias;
  return supported.find((locale) => locale === normalized.split('-')[0]) ?? 'en';
}
