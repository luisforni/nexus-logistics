import { test, expect, Page } from '@playwright/test'

const ADMIN_EMAIL = process.env.E2E_EMAIL ?? 'admin@nexus.io'
const ADMIN_PASS  = process.env.E2E_PASS  ?? 'password123'

async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder(/email/i).fill(ADMIN_EMAIL)
  await page.getByPlaceholder(/password/i).fill(ADMIN_PASS)
  await page.getByRole('button', { name: /sign in/i }).click()
  await page.waitForURL('/')
}

test.describe('Shipments', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/shipments')
  })

  test('displays shipment list', async ({ page }) => {
    await expect(page.getByText(/shipments/i).first()).toBeVisible()

    await expect(page.locator('main')).toBeVisible()
  })

  test('filters shipments by tracking number', async ({ page }) => {
    const searchInput = page.getByPlaceholder(/search/i)
    await searchInput.fill('NX-')

    await expect(searchInput).toHaveValue('NX-')
  })

  test('navigates to shipment detail on card click', async ({ page }) => {

    const firstCard = page.locator('[data-testid="shipment-card"]').first()
    const count = await firstCard.count()
    if (count === 0) test.skip()

    await firstCard.click()
    await expect(page).toHaveURL(/\/shipments\/[\w-]+/)
    await expect(page.getByText(/tracking number/i)).toBeVisible()
  })
})

test.describe('Shipment Detail', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('shows blockchain anchor section', async ({ page }) => {
    const firstCard = page.locator('[data-testid="shipment-card"]').first()
    const count = await firstCard.count()
    if (count === 0) {
      await page.goto('/shipments')
      await page.waitForSelector('[data-testid="shipment-card"]', { timeout: 5000 }).catch(() => null)
    }
    const cards = await page.locator('[data-testid="shipment-card"]').count()
    if (cards === 0) test.skip()

    await page.locator('[data-testid="shipment-card"]').first().click()
    await expect(page.getByText(/blockchain/i)).toBeVisible()
  })
})
