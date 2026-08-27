<template>
	<Section subtitle = "Connection details">
		<template #title>
			<span class = "section-title-with-icon">
				<HugeiconsIcon
					:icon = "ConnectIcon"
					width = "22"
					height = "22"
					aria-hidden = "true"
				/>
				{{ connectionId }}
			</span>
		</template>
		<div v-if = "configError" class = "list-banner-pad">
			<NotificationBlock type = "critical" role = "alert" :message = "configErrorMessage" />
		</div>

		<NotificationBlock v-else-if = "loadError" type = "critical" role = "alert" :message = "loadError" />

		<NotificationBlock v-else-if = "notFound" type = "warning" role = "alert" :message = "notFound" />

		<dl v-else-if = "connection">
			<template v-for = "row in detailRows" :key = "row.label">
				<dt>{{ row.label }}</dt>
				<dd v-if = "row.health">
					<span class = "tag" :class = "connectionHealthTagClass(connection)">{{ connectionHealthStatusLabel(connection) }}</span>
					<span v-if = "connectionHealthMessage(connection)" class = "subtle connection-health-message">{{ connectionHealthMessage(connection) }}</span>
				</dd>
				<dd v-else>{{ row.value }}</dd>
			</template>
		</dl>
	</Section>
</template>

<script setup>
	import { computed, onMounted, ref, watch } from 'vue'
	import { useRoute } from 'vue-router'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ConnectIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { getApiClient } from '../api-client.js'
	import {
		connectionBasicRows,
		connectionHealthMessage,
		connectionHealthStatusLabel,
		connectionHealthTagClass,
	} from '../connection-format.js'

	const route = useRoute()
	const connection = ref(null)
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')
	const notFound = ref('')

	const connectionId = computed(() => route.params.id)
	const detailRows = computed(() => (
		connection.value ? connectionBasicRows(connection.value) : []
	))

	const configErrorMessage = computed(() => {
		const prefix = configPath.value ? ` from ${configPath.value}` : ''
		return `Could not load config${prefix}: ${configError.value}`
	})

	async function loadConnection () {
		connection.value = null
		configError.value = ''
		configPath.value = ''
		loadError.value = ''
		notFound.value = ''

		try {
			const res = await getApiClient().getConnection({ id: connectionId.value })
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				return
			}
			if (res.error) {
				notFound.value = res.error
				return
			}
			connection.value = res.connection
		} catch (error) {
			loadError.value = 'Failed to load connection: ' + error.message
		}
	}

	onMounted(loadConnection)
	watch(connectionId, loadConnection)
</script>

<style scoped>
.connection-health-message {
	margin-left: 0.5em;
}
</style>
