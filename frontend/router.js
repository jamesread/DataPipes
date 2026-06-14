import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './views/HomeView.vue'
import ConnectionsView from './views/ConnectionsView.vue'
import ConnectionDetailView from './views/ConnectionDetailView.vue'
import TransformationsView from './views/TransformationsView.vue'

const routes = [
	{
		path: '/',
		name: 'home',
		component: HomeView,
	},
	{
		path: '/transformations',
		name: 'transformations',
		component: TransformationsView,
	},
	{
		path: '/connections',
		name: 'connections',
		component: ConnectionsView,
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
