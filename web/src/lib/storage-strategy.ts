import type { TFunction } from 'i18next'

export const storageStrategyTypes = ['local', 's3', 'tos', 'obs', 'kodo', 'oss', 'cos', 'webdav'] as const

export type StorageStrategyType = (typeof storageStrategyTypes)[number]

export function isStorageStrategyType(value: string): value is StorageStrategyType {
  return (storageStrategyTypes as readonly string[]).includes(value)
}

export function storageStrategyLabel(t: TFunction, type: string): string {
  switch (type) {
    case 'local':
      return t('admin.typeLocal', { defaultValue: '本地存储' })
    case 's3':
      return t('admin.typeS3', { defaultValue: 'S3 兼容存储' })
    case 'tos':
      return t('admin.typeTOS', { defaultValue: '火山引擎 TOS' })
    case 'obs':
      return t('admin.typeOBS', { defaultValue: '华为云 OBS' })
    case 'kodo':
      return t('admin.typeKodo', { defaultValue: '七牛云 Kodo' })
    case 'oss':
      return t('admin.typeOSS', { defaultValue: '阿里云 OSS' })
    case 'cos':
      return t('admin.typeCOS', { defaultValue: '腾讯云 COS' })
    case 'webdav':
      return t('admin.typeWebDAV', { defaultValue: 'WebDAV' })
    default:
      return type
  }
}
