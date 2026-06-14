<template>
	<section>
		<h3 class = "description">{{ issue.description }}</h3>

		<table class = "detailsTable">
			<tr>
				<th></th>
				<th v-for = "attr in issue.expected" :key = "attr.key">{{ attr.key }}</th>
			</tr>
			<tr>
				<td class = "uneditable">Expected:</td>
				<td v-for = "attr in issue.expected" :key = "'exp-' + attr.key">{{ attr.val }}</td>
			</tr>
			<tr>
				<td class = "uneditable">Actual:</td>
				<td v-for = "attr in issue.actual" :key = "'act-' + attr.key">{{ attr.val }}</td>
			</tr>
			<tr>
				<td class = "uneditable">Intermediate:</td>
				<td v-for = "attr in issue.intermediate" :key = "'int-' + attr.key">{{ attr.val }}</td>
			</tr>
		</table>

		<h4>Current Location</h4>
		<table>
			<tr>
				<td class = "uneditable">Last Filename:</td>
				<td>{{ issue.lastLocationFilename }} : {{ issue.lastLocationLineNumber }}</td>
			</tr>
			<tr>
				<td class = "uneditable">Current Filename:</td>
				<td>{{ issue.currentLocationFilename }} : {{ issue.currentLocationLineNumber }}</td>
			</tr>
		</table>
	</section>
</template>

<script setup>
	defineProps({
		issue: {
			type: Object,
			required: true,
		},
	})
</script>
