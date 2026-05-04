import { describe, expect, it } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import en from '../locales/en'
import zh from '../locales/zh'
import { routes } from '../../router'

const SRC_ROOT = join(__dirname, '../..')

function getLocaleValue(locale: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((current, segment) => {
    if (current && typeof current === 'object' && segment in current) {
      return (current as Record<string, unknown>)[segment]
    }
    return undefined
  }, locale)
}

function collectRouteLocaleKeys(): string[] {
  const keys = new Set<string>()

  for (const route of routes) {
    for (const keyName of ['titleKey', 'descriptionKey'] as const) {
      const value = route.meta?.[keyName]
      if (typeof value === 'string' && value.trim()) {
        keys.add(value)
      }
    }
  }

  return [...keys].sort()
}

function collectStaticTranslationKeysWithoutFallback(): string[] {
  const keys = new Set<string>()
  const translateCallPattern = /\bt\(\s*(['"`])([A-Za-z][A-Za-z0-9_]*(?:\.[A-Za-z0-9_-]+)+)\1\s*([,)])/g

  for (const file of collectSourceFiles(SRC_ROOT)) {
    const content = readFileSync(file, 'utf8')
    let match: RegExpExecArray | null
    while ((match = translateCallPattern.exec(content))) {
      const key = match[2]
      const terminator = match[3]
      if (terminator === ',') continue
      keys.add(`${key} (${relative(SRC_ROOT, file)})`)
    }
  }

  return [...keys].sort()
}

function collectSourceFiles(dir: string): string[] {
  const files: string[] = []

  for (const entry of readdirSync(dir)) {
    const fullPath = join(dir, entry)
    const stat = statSync(fullPath)
    if (stat.isDirectory()) {
      if (entry === 'node_modules' || entry === 'dist') continue
      files.push(...collectSourceFiles(fullPath))
      continue
    }

    if (/\.(ts|vue)$/.test(entry) && !entry.endsWith('.spec.ts')) {
      files.push(fullPath)
    }
  }

  return files
}

function stripSourceLocation(value: string): string {
  return value.replace(/\s+\(.+\)$/, '')
}

describe('admin subscription product locale keys', () => {
  it('contains zh labels for the subscription products admin page', () => {
    expect(zh.admin.subscriptionProducts.title).toBe('产品订阅')
    expect(zh.admin.subscriptionProducts.description).toBe('管理共享订阅产品、分组绑定和用户产品订阅')
    expect(zh.admin.subscriptionProducts.createProduct).toBe('创建产品')
    expect(zh.admin.subscriptionProducts.columns.defaultValidity).toBe('有效天数')
  })

  it('contains en labels for the subscription products admin page', () => {
    expect(en.admin.subscriptionProducts.title).toBe('Product Subscriptions')
    expect(en.admin.subscriptionProducts.description).toBe('Manage shared subscription products, group bindings, and user product subscriptions')
    expect(en.admin.subscriptionProducts.createProduct).toBe('Create Product')
    expect(en.admin.subscriptionProducts.columns.defaultValidity).toBe('Validity Days')
  })

  it('contains zh and en labels for every route title and description key', () => {
    const routeLocaleKeys = collectRouteLocaleKeys()
    const missingZhKeys: string[] = []
    const missingEnKeys: string[] = []

    expect(routeLocaleKeys.length).toBeGreaterThan(0)

    for (const key of routeLocaleKeys) {
      if (typeof getLocaleValue(zh, key) !== 'string') missingZhKeys.push(key)
      if (typeof getLocaleValue(en, key) !== 'string') missingEnKeys.push(key)
    }

    expect(missingZhKeys).toEqual([])
    expect(missingEnKeys).toEqual([])
  })

  it('contains zh and en labels for static translation calls without fallbacks', () => {
    const translationKeys = collectStaticTranslationKeysWithoutFallback()
    const missingZhKeys: string[] = []
    const missingEnKeys: string[] = []

    expect(translationKeys.length).toBeGreaterThan(0)

    for (const keyWithSource of translationKeys) {
      const key = stripSourceLocation(keyWithSource)
      if (typeof getLocaleValue(zh, key) !== 'string') missingZhKeys.push(keyWithSource)
      if (typeof getLocaleValue(en, key) !== 'string') missingEnKeys.push(keyWithSource)
    }

    expect(missingZhKeys).toEqual([])
    expect(missingEnKeys).toEqual([])
  })
})
