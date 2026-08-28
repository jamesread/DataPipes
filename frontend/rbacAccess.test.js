import assert from 'node:assert/strict'
import { test } from 'node:test'

import {
	canAccessControlPanelFromStatus,
	canAccessDiagnosticsFromStatus,
} from './rbacAccess.js'

test('control panel is hidden when logged out', () => {
	assert.equal(canAccessControlPanelFromStatus({ isLoggedIn: false, rbacIsSuperuser: true }), false)
})

test('control panel is hidden without privilege', () => {
	assert.equal(canAccessControlPanelFromStatus({
		isLoggedIn: true,
		rbacIsSuperuser: false,
		rbacPermissions: [],
	}), false)
})

test('superuser can open the control panel', () => {
	assert.equal(canAccessControlPanelFromStatus({
		isLoggedIn: true,
		rbacIsSuperuser: true,
		rbacPermissions: [],
	}), true)
})

test('diagnostics permission unlocks the control panel', () => {
	assert.equal(canAccessControlPanelFromStatus({
		isLoggedIn: true,
		rbacIsSuperuser: false,
		rbacPermissions: ['system.diagnostics'],
	}), true)
})

test('diagnostics tile follows the same privilege', () => {
	assert.equal(canAccessDiagnosticsFromStatus({
		isLoggedIn: true,
		rbacIsSuperuser: false,
		rbacPermissions: ['system.diagnostics'],
	}), true)
	assert.equal(canAccessDiagnosticsFromStatus({
		isLoggedIn: true,
		rbacIsSuperuser: false,
		rbacPermissions: [],
	}), false)
})
