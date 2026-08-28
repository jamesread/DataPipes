package rbac

import "testing"

func TestCanAccessControlPanel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		loggedIn  bool
		superuser bool
		perms     []string
		want      bool
	}{
		{name: "logged out", want: false},
		{name: "logged in without privilege", loggedIn: true, want: false},
		{name: "superuser", loggedIn: true, superuser: true, want: true},
		{name: "diagnostics permission", loggedIn: true, perms: []string{PermissionSystemDiagnostics}, want: true},
		{name: "unrelated permission", loggedIn: true, perms: []string{"app.access"}, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := CanAccessControlPanel(tc.loggedIn, tc.superuser, tc.perms)
			if got != tc.want {
				t.Fatalf("CanAccessControlPanel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanAccessDiagnostics(t *testing.T) {
	t.Parallel()

	if CanAccessDiagnostics(true, false, nil) {
		t.Fatal("expected no diagnostics access without privilege")
	}
	if !CanAccessDiagnostics(true, true, nil) {
		t.Fatal("expected superuser diagnostics access")
	}
	if !CanAccessDiagnostics(true, false, []string{PermissionSystemDiagnostics}) {
		t.Fatal("expected diagnostics permission to unlock the page")
	}
}
