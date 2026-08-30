import { chromium, expect } from '@playwright/test'

const browser = await chromium.launch({ headless: true })
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 } })
  for (const endpoint of ['/api/sessions', '/api/onboarding', '/api/auth/me', '/api/profiles']) {
    await page.route(`**${endpoint}`, route => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(endpoint === '/api/sessions' ? { sessions: [] } : endpoint === '/api/profiles' ? { profiles: [{ id: 'default', name: 'Default' }], active: 'default' } : endpoint === '/api/onboarding' ? { configured: false } : { authenticated: false }) }))
  }
  await page.goto('http://127.0.0.1:5173/', { waitUntil: 'networkidle' })
  await page.getByRole('button', { name: 'Tasks' }).click()
  await expect(page.getByRole('heading', { name: 'tasks' })).toBeVisible()
  await page.getByRole('button', { name: 'Skills' }).click()
  await expect(page.getByRole('heading', { name: 'skills' })).toBeVisible()
  await page.getByRole('button', { name: 'Settings' }).click()
  await expect(page.getByRole('heading', { name: 'Preferences' })).toBeVisible()
  console.log('M5 control-center browser acceptance passed.')
} finally { await browser.close() }
