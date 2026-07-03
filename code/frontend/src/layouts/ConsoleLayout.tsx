import { useState } from "react"
import { Link, Outlet, useLocation } from "react-router-dom"
import {
  LayoutDashboard,
  Bot,
  Cpu,
  MessagesSquare,
  ChevronRight,
  PanelLeftClose,
  PanelLeftOpen,
  LogOut,
  type LucideIcon,
} from "lucide-react"
import { cn } from "@/lib/utils"
import { LogoMark } from "@/components/Logo"
import { useAuth } from "@/lib/auth"

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
}
interface NavGroup {
  label: string
  items: NavItem[]
}

const NAV_GROUPS: NavGroup[] = [
  { label: "总览", items: [{ to: "/console", label: "控制台", icon: LayoutDashboard, end: true }] },
  {
    label: "编排与配置",
    items: [
      { to: "/console/builder", label: "Agent 管理", icon: Bot },
      { to: "/console/model-providers", label: "模型供应商", icon: Cpu },
    ],
  },
  { label: "调试与验证", items: [{ to: "/console/workbench", label: "Workbench", icon: MessagesSquare }] },
]

function isActive(item: NavItem, pathname: string) {
  return item.end ? pathname === item.to : pathname.startsWith(item.to)
}

function currentBreadcrumb(pathname: string) {
  for (const group of NAV_GROUPS) {
    for (const item of group.items) {
      if (isActive(item, pathname)) return { group: group.label, page: item.label }
    }
  }
  return { group: "总览", page: "控制台" }
}

// Console 布局参考阿里云控制台的经典结构：左侧固定导航（可折叠、分组）+
// 顶部面包屑与账号菜单条 + 右侧内容区，比之前单层顶部导航更适合承载持续
// 增长的管理功能（Agent 管理、模型供应商……），也更符合企业级控制台的操作习惯。
export default function ConsoleLayout() {
  const location = useLocation()
  const { logout } = useAuth()
  const [collapsed, setCollapsed] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)

  // 退出登录后不用手动跳转：登录态一清空，ProtectedRoute 自己就会在当前
  // 页面上叠一层模糊 + 登录弹窗（参考阿里云百炼），不需要导航去别的地方。
  const handleLogout = () => {
    setUserMenuOpen(false)
    logout()
  }

  const breadcrumb = currentBreadcrumb(location.pathname)

  return (
    <div className="flex min-h-screen">
      <aside
        className={cn(
          "sticky top-0 flex h-screen shrink-0 flex-col border-r border-glass-border bg-surface/60 backdrop-blur-xl transition-[width] duration-200",
          collapsed ? "w-16" : "w-60"
        )}
      >
        <div
          className={cn(
            "flex h-14 shrink-0 items-center gap-1.5 border-b border-glass-border px-3",
            collapsed && "justify-center px-0"
          )}
        >
          <button
            onClick={() => setCollapsed((c) => !c)}
            title={collapsed ? "展开导航" : "收起导航"}
            className="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
          >
            {collapsed ? <PanelLeftOpen className="size-4" /> : <PanelLeftClose className="size-4" />}
          </button>
          {!collapsed && (
            <Link to="/console" className="flex items-center gap-2 overflow-hidden">
              <LogoMark className="size-6 shrink-0 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
              <span className="truncate font-display text-base font-semibold">
                枢络
                <span className="ml-1 font-mono text-xs font-normal text-muted-foreground">Console</span>
              </span>
            </Link>
          )}
        </div>

        <nav className="flex-1 overflow-y-auto px-3 py-4">
          {NAV_GROUPS.map((group) => (
            <div key={group.label} className="mb-5">
              {!collapsed && (
                <div className="mb-1.5 px-2.5 text-xs font-medium tracking-wide text-muted-foreground/70">
                  {group.label}
                </div>
              )}
              <ul className="space-y-0.5">
                {group.items.map((item) => {
                  const active = isActive(item, location.pathname)
                  const Icon = item.icon
                  return (
                    <li key={item.to}>
                      <Link
                        to={item.to}
                        title={collapsed ? item.label : undefined}
                        className={cn(
                          "flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground",
                          collapsed && "justify-center",
                          active && "bg-primary/15 text-[#5aa6ff]"
                        )}
                      >
                        <Icon className="size-4 shrink-0" />
                        {!collapsed && <span>{item.label}</span>}
                      </Link>
                    </li>
                  )
                })}
              </ul>
            </div>
          ))}
        </nav>
      </aside>

      <div className="flex min-h-screen flex-1 flex-col">
        <header className="sticky top-0 z-40 flex h-14 shrink-0 items-center justify-between border-b border-glass-border bg-background/70 px-6 backdrop-blur-xl">
          <div className="flex items-center gap-1.5 text-sm">
            <span className="text-muted-foreground">{breadcrumb.group}</span>
            <ChevronRight className="size-3.5 text-muted-foreground/60" />
            <span className="font-medium text-foreground">{breadcrumb.page}</span>
          </div>

          <div className="relative">
            <button
              onClick={() => setUserMenuOpen((o) => !o)}
              className="flex items-center gap-2 rounded-full py-1 pl-1 pr-3 text-sm text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
            >
              <span className="flex size-7 items-center justify-center rounded-full bg-primary/20 font-display text-xs font-semibold text-[#5aa6ff]">
                A
              </span>
              开发者
            </button>
            {userMenuOpen && (
              <>
                <div className="fixed inset-0 z-40" onClick={() => setUserMenuOpen(false)} />
                <div className="glass-panel absolute right-0 top-11 z-50 w-40 overflow-hidden !rounded-xl p-1">
                  <button
                    onClick={handleLogout}
                    className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground"
                  >
                    <LogOut className="size-3.5" />
                    退出登录
                  </button>
                </div>
              </>
            )}
          </div>
        </header>

        <main className="flex-1 px-8 py-8">
          <div className="mx-auto max-w-6xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
