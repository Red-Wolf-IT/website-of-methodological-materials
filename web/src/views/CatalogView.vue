<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { fetchManuals } from '@/api/manuals'
import { fetchTags } from '@/api/tags'
import type { Manual, Tag } from '@/types'
import ManualCard from '@/components/ManualCard.vue'
import Pagination from '@/components/Pagination.vue'

const route = useRoute()

const manuals = ref<Manual[]>([])
const tags = ref<Tag[]>([])
const total = ref(0)
const page = ref(1)
const limit = 9
const isLoading = ref(true)
const error = ref('')

const searchQuery = ref('')
const authorQuery = ref('')
const selectedTagId = ref<number | ''>('')
const sort = ref<'popular' | 'newest'>('newest')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / limit)))

async function loadCatalog() {
  isLoading.value = true
  error.value = ''
  try {
    const result = await fetchManuals({
      q: searchQuery.value || undefined,
      author: authorQuery.value || undefined,
      tag_id: selectedTagId.value || undefined,
      sort: sort.value === 'popular' ? 'popular' : undefined,
      page: page.value,
      limit,
    })
    manuals.value = result.items
    total.value = result.total
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Не удалось загрузить каталог'
  } finally {
    isLoading.value = false
  }
}

function applyFilters() {
  page.value = 1
  loadCatalog()
}

function resetFilters() {
  searchQuery.value = ''
  authorQuery.value = ''
  selectedTagId.value = ''
  sort.value = 'newest'
  page.value = 1
  loadCatalog()
}

function onPageChange(newPage: number) {
  page.value = newPage
  loadCatalog()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(async () => {
  try {
    tags.value = await fetchTags()
  } catch {
    // теги опциональны для фильтра
  }
  await loadCatalog()
})

watch(
  () => route.query.refresh,
  () => loadCatalog(),
)
</script>

<template>
  <div class="space-y-8">
    <section class="rounded-2xl bg-gradient-to-br from-brand-600 to-brand-900 px-6 py-10 text-white shadow-lg sm:px-10">
      <h1 class="mb-2 text-3xl font-bold tracking-tight sm:text-4xl">Каталог методических материалов</h1>
      <p class="max-w-2xl text-brand-100">
        Учебные пособия, руководства и статьи. Ищите по названию, автору или тегу.
      </p>
    </section>

    <section class="card p-5 sm:p-6">
      <form class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4" @submit.prevent="applyFilters">
        <div class="sm:col-span-2 lg:col-span-1">
          <label class="mb-1.5 block text-sm font-medium text-slate-700">Поиск</label>
          <input
            v-model="searchQuery"
            type="search"
            class="input-field"
            placeholder="Название или содержание..."
          />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">Автор</label>
          <input v-model="authorQuery" type="text" class="input-field" placeholder="Фамилия автора" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">Тег</label>
          <select v-model="selectedTagId" class="input-field">
            <option value="">Все теги</option>
            <option v-for="tag in tags" :key="tag.id" :value="tag.id">{{ tag.name }}</option>
          </select>
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-slate-700">Сортировка</label>
          <select v-model="sort" class="input-field">
            <option value="newest">Сначала новые</option>
            <option value="popular">По популярности</option>
          </select>
        </div>
        <div class="flex flex-wrap items-end gap-2 sm:col-span-2 lg:col-span-4">
          <button type="submit" class="btn-primary">Найти</button>
          <button type="button" class="btn-secondary" @click="resetFilters">Сбросить</button>
        </div>
      </form>
    </section>

    <div v-if="isLoading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div v-for="n in 6" :key="n" class="card h-48 animate-pulse bg-slate-100" />
    </div>

    <div v-else-if="error" class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-red-700">
      {{ error }}
      <p class="mt-1 text-sm">Убедитесь, что backend запущен: <code class="rounded bg-red-100 px-1">go run ./cmd/app</code></p>
    </div>

    <div v-else-if="manuals.length === 0" class="card px-6 py-16 text-center">
      <p class="text-lg font-medium text-slate-700">Ничего не найдено</p>
      <p class="mt-1 text-slate-500">Попробуйте изменить фильтры или добавьте новый материал.</p>
      <RouterLink to="/create" class="btn-primary mt-4 inline-flex">Добавить материал</RouterLink>
    </div>

    <template v-else>
      <p class="text-sm text-slate-500">Найдено: {{ total }}</p>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <ManualCard v-for="manual in manuals" :key="manual.id" :manual="manual" />
      </div>
      <Pagination :page="page" :total-pages="totalPages" @change="onPageChange" />
    </template>
  </div>
</template>
