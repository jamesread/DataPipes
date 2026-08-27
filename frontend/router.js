import { createRouter, createWebHistory } from 'vue-router'
import { ClipboardListIcon, ConnectIcon, ShuffleIcon } from '@hugeicons/core-free-icons'

import HomeView from './views/HomeView.vue'
import JobDetailView from './views/JobDetailView.vue'
import ConnectionsView from './views/ConnectionsView.vue'
import ConnectionDetailView from './views/ConnectionDetailView.vue'
import TransformationsView from './views/TransformationsView.vue'
import {
	connectionsBreadcrumbs,
	connectionDetailBreadcrumbs,
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
]

const router = createRouter({
	history: createWebHistory(),
	routes,
})

export default router
