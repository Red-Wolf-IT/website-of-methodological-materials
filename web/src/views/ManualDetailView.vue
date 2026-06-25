<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchManual } from '@/api/manuals'
import { fileUrl, formatDate } from '@/api/client'
import type { Manual } from '@/types'
import TagBadge from '@/components/TagBadge.vue'

const props = defineProps<{
  id: string
}>()

const router = useRouter()
const manual = ref<Manual | null>(null)
const isLoading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    manual.value = await fetchManual(props.id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Материал не найден'
  } finally {
    isLoading.value = false
  }
})

function downloadUrl(path: string) {
  return fileUrl(path)
}
</script>

<template>
  <div>
    <button type="button" class="btn-secondary mb-6" @click="router.push('/')">
      ← Назад к каталогу
    </button>

    <div v-if="isLoading" class="card h-64 animate-pulse bg-slate-100" />

    <div v-else-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
      {{ error }}
    </div>

    <article v-else-if="manual" class="card overflow-hidden">
      <div class="border-b border-slate-100 bg-slate-50 px-6 py-8 sm:px-8">
        <div class="mb-4 flex flex-wrap items-center gap-2">
          <TagBadge v-for="tag in manual.tags" :key="tag.id" :name="tag.name" />
        </div>
        <h1 class="mb-3 text-3xl font-bold text-slate-900">{{ manual.title }}</h1>
        <div class="flex flex-wrap gap-4 text-sm text-slate-600">
          <span>✍️ {{ manual.author }}</span>
          <span>👁 {{ manual.views_count }} просмотров</span>
          <span>📅 {{ formatDate(manual.created_at) }}</span>
        </div>
      </div>

      <div class="px-6 py-8 sm:px-8">
        <div class="prose prose-slate max-w-none whitespace-pre-wrap text-base leading-relaxed text-slate-700">
          {{ manual.content }}
        </div>

        <div v-if="manual.file_path" class="mt-8 rounded-xl border border-dashed border-slate-300 bg-slate-50 p-5">
          <p class="mb-3 text-sm font-medium text-slate-700">Прикреплённый файл</p>
          <a
            :href="downloadUrl(manual.file_path)"
            target="_blank"
            rel="noopener noreferrer"
            class="btn-primary inline-flex"
          >
            📥 Скачать вложение
          </a>
        </div>
      </div>
    </article>
  </div>
</template>
