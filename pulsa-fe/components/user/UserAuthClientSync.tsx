"use client";

import * as React from "react";

type UserAuthClientSyncProps = {
  backendToken: string;
};

export function UserAuthClientSync({ backendToken }: UserAuthClientSyncProps) {
  React.useEffect(() => {
    localStorage.setItem("auth_token", backendToken);
    localStorage.setItem("auth_source", "google");
  }, [backendToken]);

  return null;
}

