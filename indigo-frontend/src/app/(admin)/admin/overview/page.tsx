import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

export default function AdminOverviewPage() {
  return (
    <section className="space-y-8">
      <div className="flex flex-col gap-4 rounded-xl border border-border/60 bg-card/70 p-6 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center gap-3">
          <Badge variant="secondary" className="uppercase tracking-[0.3em]">
            Welcome
          </Badge>
          <Separator className="hidden flex-1 sm:block" />
        </div>
        <div className="space-y-3">
          <h1 className="text-3xl font-semibold tracking-tight">
            Administration Overview
          </h1>
          <p className="max-w-2xl text-sm text-muted-foreground">
            Monitor organization health, keep an eye on provider connectivity,
            and orchestrate administrator tasks from a modern control surface.
            Deeper insights will populate here as core services come online.
          </p>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Next Steps
            </CardTitle>
            <CardDescription>
              Suggested work items for the upcoming implementation cycle.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>• Wire authenticated layout shell</p>
            <p>• Implement organization + project services</p>
            <p>• Integrate provider catalog view</p>
          </CardContent>
        </Card>
        <Card className="md:col-span-1 xl:col-span-2">
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Coming Soon
            </CardTitle>
            <CardDescription>
              Placeholder for metrics, system notices, and recent activity.
            </CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Surface runbooks, alerting summaries, and provider connection status
            here to give administrators a swift operational overview.
          </CardContent>
        </Card>
      </div>
    </section>
  );
}
