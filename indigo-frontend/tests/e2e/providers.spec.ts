import { expect, test } from "@playwright/test";

test.describe("Admin providers page", () => {
  test.beforeEach(async ({ page }) => {
    await page.route("**/api/jan/models/providers", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ object: "list", data: [] }),
      });
    });

    await page.route("**/api/jan/organization/providers/vendors", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          object: "list",
          data: [
            {
              key: "openai",
              name: "OpenAI",
              scope: "organization",
              default_base_url: "https://api.openai.com/v1",
              credential_hint: "sk-",
            },
          ],
        }),
      });
    });

    await page.route("**/api/jan/organization/projects", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          object: "list",
          data: [
            {
              object: "project",
              id: "proj_123",
              name: "Operations",
              status: "active",
              created_at: Math.floor(Date.now() / 1000),
            },
          ],
        }),
      });
    });
  });

  test("submits a provider with vendor defaults", async ({ page }) => {
    let createPayload: Record<string, unknown> | undefined;
    await page.route("**/api/jan/organization/models/providers", async (route) => {
      if (route.request().method() === "POST") {
        createPayload = JSON.parse(route.request().postData() ?? "{}");
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ object: "model.provider", id: "prov_123" }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ object: "list", data: [] }),
        });
      }
    });

    await page.goto("/admin/providers");
    await expect(page.getByRole("heading", { name: "Model Providers" })).toBeVisible();

    await page.getByLabel("Provider name").fill("Production OpenAI");
    await page.getByLabel("Vendor").selectOption("openai");
    await expect(page.getByLabel("Base URL")).toHaveValue("https://api.openai.com/v1");
    await page.getByLabel("API key (optional)").fill("sk-secret");

    await page.getByRole("button", { name: "Register provider" }).click();

    await expect.poll(() => createPayload).not.toBeUndefined();
    expect(createPayload).toMatchObject({
      name: "Production OpenAI",
      vendor: "openai",
      base_url: "https://api.openai.com/v1",
      api_key: "sk-secret",
    });
  });
});
