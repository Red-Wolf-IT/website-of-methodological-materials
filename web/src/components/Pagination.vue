<script setup lang="ts">
const props = defineProps<{
  page: number
  totalPages: number
}>()

const emit = defineEmits<{
  change: [page: number]
}>()

function goTo(page: number) {
  if (page >= 1 && page <= props.totalPages) emit('change', page)
}
</script>

<template>
  <div v-if="totalPages > 1" class="flex items-center justify-center gap-2">
    <button type="button" class="btn-secondary !px-3 !py-2" :disabled="page <= 1" @click="goTo(page - 1)">
      ← Назад
    </button>
    <span class="text-sm text-slate-600">Страница {{ page }} из {{ totalPages }}</span>
    <button type="button" class="btn-secondary !px-3 !py-2" :disabled="page >= totalPages" @click="goTo(page + 1)">
      Вперёд →
    </button>
  </div>
</template>
