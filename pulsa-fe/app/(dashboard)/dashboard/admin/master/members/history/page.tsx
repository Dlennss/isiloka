import { Suspense } from "react";
import AccountActivityClient from "./AccountActivityClient";

export default function Page() {
  return (
    <Suspense fallback={<div >Loading...</div>}>
      <AccountActivityClient />
    </Suspense>
  );
}
