<template>
	<Section :title = "connectionId" subtitle = "Connection details">
		<p><router-link to = "/connections">← Back to connections</router-link></p>

		<div v-if = "configError" class = "config-error" role = "alert">
			Could not load config{{ configPath ? ' from ' + configPath : '' }}: {{ configError }}
		</div>

		<p v-else-if = "loadError" class = "inline-notification">{{ loadError }}</p>

		<p v-else-if = "notFound" class = "inline-notification">{{ notFound }}</p>

		<table v-else-if = "connection" class = "hover datatable">
			<tbody>
				<tr v-for = "row in detailRows" :key = "row.label">
					<th class = "uneditable">{{ row.label }}</th>
					<td :class = "row.health ? healthClass(connection) : null">{{ row.value }}</td>
				</tr>
			</tbody>
		</table>
	</Section>
</template>

<script setup>
	import { computed, onMounted, ref, watch } from 'vue'
	import { useRoute } from 'vue-router'
	import Section from 'picocrank/vue/components/Section.vue'
	import { getApiClient } from '../api-client.js'
	import {
		connectionBasicRows,
		connectionHealthClass,
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

	function healthClass (conn) {
		return connectionHealthClass(conn)
	}

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
