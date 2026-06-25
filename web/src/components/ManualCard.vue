<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { Manual } from '@/types'
import { formatDate } from '@/api/client'
import TagBadge from './TagBadge.vue'

defineProps<{
  manual: Manual
}>()
</script>

<template>
  <RouterLink :to="`/manuals/${manual.id}`" class="card block p-5 hover:border-brand-200">
    <div class="mb-3 flex items-start justify-between gap-3">
      <h2 class="text-lg font-semibold text-slate-900 line-clamp-2">{{ manual.title }}</h2>
      <span class="shrink-0 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-600">
        👁 {{ manual.views_count }}
      </span>
    </div>

    <p class="mb-2 text-sm font-medium text-brand-700">{{ manual.author }}</p>
    <p class="mb-4 line-clamp-3 text-sm leading-relaxed text-slate-600">{{ manual.content }}</p>

    <div class="flex flex-wrap items-center justify-between gap-2">
      <div v-if="manual.tags?.length" class="flex flex-wrap gap-1.5">
        <TagBadge v-for="tag in manual.tags" :key="tag.id" :name="tag.name" />
      </div>
      <span v-if="manual.file_path" class="text-xs text-slate-500">📎 Есть вложение</span>
    </div>

    <p class="mt-3 text-xs text-slate-400">{{ formatDate(manual.created_at) }}</p>
  </RouterLink>
</template>
