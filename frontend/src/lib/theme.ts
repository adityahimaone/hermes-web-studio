export const skins = ['default', 'ares', 'catppuccin', 'charizard', 'codex', 'geist-contrast', 'github', 'graphite', 'hepburn', 'mono', 'neon', 'neon-paint', 'neon-soft', 'nous', 'poseidon', 'sienna', 'sisyphus', 'slate', 'terracotta', 'verdigris', 'zeus'] as const
export type ThemePreference = 'system' | 'dark' | 'light'

export function applyTheme(theme: ThemePreference, skin: string) {
  const root = document.documentElement
  const resolved = theme === 'system' ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark') : theme
  root.dataset.theme = resolved
  root.dataset.skin = skins.includes(skin as typeof skins[number]) ? skin : 'default'
  root.style.colorScheme = resolved
  root.classList.toggle('dark', resolved === 'dark')
}
