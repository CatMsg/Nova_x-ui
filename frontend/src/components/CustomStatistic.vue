<script setup>
defineProps({
  title: { type: String, default: '' },
  value: { type: [String, Number], default: '' },
});
</script>

<template>
  <a-statistic :title="title" :value="value">
    <template #prefix>
      <slot name="prefix" />
    </template>
    <template #suffix>
      <slot name="suffix" />
    </template>
  </a-statistic>
</template>

<style scoped>
:deep(.ant-statistic) {
  display: inline-flex;
  flex-direction: column;
  gap: 4px;
}

:deep(.ant-statistic-title) {
  margin-bottom: 0;
  color: rgba(100, 116, 139, 0.92);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

:deep(.ant-statistic-content) {
  font-size: 17px;
  font-weight: 600;
  line-height: 1.3;
}

:deep(.ant-statistic-content-prefix) {
  margin-inline-end: 8px;
  color: #1677ff;
}

:global(body.dark .ant-statistic-content) {
  color: var(--dark-color-text-primary);
}

:global(body.dark .ant-statistic-title) {
  color: rgba(255, 255, 255, 0.55);
}
</style>
