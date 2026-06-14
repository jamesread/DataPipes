<template>
	<Header
		title = "DataPipes"
		:logo-url = "logoUrl"
		:sidebar-enabled = "false"
		@logo-click = "goHome"
	>
		<template #toolbar>
			<nav>
				<ul>
					<li><router-link to = "/">Jobs</router-link></li>
					<li><router-link to = "/connections">Connections</router-link></li>
					<li><router-link to = "/transformations">Transformations</router-link></li>
				</ul>
			</nav>
		</template>
	</Header>

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
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import Header from 'picocrank/vue/components/Header.vue'
	import { getApiClient, formatRpcError } from './api-client.js'
	import logoUrl from './logo.png'

	const version = ref('')
	const configPath = ref('')
	const initReady = ref(false)
	const initError = ref('')
	const initErrors = ref([])

	const initOk = computed(() => !initError.value && initErrors.value.length === 0)

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
		}
	})

	function goHome () {
		window.location.href = '/'
	}
</script>

<style scoped>
nav ul {
	display: flex;
	gap: 0.75rem;
	list-style: none;
	margin: 0;
	padding: 0;
}

nav a {
	color: inherit;
	text-decoration: none;
}

nav a.router-link-active {
	text-decoration: underline;
	font-weight: 500;
}
</style>
