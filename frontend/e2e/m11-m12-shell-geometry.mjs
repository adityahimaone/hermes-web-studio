import { chromium, expect } from '@playwright/test'

const baseUrl = process.env.BASE_URL || 'http://127.0.0.1:5173'
const desktop = { width: 1280, height: 800 }
const utilityViews = [
  { label: 'Skills', view: 'skills', heading: 'skills', rail: 'Context skills navigation' },
  { label: 'Memory', view: 'memory', heading: 'memory', rail: 'Context memory navigation' },
  { label: 'Settings', view: 'settings', heading: 'Settings', rail: 'Context Settings navigation' },
]

function json(route, body, status = 200) {
  return route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })
}

async function mockRuntime(page) {
  await page.route('**/api/**', route => {
    const url = new URL(route.request().url())
    const path = url.pathname
    if (path === '/api/sessions') return json(route, { sessions: [{ session_id: 'm11-m12-session', title: 'Sanitized shell session', updated_at: '2026-01-01T00:00:00Z', pinned: false, archived: false, tags: [] }] })
    if (path === '/api/health/hermes') return json(route, { connected: true })
    if (path === '/api/skills') return url.searchParams.has('name') ? json(route, { name: 'apple/apple-notes/SKILL.md', content: 'A'.repeat(3200) }) : json(route, { skills: [{ name: 'sample-skill', path: 'sample-skill/SKILL.md', description: 'Sanitized skill fixture' }, { name: 'apple-notes', path: 'apple/apple-notes/SKILL.md' }, { name: 'apple-reminders', path: 'apple/apple-reminders/SKILL.md' }] })
    if (path === '/api/memory') return json(route, { notes: [{ name: 'MEMORY.md', path: 'MEMORY.md' }] })
    if (path === '/api/profiles') return json(route, { profiles: [{ id: 'default', name: 'Default', model: 'default', health: 'ready' }], active: 'default' })
    if (path === '/api/preferences') return json(route, { preferences: { theme: 'dark', skin: 'default', locale: 'en' } })
    if (path === '/api/settings/capabilities') return json(route, { sections: ['preferences', 'runtime'] })
    if (path === '/api/capabilities') return json(route, { preferences: true, voice: false, extensions: false, background: false })
    if (path === '/api/plugins') return json(route, { plugins: [] })
    if (path === '/api/operator/health') return json(route, { control: true, workspace: true })
    if (path === '/api/operator/logs') return json(route, { available: true, entries: [] })
    return json(route, {})
  })
  await page.route('**/ready', route => json(route, { ready: true }))
}

async function box(page, selector) {
  return page.locator(selector).evaluate(node => {
    const rect = node.getBoundingClientRect()
    return { width: Math.round(rect.width), height: Math.round(rect.height), overflowY: getComputedStyle(node).overflowY }
  })
}

const browser = await chromium.launch({ headless: true })
try {
  const context = await browser.newContext({ viewport: desktop, serviceWorkers: 'block' })
  const page = await context.newPage()
  await mockRuntime(page)

  await page.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })
  const primaryRail = await box(page, '[data-testid="primary-rail"]')
  const titlebar = await box(page, '[data-testid="titlebar"]')
  expect(primaryRail.width).toBe(64)
  expect(primaryRail.height).toBe(desktop.height)
  expect(titlebar.height).toBe(64)
  await expect(page.getByTestId('mobile-bottom-nav')).toBeHidden()

  for (const target of utilityViews) {
    await page.getByRole('button', { name: target.label, exact: true }).click()
    await expect(page.locator('h1').first()).toHaveText(target.heading)
    const rail = page.getByRole('complementary', { name: target.rail })
    await expect(rail).toBeVisible()
    const railBox = await rail.boundingBox()
    expect(railBox?.width ?? 0).toBeGreaterThanOrEqual(240)
    expect(railBox?.width ?? 0).toBeLessThanOrEqual(552)
    await expect(rail.locator('.context-rail__body')).toHaveCSS('overflow-y', target.view === 'skills' ? 'hidden' : 'auto')
    await expect(rail.locator('.context-rail__header')).toBeVisible()

    if (target.view === 'skills') {
      await expect(rail.getByRole('button', { name: 'Add skill' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'New skill' })).toBeHidden()
      await expect(rail.locator('.skills-rail__list')).toHaveCSS('overflow-y', 'scroll')
      await expect(rail.locator('.skills-rail__list')).toContainText('(general)')
      await expect(rail.locator('.skills-rail__list')).toContainText('apple')
      await rail.getByRole('button', { name: 'apple (2)' }).click()
      await expect(rail.getByRole('button', { name: 'apple (2)' })).toHaveAttribute('aria-expanded', 'true')
      await rail.getByRole('button', { name: 'apple-notes' }).click()
      await expect(page.locator('.skill-preview')).toBeVisible()
      const content = await page.locator('.discovery-content').boundingBox()
      expect(content?.width ?? 0).toBeGreaterThan(700)
    }

    const collapse = page.getByRole('button', { name: `Collapse ${target.label} sidebar` })
    await collapse.click()
    await expect(rail).toBeHidden()
    await expect(page.getByRole('button', { name: `Expand ${target.label} sidebar` })).toBeVisible()
    await page.getByRole('button', { name: `Expand ${target.label} sidebar` }).click()
    await expect(rail).toBeVisible()
  }

  await page.getByRole('button', { name: 'Settings', exact: true }).click()
  await expect(page.getByRole('complementary', { name: 'Context Settings navigation' })).toBeVisible()
  const settingsSearch = page.getByRole('textbox', { name: 'Search settings' }).first()
  await settingsSearch.fill('appearance')
  await expect(page.getByRole('heading', { name: 'Appearance' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Capability status' })).toBeHidden()

  await context.close()
  console.log('Design contract shell checks passed: responsive rail, utility views, single scroll owner, collapse, desktop mobile-nav boundary, and settings capability filtering.')
} finally {
  await browser.close()
}
