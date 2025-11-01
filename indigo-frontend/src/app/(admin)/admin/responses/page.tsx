export default function ResponsesPage() {
  return (
    <section className="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-12">
      <header>
        <h1 className="text-2xl font-semibold text-foreground">Responses</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          This area will monitor response jobs, including background processing,
          cancellation, and inspection.
        </p>
      </header>
    </section>
  );
}
