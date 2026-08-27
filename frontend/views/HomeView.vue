<template>
	<Section subtitle = "Configured extract, transform, and load pipelines">
		<template #title>
			<span class = "section-title-with-icon">
				<HugeiconsIcon
					:icon = "ClipboardListIcon"
					width = "22"
					height = "22"
					aria-hidden = "true"
				/>
				Jobs
			</span>
		</template>
		<p class = "subtle">
			Each job references an extract connection (CSV), optional transformations, and a load connection.
			Open a job to review its configuration, run a preview, or execute a full run.
		</p>

		<div v-if = "configError" class = "list-banner-pad">
			<NotificationBlock type = "critical" role = "alert" :message = "configErrorMessage" />
		</div>

		<NotificationBlock v-else-if = "loadError" type = "critical" role = "alert" :message = "loadError" />

		<NotificationBlock v-else-if = "loading" type = "info" message = "Loading jobs…" />

		<NotificationBlock v-else-if = "jobs.length === 0" type = "note" message = "No jobs configured." />

		<Navigation v-else ref = "jobsNavigation">
			<NavigationGrid />
		</Navigation>
	</Section>
</template>

<script setup>
	import { computed, nextTick, onMounted, ref, watch } from 'vue'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ClipboardListIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'
	import Navigation from 'picocrank/vue/components/Navigation.vue'
	import NavigationGrid from 'picocrank/vue/components/NavigationGrid.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { formatJobMeta, getApiClient } from '../api-client.js'

	const jobs = ref([])
	const loading = ref(true)
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')
	const jobsNavigation = ref(null)

	const configErrorMessage = computed(() => {
		const prefix = configPath.value ? ` from ${configPath.value}` : ''
		return `Could not load config${prefix}: ${configError.value}`
	})

	function registerJobLinks () {
		if (!jobsNavigation.value) {
			return
		}
		jobsNavigation.value.clearNavigationLinks()
		for (const job of jobs.value) {
			jobsNavigation.value.addRouterLink('job-detail', job.id, {
				params: { id: job.id },
				description: formatJobMeta(job),
			})
		}
	}

	async function loadJobs () {
		loading.value = true
		configError.value = ''
		configPath.value = ''
		loadError.value = ''

		try {
			const res = await getApiClient().listJobs({})
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				jobs.value = []
				return
			}
			jobs.value = res.jobs ?? []
		} catch (error) {
			loadError.value = 'Failed to load jobs: ' + error.message
			jobs.value = []
		} finally {
			loading.value = false
			await nextTick()
			registerJobLinks()
		}
	}

	onMounted(loadJobs)
	watch(jobs, async () => {
		await nextTick()
		registerJobLinks()
	})
</script>
