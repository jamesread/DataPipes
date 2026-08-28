package rbac

const PermissionSystemDiagnostics = "system.diagnostics"

var controlPanelPermissions = []string{
	PermissionSystemDiagnostics,
}

func CanAccessControlPanel(loggedIn, superuser bool, perms []string) bool {
	if !loggedIn {
		return false
	}
	if superuser {
		return true
	}
	return hasAnyPermission(perms, controlPanelPermissions)
}

func CanAccessDiagnostics(loggedIn, superuser bool, perms []string) bool {
	if !loggedIn {
		return false
	}
	if superuser {
		return true
	}
	return hasPermission(perms, PermissionSystemDiagnostics)
}

func hasAnyPermission(have, want []string) bool {
	for _, name := range want {
		if hasPermission(have, name) {
			return true
		}
	}
	return false
}

func hasPermission(perms []string, name string) bool {
	for _, p := range perms {
		if p == name {
			return true
		}
	}
	return false
}
