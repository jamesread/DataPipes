import { createRouter, createWebHistory } from 'vue-router'
import { ClipboardListIcon, ConnectIcon, ShuffleIcon } from '@hugeicons/core-free-icons'

import HomeView from './views/HomeView.vue'
import ConnectionsView from './views/ConnectionsView.vue'
import ConnectionDetailView from './views/ConnectionDetailView.vue'
import TransformationsView from './views/TransformationsView.vue'

const routes = [
	{
		path: '/',
		name: 'home',
		component: HomeView,
		meta: { title: 'Jobs', icon: ClipboardListIcon },
	},
	{
		path: '/transformations',
		name: 'transformations',
		component: TransformationsView,
		meta: { title: 'Transformations', icon: ShuffleIcon },
	},
	{
		path: '/connections',
		name: 'connections',
		component: ConnectionsView,
		meta: { title: 'Connections', icon: ConnectIcon },
	},
	{
		path: '/connections/:id',
		name: 'connection-detail',
		component: ConnectionDetailView,
	},
]

const router = createRouter({
	history: createWebHistory(),
	routes,
})

export default router
