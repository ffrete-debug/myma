import { describe, it, expect } from 'vitest'
import en from '../../messages/en'
import zh from '../../messages/zh'
import ptBR from '../../messages/pt-BR'

type MessageTree = { [key: string]: string | MessageTree }

/**
 * Flattens a message catalogue into dot-separated leaf keys, e.g.
 * `{ servers: { card: { running: 'Running' } } }` -> `['servers.card.running']`.
 */
function flattenKeys(tree: MessageTree, prefix = ''): string[] {
  const keys: string[] = []
  for (const [key, value] of Object.entries(tree)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'string') keys.push(path)
    else keys.push(...flattenKeys(value, path))
  }
  return keys
}

const catalogues: Record<string, MessageTree> = {
  en: en as MessageTree,
  zh: zh as MessageTree,
  'pt-BR': ptBR as MessageTree
}

const keysByLocale = Object.fromEntries(
  Object.entries(catalogues).map(([locale, tree]) => [locale, flattenKeys(tree)])
) as Record<string, string[]>

const enKeys = keysByLocale.en
const translations = Object.keys(catalogues).filter((locale) => locale !== 'en')

describe('message catalogues', () => {
  it('has no duplicate keys within a catalogue', () => {
    for (const [locale, keys] of Object.entries(keysByLocale)) {
      expect(new Set(keys).size, `${locale} has duplicate keys`).toBe(keys.length)
    }
  })

  it.each(translations)('%s is missing no key present in en', (locale) => {
    const localeKeys = new Set(keysByLocale[locale])
    const missing = enKeys.filter((key) => !localeKeys.has(key))
    expect(missing, `${locale} is missing keys present in en`).toEqual([])
  })

  it.each(translations)('%s has no key absent from en', (locale) => {
    const enKeySet = new Set(enKeys)
    const extra = keysByLocale[locale].filter((key) => !enKeySet.has(key))
    expect(extra, `${locale} has keys that do not exist in en`).toEqual([])
  })

  it('has no empty message values', () => {
    for (const [locale, tree] of Object.entries(catalogues)) {
      const empty = flattenKeys(tree).filter((path) => {
        const value = path
          .split('.')
          .reduce<string | MessageTree>((node, part) => (node as MessageTree)[part], tree)
        return typeof value === 'string' && value.trim() === ''
      })
      expect(empty, `${locale} has empty message values`).toEqual([])
    }
  })
})
