export function setupSidebarNavigation (navigation, { showControlPanel = false } = {}) {
	navigation.clearNavigationLinks()

	navigation.addRouterLink('home')
	navigation.addRouterLink('connections')
	navigation.addRouterLink('transformations')

	if (showControlPanel) {
		navigation.addSection('Control Panel', { name: 'nav-control-panel' })
		navigation.addRouterLink('controlPanel', null, {
			description: 'System administration',
		})
	}
}
