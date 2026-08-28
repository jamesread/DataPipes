import { createRouter, createWebHistory } from 'vue-router'
import { ClipboardListIcon, ComputerSettingsIcon, ConnectIcon, DashboardSquareSettingIcon, ShuffleIcon } from '@hugeicons/core-free-icons'

import HomeView from './views/HomeView.vue'
import JobDetailView from './views/JobDetailView.vue'
import ConnectionsView from './views/ConnectionsView.vue'
import ConnectionDetailView from './views/ConnectionDetailView.vue'
import TransformationsView from './views/TransformationsView.vue'
import ControlPanel from './views/ControlPanel.vue'
import DiagnosticsView from './views/DiagnosticsView.vue'
import {
	canAccessControlPanelFromStatus,
	canAccessDiagnosticsFromStatus,
	currentStatus,
} from './rbacAccess.js'
import {
	connectionsBreadcrumbs,
	connectionDetailBreadcrumbs,
	controlPanelBreadcrumbs,
	diagnosticsBreadcrumbs,
	jobDetailBreadcrumbs,
	jobsBreadcrumbs,
	transformationsBreadcrumbs,
} from './routeBreadcrumbs.js'

const routes = [
	{
		path: '/',
		name: 'home',
		component: HomeView,
		meta: {
			title: 'Jobs',
			icon: ClipboardListIcon,
			breadcrumbs: jobsBreadcrumbs(),
		},
	},
	{
		path: '/jobs/:id',
		name: 'job-detail',
		component: JobDetailView,
		meta: {
			title: 'Job',
			icon: ClipboardListIcon,
			breadcrumbs: jobDetailBreadcrumbs,
		},
	},
	{
		path: '/transformations',
		name: 'transformations',
		component: TransformationsView,
		meta: {
			title: 'Transformations',
			icon: ShuffleIcon,
			breadcrumbs: transformationsBreadcrumbs(),
		},
	},
	{
		path: '/connections',
		name: 'connections',
		component: ConnectionsView,
		meta: {
			title: 'Connections',
			icon: ConnectIcon,
			breadcrumbs: connectionsBreadcrumbs(),
		},
	},
	{
		path: '/connections/:id',
		name: 'connection-detail',
		component: ConnectionDetailView,
		meta: {
			title: 'Connection',
			icon: ConnectIcon,
			breadcrumbs: connectionDetailBreadcrumbs,
		},
	},
	{
		path: '/control-panel',
		name: 'controlPanel',
		component: ControlPanel,
		meta: {
			title: 'Control Panel',
			icon: DashboardSquareSettingIcon,
			breadcrumbs: controlPanelBreadcrumbs(),
			requiresAuth: true,
			requiresControlPanel: true,
		},
	},
	{
		path: '/control-panel/diagnostics',
		name: 'diagnostics',
		component: DiagnosticsView,
		meta: {
			title: 'Diagnostics',
			icon: ComputerSettingsIcon,
			breadcrumbs: diagnosticsBreadcrumbs(),
			requiresAuth: true,
			requiresControlPanel: true,
			requiresDiagnostics: true,
		},
	},
]

const router = createRouter({
	history: createWebHistory(),
	routes,
})

router.beforeEach((to) => {
	const st = currentStatus()
	if (to.meta.requiresAuth && !st?.isLoggedIn) {
		return '/'
	}
	if (to.meta.requiresControlPanel && !canAccessControlPanelFromStatus(st)) {
		return '/'
	}
	if (to.meta.requiresDiagnostics && !canAccessDiagnosticsFromStatus(st)) {
		return '/'
	}
})

export default router
