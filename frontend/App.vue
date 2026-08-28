<template>
	<Navigation ref = "navigation">
		<Navigation ref = "topBarNavigation">
			<Header
				title = "DataPipes"
				:logo-url = "logoUrl"
				:sidebar-enabled = "false"
				:top-bar-enabled = "true"
				:theme-toggle-enabled = "true"
				:breadcrumbs = "false"
				:navigation = "navigation"
				:top-bar-navigation = "topBarNavigation"
				@logo-click = "goHome"
			/>

			<NotificationBlock
				v-if = "initError"
				type = "critical"
				role = "alert"
				:message = "initErrorMessage"
			/>
			<NotificationBlock
				v-else-if = "initReady && initErrors.length"
				type = "critical"
				role = "alert"
			>
				Server configuration error{{ configPath ? ' (' + configPath + ')' : '' }}:
				<ul>
					<li v-for = "(err, i) in initErrors" :key = "i">{{ err }}</li>
				</ul>
			</NotificationBlock>

			<main>
				<NotificationBlock v-if = "!initReady" type = "info" message = "Loading…" />
				<router-view v-else-if = "initOk" />
			</main>

			<footer>
				<span>
					<a href = "https://github.com/jamesread/data-cleaner" target = "_blank" rel = "noopener noreferrer">
						DataPipes on GitHub
					</a>
				</span>
				<span v-if = "version"> · v{{ version }}</span>
			</footer>
		</Navigation>
	</Navigation>
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import { useRouter } from 'vue-router'
	import Header from 'picocrank/vue/components/Header.vue'
	import Navigation from 'picocrank/vue/components/Navigation.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { getApiClient, formatRpcError } from './api-client.js'
	import { canAccessControlPanelFromStatus, currentStatus } from './rbacAccess.js'
	import { setupSidebarNavigation } from './sidebarNavigation.js'
	import logoUrl from './logo.png'

	const router = useRouter()
	const navigation = ref(null)
	const topBarNavigation = ref(null)
	const version = ref('')
	const configPath = ref('')
	const initReady = ref(false)
	const initError = ref('')
	const initErrors = ref([])

	const initOk = computed(() => !initError.value && initErrors.value.length === 0)

	const initErrorMessage = computed(() => `Could not reach server: ${initError.value}`)

	onMounted(async () => {
		try {
			const res = await getApiClient().init({})
			version.value = res.version
			configPath.value = res.configPath
			if (!res.ok) {
				initErrors.value = res.errors?.length ? res.errors : ['unknown configuration error']
			}
		} catch (err) {
			initError.value = formatRpcError(err)
		} finally {
			initReady.value = true
			setupNavigation()
		}
	})

	function setupNavigation () {
		const showControlPanel = canAccessControlPanelFromStatus(currentStatus())
		for (const nav of [topBarNavigation.value, navigation.value]) {
			if (!nav) {
				continue
			}
			setupSidebarNavigation(nav, { showControlPanel })
		}
	}

	function goHome () {
		router.push({ name: 'home' })
	}
</script>
