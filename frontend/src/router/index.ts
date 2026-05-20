import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import AdminProjects from '../views/AdminProjects.vue'
import AdminReport from '../views/AdminReport.vue'
import TimesheetEntry from '../views/TimesheetEntry.vue'
import { useAuthStore } from '../store/auth'
import AdminUsers from '../views/AdminUsers.vue'
import AccountSecurity from '../views/AccountSecurity.vue'
import AdminTimesheetBackfill from '../views/AdminTimesheetBackfill.vue'
import AdminTimesheetBackfillHistory from '../views/AdminTimesheetBackfillHistory.vue'
import Force2FASetup from '../views/Force2FASetup.vue'
import Login2FA from '../views/Login2FA.vue'
import AIReport from '../views/AIReport.vue'
import WorkPlan from '../views/WorkPlan.vue'
import AdminWorkSchedule from '../views/AdminWorkSchedule.vue'

const router = createRouter({
  history: createWebHistory('/worklog/'),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: Login },
    { path: '/force-2fa-setup', component: Force2FASetup },
    { path: '/login-2fa', component: Login2FA },
    { path: '/admin/projects', component: AdminProjects, meta: { requiresAuth: true } },
    { path: '/admin/report', component: AdminReport, meta: { requiresAuth: true } },
    { path: '/timesheet', component: TimesheetEntry, meta: { requiresAuth: true } },
    { path: '/admin/users', component: AdminUsers, meta: { requiresAuth: true } },
    { path: '/account/security', component: AccountSecurity, meta: { requiresAuth: true } },
    { path: '/admin/timesheet-backfill', component: AdminTimesheetBackfill, meta: { requiresAuth: true } },
    {
      path: '/admin/timesheet-backfill-history',
      component: AdminTimesheetBackfillHistory,
      meta: { requiresAuth: true }
    },
    { path: '/ai/report', component: AIReport, meta: { requiresAuth: true } },
    { path: '/work-plan', component: WorkPlan, meta: { requiresAuth: true } },
    { path: '/work-schedule', component: AdminWorkSchedule, meta: { requiresAuth: true } },
    { path: '/admin/work-schedule', redirect: '/work-schedule' },
  ]
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  const noAuthRoutes = ['/login', '/force-2fa-setup', '/login-2fa']

  if (noAuthRoutes.includes(to.path)) {
    return next()
  }

  if (!auth.token) {
    return next('/login')
  }

  if (to.path.startsWith('/admin') && auth.role !== 'admin') {
    return next('/timesheet')
  }

  next()
})

export default router
