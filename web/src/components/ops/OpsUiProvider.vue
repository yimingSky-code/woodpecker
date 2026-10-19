<template>
  <NConfigProvider
    :theme="theme === 'dark' ? darkTheme : null"
    :theme-overrides="overrides"
    abstract
    preflight-style-disabled
  >
    <slot />
  </NConfigProvider>
</template>

<script setup lang="ts">
import { darkTheme, NConfigProvider } from 'naive-ui';
import type { GlobalThemeOverrides } from 'naive-ui';

import { useTheme } from '~/compositions/useTheme';

import './ops.css';

const { theme } = useTheme();
const overrides: GlobalThemeOverrides = {
  common: { borderRadius: '8px', borderRadiusSmall: '5px', fontFamily: 'inherit', fontSize: '14px' },
  Form: { labelFontWeight: '600' },
};
</script>
