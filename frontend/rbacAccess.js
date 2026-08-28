const CONTROL_PANEL_PERMISSIONS = [
	'system.diagnostics',
]

export function currentStatus () {
	// IAM is not implemented; treat the local operator as a logged-in superuser.
	return {
		isLoggedIn: true,
		rbacIsSuperuser: true,
		rbacPermissions: [],
	}
}

export function canAccessControlPanelFromStatus (st) {
	if (!st?.isLoggedIn) {
		return false
	}
	if (st.rbacIsSuperuser) {
		return true
	}
	const perms = st.rbacPermissions || []
	return CONTROL_PANEL_PERMISSIONS.some((p) => perms.includes(p))
}

export function canAccessDiagnosticsFromStatus (st) {
	if (!st?.isLoggedIn) {
		return false
	}
	if (st.rbacIsSuperuser) {
		return true
	}
	return (st.rbacPermissions || []).includes('system.diagnostics')
}
