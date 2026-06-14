import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { DataCleanerService } from './resources/javascript/gen/data_cleaner/api/v1/data_cleaner_pb'

let apiClient

export function getApiClient () {
  if (!apiClient) {
    const transport = createConnectTransport({ baseUrl: '/api' })
    apiClient = createClient(DataCleanerService, transport)
  }
  return apiClient
}

export function formatRpcError (err) {
  if (!err) {
    return 'unknown error'
  }
  const message = err.rawMessage || err.message
  if (message) {
    return message
  }
  if (typeof err.code === 'number') {
    return `RPC failed (code ${err.code})`
  }
  return String(err)
}

export const moneyFormatter = new Intl.NumberFormat('en-GB', {
  style: 'currency',
  currency: 'GBP',
})

export function formatJobMeta (job) {
  const parts = []
  if (job.extractConnection) {
    parts.push(`extract: ${job.extractConnection}`)
  }
  if (job.importDirectory) {
    parts.push(job.importDirectory)
  }
  if (job.loadConnection) {
    parts.push(`load → ${job.loadConnection}`)
  }
  if (!job.loadConfigured) {
    parts.push('load not configured')
  }
  return parts.join(' · ')
}

export function previewColumnNames (columnMap) {
  if (!columnMap?.length) {
    return []
  }
  return columnMap.map((e) => e.loadColumn)
}

export function previewHeaderLabels (columnMap) {
  if (!columnMap?.length) {
    return []
  }
  return columnMap.map((e) => {
    if (e.sourceColumn && e.sourceColumn !== e.loadColumn) {
      return `${e.loadColumn} (${e.sourceColumn})`
    }
    return e.loadColumn
  })
}

export function extractPreviewHeaderLabels (extractColumns) {
  if (!extractColumns?.length) {
    return []
  }
  return extractColumns.map((e) => {
    if (e.columnIndex >= 0) {
      return `${e.fieldName} (${e.columnIndex})`
    }
    return e.fieldName
  })
}

const moneyFieldPattern = /(?:^|_)(amount|value|balance|total|price)(?:_|$)/i

export function formatExtractCell (value, fieldName) {
  if (moneyFieldPattern.test(fieldName)) {
    const n = Number(value)
    if (!Number.isNaN(n) && value !== '') {
      return moneyFormatter.format(n)
    }
  }
  return value
}

export function formatPreviewCell (value, columnName) {
  if (moneyFieldPattern.test(columnName)) {
    const n = Number(value)
    if (!Number.isNaN(n) && value !== '') {
      return moneyFormatter.format(n)
    }
  }
  return value
}

export function isDownloadCsvLoad (loadConnection) {
  return loadConnection === 'download_csv'
}

export function jobDownloadCsvUrl (jobId) {
  return `/api/download/${encodeURIComponent(jobId)}.csv`
}

export function isFirstTransformationOrdinal (transformations, index) {
  const ordinal = transformations[index]?.ordinal
  if (!ordinal) {
    return false
  }
  return transformations.findIndex((t) => t.ordinal === ordinal) === index
}

export function formatLoadStats (stats) {
  if (!stats?.length) {
    return '—'
  }
  return stats.map((s) => `${s.key}: ${s.value}`).join(' · ')
}

export async function streamLoad (jobId, onProgress, signal) {
  const client = getApiClient()
  for await (const event of client.streamLoad({ jobId }, { signal })) {
    if (onProgress) {
      onProgress(event)
    }
  }
}
