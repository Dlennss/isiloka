'use client';

 

import { useEffect, useRef, useState } from 'react';
import type { GuestAdItem } from '@/lib/api.ads';

type GuestAdsCarouselProps = {
  items: GuestAdItem[];
};

const fallbackAds: GuestAdItem[] = [
  {
    id: -1,
    judul: "",
    keterangan: "",
    image_url: "/images/guest-ads/banner-ewallet.png",
    link_url: "/ewallet",
    urutan: 1,
    aktif: true,
  },
  {
    id: -2,
    judul: "",
    keterangan: "",
    image_url: "/images/guest-ads/banner-topup.png",
    link_url: "/pulsa",
    urutan: 2,
    aktif: true,
  },
  {
    id: -3,
    judul: "",
    keterangan: "",
    image_url: "/images/guest-ads/banner-operator.png",
    link_url: "/paket-data",
    urutan: 3,
    aktif: true,
  },
  {
    id: -4,
    judul: "",
    keterangan: "",
    image_url: "/images/guest-ads/banner-pln.png",
    link_url: "/listrik",
    urutan: 4,
    aktif: true,
  },
];

export function GuestAdsCarousel({ items }: GuestAdsCarouselProps) {
  const activeAds = items.filter((item) => item.aktif !== false && item.image_url);
  const ads = activeAds.length > 0 ? activeAds : fallbackAds;
  const [activeIndex, setActiveIndex] = useState(0);
  const [viewportWidth, setViewportWidth] = useState(0);
  const viewportRef = useRef<HTMLElement | null>(null);
  const touchStartXRef = useRef<number | null>(null);
  const touchDeltaXRef = useRef(0);
  const safeActiveIndex = ads.length > 0 ? Math.min(activeIndex, ads.length - 1) : 0;
  const peekSize = viewportWidth >= 768 ? 24 : 0;
  const slideGap = viewportWidth >= 768 ? 10 : 0;
  const slideWidth = Math.max(0, viewportWidth - (ads.length > 1 ? peekSize * 2 : 0));
  const trackOffset = ads.length > 1 ? safeActiveIndex * (slideWidth + slideGap) : 0;

  useEffect(() => {
    const node = viewportRef.current;
    if (!node) return;

    const observer = new ResizeObserver((entries) => {
      const width = entries[0]?.contentRect.width ?? 0;
      setViewportWidth(Math.round(width));
    });

    observer.observe(node);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (ads.length <= 1) return;
    const desktop = window.matchMedia("(min-width: 768px)");
    if (!desktop.matches) return;
    const timer = window.setInterval(() => {
      setActiveIndex((prev) => (prev + 1) % ads.length);
    }, 3800);
    return () => window.clearInterval(timer);
  }, [ads.length]);

  function moveToPrev() {
    setActiveIndex((prev) => (prev - 1 + ads.length) % ads.length);
  }

  function moveToNext() {
    setActiveIndex((prev) => (prev + 1) % ads.length);
  }

  function handleTouchStart(event: React.TouchEvent<HTMLElement>) {
    touchStartXRef.current = event.touches[0]?.clientX ?? null;
    touchDeltaXRef.current = 0;
  }

  function handleTouchMove(event: React.TouchEvent<HTMLElement>) {
    if (touchStartXRef.current == null) return;
    const currentX = event.touches[0]?.clientX ?? touchStartXRef.current;
    touchDeltaXRef.current = currentX - touchStartXRef.current;
  }

  function handleTouchEnd() {
    if (touchStartXRef.current == null || ads.length <= 1) {
      touchStartXRef.current = null;
      touchDeltaXRef.current = 0;
      return;
    }

    const delta = touchDeltaXRef.current;
    if (Math.abs(delta) >= 40) {
      if (delta > 0) {
        moveToPrev();
      } else {
        moveToNext();
      }
    }

    touchStartXRef.current = null;
    touchDeltaXRef.current = 0;
  }

  return (
    <section
      ref={viewportRef}
      className="relative overflow-hidden rounded-[18px] border border-sky-100 bg-white shadow-[0_16px_38px_rgba(15,56,104,0.08)] [touch-action:pan-y]"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
      onTouchCancel={handleTouchEnd}
    >
      <div
        className="flex transition-transform duration-700 ease-out"
        style={{
          gap: `${slideGap}px`,
          paddingLeft: ads.length > 1 ? `${peekSize}px` : 0,
          paddingRight: ads.length > 1 ? `${peekSize}px` : 0,
          transform: `translateX(-${trackOffset}px)`,
        }}
      >
        {ads.map((item) => {
          const hasCaption = Boolean(item.judul || item.keterangan);
          const content = (
            <div
              className="relative aspect-[19/9] shrink-0 overflow-hidden rounded-[16px] bg-sky-50"
              style={{ width: slideWidth > 0 ? `${slideWidth}px` : "100%" }}
            >
              <img
                src={item.image_url}
                alt={item.judul || 'Iklan Isiloka'}
                className="h-full w-full object-cover"
                loading="lazy"
              />
              {hasCaption ? (
                <>
                  <div className="absolute inset-0 bg-gradient-to-t from-slate-950/75 via-slate-900/15 to-transparent" />
                  <div className="absolute inset-x-0 bottom-0 space-y-1 px-4 pb-3 pt-10 text-white">
                    {item.judul ? <h3 className="text-sm font-bold leading-tight sm:text-base">{item.judul}</h3> : null}
                    {item.keterangan ? (
                      <p className="max-w-[92%] text-[11px] leading-4 text-white/92 sm:text-sm">{item.keterangan}</p>
                    ) : null}
                  </div>
                </>
              ) : null}
            </div>
          );

          if (item.link_url) {
            return (
              <a
                key={item.id}
                href={item.link_url}
                className="block shrink-0"
                style={{ width: slideWidth > 0 ? `${slideWidth}px` : "100%" }}
              >
                {content}
              </a>
            );
          }

          return (
            <div
              key={item.id}
              className="shrink-0"
              style={{ width: slideWidth > 0 ? `${slideWidth}px` : "100%" }}
            >
              {content}
            </div>
          );
        })}
      </div>
      {ads.length > 1 ? (
        <div className="pointer-events-none absolute inset-x-0 bottom-0 z-10 flex items-center justify-center px-4 pb-2.5">
          <div className="pointer-events-auto flex items-center gap-1.5 rounded-full bg-black/18 px-2 py-1 backdrop-blur-sm">
            {ads.map((dotItem, index) => (
              <button
                key={dotItem.id}
                type="button"
                aria-label={`Buka iklan ${index + 1}`}
                onClick={() => setActiveIndex(index)}
                className={index === safeActiveIndex ? 'h-1.5 w-5 rounded-full bg-white/95' : 'h-1.5 w-1.5 rounded-full bg-white/55'}
              />
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
}
