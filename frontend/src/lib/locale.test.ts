import { describe, expect, it } from 'vitest'
import { fallbackLocale, resolveLocale, supportedLocales } from './locale'

describe('locale contract', () => {
  it('exposes the fifteen reference locales', () => expect(supportedLocales).toHaveLength(15))
  it('falls back safely for unknown locales', () => expect(resolveLocale('xx')).toBe(fallbackLocale))
})
