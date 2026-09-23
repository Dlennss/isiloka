import { forwardRef, type ButtonHTMLAttributes } from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-2xl text-sm font-black transition-colors focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-emerald-200 disabled:pointer-events-none disabled:opacity-60 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "bg-[#13464b] text-white shadow-[0_10px_20px_rgba(6,78,59,0.16)] hover:bg-[#087e8b]",
        primary: "bg-[#13464b] text-white shadow-[0_10px_20px_rgba(6,78,59,0.16)] hover:bg-[#087e8b]",
        danger: "bg-red-600 text-white shadow-sm hover:bg-red-500",
        warning: "bg-amber-400 text-slate-950 shadow-sm hover:bg-amber-300",
        success: "bg-[#087e8b] text-white shadow-sm hover:bg-[#13464b]",
        info: "bg-[#008f6b] text-white shadow-sm hover:bg-[#087e8b]",
        destructive:
          "bg-red-600 text-white shadow-sm hover:bg-red-500",
        outline:
          "border-2 border-[#13464b] bg-white text-[#13464b] shadow-sm hover:bg-[#f1fff8]",
        secondary:
          "border border-emerald-200 bg-[#e8fff4] text-[#13464b] shadow-sm hover:bg-[#d9ffeb]",
        ghost: "text-[#13464b] hover:bg-[#e8fff4]",
        link: "text-[#087e8b] underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-8 rounded-xl px-3 text-xs",
        lg: "h-11 px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
