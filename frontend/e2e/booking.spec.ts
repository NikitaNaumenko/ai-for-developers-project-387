import { expect, test } from '@playwright/test';

test('books a meeting through the main scheduling flow', async ({ page }) => {
  const suffix = `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  const eventTitle = `Playwright booking ${suffix}`;
  const eventDescription = `End-to-end booking flow ${suffix}`;
  const guestName = `Guest ${suffix}`;
  const guestEmail = `guest-${suffix}@example.com`;
  const guestNote = `Please prepare the booking notes ${suffix}`;

  await page.goto('/');

  await page.locator('.side-link', { hasText: 'Event types' }).click();
  await expect(page.getByRole('heading', { name: 'Event types' })).toBeVisible();

  await page.getByLabel('Title').fill(eventTitle);
  await page.getByLabel('Description').fill(eventDescription);
  await page.getByLabel('Duration').fill('30');
  await page.getByRole('button', { name: 'Create event type' }).click();

  await expect(page.getByText(eventTitle)).toBeVisible();

  await page.locator('.side-link', { hasText: 'Booking page' }).click();
  await expect(page.getByRole('heading', { name: 'Booking page' })).toBeVisible();

  await page.getByRole('button', { name: new RegExp(eventTitle) }).click();
  const firstSlot = page.locator('.slot-button').first();
  await expect(firstSlot).toBeVisible();
  await firstSlot.click();

  await page.getByLabel('Name').fill(guestName);
  await page.getByLabel('Email').fill(guestEmail);
  await page.getByLabel('Note').fill(guestNote);

  const bookingResponse = page.waitForResponse(
    (response) => response.url().includes('/api/bookings') && response.status() === 201,
  );
  await page.getByRole('button', { name: 'Confirm booking' }).click();
  await bookingResponse;

  await expect(page.getByText('Booking confirmed')).toBeVisible();

  await page.locator('.side-link', { hasText: 'Bookings' }).click();
  await expect(page.getByRole('heading', { name: 'Upcoming bookings' })).toBeVisible();
  await expect(page.getByText(eventTitle)).toBeVisible();
  await expect(page.getByText(`${guestName} · ${guestEmail}`)).toBeVisible();
  await expect(page.getByText(guestNote)).toBeVisible();
});
