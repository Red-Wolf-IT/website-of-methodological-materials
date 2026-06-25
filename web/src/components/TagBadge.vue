<script setup lang="ts">
defineProps<{
  name: string
}>()
</script>

<template>
  <span class="inline-flex items-center rounded-full bg-brand-50 px-2.5 py-0.5 text-xs font-medium text-brand-700 ring-1 ring-inset ring-brand-200">
    {{ name }}
  </span>
</template>
