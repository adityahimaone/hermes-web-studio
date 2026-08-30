import { chromium, expect } from '@playwright/test'
import { mkdir } from 'node:fs/promises'

const baseUrl = process.env.BASE_URL || 'http://127.0.0.1:5173'
const outputDir = process.env.MATRIX_DIR || '/tmp/hermes-web-studio-mvp-matrix'
const viewports = [
  { name: 'desktop', width: 1440, height: 1000 },
  { name: 'laptop', width: 1280, height: 800 },
  { name: 'tablet', width: 1024, height: 768 },
  { name: 'narrow', width: 768, height: 1024 },
  { name: 'mobile', width: 390, height: 844 },
]

await mkdir(outputDir, { recursive: true })
const browser = await chromium.launch({ headless: true })
try {
  for (const viewport of viewports) {
    const context = await browser.newContext({ viewport, serviceWorkers: 'block' })
    const page = await context.newPage()
    await page.route('**/api/sessions**', route => route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ sessions: [{ session_id: 'matrix-session', title: 'Sanitized matrix session', updated_at: '2026-01-01T00:00:00Z', pinned: false, archived: false, tags: [] }] }),
    }))
    await page.route('**/api/health/hermes', route => route.fulfill({ status: 200, contentType: 'application/json', body: '{"connected":true}' }))
    await page.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })
    await expect(page.getByTestId('titlebar')).toBeVisible()
    if (viewport.width >= 1024) await expect(page.getByRole('complementary', { name: 'Recent sessions' })).toBeVisible()
    else await expect(page.getByRole('complementary', { name: 'Recent sessions' })).toBeHidden()
    if (viewport.width < 1024) await expect(page.getByTestId('mobile-bottom-nav')).toBeVisible()
    else await expect(page.getByTestId('mobile-bottom-nav')).toBeHidden()
    await expect(page.getByRole('textbox', { name: 'Message Hermes' })).toBeVisible()
    const geometry = await page.evaluate(() => {
      const pick = (selector) => {
        const node = document.querySelector(selector)
        if (!node) return null
        const box = node.getBoundingClientRect()
        const style = getComputedStyle(node)
        return { display: style.display, width: Math.round(box.width), height: Math.round(box.height), fontSize: style.fontSize }
      }
      return { rail: pick('[data-testid="primary-rail"]'), titlebar: pick('[data-testid="titlebar"]'), composer: pick('[aria-label="Message Hermes"]'), bottomNav: pick('[data-testid="mobile-bottom-nav"]') }
    })
    await page.screenshot({ path: `${outputDir}/${viewport.name}.png`, fullPage: true })
    await import('node:fs/promises').then(fs => fs.writeFile(`${outputDir}/${viewport.name}.json`, JSON.stringify({ viewport, geometry }, null, 2)))
    await context.close()
  }
  console.log(`MVP shell matrix passed for ${viewports.length} viewports: ${outputDir}`)
} finally {
  await browser.close()
}
