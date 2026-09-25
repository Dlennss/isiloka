import { getAppServerSession } from "@/lib/server-auth";
import { getUserProfile } from "@/lib/api.auth";
import type { UserAppOrder, UserProfile, UserSession } from "@/components/user/types";
import { GuestConceptHome } from "@/components/guest/GuestConceptHome";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type OrdersResponse = {
  ok?: boolean;
  items?: UserAppOrder[];
};

const apiBase = () => (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");

async function getUserOrders(token?: string): Promise<UserAppOrder[]> {
  if (!token) return [];
  try {
    const res = await fetch(`${apiBase()}/v1/app/me/orders?limit=3`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return [];
    const json = (await res.json().catch(() => ({}))) as OrdersResponse;
    return Array.isArray(json.items) ? json.items : [];
  } catch (error) {
    console.error("[user-home] gagal mengambil aktivitas", error);
    return [];
  }
}

export default async function UserAppHomePage() {
  const session = (await getAppServerSession()) as SessionShape | null;
  const token = session?.backendToken;
  const sessionUser = session?.user;
  const [profile, recentOrders] = await Promise.all([
    token ? getUserProfile(token) : Promise.resolve(null),
    getUserOrders(token),
  ]);
  const isLoggedIn = Boolean(token);
  const displayName = profile?.nama || sessionUser?.name || profile?.email || sessionUser?.email || "Pengguna";
  const displayEmail = profile?.email || sessionUser?.email || "";

  return (
    <main className="brand-retail-main bg-[#f0f8f8]">
      {token ? <UserAuthClientSync backendToken={token} /> : null}
      <GuestConceptHome
        user={isLoggedIn ? {
          isLoggedIn: true,
          name: displayName,
          email: displayEmail,
          image: profile?.profile_photo_url || sessionUser?.image || null,
          balance: Number(profile?.saldo || 0),
        } : { isLoggedIn: false }}
        recentOrders={recentOrders}
        userMode
      />
      <UserBottomNav />
    </main>
  );
}
