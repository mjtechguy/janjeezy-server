"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

export default function GoogleCallbackPage() {
  const router = useRouter();
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const authCode = params.get("code");
    const authState = params.get("state");

    if (!authCode) {
      toast.error("Missing authorization code");
      const timer = setTimeout(() => setHasError(true), 0);
      return () => {
        clearTimeout(timer);
      };
      return;
    }

    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;

    async function finalize() {
      const res = await fetch("/api/jan/auth/google/callback", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: authCode, state: authState }),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        if (!cancelled && !timer) {
          timer = setTimeout(() => setHasError(true), 0);
        }
        toast.error(body?.error ?? "Unable to sign in with Google");
        return;
      }
      toast.success("Signed in with Google");
      router.replace("/admin/overview");
      router.refresh();
    }

    void finalize();
    return () => {
      cancelled = true;
      if (timer) {
        clearTimeout(timer);
      }
    };
  }, [router]);

  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-background text-foreground">
      <div className="w-full max-w-md space-y-4 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          Completing Sign In
        </h1>
        {!hasError ? (
          <p className="text-sm text-white/70">
            Please wait while we verify your Google account…
          </p>
        ) : (
          <p className="text-sm text-warning">
            Something went wrong. You can close this tab and try again.
          </p>
        )}
      </div>
    </main>
  );
}
