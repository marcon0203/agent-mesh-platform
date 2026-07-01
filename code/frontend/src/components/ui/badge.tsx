import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 font-mono text-[10.5px] tracking-wide",
  {
    variants: {
      variant: {
        tool: "bg-primary/15 text-[#5aa6ff]",
        skill: "bg-accent/15 text-accent",
        agent: "bg-[#b48cff]/15 text-[#b48cff]",
        neutral: "bg-glass border border-glass-border text-muted-foreground",
      },
    },
    defaultVariants: { variant: "neutral" },
  }
)

function Badge({
  className,
  variant,
  ...props
}: React.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return (
    <span
      data-slot="badge"
      className={cn(badgeVariants({ variant, className }))}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
