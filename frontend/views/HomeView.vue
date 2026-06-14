<template>
	<p class = "subtle">
		Configured jobs are listed below. Each job references an <code>extract</code> connection (CSV extract)
		and a <code>load</code> connection (e.g. MySQL). Run preview to sample extract and transformations, or
		full run to execute the complete pipeline including load.
	</p>

	<JobsList />
</template>

<script setup>
	import JobsList from '../JobsList.vue'
</script>
