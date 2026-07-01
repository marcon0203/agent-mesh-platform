import { Link, Outlet, useLocation } from "react-router-dom"
import { cn } from "@/lib/utils"

const NAV_ITEMS = [
  { to: "/marketplace", label: "能力市场" },
  { to: "/builder", label: "构建 Agent" },
  { to: "/workbench", label: "Workbench" },
  { to: "/dashboard", label: "控制台" },
]

export default function AppLayout() {
  const location = useLocation()

  return (
    <div className="min-h-screen">
      <nav className="sticky top-0 z-50 border-b border-glass-border bg-background/70 backdrop-blur-xl">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-8 py-4">
          <Link to="/" className="font-display text-lg font-semibold">
            枢络<span className="ml-1 font-mono text-xs text-muted-foreground">AgentMesh</span>
          </Link>
          <ul className="flex gap-8">
            {NAV_ITEMS.map((item) => (
              <li key={item.to}>
                <Link
                  to={item.to}
                  className={cn(
                    "text-sm font-medium text-muted-foreground transition-colors hover:text-foreground",
                    location.pathname.startsWith(item.to) && "text-foreground"
                  )}
                >
                  {item.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </nav>
      <main className="mx-auto max-w-6xl px-8 py-12">
        <Outlet />
      </main>
    </div>
  )
}
