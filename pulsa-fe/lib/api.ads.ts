import { fetchAPI } from "./api";

export type GuestAdItem = {
  id: number;
  judul: string;
  keterangan: string;
  image_url: string;
  link_url: string;
  urutan: number;
  aktif: boolean;
};

export async function getGuestAds(): Promise<GuestAdItem[]> {
  return fetchAPI<GuestAdItem>("/v1/app/ads", {
    revalidate: 30,
  });
}
