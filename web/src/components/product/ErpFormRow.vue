<script setup lang="ts">
defineProps<{
  label: string
  required?: boolean
  hint?: string
  inline?: boolean
}>()
</script>

<template>
  <div class="erp-form-row">
    <div class="erp-form-label" :class="{ required }">{{ label }}：</div>
    <div class="erp-form-control">
      <p v-if="hint" class="erp-hint">{{ hint }}</p>
      <slot />
    </div>
  </div>
</template>

<style scoped>
.erp-form-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.erp-form-label {
  width: 120px;
  flex-shrink: 0;
  padding-top: 6px;
  font-size: 15px;
  color: rgba(0, 0, 0, 0.85);
  text-align: right;
  padding-right: 4px;
  line-height: 22px;
  white-space: nowrap;
}

.erp-form-label.required::before {
  content: '*';
  color: #ff4d4f;
  margin-right: 2px;
}

.erp-form-control {
  flex: 1;
  min-width: 0;
}

.erp-hint {
  margin: 0 0 8px;
  font-size: 13px;
  color: rgba(0, 0, 0, 0.45);
  line-height: 1.5;
}
</style>
