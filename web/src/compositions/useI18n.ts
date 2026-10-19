import { useStorage } from '@vueuse/core';
import { SUPPORTED_LOCALES } from 'virtual:vue-i18n-supported-locales';
import { nextTick } from 'vue';
import { createI18n } from 'vue-i18n';

import { resolveLocale } from '~/lib/locale';

import { useDate } from './useDate';

export function getUserLanguage(): string {
  const browserLocale = resolveLocale(navigator.language, SUPPORTED_LOCALES);
  const selectedLocale = useStorage('woodpecker:locale', browserLocale).value;

  return resolveLocale(selectedLocale, SUPPORTED_LOCALES);
}

const { setDateLocale } = useDate();
const userLanguage = getUserLanguage();
const fallbackLocale = 'en';
export const i18n = createI18n({
  locale: userLanguage,
  legacy: false,
  globalInjection: true,
  fallbackLocale,
});

const loadLocaleMessages = async (locale: string) => {
  const messages = (await import(`~/assets/locales/${locale}.json`)) as { default: any };

  i18n.global.setLocaleMessage(locale, messages.default);

  return nextTick();
};

export const setI18nLanguage = async (lang: string): Promise<void> => {
  if (!i18n.global.availableLocales.includes(lang)) {
    await loadLocaleMessages(lang);
  }
  i18n.global.locale.value = lang;
  await setDateLocale(lang);
};

loadLocaleMessages(fallbackLocale).catch(console.error);
loadLocaleMessages(userLanguage).catch(console.error);
setDateLocale(userLanguage).catch(console.error);
