<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { attachTags, createManual } from '@/api/manuals'
import { createTag, fetchTags } from '@/api/tags'
import { ApiClientError } from '@/api/client'
import type { Tag, ValidationFieldError } from '@/types'

const router = useRouter()

const title = ref('')
const author = ref('')
const content = ref('')
const selectedTagIds = ref<number[]>([])
const tags = ref<Tag[]>([])
const newTagName = ref('')

const isSubmitting = ref(false)
const error = ref('')
const fieldErrors = ref<ValidationFieldError[]>([])
const success = ref('')

onMounted(async () => {
  try {
    tags.value = await fetchTags()
  } catch {
    // можно создать материал без тегов
  }
})

function toggleTag(tagId: number) {
  const idx = selectedTagIds.value.indexOf(tagId)
  if (idx === -1) selectedTagIds.value.push(tagId)
  else selectedTagIds.value.splice(idx, 1)
}

async function addNewTag() {
  const name = newTagName.value.trim()
  if (!name) return
  try {
    const tag = await createTag(name)
    tags.value = [...tags.value, tag].sort((a, b) => a.name.localeCompare(b.name))
    selectedTagIds.value.push(tag.id)
    newTagName.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Не удалось создать тег'
  }
}

async function onSubmit() {
  isSubmitting.value = true
  error.value = ''
  fieldErrors.value = []
  success.value = ''

  try {
    const manual = await createManual({
      title: title.value.trim(),
      author: author.value.trim(),
      content: content.value.trim(),
    })

    if (selectedTagIds.value.length > 0) {
      await attachTags(manual.id, selectedTagIds.value)
    }

    success.value = 'Материал успешно добавлен!'
    setTimeout(() => router.push({ path: `/manuals/${manual.id}` }), 800)
  } catch (e) {
    if (e instanceof ApiClientError) {
      error.value = e.message
      fieldErrors.value = e.fields ?? []
    } else {
      error.value = e instanceof Error ? e.message : 'Ошибка при сохранении'
    }
  } finally {
    isSubmitting.value = false
  }
}

function fieldError(name: string): string | undefined {
  return fieldErrors.value.find((f) => f.field === name)?.message
}
</script>

<template>
  <div class="mx-auto max-w-2xl">
    <h1 class="mb-2 text-2xl font-bold text-slate-900">Добавить материал</h1>
    <p class="mb-8 text-slate-600">Заполните форму — материал появится в каталоге для всех пользователей.</p>

    <form class="card space-y-5 p-6 sm:p-8" @submit.prevent="onSubmit">
      <div v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {{ error }}
      </div>
      <div v-if="success" class="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
        {{ success }}
      </div>

      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700" for="title">Название *</label>
        <input id="title" v-model="title" type="text" class="input-field" :class="{ 'border-red-400': fieldError('title') }" />
        <p v-if="fieldError('title')" class="mt-1 text-xs text-red-600">{{ fieldError('title') }}</p>
      </div>

      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700" for="author">Автор *</label>
        <input id="author" v-model="author" type="text" class="input-field" :class="{ 'border-red-400': fieldError('author') }" />
        <p v-if="fieldError('author')" class="mt-1 text-xs text-red-600">{{ fieldError('author') }}</p>
      </div>

      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700" for="content">Содержание *</label>
        <textarea
          id="content"
          v-model="content"
          rows="10"
          class="input-field resize-y"
          :class="{ 'border-red-400': fieldError('content') }"
        />
        <p v-if="fieldError('content')" class="mt-1 text-xs text-red-600">{{ fieldError('content') }}</p>
      </div>

      <div v-if="tags.length">
        <label class="mb-2 block text-sm font-medium text-slate-700">Теги</label>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="tag in tags"
            :key="tag.id"
            type="button"
            class="rounded-full px-3 py-1.5 text-sm font-medium ring-1 ring-inset transition"
            :class="
              selectedTagIds.includes(tag.id)
                ? 'bg-brand-600 text-white ring-brand-600'
                : 'bg-white text-slate-700 ring-slate-300 hover:bg-slate-50'
            "
            @click="toggleTag(tag.id)"
          >
            {{ tag.name }}
          </button>
        </div>
      </div>

      <div>
        <label class="mb-1.5 block text-sm font-medium text-slate-700" for="newTag">Новый тег</label>
        <div class="flex gap-2">
          <input id="newTag" v-model="newTagName" type="text" class="input-field" placeholder="например, docker" />
          <button type="button" class="btn-secondary shrink-0" @click="addNewTag">Добавить</button>
        </div>
      </div>

      <div class="flex gap-3 pt-2">
        <button type="submit" class="btn-primary" :disabled="isSubmitting">
          {{ isSubmitting ? 'Сохранение...' : 'Опубликовать' }}
        </button>
        <RouterLink to="/" class="btn-secondary">Отмена</RouterLink>
      </div>
    </form>
  </div>
</template>
