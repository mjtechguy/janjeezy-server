import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ProvidersPage from "../page";
import { renderWithProviders } from "@/test-utils/render";

vi.mock("@/services/providers", () => ({
  fetchProviders: vi.fn(),
  fetchProviderVendors: vi.fn(),
  createProvider: vi.fn(),
  updateProvider: vi.fn(),
}));

vi.mock("@/services/projects", () => ({
  fetchProjects: vi.fn(),
}));

const mockFetchProviders = vi.mocked(
  await import("@/services/providers").then((m) => m.fetchProviders)
);
const mockFetchProviderVendors = vi.mocked(
  await import("@/services/providers").then((m) => m.fetchProviderVendors)
);
const mockCreateProvider = vi.mocked(
  await import("@/services/providers").then((m) => m.createProvider)
);
const mockFetchProjects = vi.mocked(
  await import("@/services/projects").then((m) => m.fetchProjects)
);

describe("ProvidersPage", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    mockFetchProviders.mockResolvedValue({
      object: "list",
      data: [],
    });
    mockFetchProviderVendors.mockResolvedValue({
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
    });
    mockFetchProjects.mockResolvedValue({
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
    });
    mockCreateProvider.mockResolvedValue({
      object: "model.provider",
      id: "prov_123",
    });
  });

  it("shows validation errors when form is submitted empty", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProvidersPage />);

    await waitFor(() =>
      expect(screen.getByText("Model Providers")).toBeInTheDocument()
    );

    await user.click(
      screen.getByRole("button", { name: /register provider/i })
    );

    expect(screen.getByText("Name is required")).toBeInTheDocument();
    expect(screen.getByText("Select a vendor")).toBeInTheDocument();
    expect(screen.getByText("Base URL is required")).toBeInTheDocument();
    expect(mockCreateProvider).not.toHaveBeenCalled();
  });

  it("auto-populates base URL for known vendor and submits payload", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProvidersPage />);

    await waitFor(() =>
      expect(screen.getByText("Model Providers")).toBeInTheDocument()
    );

    await user.type(
      screen.getByLabelText(/provider name/i),
      " Production OpenAI "
    );
    await user.selectOptions(screen.getByLabelText(/vendor/i), "openai");
    await user.type(screen.getByLabelText(/api key/i), "sk-secret");

    await user.click(
      screen.getByRole("button", { name: /register provider/i })
    );

    await waitFor(() =>
      expect(mockCreateProvider).toHaveBeenCalledWith({
        name: "Production OpenAI",
        vendor: "openai",
        base_url: "https://api.openai.com/v1",
        api_key: "sk-secret",
      })
    );
  });

  it("requires project selection when scope is project", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProvidersPage />);

    await waitFor(() =>
      expect(screen.getByText("Model Providers")).toBeInTheDocument()
    );

    await user.type(screen.getByLabelText(/provider name/i), "Project Vendor");
    await user.selectOptions(screen.getByLabelText(/vendor/i), "openai");
    await user.selectOptions(screen.getByLabelText(/scope/i), "project");

    await user.click(
      screen.getByRole("button", { name: /register provider/i })
    );

    expect(screen.getByText("Select a project")).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText(/project/i), "proj_123");
    await user.click(
      screen.getByRole("button", { name: /register provider/i })
    );

    await waitFor(() =>
      expect(mockCreateProvider).toHaveBeenCalledWith({
        name: "Project Vendor",
        vendor: "openai",
        base_url: "https://api.openai.com/v1",
        api_key: undefined,
        project_public_id: "proj_123",
      })
    );
  });
});
