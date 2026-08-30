import { chromium, expect } from '@playwright/test'

const browser = await chromium.launch({ headless: true })
try {
  const mobile = await browser.newPage({ viewport: { width: 390, height: 844 } })
  let starts = 0
  let streamRequests = 0
  const startPayloads = []
  await mobile.route('**/api/sessions', (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ sessions: [] }) }))
  await mobile.route('**/api/attachments', (route) => route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify({ id: 'browser-attachment', name: 'note.txt', mime: 'text/plain', size: 12 }) }))
  await mobile.route('**/api/chat/start', (route) => { starts += 1; startPayloads.push(route.request().postDataJSON()); return route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ stream_id: `browser-stream-${starts}`, session_id: 'browser-session' }) }) })
  await mobile.route('**/api/chat/stream*', (route) => {
    streamRequests += 1
    const body = streamRequests === 1
      ? 'id: 1\nevent: token\ndata: {"text":"browser "}\n\n'
      : 'id: 2\nevent: token\ndata: {"text":"reply"}\n\nid: 3\nevent: done\ndata: {"answer":"browser reply"}\n\n'
    return route.fulfill({ status: 200, contentType: 'text/event-stream', body })
  })
  await mobile.goto('http://127.0.0.1:5173/', { waitUntil: 'networkidle' })
  await mobile.getByRole('textbox', { name: 'Message Hermes' }).click()
  await mobile.getByRole('button', { name: 'Open navigation' }).click()
  await mobile.getByRole('button', { name: 'Close navigation' }).click()
  await mobile.getByRole('textbox', { name: 'Message Hermes' }).press('Shift+Enter')
  await mobile.getByRole('button', { name: 'Send message' }).isVisible()
  await mobile.getByRole('textbox', { name: 'Message Hermes' }).fill('keyboard check')
  await mobile.getByRole('textbox', { name: 'Message Hermes' }).press('Enter')
  await mobile.getByText('keyboard check').waitFor()
  await mobile.getByText('browser', { exact: true }).waitFor()
  await mobile.getByRole('textbox', { name: 'Message Hermes' }).fill('queued with file')
  await mobile.locator('input[type="file"]').setInputFiles({ name: 'note.txt', mimeType: 'text/plain', buffer: Buffer.from('queued file') })
  await mobile.getByRole('button', { name: 'Queue message' }).click()
  await mobile.getByText('queued with file').waitFor()
  await expect.poll(() => streamRequests, { timeout: 10000 }).toBeGreaterThan(1)
  await mobile.getByText('browser reply').waitFor()
  await expect.poll(() => starts, { timeout: 10000 }).toBe(2)
  if (startPayloads[1]?.attachment_ids?.[0] !== 'browser-attachment') throw new Error(`queued attachment was not forwarded: ${JSON.stringify(startPayloads)}`)
  await mobile.close()

  const desktop = await browser.newPage({ viewport: { width: 1280, height: 900 } })
  await desktop.goto('http://127.0.0.1:5173/', { waitUntil: 'networkidle' })
  if (!(await desktop.locator('aside').isVisible())) throw new Error('desktop sidebar is not visible')
  if (!(await desktop.getByRole('textbox', { name: 'Message Hermes' }).isVisible())) throw new Error('composer is not visible')
  console.log('M1 browser acceptance passed.')
} finally {
  await browser.close()
}
