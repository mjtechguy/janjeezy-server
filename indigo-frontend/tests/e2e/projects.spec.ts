import { expect, test } from "@playwright/test";

test.describe("Admin projects page", () => {
  test.beforeEach(async ({ page }) => {
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
              name: "Automation",
              status: "active",
              created_at: Math.floor(Date.now() / 1000),
            },
          ],
        }),
      });
    });
  });

  test("renames a project through the modal workflow", async ({ page }) => {
    let updatePayload: Record<string, unknown> | undefined;
    await page.route("**/api/jan/organization/projects/proj_123", async (route) => {
      if (route.request().method() === "POST") {
        updatePayload = JSON.parse(route.request().postData() ?? "{}");
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ object: "project", id: "proj_123" }),
        });
      } else {
        await route.continue();
      }
    });

    await page.goto("/admin/projects");
    await expect(page.getByRole("heading", { name: "Projects" })).toBeVisible();

    await page.getByRole("button", { name: "Rename" }).click();
    const modalInput = page.getByLabel("New name");
    await expect(modalInput).toBeVisible();

    await modalInput.fill("Automation v2");
    await page.getByRole("button", { name: "Save changes" }).click();

    await expect.poll(() => updatePayload).not.toBeUndefined();
    expect(updatePayload).toMatchObject({ name: "Automation v2" });
  });
});
