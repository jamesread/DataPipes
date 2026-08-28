<template>
	<Section
		v-if = "canAccessControlPanel"
		title = "Control Panel"
		subtitle = "System functionality"
	>
		<template #toolbar>
			<button
				type = "button"
				class = "neutral"
				:disabled = "!clientReady"
				@click = "refresh"
			>Refresh</button>
		</template>
		<Navigation ref = "localNavigation">
			<NavigationGrid />
		</Navigation>
	</Section>
</template>

<script setup>
	import { computed, nextTick, onMounted, ref } from 'vue'
	import { useRouter } from 'vue-router'
	import Section from 'picocrank/vue/components/Section.vue'
	import Navigation from 'picocrank/vue/components/Navigation.vue'
	import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
	import { ComputerSettingsIcon } from '@hugeicons/core-free-icons'
	import {
		canAccessControlPanelFromStatus,
		canAccessDiagnosticsFromStatus,
		currentStatus,
	} from '../rbacAccess.js'

	const router = useRouter()
	const clientReady = ref(false)
	const status = ref(currentStatus())
	const localNavigation = ref(null)
	const canAccessControlPanel = computed(() => canAccessControlPanelFromStatus(status.value))

	async function refresh () {
		status.value = currentStatus()
		if (!canAccessControlPanelFromStatus(status.value)) {
			router.push('/')
			return
		}
		clientReady.value = true
		await nextTick()
		populateHubTiles()
	}

	function populateHubTiles () {
		const nav = localNavigation.value
		if (!nav) {
			return
		}
		nav.clearNavigationLinks()
		if (canAccessDiagnosticsFromStatus(status.value)) {
			nav.addCallback('Diagnostics', () => router.push({ name: 'diagnostics' }), {
				icon: ComputerSettingsIcon,
				name: 'diagnostics',
				description: 'Version, configuration path, and reload',
			})
		}
	}

	onMounted(refresh)
</script>
