export const supportedLocales = [
  'en', 'id', 'de', 'es', 'fr', 'it', 'ja', 'ko', 'pt-BR', 'ru', 'zh-CN', 'zh-TW', 'ar', 'hi', 'tr',
] as const

export type Locale = typeof supportedLocales[number]

export const localeLabels: Record<Locale, string> = {
  en: 'English', id: 'Bahasa Indonesia', de: 'Deutsch', es: 'Español', fr: 'Français', it: 'Italiano',
  ja: '日本語', ko: '한국어', 'pt-BR': 'Português (Brasil)', ru: 'Русский', 'zh-CN': '简体中文',
  'zh-TW': '繁體中文', ar: 'العربية', hi: 'हिन्दी', tr: 'Türkçe',
}

export const fallbackLocale: Locale = 'en'

export function resolveLocale(value: string | null | undefined): Locale {
  return supportedLocales.includes(value as Locale) ? value as Locale : fallbackLocale
}
