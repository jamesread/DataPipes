<template>
	<Section
		subtitle = "Configured extract and load connections"
		classes = "connections-list settings-list"
		:padding = "false"
	>
		<template #title>
			<span class = "section-title-with-icon">
				<HugeiconsIcon
					:icon = "ConnectIcon"
					width = "22"
					height = "22"
					aria-hidden = "true"
				/>
				Connections
			</span>
		</template>

		<template #toolbar>
			<button
				type = "button"
				class = "neutral"
				title = "Refresh"
				aria-label = "Refresh"
				:disabled = "loading"
				@click = "loadConnections"
			>
				<HugeiconsIcon
					:icon = "RefreshIcon"
					width = "1em"
					height = "1em"
					aria-hidden = "true"
				/>
			</button>
		</template>

		<NotificationBlock
			v-if = "configError"
			type = "critical"
			role = "alert"
			:message = "configErrorMessage"
			class = "list-banner-pad"
		/>

		<NotificationBlock
			v-else-if = "loadError"
			type = "critical"
			role = "alert"
			:message = "loadError"
			class = "list-banner-pad"
		/>

		<div v-else-if = "loading && !connections.length" class = "list-banner-pad muted">Loading connections…</div>

		<template v-else>
			<NotificationBlock
				v-if = "!connections.length"
				type = "note"
				message = "No connections configured."
				class = "list-banner-pad"
			/>

			<Table
				v-else
				class = "list-table-wrap"
				row-clickable
				:headers = "tableHeaders"
				:data = "tableRows"
				@row-click = "openConnection"
			>
				<template #cell-name = "{ value }">
					<strong>{{ value }}</strong>
				</template>
				<template #cell-health = "{ row }">
					<span class = "tag" :class = "connectionHealthTagClass(row)">{{ row.health }}</span>
					<span v-if = "row.healthMessage" class = "subtle connection-health-message">{{ row.healthMessage }}</span>
				</template>
			</Table>
		</template>
	</Section>
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import { useRouter } from 'vue-router'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ConnectIcon, RefreshIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'
	import Table from 'picocrank/vue/components/Table.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { getApiClient } from '../api-client.js'
	import {
		connectionHealthMessage,
		connectionHealthStatusLabel,
		connectionHealthTagClass,
		formatConnectionDetails,
	} from '../connection-format.js'

	const router = useRouter()
	const connections = ref([])
	const loading = ref(false)
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')

	const configErrorMessage = computed(() => {
		const prefix = configPath.value ? ` from ${configPath.value}` : ''
		return `Could not load config${prefix}: ${configError.value}`
	})

	const tableHeaders = [
		{ key: 'name', label: 'Name', sortable: true },
		{ key: 'type', label: 'Type', sortable: true },
		{ key: 'details', label: 'Details', sortable: true },
		{ key: 'health', label: 'Health', sortable: true },
	]

	const tableRows = computed(() => (
		connections.value.map((conn) => ({
			name: conn.id,
			type: conn.type,
			details: formatConnectionDetails(conn),
			health: connectionHealthStatusLabel(conn),
			healthMessage: connectionHealthMessage(conn),
			healthOk: conn.healthOk,
			id: conn.id,
		}))
	))

	function openConnection ({ row }) {
		router.push({
			name: 'connection-detail',
			params: { id: row.id },
		})
	}

	async function loadConnections () {
		loading.value = true
		configError.value = ''
		configPath.value = ''
		loadError.value = ''

		try {
			const res = await getApiClient().listConnections({})
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				connections.value = []
				return
			}
			connections.value = res.connections ?? []
		} catch (error) {
			loadError.value = 'Failed to load connections: ' + error.message
			connections.value = []
		} finally {
			loading.value = false
		}
	}

	onMounted(loadConnections)
</script>

<style scoped>
.connection-health-message {
	margin-left: 0.5em;
}
</style>
