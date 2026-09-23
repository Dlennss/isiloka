import * as React from "react";
import { CalendarRange } from "lucide-react";

import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";

type BaseProps = Omit<React.ComponentProps<typeof Input>, "type"> & {
  label?: string;
  wrapperClassName?: string;
};

export function DateField({ label, className, wrapperClassName, ...props }: BaseProps) {
  return (
    <div className={cn("space-y-1", wrapperClassName)}>
      {label ? <div className="text-xs uppercase tracking-wide text-slate-400">{label}</div> : null}
      <div className="relative">
        <Input type="date" className={cn("pr-11", className)} {...props} />
        <span className="pointer-events-none absolute inset-y-0 right-0 flex w-11 items-center justify-center rounded-r-md border-l border-white/10 bg-white/5 text-slate-300">
          <CalendarRange className="h-4 w-4" />
        </span>
      </div>
    </div>
  );
}

export function MonthField({ label, className, wrapperClassName, ...props }: BaseProps) {
  return (
    <div className={cn("space-y-1", wrapperClassName)}>
      {label ? <div className="text-xs uppercase tracking-wide text-slate-400">{label}</div> : null}
      <div className="relative">
        <Input type="month" className={cn("pr-11", className)} {...props} />
        <span className="pointer-events-none absolute inset-y-0 right-0 flex w-11 items-center justify-center rounded-r-md border-l border-white/10 bg-white/5 text-slate-300">
          <CalendarRange className="h-4 w-4" />
        </span>
      </div>
    </div>
  );
}
