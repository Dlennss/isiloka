import { getAppServerSession } from "@/lib/server-auth";
import type { UserAppOrder, UserProfile, UserSession } from "@/components/user/types";
import { GuestConceptHome } from "@/components/guest/GuestConceptHome";
import { UserBottomNav } from "@/components/user/UserBottomNav";
import { UserAuthClientSync } from "@/components/user/UserAuthClientSync";

type SessionShape = {
  user?: UserSession;
  backendToken?: string;
};

type ProfileResponse = {
  ok?: boolean;
  profile?: UserProfile;
};

type OrdersResponse = {
  ok?: boolean;
  items?: UserAppOrder[];
};

const apiBase = () => (process.env.NEXT_PUBLIC_API_BASE || process.env.API_BASE || "http://127.0.0.1:8083").replace(/\/+$/, "");

async function getUserProfile(token?: string): Promise<UserProfile | null> {
  if (!token) return null;
  try {
    const res = await fetch(`${apiBase()}/v1/me/profile`, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return null;
    const json = (await res.json().catch(() => ({}))) as ProfileResponse;
    return json.ok && json.profile ? json.profile : null;
  } catch (error) {
    console.error("[user-home] gagal mengambil profile", error);
    return null;
  }
}

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
  const [profile, recentOrders] = await Promise.all([
    getUserProfile(token),
    getUserOrders(token),
  ]);

  return (
    <main className="brand-retail-main bg-[#f0f8f8]">
      {token ? <UserAuthClientSync backendToken={token} /> : null}
      <GuestConceptHome
        user={profile ? {
          isLoggedIn: true,
          name: profile.nama || profile.email,
          email: profile.email,
          image: profile.profile_photo_url || session?.user?.image || null,
          balance: profile.saldo,
        } : { isLoggedIn: false }}
        recentOrders={recentOrders}
        userMode
      />
      <UserBottomNav />
    </main>
  );
}
