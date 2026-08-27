<template>
	<div class = "job-run" :class = "{ embedded }">
		<Tabs :key = "tabsKey" :tabs = "tabs" :default-tab = "initialTab" :padding = "true">
			<template #tab-extraction>
				<div class = "tab-actions">
					<button
						type = "button"
						class = "good"
						:disabled = "pipeRunning"
						@click = "emitExtract"
					>{{ extractLabel }}</button>
				</div>

				<p v-if = "extractSourceLoading" class = "subtle">Loading extract source…</p>
				<NotificationBlock
					v-else-if = "extractSourceError"
					type = "critical"
					role = "alert"
					:message = "extractSourceError"
				/>
				<dl v-else-if = "job.extractConnection && extractSource">
					<dt>Connection</dt>
					<dd>
						<router-link :to = "extractConnectionPath">{{ job.extractConnection }}</router-link>
					</dd>
					<dt>Import directory</dt>
					<dd>{{ job.importDirectory || extractSource.importDirectory || '—' }}</dd>
					<template v-for = "row in extractSourceRows" :key = "row.label">
						<dt>{{ row.label }}</dt>
						<dd v-if = "row.health">
							<span class = "tag" :class = "connectionHealthTagClass(extractSource)">{{ connectionHealthStatusLabel(extractSource) }}</span>
							<span v-if = "connectionHealthMessage(extractSource)" class = "subtle extract-source-health-message">{{ connectionHealthMessage(extractSource) }}</span>
						</dd>
						<dd v-else>{{ row.value }}</dd>
					</template>
				</dl>
				<p v-else-if = "!job.extractConnection" class = "subtle">No extract connection configured for this job.</p>

				<div v-if = "showIssues" class = "issuesList">
					<NotificationBlock
						v-if = "previewClean"
						type = "good"
						message = "No validation issues found"
					/>
					<p v-if = "previewClean" class = "subtle issues-summary">
						Extract and transformation checks passed for this run.
					</p>
					<IssueDetail v-for = "(issue, i) in result.issues" :key = "i" :issue = "issue" />
				</div>

				<NotificationBlock
					v-if = "!runCompleted"
					type = "info"
					message = "Click Extract to load column mapping, preview rows, and file stats."
				/>

				<template v-else>
					<p v-if = "hasSourceFields || hasExtractPreview" class = "subtle">
						Source fields are read from CSV columns configured on the extract connection
						under <code>connections.&lt;name&gt;.columns</code>.
					</p>

					<template v-if = "hasSourceFields">
						<h4>Source fields</h4>
						<table class = "hover datatable">
							<thead>
								<tr>
									<th>Field</th>
									<th>CSV column</th>
								</tr>
							</thead>
							<tbody>
								<tr v-for = "col in result.extractColumns" :key = "col.fieldName">
									<td><code>{{ col.fieldName }}</code></td>
									<td>{{ col.columnIndex }}</td>
								</tr>
							</tbody>
						</table>
					</template>
					<p v-else class = "subtle">No extract column mapping configured.</p>

					<template v-if = "hasExtractPreview">
						<h4>Extract preview</h4>
						<p class = "subtle">First 10 rows after extract, before any transformations.</p>
						<Table
							class = "preview-data-table"
							:headers = "extractPreviewTableHeaders"
							:data = "extractPreviewTableData"
							:show-pagination = "false"
						>
							<template
								v-for = "(fieldName, i) in extractPreviewFieldNames"
								:key = "fieldName"
								#[`cell-col_${i}`] = "{ value }"
							>{{ formatExtractCell(value, fieldName) }}</template>
						</Table>
					</template>
					<p v-else-if = "hasSourceFields" class = "subtle">No extracted rows to preview.</p>

					<template v-if = "hasSourceFiles">
						<h4>File list</h4>
						<table class = "hover datatable">
							<thead>
								<tr>
									<th>Filename</th>
									<th>Lines</th>
								</tr>
							</thead>
							<tbody>
								<tr v-for = "file in result.sourceFiles" :key = "file.filename">
									<td>{{ file.filename }}</td>
									<td>{{ file.lineCount }}</td>
								</tr>
							</tbody>
						</table>
					</template>
					<p v-else-if = "hasSourceFields || hasExtractPreview" class = "subtle">No files extracted.</p>

					<template v-if = "hasExtractStats">
						<h4>Stats</h4>
						<div class = "grid-boxed">
							<div class = "stat-display">
								<span class = "subtle">Total lines</span>
								<span class = "stat">{{ result.totalLines ?? 0 }}</span>
							</div>
							<div class = "stat-display">
								<span class = "subtle">Total files</span>
								<span class = "stat">{{ result.sourceFiles?.length ?? 0 }}</span>
							</div>
						</div>
					</template>

					<template v-if = "hasExtractExport">
						<h4>Export</h4>
						<a :href = "downloadUrl" class = "button good">Download CSV</a>
					</template>
				</template>
			</template>

			<template #tab-transformations>
				<NotificationBlock
					v-if = "!transformations.length"
					type = "note"
					message = "No transformations configured."
					class = "list-banner-pad"
				/>
				<Table
					v-else
					class = "list-table-wrap"
					:headers = "transformationTableHeaders"
					:data = "transformationTableRows"
				>
					<template #cell-name = "{ value }">
						<strong>{{ value }}</strong>
					</template>
					<template #cell-actions = "{ row }">
						<div class = "actions-cell">
							<button
								v-if = "row.showStep"
								type = "button"
								class = "good small"
								:disabled = "stepping"
								@click = "emitStep(row.stepOrdinal)"
							>Step</button>
						</div>
					</template>
				</Table>
			</template>

			<template #tab-preview>
				<template v-if = "!previewTabEnabled">
					<div class = "tab-actions">
						<button
							type = "button"
							class = "good"
							:disabled = "pipeRunning"
							@click = "emitPreview"
						>{{ previewLabel }}</button>
					</div>
					<NotificationBlock
						type = "info"
						message = "Run preview to extract data, apply transformations, and show transformed rows here."
					/>
				</template>
				<template v-else>
				<p class = "subtle">
					<span v-if = "result.appliedStepOrdinal">Showing results through transformation step {{ result.appliedStepOrdinal }}.</span>
					<span v-else>Extract and transformations applied; load has not run.</span>
					<span v-if = "result.truncated"> Showing first {{ result.rowLimit }} rows.</span>
				</p>
				<Table
					class = "preview-data-table"
					:headers = "previewTableHeaders"
					:data = "previewTableData"
					:show-pagination = "false"
				>
					<template
						v-for = "(columnName, i) in previewColumnNamesList"
						:key = "columnName + i"
						#[`cell-col_${i}`] = "{ value }"
					>{{ formatPreviewCell(value, columnName) }}</template>
				</Table>
				</template>
			</template>

			<template #tab-load>
				<template v-if = "downloadCsvLoad">
					<template v-if = "loaded">
						<NotificationBlock type = "good" :message = "loadMessage" />
					</template>
					<template v-else-if = "loadMessage">
						<NotificationBlock type = "info" :message = "loadMessage" />
					</template>
					<template v-else>
						<p class = "subtle">Run preview or full job to prepare transformed data for download.</p>
					</template>
					<p v-if = "showDownloadLink">
						Download URL:
						<code><a :href = "downloadUrl">{{ downloadUrlAbsolute }}</a></code>
					</p>
					<p v-if = "showDownloadLink">
						<a :href = "downloadUrl" class = "button good">Download CSV</a>
					</p>
					<p v-else-if = "!previewClean" class = "subtle">Resolve preview issues before downloading.</p>
				</template>
				<template v-else>
					<template v-if = "loadProgress.active || loadProgress.rows.length">
						<div class = "load-progress">
							<label class = "subtle">Load progress</label>
							<progress
								:max = "loadProgress.total || 1"
								:value = "loadProgress.current"
							></progress>
							<p class = "subtle">{{ loadProgress.message }}</p>
							<div v-if = "loadProgress.total" class = "grid-boxed">
								<div class = "stat-display">
									<span class = "subtle">Loaded</span>
									<span class = "stat">{{ loadProgress.succeeded }}</span>
								</div>
								<div class = "stat-display">
									<span class = "subtle">Failed</span>
									<span class = "stat">{{ loadProgress.failed }}</span>
								</div>
								<div class = "stat-display">
									<span class = "subtle">Total</span>
									<span class = "stat">{{ loadProgress.total }}</span>
								</div>
							</div>
						</div>
						<table v-if = "loadProgress.rows.length" class = "hover datatable load-results">
							<thead>
								<tr>
									<th>Row</th>
									<th>Status</th>
									<th>Details</th>
									<th>Error</th>
								</tr>
							</thead>
							<tbody>
								<tr v-for = "row in loadProgress.rows" :key = "row.rowNumber">
									<td>{{ row.rowNumber }}</td>
									<td :class = "row.success ? 'karma-good' : 'karma-bad'">{{ row.success ? 'OK' : 'Failed' }}</td>
									<td>{{ row.details }}</td>
									<td>{{ row.error || '—' }}</td>
								</tr>
							</tbody>
						</table>
					</template>
					<template v-if = "loaded">
						<NotificationBlock type = "good" :message = "loadMessage" />
					</template>
					<template v-else-if = "loadMessage && !loadProgress.active">
						<NotificationBlock type = "info" :message = "loadMessage" />
					</template>
					<template v-else-if = "!loadProgress.active && !loadProgress.rows.length">
						<p class = "subtle">Load has not run.</p>
						<p v-if = "previewMode">Run a full job or use the button below to load transformed rows to the configured destination.</p>
						<p v-else-if = "job.loadConfigured">Load was not completed for this run.</p>
						<p v-else>Load is not configured for this job.</p>
					</template>
					<button
						v-if = "!loaded && job.loadConfigured && !loadProgress.active"
						type = "button"
						class = "good"
						:disabled = "loadDisabled"
						@click = "runLoad"
					>{{ loadLabel }}</button>
				</template>
			</template>
		</Tabs>
	</div>
</template>

<script setup>
	import { computed, onMounted, ref, watch } from 'vue'
	import { RouterLink } from 'vue-router'
	import Tabs from 'picocrank/vue/components/Tabs.vue'
	import Table from 'picocrank/vue/components/Table.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import IssueDetail from './IssueDetail.vue'
	import {
		connectionBasicRows,
		connectionHealthMessage,
		connectionHealthStatusLabel,
		connectionHealthTagClass,
	} from './connection-format.js'
	import {
		extractPreviewHeaderLabels,
		formatExtractCell,
		formatLoadStats,
		formatPreviewCell,
		formatRpcError,
		getApiClient,
		isDownloadCsvLoad,
		isFirstTransformationOrdinal,
		jobDownloadCsvUrl,
		previewColumnNames,
		previewHeaderLabels,
		streamLoad,
	} from './api-client.js'

	const props = defineProps({
		job: {
			type: Object,
			required: true,
		},
		result: {
			type: Object,
			required: true,
		},
		previewMode: {
			type: Boolean,
			default: false,
		},
		initialTab: {
			type: String,
			default: 'extraction',
		},
		stepping: {
			type: Boolean,
			default: false,
		},
		extractLabel: {
			type: String,
			default: 'Extract',
		},
		previewLabel: {
			type: String,
			default: 'Run preview',
		},
		pipeRunning: {
			type: Boolean,
			default: false,
		},
		loadMessage: {
			type: String,
			default: '',
		},
		embedded: {
			type: Boolean,
			default: false,
		},
	})

	const emit = defineEmits(['update:loadMessage', 'step', 'extract', 'preview', 'history-changed'])

	const loading = ref(false)
	const loadLabel = ref('Load')
	const loadProgress = ref({
		active: false,
		total: 0,
		current: 0,
		succeeded: 0,
		failed: 0,
		message: '',
		rows: [],
	})
	const loadAbort = ref(null)
	const extractSource = ref(null)
	const extractSourceLoading = ref(false)
	const extractSourceError = ref('')

	const extractConnectionPath = computed(() => (
		`/connections/${encodeURIComponent(props.job.extractConnection)}`
	))
	const extractSourceRows = computed(() => (
		extractSource.value ? connectionBasicRows(extractSource.value) : []
	))

	const previewClean = computed(() => (props.result.issues?.length ?? 0) === 0)
	const runCompleted = computed(() => {
		const r = props.result
		return Boolean(r.completedDate)
			|| r.totalLines !== undefined
			|| (r.extractColumns?.length ?? 0) > 0
	})
	const showIssues = computed(() => {
		if ((props.result.issues?.length ?? 0) > 0) {
			return true
		}
		return runCompleted.value
	})
	const hasSourceFields = computed(() => (props.result.extractColumns?.length ?? 0) > 0)
	const hasExtractPreview = computed(() => (props.result.extractPreviewRows?.length ?? 0) > 0)
	const hasSourceFiles = computed(() => (props.result.sourceFiles?.length ?? 0) > 0)
	const extractPreviewFieldNames = computed(() => (
		(props.result.extractColumns ?? []).map((col) => col.fieldName)
	))
	const extractPreviewTableHeaders = computed(() => {
		const headers = extractPreviewHeaderLabels(props.result.extractColumns).map((label, i) => ({
			key: `col_${i}`,
			label,
			sortable: false,
		}))
		headers.push(
			{ key: 'sourceFilename', label: 'File', sortable: false },
			{ key: 'sourceLineNumber', label: 'Line', sortable: false },
		)
		return headers
	})
	const extractPreviewTableData = computed(() => (
		(props.result.extractPreviewRows ?? []).map((line) => {
			const row = {
				sourceFilename: line.sourceFilename,
				sourceLineNumber: line.sourceLineNumber,
			}
			line.cells.forEach((cell, i) => {
				row[`col_${i}`] = cell
			})
			return row
		})
	))
	const previewColumnNamesList = computed(() => previewColumnNames(props.result.columnMap))
	const previewHeaders = computed(() => previewHeaderLabels(props.result.columnMap))
	const previewTableHeaders = computed(() => {
		const headers = previewColumnNamesList.value.map((columnName, i) => ({
			key: `col_${i}`,
			label: previewHeaders.value[i] || columnName,
			sortable: false,
		}))
		headers.push(
			{ key: 'sourceFilename', label: 'File', sortable: false },
			{ key: 'sourceLineNumber', label: 'Line', sortable: false },
		)
		return headers
	})
	const previewTableData = computed(() => (
		(props.result.previewRows ?? []).map((line) => {
			const row = {
				sourceFilename: line.sourceFilename,
				sourceLineNumber: line.sourceLineNumber,
			}
			line.cells.forEach((cell, i) => {
				row[`col_${i}`] = cell
			})
			return row
		})
	))
	const transformations = computed(() => props.result.transformations ?? props.job.transformations ?? [])
	const transformationTableHeaders = [
		{ key: 'ordinal', label: 'Ordinal', sortable: true, width: '6rem' },
		{ key: 'name', label: 'Type', sortable: true, width: '10rem' },
		{ key: 'description', label: 'Details', sortable: true },
		{ key: 'actions', label: 'Actions', sortable: false, width: '8rem' },
	]
	const transformationTableRows = computed(() => {
		const list = transformations.value
		return list.map((t, i) => ({
			ordinal: t.ordinal || '—',
			name: t.name || '—',
			description: t.description || '—',
			stepOrdinal: t.ordinal,
			showStep: isFirstTransformationOrdinal(list, i),
		}))
	})
	const previewTabEnabled = computed(() => {
		const r = props.result
		if ((r.previewRows?.length ?? 0) > 0) {
			return true
		}
		if (r.columnMap && Object.keys(r.columnMap).length > 0) {
			return true
		}
		if ((r.appliedStepOrdinal ?? 0) > 0) {
			return true
		}
		return false
	})
	const tabsKey = computed(() => `${previewTabEnabled.value}-${props.result.appliedStepOrdinal ?? 0}-${props.result.completedDate ?? ''}-${props.initialTab}`)
	const downloadCsvLoad = computed(() => isDownloadCsvLoad(props.job.loadConnection))
	const downloadUrl = computed(() => jobDownloadCsvUrl(props.job.id))
	const downloadUrlAbsolute = computed(() => {
		if (typeof window === 'undefined') {
			return downloadUrl.value
		}
		return `${window.location.origin}${downloadUrl.value}`
	})
	const showDownloadLink = computed(() => previewClean.value && (props.result.totalLines ?? 0) > 0)
	const hasExtractStats = computed(() => {
		if (!runCompleted.value) {
			return false
		}
		return (props.result.totalLines ?? 0) > 0 || hasSourceFiles.value
	})
	const hasExtractExport = computed(() => showDownloadLink.value)
	const loaded = computed(() => {
		if (downloadCsvLoad.value && props.loadMessage.startsWith('Load complete')) {
			return true
		}
		return props.loadMessage.startsWith('Load completed') || props.loadMessage.startsWith('Load finished')
	})
	const tabs = computed(() => [
		{ id: 'extraction', label: 'Extraction' },
		{ id: 'transformations', label: 'Transformations' },
		{ id: 'preview', label: 'Preview' },
		{ id: 'load', label: 'Load' },
	])
	const loadDisabled = computed(() => {
		if (loading.value) {
			return true
		}
		if (!previewClean.value || !props.job.loadConfigured) {
			return true
		}
		return false
	})

	function emitStep (ordinal) {
		emit('step', ordinal)
	}

	function emitExtract () {
		emit('extract')
	}

	function emitPreview () {
		emit('preview')
	}

	async function loadExtractSource () {
		const connectionId = props.job.extractConnection
		extractSource.value = null
		extractSourceError.value = ''
		if (!connectionId) {
			return
		}
		extractSourceLoading.value = true
		try {
			const res = await getApiClient().getConnection({ id: connectionId })
			if (res.configError) {
				extractSourceError.value = `Could not load config: ${res.configError}`
				return
			}
			if (res.error) {
				extractSourceError.value = res.error
				return
			}
			extractSource.value = res.connection ?? null
		} catch (error) {
			extractSourceError.value = 'Failed to load extract source: ' + error.message
		} finally {
			extractSourceLoading.value = false
		}
	}

	onMounted(loadExtractSource)
	watch(() => props.job.extractConnection, loadExtractSource)

	function resetLoadProgress () {
		loadProgress.value = {
			active: false,
			total: 0,
			current: 0,
			succeeded: 0,
			failed: 0,
			message: '',
			rows: [],
		}
	}

	function applyLoadProgress (event) {
		if (event.phase === 'started') {
			loadProgress.value.active = true
			loadProgress.value.total = event.totalRows ?? 0
			loadProgress.value.message = event.message || 'Loading…'
			return
		}
		if (event.phase === 'row') {
			loadProgress.value.current = event.rowNumber ?? loadProgress.value.current
			loadProgress.value.total = event.totalRows ?? loadProgress.value.total
			loadProgress.value.succeeded = event.succeeded ?? loadProgress.value.succeeded
			loadProgress.value.failed = event.failed ?? loadProgress.value.failed
			loadProgress.value.message = event.message || loadProgress.value.message
			loadProgress.value.rows.push({
				rowNumber: event.rowNumber,
				success: event.rowSuccess,
				details: formatLoadStats(event.stats),
				error: event.error,
			})
			return
		}
		if (event.phase === 'complete') {
			loadProgress.value.active = false
			loadProgress.value.succeeded = event.succeeded ?? loadProgress.value.succeeded
			loadProgress.value.failed = event.failed ?? loadProgress.value.failed
			loadProgress.value.current = loadProgress.value.total
			loadProgress.value.message = event.message || 'Load complete'
			return
		}
		if (event.phase === 'failed') {
			loadProgress.value.active = false
			loadProgress.value.message = event.message || 'Load failed'
		}
	}

	async function runLoad () {
		loading.value = true
		loadLabel.value = 'Loading'
		emit('update:loadMessage', '')
		resetLoadProgress()
		loadProgress.value.active = true

		if (loadAbort.value) {
			loadAbort.value.abort()
		}
		const abort = new AbortController()
		loadAbort.value = abort

		try {
			await streamLoad(props.job.id, applyLoadProgress, abort.signal)
			const failed = loadProgress.value.failed
			const succeeded = loadProgress.value.succeeded
			if (failed > 0) {
				emit('update:loadMessage', `Load finished with ${failed} failed row(s) (${succeeded} succeeded).`)
			} else {
				emit('update:loadMessage', 'Load completed successfully.')
				loadLabel.value = 'Loaded successfully!'
			}
			setTimeout(() => {
				loading.value = false
				loadLabel.value = 'Load'
			}, 1000)
			emit('history-changed')
		} catch (error) {
			loading.value = false
			loadLabel.value = 'Load'
			loadProgress.value.active = false
			showError('Failed to load: ' + formatRpcError(error))
			emit('history-changed')
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
.job-run {
	margin-top: 1.5rem;
}

.job-run.embedded {
	margin-top: 0;
}

.tab-actions {
	margin-bottom: 1rem;
}

.extract-source-health-message {
	margin-left: 0.5em;
}

.issuesList {
	margin-bottom: 1rem;
}

.issues-summary {
	margin-top: 0.25rem;
}

h4 {
	margin-top: 1.25rem;
}

.preview-data-table :deep(th:nth-last-child(-n+2)) {
	color: #999;
}

.load-progress progress {
	width: 100%;
	display: block;
	margin: 0.5rem 0;
}

.load-results {
	max-height: 24rem;
	overflow-y: auto;
	display: block;
}

.actions-cell {
	text-align: right;
}
</style>
