import { Outlet } from "react-router-dom"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { LogoMark } from "@/components/Logo"
import { useAuth } from "@/lib/auth"

// 挡住 /console/* 路由：参考阿里云百炼的交互——未登录时不整页跳走，
// 而是照常渲染目标页面（当"预览"），叠一层模糊 + 登录弹窗上去，登录成功
// 弹窗直接消失、露出刚才那个页面，不用再跳一次。
export function ProtectedRoute() {
  const { isAuthenticated, login } = useAuth()

  if (!isAuthenticated) {
    return (
      <div className="relative min-h-screen">
        <div aria-hidden className="pointer-events-none max-h-screen select-none overflow-hidden blur-md">
          <Outlet />
        </div>
        <Dialog open modal>
          <DialogContent
            showCloseButton={false}
            onPointerDownOutside={(e) => e.preventDefault()}
            onEscapeKeyDown={(e) => e.preventDefault()}
          >
            <DialogHeader>
              <div className="mb-2 flex items-center gap-2.5">
                <LogoMark className="size-7 drop-shadow-[0_0_6px_rgba(90,166,255,0.5)]" />
                <span className="font-display text-base font-semibold">
                  枢络<span className="ml-1 font-mono text-xs font-normal text-muted-foreground">AgentMesh</span>
                </span>
              </div>
              <DialogTitle>登录后查看控制台</DialogTitle>
              <DialogDescription>
                登录后即可管理你的 Agent、模型供应商和调用数据。平台账号体系还没接入，这里先用一个假登录跑通交互。
              </DialogDescription>
            </DialogHeader>
            <Button size="lg" className="w-full" onClick={login}>
              登录
            </Button>
          </DialogContent>
        </Dialog>
      </div>
    )
  }

  return <Outlet />
}
