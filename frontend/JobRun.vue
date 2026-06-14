<template>
	<div class = "job-run">
		<div class = "issuesList">
			<p v-if = "previewClean" class = "inline-notification good">No issues found!</p>
			<IssueDetail v-for = "(issue, i) in result.issues" :key = "i" :issue = "issue" />
		</div>

		<Tabs :key = "tabsKey" :tabs = "tabs" :default-tab = "initialTab">
			<template #tab-extraction>
				<p class = "subtle">
					Source fields are read from CSV columns configured on the extract connection
					under <code>connections.&lt;name&gt;.columns</code>.
				</p>

				<h4>Source fields</h4>
				<table v-if = "result.extractColumns?.length" class = "hover datatable">
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
				<p v-else class = "subtle">No extract column mapping configured.</p>

				<h4>Extract preview</h4>
				<p class = "subtle">First 10 rows after extract, before any transformations.</p>
				<Table
					v-if = "extractPreviewTableData.length"
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
				<p v-else class = "subtle">No extracted rows to preview.</p>

				<h4>File list</h4>
				<table class = "hover datatable">
					<thead>
						<tr>
							<th>Filename</th>
							<th>Lines</th>
						</tr>
					</thead>
					<tbody>
						<tr v-if = "!result.sourceFiles?.length">
							<td colspan = "2" class = "subtle">No files extracted.</td>
						</tr>
						<tr v-for = "file in result.sourceFiles" :key = "file.filename">
							<td>{{ file.filename }}</td>
							<td>{{ file.lineCount }}</td>
						</tr>
					</tbody>
				</table>

				<h4>Stats</h4>
				<div class = "grid-boxed">
					<div class = "stat-display">
						<span class = "subtle">Total lines</span>
						<span class = "stat">{{ result.totalLines }}</span>
					</div>
					<div class = "stat-display">
						<span class = "subtle">Total files</span>
						<span class = "stat">{{ result.sourceFiles?.length ?? 0 }}</span>
					</div>
				</div>

				<h4>Export</h4>
				<a v-if = "showDownloadLink" :href = "downloadUrl" class = "button good">Download CSV</a>
				<p v-else class = "subtle">Run preview to generate a downloadable CSV.</p>
			</template>

			<template #tab-transformations>
				<p v-if = "!transformations.length" class = "subtle">No transformations configured.</p>
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
						<tr v-for = "(t, i) in transformations" :key = "i">
							<td>{{ t.ordinal || '—' }}</td>
							<td>{{ t.name || '—' }}</td>
							<td>{{ t.description }}</td>
							<td>
								<button
									v-if = "isFirstTransformationOrdinal(transformations, i)"
									type = "button"
									class = "good"
									:disabled = "stepping"
									@click = "emitStep(t.ordinal)"
								>Step</button>
							</td>
						</tr>
					</tbody>
				</table>
			</template>

			<template v-if = "showPreviewTab" #tab-preview>
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

			<template #tab-load>
				<template v-if = "downloadCsvLoad">
					<template v-if = "loaded">
						<p class = "inline-notification good">{{ loadMessage }}</p>
					</template>
					<template v-else-if = "loadMessage">
						<p class = "inline-notification">{{ loadMessage }}</p>
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
						<p class = "inline-notification good">{{ loadMessage }}</p>
					</template>
					<template v-else-if = "loadMessage && !loadProgress.active">
						<p class = "inline-notification">{{ loadMessage }}</p>
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
	import { computed, ref } from 'vue'
	import Tabs from 'picocrank/vue/components/Tabs.vue'
	import Table from 'picocrank/vue/components/Table.vue'
	import IssueDetail from './IssueDetail.vue'
	import {
		extractPreviewHeaderLabels,
		formatExtractCell,
		formatLoadStats,
		formatPreviewCell,
		formatRpcError,
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
		loadMessage: {
			type: String,
			default: '',
		},
	})

	const emit = defineEmits(['update:loadMessage', 'step'])

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

	const previewClean = computed(() => (props.result.issues?.length ?? 0) === 0)
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
	const showPreviewTab = computed(() => props.previewMode || (props.result.appliedStepOrdinal ?? 0) > 0)
	const tabsKey = computed(() => `${props.result.appliedStepOrdinal ?? 0}-${props.result.completedDate ?? ''}-${props.initialTab}`)
	const downloadCsvLoad = computed(() => isDownloadCsvLoad(props.job.loadConnection))
	const downloadUrl = computed(() => jobDownloadCsvUrl(props.job.id))
	const downloadUrlAbsolute = computed(() => {
		if (typeof window === 'undefined') {
			return downloadUrl.value
		}
		return `${window.location.origin}${downloadUrl.value}`
	})
	const showDownloadLink = computed(() => previewClean.value && (props.result.totalLines ?? 0) > 0)
	const loaded = computed(() => {
		if (downloadCsvLoad.value && props.loadMessage.startsWith('Load complete')) {
			return true
		}
		return props.loadMessage.startsWith('Load completed') || props.loadMessage.startsWith('Load finished')
	})
	const tabs = computed(() => {
		const items = [
			{ id: 'extraction', label: 'Extraction Results' },
			{ id: 'transformations', label: 'Transformations' },
		]
		if (showPreviewTab.value) {
			items.push({ id: 'preview', label: 'Preview' })
		}
		items.push({ id: 'load', label: 'Load' })
		return items
	})
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
		} catch (error) {
			loading.value = false
			loadLabel.value = 'Load'
			loadProgress.value.active = false
			showError('Failed to load: ' + formatRpcError(error))
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

.issuesList {
	margin-bottom: 1rem;
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
</style>
