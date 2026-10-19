<template>
  <OpsUiProvider>
    <div v-if="pipeline && repo" class="pipeline-workbench">
      <Scaffold enable-tabs :go-back="goBack" :fluid-content="route.name === 'repo-pipeline'" full-width-header>
        <template #title>
          <span class="pipeline-breadcrumb">
            <router-link :to="{ name: 'org', params: { orgId: repo.org_id } }" class="hover:underline">{{
              repo.owner
              /* eslint-disable-next-line @intlify/vue-i18n/no-raw-text */
            }}</router-link>
            /
            <router-link :to="{ name: 'repo' }" class="pipeline-repo-name hover:underline" :title="repo.name">{{
              repo.name
            }}</router-link>
          </span>
        </template>

        <template #headerActions>
          <div class="pipeline-summary">
            <div class="pipeline-summary-main">
              <div class="pipeline-run-heading">
                <h2>{{ $t('repo.pipeline.pipeline', { pipelineId }) }}</h2>
                <NTag :bordered="false" :type="statusTagType" round>
                  <template #icon><PipelineStatusIcon :status="pipeline.status" class="h-4! w-4!" /></template>
                  {{ statusLabels[pipeline.status] }}
                </NTag>
                <NTag :bordered="false" round>{{ eventLabels[pipeline.event] || pipeline.event }}</NTag>
              </div>
              <RenderMarkdown class="pipeline-run-message" :title="message" :content="shortMessage" inline />
            </div>
            <div v-if="repoPermissions!.push && pipeline.status !== 'blocked'" class="pipeline-primary-actions">
              <NButton
                v-if="pipeline.status === 'pending' || pipeline.status === 'running'"
                :loading="isCancelingPipeline"
                @click="cancelPipeline"
              >
                {{ $t('repo.pipeline.actions.cancel') }}
              </NButton>
              <NButton :loading="isRestartingPipeline" @click="restartPipeline">
                <template #icon><Icon name="refresh" /></template>
                {{ $t('repo.pipeline.actions.restart') }}
              </NButton>
              <NButton
                v-if="pipeline.status === 'success' && repo.allow_deploy"
                type="primary"
                @click="showDeployPipelinePopup = true"
              >
                <template #icon><Icon name="deployment" /></template>
                {{ $t('repo.pipeline.actions.deploy') }}
              </NButton>
            </div>
          </div>
          <DeployPipelinePopup
            v-if="repoPermissions!.push && pipeline.status !== 'blocked'"
            :pipeline-number="pipelineId"
            :open="showDeployPipelinePopup"
            @close="showDeployPipelinePopup = false"
          />
        </template>

        <template #tabActions>
          <div class="pipeline-run-metrics">
            <div class="flex shrink-0 items-center gap-2" :title="$t('repo.pipeline.created', { created })">
              <Icon name="since" />
              <span>{{ since }}</span>
            </div>
            <div
              class="flex shrink-0 items-center gap-2"
              :title="
                durationElapsed > 0 ? $t('repo.pipeline.duration', { duration: durationAsNumber(durationElapsed) }) : ''
              "
            >
              <Icon name="duration" />
              <span>{{ duration }}</span>
            </div>
            <div v-if="pipeline.status === 'killed' && pipeline.cancel_info" class="flex shrink-0 items-center gap-2">
              <Icon name="status-killed" />
              <span class="truncate">
                <router-link
                  v-if="pipeline.cancel_info.superseded_by"
                  :to="{ name: 'repo-pipeline', params: { pipelineId: pipeline.cancel_info.superseded_by } }"
                  class="hover:underline"
                >
                  {{
                    $t('repo.pipeline.cancel_info.superseded_by', { pipelineId: pipeline.cancel_info.superseded_by })
                  }}
                </router-link>
                <template v-else-if="pipeline.cancel_info.canceled_by_user">
                  {{
                    $t('repo.pipeline.cancel_info.canceled_by_user', { user: pipeline.cancel_info.canceled_by_user })
                  }}
                </template>
                <template v-else-if="pipeline.cancel_info.canceled_by_step">
                  {{
                    $t('repo.pipeline.cancel_info.canceled_by_step', { user: pipeline.cancel_info.canceled_by_step })
                  }}
                </template>
              </span>
            </div>
          </div>
        </template>

        <Tab icon="tray-full" :to="{ name: 'repo-pipeline' }" :title="$t('repo.pipeline.tasks')" />
        <Tab
          v-if="errorsTabCount > 0"
          :to="{ name: 'repo-pipeline-errors' }"
          icon="alert"
          :title="pipelineHasErrorsToShow(pipeline) ? $t('repo.pipeline.errors') : $t('repo.pipeline.warnings')"
          :count="errorsTabCount"
          :icon-class="pipelineHasErrorsToShow(pipeline) ? 'text-wp-error-100' : 'text-wp-state-warn-100'"
        />
        <Tab icon="file-cog-outline" :to="{ name: 'repo-pipeline-config' }" :title="$t('repo.pipeline.config')" />
        <Tab
          v-if="pipeline.changed_files && pipeline.changed_files.length > 0"
          :to="{ name: 'repo-pipeline-changed-files' }"
          icon="file-edit-outline"
          :title="$t('repo.pipeline.files')"
          :count="pipeline.changed_files?.length"
        />
        <Tab
          v-if="repoPermissions && repoPermissions.push"
          icon="bug-outline"
          :to="{ name: 'repo-pipeline-debug' }"
          :title="$t('repo.pipeline.debug.title')"
        />

        <router-view />
      </Scaffold>
    </div>
  </OpsUiProvider>
</template>

<script lang="ts" setup>
import { NButton, NTag } from 'naive-ui';
import { computed, onMounted, ref, toRef, watch } from 'vue';
import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import RenderMarkdown from '~/components/atomic/RenderMarkdown.vue';
import DeployPipelinePopup from '~/components/layout/popups/DeployPipelinePopup.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import OpsUiProvider from '~/components/ops/OpsUiProvider.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useDate } from '~/compositions/useDate';
import { useFavicon } from '~/compositions/useFavicon';
import { provide, requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import usePipeline from '~/compositions/usePipeline';
import { useRouteBack } from '~/compositions/useRouteBack';
import type { Pipeline, PipelineConfig } from '~/lib/api/types';
import { pipelineHasErrorsToShow, workflowsWithErrors } from '~/lib/pipeline';
import { usePipelineStore } from '~/store/pipelines';

import '~/components/repo/pipeline/pipeline-workbench.css';

const props = defineProps<{
  repoId: string;
  pipelineId: string;
}>();

const apiClient = useApiClient();
const route = useRoute();
const router = useRouter();
const notifications = useNotifications();
const favicon = useFavicon();
const i18n = useI18n();

const pipelineStore = usePipelineStore();
const { durationAsNumber } = useDate();
const pipelineId = toRef(props, 'pipelineId');
const _repoId = toRef(props, 'repoId');
const repositoryId = computed(() => Number.parseInt(_repoId.value, 10));
const repo = requiredInject('repo');
const repoPermissions = requiredInject('repo-permissions');

const pipeline = pipelineStore.getPipeline(repositoryId, pipelineId);
const { since, duration, durationElapsed, created, message, shortMessage } = usePipeline(pipeline);

const statusLabels = computed(() => ({
  blocked: i18n.t('repo.pipeline.status.blocked'),
  declined: i18n.t('repo.pipeline.status.declined'),
  error: i18n.t('repo.pipeline.status.error'),
  failure: i18n.t('repo.pipeline.status.failure'),
  killed: i18n.t('repo.pipeline.status.killed'),
  pending: i18n.t('repo.pipeline.status.pending'),
  running: i18n.t('repo.pipeline.status.running'),
  skipped: i18n.t('repo.pipeline.status.skipped'),
  canceled: i18n.t('repo.pipeline.status.canceled'),
  started: i18n.t('repo.pipeline.status.started'),
  success: i18n.t('repo.pipeline.status.success'),
}));
const eventLabels = computed<Record<string, string>>(() => ({
  manual: i18n.t('repo.pipeline.event.manual'),
  deployment: i18n.t('repo.pipeline.event.deploy'),
  push: i18n.t('repo.pipeline.event.push'),
  tag: i18n.t('repo.pipeline.event.tag'),
  cron: i18n.t('repo.pipeline.event.cron'),
  release: i18n.t('repo.pipeline.event.release'),
  pull_request: i18n.t('repo.pipeline.event.pr'),
  pull_request_closed: i18n.t('repo.pipeline.event.pr_closed'),
  pull_request_metadata: i18n.t('repo.pipeline.event.pr_metadata'),
}));
const statusTagType = computed(() => {
  const status = pipeline.value?.status;
  if (status === 'success') return 'success';
  if (status === 'error' || status === 'failure' || status === 'killed') return 'error';
  if (status === 'running' || status === 'started') return 'info';
  return 'default';
});

const errorsTabCount = computed(
  () => (pipeline.value?.errors?.length ?? 0) + workflowsWithErrors(pipeline.value).length,
);
provide('pipeline', pipeline as Ref<Pipeline>); // can't be undefined because of v-if in template

const pipelineConfigs = ref<PipelineConfig[]>();
provide('pipeline-configs', pipelineConfigs);

watch(
  pipeline,
  () => {
    favicon.updateStatus(pipeline.value?.status);
  },
  { immediate: true },
);

const showDeployPipelinePopup = ref(false);

async function loadPipeline(): Promise<void> {
  await pipelineStore.loadPipeline(repo.value.id, Number.parseInt(pipelineId.value, 10));

  if (!pipeline.value?.number) {
    throw new Error('Unexpected: Pipeline number not found');
  }

  pipelineConfigs.value = await apiClient.getPipelineConfig(repo.value.id, pipeline.value.number);
}

const { doSubmit: cancelPipeline, isLoading: isCancelingPipeline } = useAsyncAction(async () => {
  if (!pipeline.value?.number) {
    throw new Error('Unexpected: Pipeline number not found');
  }

  await apiClient.cancelPipeline(repo.value.id, pipeline.value.number);
  notifications.notify({ title: i18n.t('repo.pipeline.actions.cancel_success'), type: 'success' });
});

const { doSubmit: restartPipeline, isLoading: isRestartingPipeline } = useAsyncAction(async () => {
  const newPipeline = await apiClient.restartPipeline(repo.value.id, pipelineId.value, {
    fork: true,
  });
  notifications.notify({ title: i18n.t('repo.pipeline.actions.restart_success'), type: 'success' });
  await router.push({
    name: 'repo-pipeline',
    params: { pipelineId: newPipeline.number },
  });
});

onMounted(loadPipeline);
watch([repositoryId, pipelineId], loadPipeline);

const goBack = useRouteBack({ name: 'repo' });
</script>
