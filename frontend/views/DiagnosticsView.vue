<template>
	<Section
		v-if = "canAccessDiagnostics"
		title = "Diagnostics"
		subtitle = "Version, configuration path, and reload"
	>
		<template #toolbar>
			<router-link :to = "{ name: 'controlPanel' }" class = "button inline-icon neutral">
				<span>Control Panel</span>
			</router-link>
			<button
				type = "button"
				class = "neutral"
				:disabled = "busy"
				@click = "load"
			>Refresh</button>
			<button
				type = "button"
				class = "neutral"
				:disabled = "busy"
				@click = "reloadConfig"
			>Reload config</button>
		</template>

		<NotificationBlock
			v-if = "loadError"
			type = "critical"
			role = "alert"
			:message = "loadError"
		/>

		<dl v-else-if = "loaded">
			<dt>Version</dt>
			<dd>{{ version || '—' }}</dd>
			<dt>Config path</dt>
			<dd>{{ configPath || '—' }}</dd>
			<dt>Status</dt>
			<dd>{{ ok ? 'OK' : 'Configuration error' }}</dd>
		</dl>

		<NotificationBlock
			v-if = "loaded && errors.length"
			type = "critical"
			role = "alert"
		>
			Server configuration error{{ configPath ? ' (' + configPath + ')' : '' }}:
			<ul>
				<li v-for = "(err, i) in errors" :key = "i">{{ err }}</li>
			</ul>
		</NotificationBlock>
	</Section>
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import { useRouter } from 'vue-router'
	import Section from 'picocrank/vue/components/Section.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { canAccessDiagnosticsFromStatus, currentStatus } from '../rbacAccess.js'
	import { formatRpcError, getApiClient } from '../api-client.js'

	const router = useRouter()
	const status = ref(currentStatus())
	const canAccessDiagnostics = computed(() => canAccessDiagnosticsFromStatus(status.value))
	const busy = ref(false)
	const loaded = ref(false)
	const loadError = ref('')
	const version = ref('')
	const configPath = ref('')
	const ok = ref(true)
	const errors = ref([])

	async function load () {
		status.value = currentStatus()
		if (!canAccessDiagnosticsFromStatus(status.value)) {
			router.push('/')
			return
		}
		busy.value = true
		loadError.value = ''
		try {
			const res = await getApiClient().init({})
			version.value = res.version
			configPath.value = res.configPath
			ok.value = res.ok
			errors.value = res.errors ?? []
			loaded.value = true
		} catch (err) {
			loadError.value = formatRpcError(err)
		} finally {
			busy.value = false
		}
	}

	async function reloadConfig () {
		status.value = currentStatus()
		if (!canAccessDiagnosticsFromStatus(status.value)) {
			router.push('/')
			return
		}
		busy.value = true
		loadError.value = ''
		try {
			await getApiClient().reload({})
		} catch (err) {
			loadError.value = formatRpcError(err)
			busy.value = false
			return
		}
		await load()
	}

	onMounted(load)
</script>
