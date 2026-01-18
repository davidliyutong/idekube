import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { setupRouterGuards } from './guards'

// Layouts
const AuthLayout = () => import('@/layouts/AuthLayout.vue')
const AppLayout = () => import('@/layouts/AppLayout.vue')

// Auth pages
const LoginPage = () => import('@/pages/auth/LoginPage.vue')
const RegisterPage = () => import('@/pages/auth/RegisterPage.vue')
const ForgotPasswordPage = () => import('@/pages/auth/ForgotPasswordPage.vue')

// App pages
const DashboardPage = () => import('@/pages/dashboard/DashboardPage.vue')
const WorkspacesPage = () => import('@/pages/workspaces/WorkspacesPage.vue')
const WorkspaceDetailPage = () => import('@/pages/workspaces/WorkspaceDetailPage.vue')
const TemplatesPage = () => import('@/pages/templates/TemplatesPage.vue')
const TemplateDetailPage = () => import('@/pages/templates/TemplateDetailPage.vue')
const UsersPage = () => import('@/pages/users/UsersPage.vue')
const UserDetailPage = () => import('@/pages/users/UserDetailPage.vue')
const OrganizationsPage = () => import('@/pages/organizations/OrganizationsPage.vue')
const OrganizationDetailPage = () => import('@/pages/organizations/OrganizationDetailPage.vue')
const VolumesPage = () => import('@/pages/volumes/VolumesPage.vue')
const VolumeDetailPage = () => import('@/pages/volumes/VolumeDetailPage.vue')
const SettingsPage = () => import('@/pages/settings/SettingsPage.vue')
const ProfilePage = () => import('@/pages/profile/ProfilePage.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/auth',
    component: AuthLayout,
    children: [
      {
        path: 'login',
        name: 'login',
        component: LoginPage,
        meta: { requiresGuest: true },
      },
      {
        path: 'register',
        name: 'register',
        component: RegisterPage,
        meta: { requiresGuest: true },
      },
      {
        path: 'forgot-password',
        name: 'forgot-password',
        component: ForgotPasswordPage,
        meta: { requiresGuest: true },
      },
    ],
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: DashboardPage,
      },
      {
        path: 'workspaces',
        name: 'workspaces',
        component: WorkspacesPage,
      },
      {
        path: 'workspaces/:id',
        name: 'workspace-detail',
        component: WorkspaceDetailPage,
      },
      {
        path: 'templates',
        name: 'templates',
        component: TemplatesPage,
      },
      {
        path: 'templates/:id',
        name: 'template-detail',
        component: TemplateDetailPage,
      },
      {
        path: 'users',
        name: 'users',
        component: UsersPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'users/:id',
        name: 'user-detail',
        component: UserDetailPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'organizations',
        name: 'organizations',
        component: OrganizationsPage,
      },
      {
        path: 'organizations/:id',
        name: 'organization-detail',
        component: OrganizationDetailPage,
      },
      {
        path: 'volumes',
        name: 'volumes',
        component: VolumesPage,
      },
      {
        path: 'volumes/:id',
        name: 'volume-detail',
        component: VolumeDetailPage,
      },
      {
        path: 'settings',
        name: 'settings',
        component: SettingsPage,
        meta: { requiresAdmin: true },
      },
      {
        path: 'profile',
        name: 'profile',
        component: ProfilePage,
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard',
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

setupRouterGuards(router)

export default router
