<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <header class="sticky top-0 z-10 border-b border-slate-200 bg-white/90 backdrop-blur">
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
        <RouterLink to="/" class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-600 text-lg font-bold text-white">
            М
          </div>
          <div>
            <div class="text-base font-semibold text-slate-900">Методические материалы</div>
            <div class="text-xs text-slate-500">Каталог учебных пособий</div>
          </div>
        </RouterLink>

        <nav class="flex items-center gap-2">
          <RouterLink
            to="/"
            class="rounded-lg px-3 py-2 text-sm font-medium text-slate-600 transition hover:bg-slate-100 hover:text-slate-900"
            active-class="!bg-brand-50 !text-brand-700"
          >
            Каталог
          </RouterLink>
          <RouterLink to="/create" class="btn-primary">
            + Добавить материал
          </RouterLink>
        </nav>
      </div>
    </header>

    <main class="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6">
      <RouterView />
    </main>

    <footer class="border-t border-slate-200 bg-white py-6 text-center text-sm text-slate-500">
      Каталог методических материалов · Backend API на Go
    </footer>
  </div>
</template>
