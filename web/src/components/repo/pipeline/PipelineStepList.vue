<template>
  <aside class="pipeline-step-list">
    <div class="pipeline-source-card">
      <h3 class="pipeline-section-label">{{ $t('repo.pipeline.source_title') }}</h3>
      <div class="pipeline-source-author">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="w-6 rounded-md" :src="pipeline.author_avatar" />
        </div>
        <span>{{ pipeline.author }}</span>
      </div>
      <a
        v-if="
          // eslint-disable vue/html-indent
          pipeline.event === 'pull_request' ||
          pipeline.event === 'pull_request_closed' ||
          pipeline.event === 'pull_request_metadata'
          // eslint-enable vue/html-indent
        "
        class="pipeline-source-link"
        :href="pipeline.forge_url"
      >
        <Icon name="pull-request" />
        <span class="truncate">{{ prettyRef }}</span>
      </a>
      <router-link
        v-else-if="pipeline.event === 'push' || pipeline.event === 'manual' || pipeline.event === 'deployment'"
        class="pipeline-source-link"
        :to="{ name: 'repo-branch', params: { branch: prettyRef } }"
      >
        <Icon v-if="pipeline.event === 'manual'" name="manual-pipeline" />
        <Icon v-else-if="pipeline.event === 'push'" name="branch" />
        <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
        <span class="truncate">{{ prettyRef }}</span>
      </router-link>
      <div v-else class="pipeline-source-link">
        <Icon v-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />

        <span class="truncate">{{ prettyRef }}</span>
      </div>
      <div class="pipeline-source-link">
        <template v-if="pipeline.event === 'pull_request'">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </template>
        <a v-else class="pipeline-source-link" :href="pipeline.forge_url" target="_blank">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <Panel v-if="pipeline.workflows === undefined || pipeline.workflows.length === 0" class="pipeline-empty-steps">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </Panel>

    <h3 v-if="pipeline.workflows?.length" class="pipeline-section-label pipeline-steps-title">
      {{ $t('repo.pipeline.steps_title') }}
    </h3>
    <div class="pipeline-workflows">
      <div class="pipeline-workflows-scroll">
        <div v-for="workflow in pipeline.workflows" :key="workflow.id" class="pipeline-workflow">
          <div class="flex flex-col gap-2">
            <div v-if="workflow.environ" class="flex flex-wrap justify-end gap-x-1 gap-y-2 pt-1 pr-1 text-xs">
              <div v-for="(value, key) in workflow.environ" :key="key">
                <Badge :label="key" :value="value" />
              </div>
            </div>
            <button
              v-if="!singleConfig"
              type="button"
              :title="workflow.name"
              class="pipeline-workflow-heading"
              :aria-expanded="!workflowsCollapsed[workflow.id]"
              @click="workflowsCollapsed[workflow.id] = !workflowsCollapsed[workflow.id]"
            >
              <Icon
                name="chevron-right"
                class="h-6 min-w-6 transition-transform duration-150"
                :class="{ 'rotate-90 transform': !workflowsCollapsed[workflow.id] }"
              />
              <PipelineStatusIcon :status="workflow.state" class="h-4! w-4!" />
              <span class="truncate">{{ workflow.name }}</span>
              <PipelineStepDuration
                v-if="workflow.started !== workflow.finished"
                :workflow="workflow"
                class="pr-2px mr-1"
              />
            </button>
          </div>
          <div
            class="pipeline-workflow-steps transition-height overflow-hidden duration-150"
            :class="{ 'max-h-0': workflowsCollapsed[workflow.id] }"
          >
            <button
              v-for="step in workflow.children"
              ref="steps"
              :key="step.pid"
              :data-step-id="step.pid"
              type="button"
              :title="step.name"
              class="pipeline-step-button"
              :aria-pressed="selectedStepId === step.pid"
              :class="{
                'is-selected': selectedStepId && selectedStepId === step.pid,
                'mt-1': !singleConfig || step.pid !== workflow.children?.[0]?.pid,
              }"
              @click="$emit('update:selectedStepId', step.pid)"
            >
              <PipelineStatusIcon :service="step.type === StepType.Service" :status="step.state" class="h-4! w-4!" />
              <span class="truncate">{{ step.name }}</span>
              <PipelineStepDuration :step="step" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, toRef, useTemplateRef, watch } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import PipelineStepDuration from '~/components/repo/pipeline/PipelineStepDuration.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import usePipeline from '~/compositions/usePipeline';
import { StepType } from '~/lib/api/types';
import type { Pipeline, PipelineStep } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
  selectedStepId?: number | null;
}>();

defineEmits<{
  (event: 'update:selectedStepId', selectedStepId: number): void;
}>();

const pipeline = toRef(props, 'pipeline');
const selectedStepId = toRef(props, 'selectedStepId');
const { prettyRef } = usePipeline(pipeline);
const pipelineConfigs = requiredInject('pipeline-configs');

const workflowsCollapsed = ref<Record<PipelineStep['id'], boolean>>(
  pipeline.value.workflows && pipeline.value.workflows.length > 1
    ? (pipeline.value.workflows || []).reduce(
        (collapsed, workflow) => ({
          ...collapsed,
          [workflow.id]:
            ['success', 'skipped', 'blocked'].includes(workflow.state) &&
            !(workflow.children ?? []).some((child) => child.pid === selectedStepId.value),
        }),
        {},
      )
    : {},
);

const singleConfig = computed(
  () => pipelineConfigs?.value?.length === 1 && pipeline.value.workflows && pipeline.value.workflows.length === 1,
);

const steps = useTemplateRef('steps');
watch(selectedStepId, async (newSelectedStepId, oldSelectedStepId) => {
  if (!oldSelectedStepId && newSelectedStepId) {
    await nextTick();
    const step = steps.value?.find((s) => s.dataset.stepId === newSelectedStepId.toString());
    if (step) {
      step.scrollIntoView({
        behavior: 'auto',
        block: 'start',
      });
    }
  }
});
</script>
