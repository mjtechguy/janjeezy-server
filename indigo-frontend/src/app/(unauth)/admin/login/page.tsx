"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { LocalLoginSchema } from "@/schemas/auth";
import { z } from "zod";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ThemeToggle } from "@/components/ui/theme-toggle";

type LocalLoginValues = z.infer<typeof LocalLoginSchema>;

export default function AdminLoginPage() {
  const router = useRouter();
  const [formState, setFormState] = useState<LocalLoginValues>({
    email: "",
    password: "",
  });
  const [isPending, setPending] = useState(false);
  const [nextPath, setNextPath] = useState("/admin/overview");

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const next = params.get("next");
    if (next) {
      setNextPath(next);
    }
  }, []);

  const handleGoogleLogin = async () => {
    try {
      setPending(true);
      const res = await fetch("/api/jan/auth/google/login", { method: "GET" });
      if (!res.ok) {
        toast.error("Unable to start Google login");
        return;
      }
      const { redirectUrl } = (await res.json()) as { redirectUrl: string };
      window.location.href = redirectUrl;
    } catch {
      toast.error("Unexpected error starting Google login");
    } finally {
      setPending(false);
    }
  };

  const handleLocalLogin = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const parseResult = LocalLoginSchema.safeParse(formState);
    if (!parseResult.success) {
      toast.error(parseResult.error.issues[0]?.message ?? "Invalid credentials");
      return;
    }

    try {
      setPending(true);
      const res = await fetch("/api/jan/auth/local/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(parseResult.data),
      });

      if (!res.ok) {
        const body = await res.json().catch(() => null);
        toast.error(body?.error ?? "Invalid email or password");
        return;
      }

      toast.success("Signed in");
      router.replace(nextPath);
      router.refresh();
    } catch {
      toast.error("Unexpected error signing in");
    } finally {
      setPending(false);
    }
  };

  return (
    <main className="relative flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background to-muted">
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      <Card className="w-full max-w-md border-border/60 bg-card/80 backdrop-blur">
        <CardHeader className="space-y-4 text-center">
          <div className="flex justify-center">
            <span className="rounded-full bg-primary/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.45em] text-primary">
              Jan Admin
            </span>
          </div>
          <div className="space-y-2">
            <CardTitle className="text-2xl font-semibold tracking-tight">
              Secure Administration
            </CardTitle>
            <CardDescription>
              Sign in with Google or your local admin credentials to access the
              control plane.
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <Button
            type="button"
            variant="secondary"
            className="w-full"
            onClick={handleGoogleLogin}
            disabled={isPending}
          >
            Continue with Google
          </Button>

          <div className="flex items-center gap-2 text-xs uppercase tracking-[0.3em] text-muted-foreground">
            <div className="h-px flex-1 bg-border" />
            <span>or</span>
            <div className="h-px flex-1 bg-border" />
          </div>

          <form onSubmit={handleLocalLogin} className="space-y-4">
            <div className="space-y-2 text-left">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                value={formState.email}
                onChange={(e) =>
                  setFormState((prev) => ({ ...prev, email: e.target.value }))
                }
                placeholder="admin@example.com"
                disabled={isPending}
                required
              />
            </div>

            <div className="space-y-2 text-left">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                value={formState.password}
                onChange={(e) =>
                  setFormState((prev) => ({ ...prev, password: e.target.value }))
                }
                placeholder="••••••••••••"
                disabled={isPending}
                required
              />
            </div>

            <Button type="submit" className="w-full" disabled={isPending}>
              Sign in
            </Button>
          </form>

          <p className="text-center text-xs text-muted-foreground">
            Need access? Contact an organization owner to enable a local admin
            account.
          </p>
        </CardContent>
      </Card>
    </main>
  );
}
