<template>
	<template v-if = "loading">
		<NotificationBlock type = "info" message = "Loading job…" />
	</template>

	<template v-else-if = "configError">
		<NotificationBlock type = "critical" role = "alert" :message = "configErrorMessage" />
	</template>

	<template v-else-if = "loadError">
		<NotificationBlock type = "critical" role = "alert" :message = "loadError" />
	</template>

	<template v-else-if = "notFound">
		<NotificationBlock type = "warning" role = "alert" :message = "notFound" />
	</template>

	<template v-else-if = "job">
		<Section
			:id = "'job-' + job.id"
			:subtitle = "jobMeta"
			classes = "job-detail-section"
		>
			<template #title>
				<span class = "section-title-with-icon">
					<HugeiconsIcon
						:icon = "ClipboardListIcon"
						width = "22"
						height = "22"
						aria-hidden = "true"
					/>
					Overview
				</span>
			</template>

			<dl>
				<dt>Job ID</dt>
				<dd><code>{{ job.id }}</code></dd>
				<dt>Extract connection</dt>
				<dd>
					<router-link v-if = "job.extractConnection" :to = "connectionPath(job.extractConnection)">{{ job.extractConnection }}</router-link>
					<span v-else>—</span>
				</dd>
				<dt>Import directory</dt>
				<dd>{{ job.importDirectory || '—' }}</dd>
				<dt>Load connection</dt>
				<dd>
					<router-link v-if = "job.loadConnection" :to = "connectionPath(job.loadConnection)">{{ job.loadConnection }}</router-link>
					<span v-else>—</span>
				</dd>
				<dt>Load configured</dt>
				<dd>{{ job.loadConfigured ? 'Yes' : 'No' }}</dd>
			</dl>
		</Section>

		<Section
			subtitle = "Extraction, transformation, preview, and load"
			classes = "job-detail-section job-run-section"
			:padding = "false"
		>
			<template #title>
				<span class = "section-title-with-icon">
					<HugeiconsIcon
						:icon = "ShuffleIcon"
						width = "22"
						height = "22"
						aria-hidden = "true"
					/>
					Pipe
				</span>
			</template>

			<template #toolbar>
				<button
					type = "button"
					class = "good"
					:disabled = "jobRunning"
					@click = "runPipeline"
				>{{ pipelineLabel }}</button>
			</template>

			<JobRun
				:job = "job"
				:result = "jobRunResult"
				:preview-mode = "showPreviewTab"
				:initial-tab = "jobRunInitialTab"
				:stepping = "stepping"
				:extract-label = "extractLabel"
				:preview-label = "previewLabel"
				:pipe-running = "jobRunning"
				v-model:load-message = "fullRunLoadMessage"
				embedded
				@extract = "runExtract"
				@preview = "runPreview"
				@step = "runStep"
				@history-changed = "loadExecutions"
			/>
		</Section>

		<Section
			subtitle = "Past extract, preview, step, pipeline, and load runs for this job"
			classes = "job-detail-section job-history-section settings-list"
			:padding = "false"
		>
			<template #title>
				<span class = "section-title-with-icon">
					<HugeiconsIcon
						:icon = "ClockIcon"
						width = "22"
						height = "22"
						aria-hidden = "true"
					/>
					History
				</span>
			</template>

			<template #toolbar>
				<button
					type = "button"
					class = "neutral"
					title = "Refresh"
					aria-label = "Refresh"
					:disabled = "executionsLoading"
					@click = "loadExecutions"
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
				v-if = "executionsConfigError"
				type = "critical"
				role = "alert"
				:message = "executionsConfigErrorMessage"
				class = "list-banner-pad"
			/>

			<NotificationBlock
				v-else-if = "executionsLoadError"
				type = "critical"
				role = "alert"
				:message = "executionsLoadError"
				class = "list-banner-pad"
			/>

			<div v-else-if = "executionsLoading && !executions.length" class = "list-banner-pad muted">Loading history…</div>

			<template v-else>
				<NotificationBlock
					v-if = "!executions.length"
					type = "note"
					message = "No executions recorded yet."
					class = "list-banner-pad"
				/>

				<Table
					v-else
					class = "list-table-wrap"
					:headers = "executionTableHeaders"
					:data = "executionTableRows"
				>
					<template #cell-runKind = "{ value }">
						<strong>{{ value }}</strong>
					</template>
					<template #cell-status = "{ row }">
						<span class = "tag" :class = "executionStatusTagClass(row)">{{ executionStatusLabel(row) }}</span>
					</template>
					<template #cell-message = "{ value }">
						<span class = "subtle">{{ value || '—' }}</span>
					</template>
				</Table>
			</template>
		</Section>
	</template>
</template>

<script setup>
	import { computed, onMounted, ref, watch } from 'vue'
	import { useRoute } from 'vue-router'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ClipboardListIcon, ClockIcon, RefreshIcon, ShuffleIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'
	import Table from 'picocrank/vue/components/Table.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import JobRun from '../JobRun.vue'
	import {
		executionRunKindLabel,
		executionStatusLabel,
		executionStatusTagClass,
		formatExecutionStartedAt,
	} from '../execution-format.js'
	import {
		formatJobMeta,
		formatRpcError,
		getApiClient,
		isDownloadCsvLoad,
	} from '../api-client.js'

	const route = useRoute()
	const job = ref(null)
	const loading = ref(true)
	const configError = ref('')
	const configPath = ref('')
	const loadError = ref('')
	const notFound = ref('')

	const previewResult = ref(null)
	const extracting = ref(false)
	const extractLabel = ref('Extract')
	const previewing = ref(false)
	const previewLabel = ref('Run preview')
	const fullRunning = ref(false)
	const pipelineLabel = ref('Run Pipeline')
	const fullRunLoadMessage = ref('')
	const runMode = ref(null)
	const stepping = ref(false)
	const jobRunInitialTab = ref('extraction')
	const executions = ref([])
	const executionsLoading = ref(false)
	const executionsConfigError = ref('')
	const executionsConfigPath = ref('')
	const executionsLoadError = ref('')

	const executionTableHeaders = [
		{ key: 'startedAt', label: 'Started', sortable: true, width: '12rem' },
		{ key: 'runKind', label: 'Run', sortable: true, width: '8rem' },
		{ key: 'stepDetail', label: 'Step', sortable: true },
		{ key: 'status', label: 'Status', sortable: true, width: '8rem' },
		{ key: 'triggeredBy', label: 'Triggered by', sortable: true, width: '9rem' },
		{ key: 'message', label: 'Message', sortable: false },
	]

	const jobId = computed(() => route.params.id)
	const jobMeta = computed(() => (job.value ? formatJobMeta(job.value) : ''))
	const jobRunning = computed(() => extracting.value || previewing.value || fullRunning.value || stepping.value)
	const showPreviewTab = computed(() => runMode.value === 'preview' || runMode.value === 'step')

	const jobRunResult = computed(() => {
		if (previewResult.value) {
			return previewResult.value
		}
		return {
			issues: [],
			transformations: job.value?.transformations ?? [],
		}
	})

	const configErrorMessage = computed(() => {
		const prefix = configPath.value ? ` from ${configPath.value}` : ''
		return `Could not load config${prefix}: ${configError.value}`
	})

	const executionsConfigErrorMessage = computed(() => {
		const prefix = executionsConfigPath.value ? ` from ${executionsConfigPath.value}` : ''
		return `Could not load config${prefix}: ${executionsConfigError.value}`
	})

	const executionTableRows = computed(() => (
		executions.value.map((entry) => ({
			startedAt: formatExecutionStartedAt(entry.startedAt),
			startedAtSort: entry.startedAt,
			runKind: executionRunKindLabel(entry.runKind),
			stepDetail: entry.stepDetail || '—',
			success: entry.success,
			triggeredBy: entry.triggeredBy || 'manual',
			message: entry.message || '',
		}))
	))

	function connectionPath (id) {
		return `/connections/${encodeURIComponent(id)}`
	}

	async function loadJob () {
		job.value = null
		loading.value = true
		configError.value = ''
		configPath.value = ''
		loadError.value = ''
		notFound.value = ''
		previewResult.value = null
		extractLabel.value = 'Extract'
		extracting.value = false
		previewLabel.value = 'Run preview'
		pipelineLabel.value = 'Run Pipeline'
		previewing.value = false
		fullRunning.value = false
		stepping.value = false
		runMode.value = null
		executions.value = []

		try {
			const res = await getApiClient().listJobs({})
			if (res.configError) {
				configError.value = res.configError
				configPath.value = res.configPath
				return
			}
			const found = (res.jobs ?? []).find((entry) => entry.id === jobId.value)
			if (!found) {
				notFound.value = `Job "${jobId.value}" not found.`
				return
			}
			job.value = found
			await loadExecutions()
		} catch (error) {
			loadError.value = 'Failed to load job: ' + error.message
		} finally {
			loading.value = false
		}
	}

	async function loadExecutions () {
		if (!job.value?.id) {
			executions.value = []
			return
		}
		executionsLoading.value = true
		executionsConfigError.value = ''
		executionsConfigPath.value = ''
		executionsLoadError.value = ''

		try {
			const res = await getApiClient().listJobExecutions({ jobId: job.value.id })
			if (res.configError) {
				executionsConfigError.value = res.configError
				executionsConfigPath.value = res.configPath
				executions.value = []
				return
			}
			executions.value = res.executions ?? []
		} catch (error) {
			executionsLoadError.value = 'Failed to load history: ' + error.message
			executions.value = []
		} finally {
			executionsLoading.value = false
		}
	}

	async function runExtract () {
		if (!job.value) {
			return
		}
		extracting.value = true
		extractLabel.value = 'Extracting…'
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'extract'
			jobRunInitialTab.value = 'extraction'
			previewResult.value = await getApiClient().preview({
				jobId: job.value.id,
				stepOrdinal: -1,
			})
			extractLabel.value = 'Extract complete'
			setTimeout(() => {
				extractLabel.value = 'Re-extract'
				extracting.value = false
			}, 1000)
			await loadExecutions()
		} catch (error) {
			extracting.value = false
			extractLabel.value = 'Extract'
			showError('Failed to extract: ' + formatRpcError(error))
		}
	}

	async function runStep (ordinal) {
		if (!ordinal || !job.value) {
			return
		}
		stepping.value = true
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'step'
			jobRunInitialTab.value = 'preview'
			previewResult.value = await getApiClient().preview({
				jobId: job.value.id,
				rowLimit: 10,
				stepOrdinal: ordinal,
			})
			await loadExecutions()
		} catch (error) {
			showError('Failed to run step: ' + formatRpcError(error))
		} finally {
			stepping.value = false
		}
	}

	async function runPreview () {
		if (!job.value) {
			return
		}
		previewing.value = true
		previewLabel.value = 'Running preview…'
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'preview'
			jobRunInitialTab.value = 'preview'
			previewResult.value = await getApiClient().preview({ jobId: job.value.id })
			previewLabel.value = 'Preview complete'
			setTimeout(() => {
				previewLabel.value = 'Re-run preview'
				previewing.value = false
			}, 1000)
			await loadExecutions()
		} catch (error) {
			previewing.value = false
			previewLabel.value = 'Run preview'
			showError('Failed to run preview: ' + error.message)
		}
	}

	async function runPipeline () {
		if (!job.value) {
			return
		}
		fullRunning.value = true
		pipelineLabel.value = 'Running pipeline…'
		fullRunLoadMessage.value = ''

		try {
			runMode.value = 'fullRun'
			const res = await getApiClient().fullRun({ jobId: job.value.id })
			previewResult.value = res.preview
			if (res.loadSucceeded) {
				if (isDownloadCsvLoad(job.value.loadConnection)) {
					fullRunLoadMessage.value = 'Load complete. Download your CSV from the Load tab.'
				} else {
					fullRunLoadMessage.value = 'Load completed successfully.'
				}
			} else if (res.loadAttempted) {
				fullRunLoadMessage.value = res.loadError || 'Load failed.'
			} else if (res.loadError) {
				fullRunLoadMessage.value = res.loadError
			} else if (!job.value.loadConfigured) {
				fullRunLoadMessage.value = 'Load not configured.'
			}
			pipelineLabel.value = 'Pipeline complete'
			setTimeout(() => {
				pipelineLabel.value = 'Run Pipeline'
				fullRunning.value = false
			}, 1000)
			await loadExecutions()
		} catch (error) {
			fullRunning.value = false
			pipelineLabel.value = 'Run Pipeline'
			showError('Failed to run pipeline: ' + error.message)
		}
	}

	function showError (message) {
		const dialog = document.createElement('dialog')
		dialog.classList.add('alert', 'critical')
		dialog.innerText = 'An error occurred: ' + message
		document.body.appendChild(dialog)
		dialog.showModal()
	}

	onMounted(loadJob)
	watch(jobId, loadJob)
</script>

<style scoped>
.job-detail-section :deep(.toolbar) {
	display: flex;
	gap: 0.5rem;
}
</style>
