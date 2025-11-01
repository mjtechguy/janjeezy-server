"use client";

import { useEffect, useState, useTransition } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  fetchSmtpSettings,
  updateSmtpSettings,
} from "@/services/settings";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

type FormState = {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  fromEmail: string;
  password: string;
  passwordTouched: boolean;
};

export default function SmtpSettingsPage() {
  const queryClient = useQueryClient();
  const [form, setForm] = useState<FormState>({
    enabled: false,
    host: "",
    port: 587,
    username: "",
    fromEmail: "",
    password: "",
    passwordTouched: false,
  });

  const [_isTransitionPending, startTransition] = useTransition();

  const smtpQuery = useQuery({
    queryKey: ["smtp-settings"],
    queryFn: fetchSmtpSettings,
  });

  useEffect(() => {
    if (!smtpQuery.data) {
      return;
    }
    startTransition(() => {
      setForm({
        enabled: smtpQuery.data.enabled,
        host: smtpQuery.data.host,
        port: smtpQuery.data.port,
        username: smtpQuery.data.username,
        fromEmail: smtpQuery.data.from_email,
        password: "",
        passwordTouched: false,
      });
    });
  }, [smtpQuery.data, startTransition]);

  const mutation = useMutation({
    mutationFn: updateSmtpSettings,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["smtp-settings"] });
      toast.success("SMTP settings updated");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!form.host.trim() || !form.fromEmail.trim()) {
      toast.error("Host and From email are required");
      return;
    }
    const payload: {
      enabled: boolean;
      host: string;
      port: number;
      username: string;
      password?: string;
      from_email: string;
    } = {
      enabled: form.enabled,
      host: form.host.trim(),
      port: Number(form.port) || 587,
      username: form.username.trim(),
      from_email: form.fromEmail.trim(),
    };
    if (form.passwordTouched) {
      payload.password = form.password;
    }
    mutation.mutate(payload);
  };

  return (
    <section className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          SMTP Settings
        </h1>
        <p className="text-sm text-muted-foreground">
          Configure mail delivery for invitations and notifications. Updates
          apply immediately and persist across restarts.
        </p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>Email transport</CardTitle>
          <CardDescription>
            Provide connection details for your SMTP relay. Leave the password
            field blank to keep the existing secret.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex items-center gap-3 sm:col-span-2">
                <input
                  id="smtp-enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      enabled: event.target.checked,
                    }))
                  }
                  className="h-4 w-4 rounded border-input text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  disabled={mutation.isPending || smtpQuery.isLoading}
                />
                <Label htmlFor="smtp-enabled" className="text-sm">
                  Enable SMTP delivery
                </Label>
              </div>

              <div className="space-y-2">
                <Label htmlFor="smtp-host">Host</Label>
                <Input
                  id="smtp-host"
                  value={form.host}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, host: event.target.value }))
                  }
                  placeholder="smtp.example.com"
                  disabled={mutation.isPending || smtpQuery.isLoading}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="smtp-port">Port</Label>
                <Input
                  id="smtp-port"
                  type="number"
                  value={form.port}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      port: Number(event.target.value),
                    }))
                  }
                  disabled={mutation.isPending || smtpQuery.isLoading}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="smtp-username">Username</Label>
                <Input
                  id="smtp-username"
                  value={form.username}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      username: event.target.value,
                    }))
                  }
                  placeholder="smtp-user"
                  disabled={mutation.isPending || smtpQuery.isLoading}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="smtp-from">From email</Label>
                <Input
                  id="smtp-from"
                  type="email"
                  value={form.fromEmail}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      fromEmail: event.target.value,
                    }))
                  }
                  placeholder="notifications@example.com"
                  disabled={mutation.isPending || smtpQuery.isLoading}
                  required
                />
              </div>

              <div className="space-y-2 sm:col-span-2">
                <Label htmlFor="smtp-password">
                  Password {smtpQuery.data?.has_password ? "(update to change)" : ""}
                </Label>
                <Input
                  id="smtp-password"
                  type="password"
                  placeholder={smtpQuery.data?.has_password ? "••••••••" : "Optional"}
                  value={form.password}
                  onChange={(event) =>
                    setForm((prev) => ({
                      ...prev,
                      password: event.target.value,
                      passwordTouched: true,
                    }))
                  }
                  onBlur={() =>
                    setForm((prev) => ({ ...prev, passwordTouched: prev.passwordTouched || prev.password.length > 0 }))
                  }
                  disabled={mutation.isPending || smtpQuery.isLoading}
                />
                <p className="text-xs text-muted-foreground">
                  Leave blank to keep the existing secret. Clear the field and
                  submit to remove the stored password.
                </p>
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <Button
                type="submit"
                disabled={mutation.isPending || smtpQuery.isLoading}
              >
                {mutation.isPending ? "Saving…" : "Save changes"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </section>
  );
}
