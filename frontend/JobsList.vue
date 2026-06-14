<template>
	<div v-if = "configError" class = "config-error" role = "alert">
		Could not load config{{ configPath ? ' from ' + configPath : '' }}: {{ configError }}
	</div>

	<p v-else-if = "loadError" class = "inline-notification">{{ loadError }}</p>

	<p v-else-if = "jobs.length === 0" class = "inline-notification">No jobs configured.</p>

	<JobSection v-for = "job in jobs" :key = "job.id" :job = "job" />
</template>

<script setup>
	import { onMounted, ref } from 'vue'
	import JobSection from './JobSection.vue'
	import { getApiClient } from './api-client.js'

	const jobs = ref([])
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')

	onMounted(async () => {
		try {
			const res = await getApiClient().listJobs({})
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				return
			}
			jobs.value = res.jobs ?? []
		} catch (error) {
			loadError.value = 'Failed to load jobs: ' + error.message
		}
	})
</script>
