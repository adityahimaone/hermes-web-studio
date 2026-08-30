import { chromium, expect } from '@playwright/test'

const browser = await chromium.launch({ headless: true })
const baseUrl = process.env.BASE_URL || 'http://127.0.0.1:5173'
try {
  const page = await browser.newPage({ viewport: { width: 1280, height: 900 } })
  await page.route('**/api/sessions**', route => route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ sessions: [{ session_id: 'shell-session', title: 'Shell contract session', updated_at: new Date().toISOString(), pinned: false, archived: false, tags: [] }] }),
  }))
  await page.route('**/api/health/hermes', route => route.fulfill({ status: 200, contentType: 'application/json', body: '{"connected":true}' }))
  await page.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })
  await page.evaluate(() => localStorage.setItem('hermes-theme', 'light'))
  await page.reload({ waitUntil: 'domcontentloaded' })
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light')

  const rail = page.getByTestId('primary-rail')
  await expect(rail).toBeVisible()
  await expect(page.getByTestId('mobile-bottom-nav')).toBeHidden()
  await expect(page.getByTestId('titlebar')).toBeVisible()
  const chatButton = page.getByTestId('primary-navigation').getByRole('button', { name: 'Chat' })
  await expect(chatButton).toHaveAttribute('aria-describedby', /.+/)
  await expect(page.locator(`[id="${await chatButton.getAttribute('aria-describedby')}"]`)).toHaveAttribute('role', 'tooltip')

  const sessionRail = page.getByRole('complementary', { name: 'Recent sessions' })
  await expect(sessionRail).toBeVisible()
  await chatButton.click()
  await expect(sessionRail).toBeHidden()
  await chatButton.click()
  await expect(sessionRail).toBeVisible()

  const actions = page.getByRole('button', { name: 'Actions for Shell contract session' })
  await actions.click()
  await expect(page.getByRole('menu')).toBeVisible()
  await page.getByRole('menuitem', { name: 'Rename conversation' }).press('ArrowDown')
  await expect(page.getByRole('menuitem', { name: 'Duplicate conversation' })).toBeFocused()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('menu')).toBeHidden()
  await actions.click()
  await page.getByRole('menuitem', { name: 'Rename conversation' }).click()
  await expect(page.getByRole('dialog', { name: 'Rename session' })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog', { name: 'Rename session' })).toBeHidden()

  await page.getByRole('button', { name: 'Customize navigation' }).click()
  const navigationDialog = page.getByRole('dialog', { name: 'Customize navigation' })
  await expect(navigationDialog).toBeVisible()
  const tasksToggle = navigationDialog.getByRole('checkbox', { name: /Tasks/ })
  await expect(tasksToggle).toHaveAttribute('aria-checked', 'true')
  await tasksToggle.click()
  await expect(tasksToggle).toHaveAttribute('aria-checked', 'false')
  await navigationDialog.getByRole('button', { name: 'Done' }).click()
  await expect(navigationDialog).toBeHidden()

  const mobile = await browser.newPage({ viewport: { width: 390, height: 844 } })
  await mobile.route('**/api/sessions**', route => route.fulfill({ status: 200, contentType: 'application/json', body: '{"sessions":[]}' }))
  await mobile.goto(`${baseUrl}/`, { waitUntil: 'networkidle' })
  const bottomNav = mobile.getByTestId('mobile-bottom-nav')
  await expect(bottomNav).toBeVisible()
  await expect(bottomNav.getByRole('button', { name: 'Skills' })).toHaveCSS('min-height', '44px')
  await bottomNav.getByRole('button', { name: 'Skills' }).click()
  await expect(mobile.locator('h1')).toHaveText('skills')
  await expect(mobile.getByRole('button', { name: 'Open navigation' })).toBeVisible()
  await mobile.getByRole('button', { name: 'Open navigation' }).click()
  await expect(mobile.getByTestId('primary-rail')).toBeVisible()
  await mobile.setViewportSize({ width: 1280, height: 900 })
  await expect(mobile.getByTestId('primary-rail')).toHaveCSS('width', '72px')
  await expect(mobile.getByRole('button', { name: 'Open navigation' })).toBeHidden()
  console.log('M7 shell browser acceptance passed.')
} finally {
  await browser.close()
}
