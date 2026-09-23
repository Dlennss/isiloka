"use client";

import { LogOut } from "lucide-react";
import { signOut } from "next-auth/react";
import { Button } from "@/components/ui/button";

type UserLogoutButtonProps = {
  className?: string;
  iconOnly?: boolean;
};

export function UserLogoutButton({ className, iconOnly = false }: UserLogoutButtonProps) {
  async function onLogout() {
    localStorage.removeItem("auth_token");
    localStorage.removeItem("auth_source");
    await fetch("/api/auth/logout", { method: "POST" }).catch(() => undefined);
    await signOut({ callbackUrl: "/" });
  }

  if (iconOnly) {
    return (
      <Button
        type="button"
        size="icon"
        className={className}
        onClick={onLogout}
        aria-label="Keluar"
        title="Keluar"
      >
        <LogOut className="h-4 w-4" />
      </Button>
    );
  }

  return (
    <Button type="button" className={className} onClick={onLogout}>
      <LogOut className="mr-2 h-4 w-4" />
      Keluar
    </Button>
  );
}
