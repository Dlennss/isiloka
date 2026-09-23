import Link from "next/link";
import type { UserMenuItem } from "@/components/user/types";

type UserMenuGridProps = {
  menuItems: UserMenuItem[];
};

export function UserMenuGrid({ menuItems }: UserMenuGridProps) {
  return (
    <section>
      <div className="rounded-xl bg-white p-2 shadow-[0_8px_22px_rgba(15,23,42,0.13)]">
        <div className="grid grid-cols-3 gap-x-1 gap-y-4">
          {menuItems.map((item) => {
            const Icon = item.icon;
            return (
              <Link key={item.title} href={item.href} className="flex flex-col items-center gap-2 text-center">
                <span className="grid h-18 w-18 place-items-center rounded-xl bg-sky-50 text-[#2d8fdc]">
                  <Icon className="h-6 w-6" />
                </span>
                <span className="text-[11px] leading-tight font-medium text-neutral-800">{item.title}</span>
              </Link>
            );
          })}
        </div>
      </div>
    </section>
  );
}
