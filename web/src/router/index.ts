import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'catalog',
      component: () => import('@/views/CatalogView.vue'),
    },
    {
      path: '/manuals/:id',
      name: 'manual',
      component: () => import('@/views/ManualDetailView.vue'),
      props: true,
    },
    {
      path: '/create',
      name: 'create',
      component: () => import('@/views/CreateManualView.vue'),
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

export default router
