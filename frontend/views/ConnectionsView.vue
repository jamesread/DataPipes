<template>
	<Section title = "Connections" subtitle = "Configured extract and load connections">
		<div v-if = "configError" class = "config-error" role = "alert">
			Could not load config{{ configPath ? ' from ' + configPath : '' }}: {{ configError }}
		</div>

		<p v-else-if = "loadError" class = "inline-notification">{{ loadError }}</p>

		<p v-else-if = "connections.length === 0" class = "inline-notification">No connections configured.</p>

		<table v-else class = "hover datatable">
			<thead>
				<tr>
					<th>Name</th>
					<th>Type</th>
					<th>Details</th>
					<th>Health</th>
				</tr>
			</thead>
			<tbody>
				<tr v-for = "conn in connections" :key = "conn.id">
					<td><router-link :to = "'/connections/' + encodeURIComponent(conn.id)">{{ conn.id }}</router-link></td>
					<td>{{ conn.type }}</td>
					<td>{{ formatDetails(conn) }}</td>
					<td :class = "healthClass(conn)">{{ conn.healthMessage || '—' }}</td>
				</tr>
			</tbody>
		</table>
	</Section>
</template>

<script setup>
	import { onMounted, ref } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'
	import { getApiClient } from '../api-client.js'
	import {
		connectionHealthClass,
		formatConnectionDetails,
	} from '../connection-format.js'

	const connections = ref([])
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')

	function healthClass (conn) {
		return connectionHealthClass(conn)
	}

	function formatDetails (conn) {
		return formatConnectionDetails(conn)
	}

	onMounted(async () => {
		try {
			const res = await getApiClient().listConnections({})
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				return
			}
			connections.value = res.connections ?? []
		} catch (error) {
			loadError.value = 'Failed to load connections: ' + error.message
		}
	})
</script>
