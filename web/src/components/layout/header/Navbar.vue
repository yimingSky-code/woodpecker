<template>
  <nav class="navbar-shell">
    <div class="navbar-main">
      <router-link :to="{ name: 'home' }" class="navbar-brand">
        <span class="navbar-brand-mark"><WoodpeckerLogo aria-hidden="true" /></span>
        <span class="navbar-brand-title">{{ $t('branding.name') }}</span>
        <span class="navbar-version" :title="version?.current">{{ version?.currentShort }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-clickable navbar-link">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories.title') }}</span>
      </router-link>
      <a href="https://woodpecker-ci.org/" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('docs')
      }}</a>
      <a v-if="enableSwagger" :href="apiUrl" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('api')
      }}</a>
    </div>
    <div class="navbar-actions">
      <IconButton
        v-if="user?.admin"
        class="navbar-icon relative"
        :title="$t('settings')"
        :to="{ name: 'admin-settings' }"
      >
        <Icon name="settings" />
        <div v-if="version?.needsUpdate" class="bg-wp-error-100 absolute top-2 right-2 h-3 w-3 rounded-full" />
      </IconButton>

      <ActivePipelines v-if="user" class="navbar-icon p-1.5!" />
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings.settings')" class="navbar-icon p-1.5!">
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
      <Button v-else :text="$t('login')" :to="{ name: 'login' }" class="navbar-link" @click="saveRedirect" />
    </div>
  </nav>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useUserConfig from '~/compositions/useUserConfig';
import { useVersion } from '~/compositions/useVersion';

import ActivePipelines from './ActivePipelines.vue';

const version = useVersion();
const config = useConfig();
const userConfig = useUserConfig();
const route = useRoute();
const authentication = useAuthentication();
const { user } = authentication;
const apiUrl = `${config.rootPath ?? ''}/swagger/index.html`;

const { enableSwagger } = config;

function saveRedirect() {
  userConfig.setUserConfig('redirectUrl', route.fullPath);
}
</script>

<style scoped>
.navbar-shell {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  gap: 24px;
  min-height: 66px;
  padding: 10px 28px;
  border-bottom: 1px solid var(--wp-control-neutral-200);
  background: var(--wp-background-100);
  color: var(--wp-text-200);
  font-size: 13px;
  font-weight: 500;
}
.navbar-main,
.navbar-actions,
.navbar-brand {
  display: flex;
  align-items: center;
}
.navbar-main {
  gap: 8px;
  min-width: 0;
}
.navbar-actions {
  margin-left: auto;
  gap: 6px;
}
.navbar-brand {
  gap: 10px;
  margin-right: 22px;
  flex-shrink: 0;
}
.navbar-brand-mark {
  display: grid;
  place-items: center;
  width: 34px;
  height: 36px;
  border-radius: 10px;
  background: var(--color-int-wp-primary-600);
}
.navbar-brand-mark :deep(svg) {
  width: 22px;
  height: 26px;
}
.navbar-brand-title {
  font-size: 16px;
  font-weight: 650;
  letter-spacing: -0.035em;
}
.navbar-version {
  color: var(--wp-text-alt-100);
  font-size: 10px;
  font-weight: 500;
  padding: 2px 5px;
  border: 1px solid var(--wp-control-neutral-200);
  border-radius: 4px;
}
.navbar-icon {
  width: 34px;
  height: 34px;
  padding: 7px;
  border-radius: 8px;
  color: var(--wp-text-alt-100);
}
.navbar-icon:hover,
.navbar-link:hover {
  background: var(--wp-control-neutral-100);
  color: var(--wp-text-200);
}
.navbar-icon :deep(svg) {
  width: 100%;
  height: 100%;
}
.navbar-link {
  padding: 8px 12px;
  border-radius: 7px;
  color: var(--wp-text-alt-100);
}
.navbar-main :deep(.navbar-link.router-link-active) {
  color: var(--color-int-wp-primary-600);
  background: color-mix(in srgb, var(--wp-background-100) 93%, var(--color-int-wp-primary-400));
}
@media (max-width: 600px) {
  .navbar-shell {
    gap: 8px;
    min-height: 58px;
    padding: 8px 16px;
  }
  .navbar-brand {
    gap: 8px;
    margin-right: 4px;
  }
  .navbar-brand-title {
    font-size: 14px;
  }
  .navbar-version {
    display: none;
  }
  .navbar-link {
    padding-inline: 8px;
  }
}
</style>
