function routeCrumb (routeName, title, params = undefined) {
	return {
		name: title,
		to: params ? { name: routeName, params } : { name: routeName },
	}
}

export function jobsBreadcrumbs () {
	return () => [routeCrumb('home', 'Jobs')]
}

export function jobDetailBreadcrumbs (route) {
	return [
		routeCrumb('home', 'Jobs'),
		routeCrumb('job-detail', route.params.id, { id: route.params.id }),
	]
}

export function connectionsBreadcrumbs () {
	return () => [routeCrumb('connections', 'Connections')]
}

export function connectionDetailBreadcrumbs (route) {
	return [
		routeCrumb('connections', 'Connections'),
		routeCrumb('connection-detail', route.params.id, { id: route.params.id }),
	]
}

export function transformationsBreadcrumbs () {
	return () => [routeCrumb('transformations', 'Transformations')]
}

export function controlPanelBreadcrumbs () {
	return () => [routeCrumb('controlPanel', 'Control Panel')]
}

export function diagnosticsBreadcrumbs () {
	return () => [
		routeCrumb('controlPanel', 'Control Panel'),
		routeCrumb('diagnostics', 'Diagnostics'),
	]
}
