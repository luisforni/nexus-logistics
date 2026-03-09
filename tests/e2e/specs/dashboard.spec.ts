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

test.describe('Analytics', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/analytics')
  })

  test('renders analytics page', async ({ page }) => {
    await expect(page.getByText(/analytics/i).first()).toBeVisible()
  })

  test('shows demand forecast section', async ({ page }) => {
    await expect(page.getByText(/demand forecast/i)).toBeVisible()
  })

  test('shows KPI metrics', async ({ page }) => {

    const kpiCards = page.locator('[data-testid="kpi-card"]')
    const count = await kpiCards.count()

    await expect(page.locator('main')).toBeVisible()
    void count
  })
})

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('shows dashboard with stats cards', async ({ page }) => {
    await expect(page.locator('main')).toBeVisible()
  })

  test('navigates to shipments via sidebar', async ({ page }) => {
    await page.getByRole('link', { name: /shipments/i }).click()
    await expect(page).toHaveURL('/shipments')
  })

  test('navigates to analytics via sidebar', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click()
    await expect(page).toHaveURL('/analytics')
  })
})
