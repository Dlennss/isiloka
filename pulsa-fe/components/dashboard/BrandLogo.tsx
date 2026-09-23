import Image from "next/image";

type BrandLogoProps = {
  variant?: "light" | "dark";
};

export function BrandLogo({ variant = "light" }: BrandLogoProps) {
  const isDark = variant === "dark";
  return (
    <div
      className={`inline-flex items-center justify-center rounded-2xl ${
        isDark
          ? "bg-white px-3 py-2 shadow-[0_14px_30px_rgba(0,0,0,0.20)] ring-1 ring-white/70"
          : "px-1 py-1"
      }`}
    >
      <Image
        src="/images/logo-pulsakilat-header.svg"
        alt="Isiloka"
        width={168}
        height={38}
        priority
        className="h-auto w-[148px] object-contain sm:w-[164px]"
      />
    </div>
  );
}

