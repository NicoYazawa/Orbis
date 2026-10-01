import { expect, test } from "@playwright/test";
import { mkdir } from "node:fs/promises";

test("same-origin status page reflects the API", async ({ page, request }) => {
  const system = await request.get("/api/v1/system");
  expect(system.ok()).toBeTruthy();
  const data = await system.json();

  await page.goto("/");
  await expect(page.getByRole("heading", { name: /Orbis/ })).toBeVisible();
  await expect(page.getByText(data.version, { exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "刷新状态" })).toBeVisible();
  await mkdir("test-results", { recursive: true });
  await page.screenshot({ path: "test-results/phase0-status.png", fullPage: true });
});

test("browser can fetch a presigned object when one is supplied", async ({ page }) => {
  test.skip(!process.env.ORBIS_S3_PRESIGN_URL, "Set ORBIS_S3_PRESIGN_URL after creating a real presigned object");
  await page.goto("/");
  const result = await page.evaluate(async (url) => {
    const response = await fetch(url);
    return { ok: response.ok, status: response.status, size: (await response.arrayBuffer()).byteLength };
  }, process.env.ORBIS_S3_PRESIGN_URL!);
  expect(result.ok, `HTTP ${result.status}`).toBeTruthy();
  expect(result.size).toBeGreaterThan(0);
});
