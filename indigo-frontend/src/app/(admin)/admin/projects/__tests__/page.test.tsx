import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ProjectsPage from "../page";
import { renderWithProviders } from "@/test-utils/render";

vi.mock("@/services/projects", () => ({
  fetchProjects: vi.fn(),
  createProject: vi.fn(),
  updateProjectName: vi.fn(),
  archiveProject: vi.fn(),
}));

const mockFetchProjects = vi.mocked(
  await import("@/services/projects").then((m) => m.fetchProjects)
);
const mockCreateProject = vi.mocked(
  await import("@/services/projects").then((m) => m.createProject)
);
const mockUpdateProjectName = vi.mocked(
  await import("@/services/projects").then((m) => m.updateProjectName)
);

describe("ProjectsPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockFetchProjects.mockResolvedValue({
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
    });
    mockCreateProject.mockResolvedValue({ object: "project" });
    mockUpdateProjectName.mockResolvedValue({ object: "project" });
  });

  it("requires project name when creating", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProjectsPage />);

    await waitFor(() =>
      expect(screen.getByText("Projects")).toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /create project/i }));
    expect(mockCreateProject).not.toHaveBeenCalled();

    await user.type(
      screen.getByLabelText(/project name/i),
      "  Internal Ops  "
    );
    await user.click(screen.getByRole("button", { name: /create project/i }));

    await waitFor(() =>
      expect(mockCreateProject).toHaveBeenCalledWith({ name: "Internal Ops" })
    );
  });

  it("validates rename modal and submits trimmed update", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProjectsPage />);

    await waitFor(() =>
      expect(screen.getByText("Automation")).toBeInTheDocument()
    );

    await user.click(screen.getByRole("button", { name: /rename/i }));

    const modalInput = await screen.findByLabelText(/new name/i);
    await user.clear(modalInput);
    await user.type(modalInput, "Automation");
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    expect(
      await screen.findByText("Enter a different name to update")
    ).toBeInTheDocument();

    await user.clear(modalInput);
    await user.type(modalInput, "  Automation v2  ");
    await user.click(screen.getByRole("button", { name: /save changes/i }));

    await waitFor(() =>
      expect(mockUpdateProjectName).toHaveBeenCalledWith(
        "proj_123",
        "Automation v2"
      )
    );
  });
});
