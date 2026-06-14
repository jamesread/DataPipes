<template>
	<Section
		:id = "'job-' + job.id"
		:title = "job.id"
		:subtitle = "jobMeta"
		classes = "job-section"
	>
		<template #toolbar>
			<button
				type = "button"
				class = "good"
				:disabled = "jobRunning"
				@click = "runPreview"
			>{{ previewLabel }}</button>
			<button
				type = "button"
				class = "good"
				:disabled = "fullRunDisabled"
				@click = "runFullRun"
			>{{ fullRunLabel }}</button>
		</template>

		<table class = "hover datatable">
			<tbody>
				<tr>
					<th class = "uneditable">Extract connection</th>
					<td>
						<router-link v-if = "job.extractConnection" :to = "connectionPath(job.extractConnection)">{{ job.extractConnection }}</router-link>
						<span v-else>—</span>
					</td>
				</tr>
				<tr>
					<th class = "uneditable">Import directory</th>
					<td>{{ job.importDirectory || '—' }}</td>
				</tr>
				<tr>
					<th class = "uneditable">Load connection</th>
					<td>
						<router-link v-if = "job.loadConnection" :to = "connectionPath(job.loadConnection)">{{ job.loadConnection }}</router-link>
						<span v-else>—</span>
					</td>
				</tr>
				<tr>
					<th class = "uneditable">Load configured</th>
					<td>{{ job.loadConfigured ? 'Yes' : 'No' }}</td>
				</tr>
			</tbody>
		</table>

		<h3>Transformations</h3>
		<p v-if = "!job.transformations?.length" class = "subtle">No transformations configured.</p>
		<table v-else class = "hover datatable">
			<thead>
				<tr>
					<th>Ordinal</th>
					<th>Type</th>
					<th>Details</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				<tr v-for = "(t, i) in job.transformations" :key = "i">
					<td>{{ t.ordinal || '—' }}</td>
					<td>{{ t.name || '—' }}</td>
					<td>{{ t.description }}</td>
					<td>
						<button
							v-if = "isFirstTransformationOrdinal(job.transformations, i)"
							type = "button"
							class = "good"
							:disabled = "jobRunning"
							@click = "runStep(t.ordinal)"
						>Step</button>
					</td>
				</tr>
			</tbody>
		</table>

		<JobRun
			v-if = "previewResult"
			:job = "job"
			:result = "previewResult"
			:preview-mode = "showPreviewTab"
			:initial-tab = "jobRunInitialTab"
			:stepping = "stepping"
			v-model:load-message = "fullRunLoadMessage"
			@step = "runStep"
		/>
	</Section>
</template>

<script setup>
	import { computed, ref } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'
	import JobRun from './JobRun.vue'
	import {
		formatJobMeta,
		formatRpcError,
		getApiClient,
		isDownloadCsvLoad,
		isFirstTransformationOrdinal,
	} from './api-client.js'

	const props = defineProps({
		job: {
			type: Object,
			required: true,
		},
	})

	const previewResult = ref(null)
	const previewing = ref(false)
	const previewLabel = ref('Run preview')
	const fullRunning = ref(false)
	const fullRunLabel = ref('Full run')
	const fullRunLoadMessage = ref('')
	const runMode = ref(null)
	const stepping = ref(false)
	const jobRunInitialTab = ref('extraction')

	const jobMeta = computed(() => formatJobMeta(props.job))
	const jobRunning = computed(() => previewing.value || fullRunning.value || stepping.value)
	const fullRunDisabled = computed(() => jobRunning.value)
	const showPreviewTab = computed(() => runMode.value === 'preview' || runMode.value === 'step')

	function connectionPath (id) {
		return `/connections/${encodeURIComponent(id)}`
	}

	async function runStep (ordinal) {
		if (!ordinal) {
			return
		}
		stepping.value = true
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'step'
			jobRunInitialTab.value = 'preview'
			previewResult.value = await getApiClient().preview({
				jobId: props.job.id,
				rowLimit: 10,
				stepOrdinal: ordinal,
			})
		} catch (error) {
			showError('Failed to run step: ' + formatRpcError(error))
		} finally {
			stepping.value = false
		}
	}

	async function runPreview () {
		previewing.value = true
		previewLabel.value = 'Running preview…'
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'preview'
			jobRunInitialTab.value = 'preview'
			previewResult.value = await getApiClient().preview({ jobId: props.job.id })
			previewLabel.value = 'Preview complete'
			setTimeout(() => {
				previewLabel.value = 'Re-run preview'
				previewing.value = false
			}, 1000)
		} catch (error) {
			previewing.value = false
			previewLabel.value = 'Run preview'
			showError('Failed to run preview: ' + error.message)
		}
	}

	async function runFullRun () {
		fullRunning.value = true
		fullRunLabel.value = 'Running full job…'
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'fullRun'
			const res = await getApiClient().fullRun({ jobId: props.job.id })
			previewResult.value = res.preview
			if (res.loadSucceeded) {
				if (isDownloadCsvLoad(props.job.loadConnection)) {
					fullRunLoadMessage.value = 'Load complete. Download your CSV from the Load tab.'
				} else {
					fullRunLoadMessage.value = 'Load completed successfully.'
				}
			} else if (res.loadAttempted) {
				fullRunLoadMessage.value = res.loadError || 'Load failed.'
			} else if (res.loadError) {
				fullRunLoadMessage.value = res.loadError
			} else if (!props.job.loadConfigured) {
				fullRunLoadMessage.value = 'Load not configured.'
			}
			fullRunLabel.value = 'Full run complete'
			setTimeout(() => {
				fullRunLabel.value = 'Full run'
				fullRunning.value = false
			}, 1000)
		} catch (error) {
			fullRunning.value = false
			fullRunLabel.value = 'Full run'
			showError('Failed to run full job: ' + error.message)
		}
	}

	function showError (message) {
		const dialog = document.createElement('dialog')
		dialog.classList.add('alert', 'critical')
		dialog.innerText = 'An error occurred: ' + message
		document.body.appendChild(dialog)
		dialog.showModal()
	}
</script>

<style scoped>
.job-section :deep(.toolbar) {
	display: flex;
	gap: 0.5rem;
}
</style>
