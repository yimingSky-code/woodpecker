<template>
  <NConfigProvider
    :theme="theme === 'dark' ? darkTheme : null"
    :theme-overrides="overrides"
    :locale="locale.startsWith('zh') ? zhCN : enUS"
    abstract
    preflight-style-disabled
  >
    <slot />
  </NConfigProvider>
</template>

<script setup lang="ts">
import { darkTheme, enUS, NConfigProvider, zhCN } from 'naive-ui';
import type { GlobalThemeOverrides } from 'naive-ui';
import { useI18n } from 'vue-i18n';

import { useTheme } from '~/compositions/useTheme';

import './ops.css';

const { theme } = useTheme();
const { locale } = useI18n();
const overrides: GlobalThemeOverrides = {
  common: { borderRadius: '8px', borderRadiusSmall: '5px', fontFamily: 'inherit', fontSize: '14px' },
  Form: { labelFontWeight: '600' },
};
</script>
