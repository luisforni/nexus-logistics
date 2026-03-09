import { test, expect } from '@playwright/test'

const ADMIN_EMAIL = process.env.E2E_EMAIL ?? 'admin@nexus.io'
const ADMIN_PASS  = process.env.E2E_PASS  ?? 'password123'

test.describe('Authentication', () => {
  test('redirects unauthenticated users to /login', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL(/\/login/)
  })

  test('shows error on bad credentials', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder(/email/i).fill('bad@example.com')
    await page.getByPlaceholder(/password/i).fill('wrongpassword')
    await page.getByRole('button', { name: /sign in/i }).click()

    await expect(page.getByText(/invalid credentials/i)).toBeVisible()
  })

  test('logs in successfully and shows dashboard', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder(/email/i).fill(ADMIN_EMAIL)
    await page.getByPlaceholder(/password/i).fill(ADMIN_PASS)
    await page.getByRole('button', { name: /sign in/i }).click()

    await expect(page).toHaveURL('/')
    await expect(page.getByText(/dashboard/i)).toBeVisible()
  })

  test('logs out and returns to login', async ({ page }) => {

    await page.goto('/login')
    await page.getByPlaceholder(/email/i).fill(ADMIN_EMAIL)
    await page.getByPlaceholder(/password/i).fill(ADMIN_PASS)
    await page.getByRole('button', { name: /sign in/i }).click()
    await expect(page).toHaveURL('/')

    await page.getByRole('button', { name: /logout/i }).click()
    await expect(page).toHaveURL(/\/login/)
  })
})
