<template>
	<Navigation ref = "navigation">
		<Header
			title = "DataPipes"
			:logo-url = "logoUrl"
			:sidebar-enabled = "false"
			:top-bar-enabled = "true"
			:theme-toggle-enabled = "true"
			:navigation = "navigation"
			@logo-click = "goHome"
		/>

		<div v-if = "initError" class = "config-error" role = "alert">
			Could not reach server: {{ initError }}
		</div>
		<div v-else-if = "initReady && initErrors.length" class = "config-error" role = "alert">
			<p>Server configuration error{{ configPath ? ' (' + configPath + ')' : '' }}:</p>
			<ul>
				<li v-for = "(err, i) in initErrors" :key = "i">{{ err }}</li>
			</ul>
		</div>

		<main>
			<p v-if = "!initReady" class = "inline-notification">Loading…</p>
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
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import Header from 'picocrank/vue/components/Header.vue'
	import Navigation from 'picocrank/vue/components/Navigation.vue'
	import { getApiClient, formatRpcError } from './api-client.js'
	import logoUrl from './logo.png'

	const navigation = ref(null)
	const version = ref('')
	const configPath = ref('')
	const initReady = ref(false)
	const initError = ref('')
	const initErrors = ref([])

	const initOk = computed(() => !initError.value && initErrors.value.length === 0)

	onMounted(async () => {
		setupNavigation()

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
		}
	})

	function setupNavigation () {
		if (!navigation.value) {
			return
		}
		navigation.value.clearNavigationLinks()
		navigation.value.addRouterLink('home')
		navigation.value.addRouterLink('connections')
		navigation.value.addRouterLink('transformations')
	}

	function goHome () {
		window.location.href = '/'
	}
</script>
