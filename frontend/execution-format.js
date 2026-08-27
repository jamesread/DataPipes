export function executionRunKindLabel (runKind) {
  switch (runKind) {
    case 'extract':
      return 'Extract'
    case 'preview':
      return 'Preview'
    case 'step':
      return 'Step'
    case 'pipeline':
      return 'Pipeline'
    case 'load':
      return 'Load'
    default:
      return runKind || '—'
  }
}

export function executionStatusLabel (row) {
  return row.success ? 'Success' : 'Failed'
}

export function executionStatusTagClass (row) {
  return row.success ? 'good' : 'bad'
}

export function formatExecutionStartedAt (startedAt) {
  if (!startedAt) {
    return '—'
  }
  const date = new Date(startedAt)
  if (Number.isNaN(date.getTime())) {
    return startedAt
  }
  return date.toLocaleString()
}
