import { chromium, expect } from '@playwright/test'
import { mkdir, writeFile } from 'node:fs/promises'

const baseUrl = process.env.BASE_URL || 'http://127.0.0.1:5173'
const outputDir = process.env.MATRIX_DIR || '/tmp/hermes-web-studio-p057-visual'
const viewports = [
  { name: 'desktop', width: 1440, height: 1000 },
  { name: 'laptop', width: 1280, height: 800 },
  { name: 'tablet', width: 1024, height: 768 },
  { name: 'mobile', width: 390, height: 844 },
]
const themes = ['dark', 'light', 'system']
const skins = ['default', 'ares', 'catppuccin', 'charizard', 'codex', 'geist-contrast', 'github', 'graphite', 'hepburn', 'mono', 'neon', 'neon-paint', 'neon-soft', 'nous', 'poseidon', 'sienna', 'sisyphus', 'slate', 'terracotta', 'verdigris', 'zeus']

await mkdir(outputDir, { recursive: true })
const browser = await chromium.launch({ headless: true })
try {
  const context = await browser.newContext({ viewport: viewports[0], serviceWorkers: 'block' })
  const page = await context.newPage()
  await page.route('**/api/sessions**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ sessions: [{ session_id: 'visual-session', title: 'Sanitized visual session', updated_at: '2026-01-01T00:00:00Z', pinned: false, archived: false, tags: [] }] }) }))
  await page.route('**/api/health/hermes', route => route.fulfill({ status: 200, contentType: 'application/json', body: '{"connected":true}' }))
  await page.route('**/api/preferences', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ preferences: { theme: 'dark', skin: 'default', locale: 'en' } }) }))

  const rows = []
  for (const viewport of viewports) {
    await page.setViewportSize(viewport)
    await page.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })
    for (const theme of themes) for (const skin of skins) {
      await page.evaluate(({ theme, skin }) => { document.documentElement.dataset.theme = theme; document.documentElement.dataset.skin = skin }, { theme, skin })
      await expect(page.getByTestId('titlebar')).toBeVisible()
      await expect(page.getByRole('textbox', { name: 'Message Hermes' })).toBeVisible()
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth || document.body.scrollWidth > window.innerWidth)
      if (overflow) throw new Error(`horizontal overflow: ${viewport.name}/${theme}/${skin}`)
      const slug = `${viewport.name}-${theme}-${skin}`
      const geometry = await page.evaluate(() => Object.fromEntries(['[data-testid="titlebar"]', '[aria-label="Message Hermes"]', '[data-testid="primary-rail"]', '[data-testid="mobile-bottom-nav"]'].map(selector => { const node = document.querySelector(selector); if (!node) return [selector, null]; const box = node.getBoundingClientRect(); return [selector, { width: Math.round(box.width), height: Math.round(box.height), display: getComputedStyle(node).display }] })))
      await page.screenshot({ path: `${outputDir}/${slug}.png`, fullPage: true })
      rows.push({ viewport, theme, skin, geometry })
    }
  }
  await writeFile(`${outputDir}/matrix.json`, JSON.stringify({ rows, note: 'Local deterministic screenshots and geometry only.' }, null, 2))
  await context.close()
  console.log(`P057 local visual matrix passed: ${rows.length} rows in ${outputDir}`)
} finally { await browser.close() }
