export function connectionHealthClass (conn) {
  if (conn?.healthOk) {
    return 'karma-good'
  }
  return 'karma-bad'
}

export function formatConnectionDetails (conn) {
  if (conn.type === 'csv') {
    return conn.importDirectory || '—'
  }
  if (conn.type === 'mysql') {
    const parts = []
    if (conn.host) {
      parts.push(conn.port ? `${conn.host}:${conn.port}` : conn.host)
    }
    if (conn.database) {
      parts.push(conn.database)
    }
    if (conn.table) {
      parts.push(conn.table)
    }
    if (conn.user) {
      parts.push(`user=${conn.user}`)
    }
    return parts.length ? parts.join(' · ') : '—'
  }
  if (conn.type === 'download_csv') {
    return 'Download transformed CSV (no configuration required)'
  }
  if (conn.type === 'firefly_iii') {
    const parts = []
    if (conn.host) {
      parts.push(conn.host)
    }
    if (conn.user) {
      parts.push(`source=${conn.user}`)
    }
    if (conn.database) {
      parts.push(`type_hint=${conn.database}`)
    }
    return parts.length ? parts.join(' · ') : '—'
  }
  return '—'
}

export function connectionBasicRows (conn) {
  return [
    { label: 'Type', value: conn.type },
    { label: 'Details', value: formatConnectionDetails(conn) },
    { label: 'Health', value: conn.healthMessage || '—', health: true },
  ]
}
