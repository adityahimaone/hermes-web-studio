import { chromium, expect } from '@playwright/test'

const browser = await chromium.launch({ headless: true })
const baseUrl = process.env.BASE_URL || 'http://127.0.0.1:5173'
const sessions = [
  { session_id: 'session-alpha', title: 'Alpha fixture', updated_at: '2026-08-30T10:00:00Z', pinned: false, archived: false, tags: [] },
  { session_id: 'session-beta', title: 'Beta fixture', updated_at: '2026-08-30T09:00:00Z', pinned: false, archived: false, tags: [] },
  { session_id: 'session-gamma', title: 'Gamma fixture', updated_at: '2026-08-29T09:00:00Z', pinned: false, archived: false, tags: [] },
]
try {
  const context = await browser.newContext({ viewport: { width: 1280, height: 900 }, serviceWorkers: 'block' })
  const page = await context.newPage()
  await page.route('**/api/sessions**', route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ sessions }) }))
  await page.route('**/api/health/hermes', route => route.fulfill({ status: 200, contentType: 'application/json', body: '{"connected":true}' }))
  await page.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })

  const alpha = page.locator('[data-session-id="session-alpha"]')
  const beta = page.locator('[data-session-id="session-beta"]')
  const gamma = page.locator('[data-session-id="session-gamma"]')
  await expect(alpha).toBeVisible()
  await alpha.focus()
  await alpha.press('ArrowDown')
  await expect(beta).toBeFocused()
  await beta.press('ArrowDown')
  await expect(gamma).toBeFocused()
  await gamma.press('ArrowDown')
  await expect(alpha).toBeFocused()
  await alpha.press('ArrowUp')
  await expect(gamma).toBeFocused()
  await gamma.press('Enter')
  await expect(gamma).toHaveAttribute('aria-current', 'page')
  console.log('Session keyboard browser acceptance passed.')
} finally {
  await browser.close()
}
