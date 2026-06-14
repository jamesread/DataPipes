<template>
	<Section title = "Transformations" subtitle = "Supported job transform steps">
		<p class = "subtle">
			Configure transformations under <code>jobs.&lt;name&gt;.transform</code> in the config file.
		</p>

		<p v-if = "loadError" class = "inline-notification">{{ loadError }}</p>

		<article v-for = "t in types" :key = "t.id" class = "transformation-type">
			<h3>{{ t.name }}</h3>
			<p>{{ t.description }}</p>
			<pre class = "yaml-example"><code>{{ t.yamlExample }}</code></pre>
		</article>
	</Section>
</template>

<script setup>
	import { onMounted, ref } from 'vue'
	import Section from 'picocrank/vue/components/Section.vue'
	import { getApiClient } from '../api-client.js'

	const types = ref([])
	const loadError = ref('')

	onMounted(async () => {
		try {
			const res = await getApiClient().listTransformationTypes({})
			types.value = res.types ?? []
		} catch (error) {
			loadError.value = 'Failed to load transformation types: ' + error.message
		}
	})
</script>

<style scoped>
.transformation-type {
	margin-top: 1.5rem;
}

.transformation-type h3 {
	margin: 0 0 0.5rem;
}

.transformation-type p {
	margin: 0 0 0.75rem;
}

.yaml-example {
	margin: 0;
	padding: 1rem;
	overflow-x: auto;
	background: var(--hover-background-color, #f5f5f5);
	border-radius: 4px;
	font-size: 0.9em;
	line-height: 1.5;
}
</style>
