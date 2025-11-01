import { getOrganizationOverview } from "@/lib/server/organization";
import type { OrganizationOverview } from "@/schemas/organization";

export default async function OrganizationPage() {
  let overview: OrganizationOverview | null = null;
  let errorMessage: string | null = null;

  try {
    overview = await getOrganizationOverview();
  } catch (error) {
    errorMessage = error instanceof Error ? error.message : "";
  }

  if (errorMessage) {
    return (
      <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
        <header>
          <h1 className="text-2xl font-semibold text-foreground">
            Organization Overview
          </h1>
        </header>
        <div className="rounded-2xl border border-danger/30 bg-danger/10 p-6 text-danger">
          Failed to load organization metrics. {errorMessage}
        </div>
      </section>
    );
  }

  if (!overview) {
    return null;
  }

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header>
        <h1 className="text-2xl font-semibold text-foreground">
          Organization Overview
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          High-level metrics across projects, members, and providers. More
          detailed insights will roll out alongside upcoming admin features.
        </p>
      </header>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
            <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
              Total Projects
            </h2>
            <p className="mt-3 text-3xl font-semibold text-foreground">
              {overview.projects.total}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Across organization and project scopes.
            </p>
          </article>

          <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
            <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
              Active Projects
            </h2>
            <p className="mt-3 text-3xl font-semibold text-success">
              {overview.projects.active}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Ready for provider assignment.
            </p>
          </article>

          <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
            <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
              Members
            </h2>
            <p className="mt-3 text-3xl font-semibold text-foreground">
              {overview.members.total}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Including owners and readers.
            </p>
          </article>

          <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
            <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
              Pending Invites
            </h2>
            <p className="mt-3 text-3xl font-semibold text-warning">
              {overview.invites.pending}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              Awaiting acceptance.
            </p>
          </article>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
          <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
            Provider Health
          </h2>
          <div className="mt-4 flex items-center gap-10 text-muted-foreground">
            <div>
              <p className="text-sm uppercase tracking-[0.3em] text-success">
                Active
              </p>
              <p className="mt-1 text-2xl font-semibold text-foreground">
                {overview.providers.active}
              </p>
            </div>
            <div>
              <p className="text-sm uppercase tracking-[0.3em] text-warning">
                Inactive
              </p>
              <p className="mt-1 text-2xl font-semibold text-foreground">
                {overview.providers.inactive}
              </p>
            </div>
          </div>
        </article>

        <article className="rounded-2xl border border-border bg-card p-6 shadow-sm">
          <h2 className="text-xs font-semibold uppercase tracking-[0.4em] text-muted-foreground">
            Archived Projects
          </h2>
          <p className="mt-3 text-3xl font-semibold text-foreground">
            {overview.projects.archived}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            Preserved for record keeping.
          </p>
        </article>
      </div>
    </section>
  );
}
