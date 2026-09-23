import { getGuestAds } from "@/lib/api.ads";
import { GuestAdsCarousel } from "@/components/guest/GuestAdsCarousel";

export async function GuestAdsSection() {
  const ads = await getGuestAds();
  return <GuestAdsCarousel items={ads} />;
}
