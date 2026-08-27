<template>
	<Section
		subtitle = "Supported job transform steps"
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
				Transformations
			</span>
		</template>

		<NotificationBlock v-if = "loadError" type = "critical" role = "alert" :message = "loadError" class = "list-banner-pad" />

		<NotificationBlock
			v-else-if = "!sortedTypes.length"
			type = "note"
			message = "No transformation types available."
			class = "list-banner-pad"
		/>

		<Tabs
			v-else
			:tabs = "transformationTabs"
			orientation = "vertical"
			:padding = "true"
			:default-tab = "transformationTabs[0]?.id"
		>
			<template v-for = "t in sortedTypes" :key = "t.id" #[`tab-${t.id}`]>
				<p>{{ t.description }}</p>
				<pre class = "yaml-example"><code>{{ t.yamlExample }}</code></pre>
			</template>
		</Tabs>
	</Section>
</template>

<script setup>
	import { computed, onMounted, ref } from 'vue'
	import { HugeiconsIcon } from '@hugeicons/vue'
	import { ShuffleIcon } from '@hugeicons/core-free-icons'
	import Section from 'picocrank/vue/components/Section.vue'
	import Tabs from 'picocrank/vue/components/Tabs.vue'
	import NotificationBlock from 'picocrank/vue/components/NotificationBlock.vue'
	import { getApiClient } from '../api-client.js'

	const types = ref([])
	const loadError = ref('')

	const sortedTypes = computed(() => (
		[...types.value].sort((a, b) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base' }))
	))

	const transformationTabs = computed(() => (
		sortedTypes.value.map((t) => ({
			id: t.id,
			label: t.name,
		}))
	))

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
