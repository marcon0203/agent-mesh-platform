import { Navigate, Route, Routes } from "react-router-dom"
import PortalLayout from "@/layouts/PortalLayout"
import ConsoleLayout from "@/layouts/ConsoleLayout"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import PortalHomePage from "@/pages/PortalHomePage"
import MarketplacePage from "@/pages/MarketplacePage"
import LoginPage from "@/pages/LoginPage"
import AgentListPage from "@/pages/AgentListPage"
import AgentEditorPage from "@/pages/AgentEditorPage"
import ModelProvidersPage from "@/pages/ModelProvidersPage"
import WorkbenchPage from "@/pages/WorkbenchPage"
import DashboardPage from "@/pages/DashboardPage"

export default function App() {
  return (
    <Routes>
      {/* Portal：对外宣传 + 能力市场浏览，不需要登录 */}
      <Route element={<PortalLayout />}>
        <Route path="/" element={<PortalHomePage />} />
        <Route path="/marketplace" element={<MarketplacePage />} />
      </Route>

      <Route path="/login" element={<LoginPage />} />

      {/* Console：登录后的管理后台 */}
      <Route element={<ProtectedRoute />}>
        <Route element={<ConsoleLayout />}>
          <Route path="/console" element={<DashboardPage />} />
          <Route path="/console/builder" element={<AgentListPage />} />
          <Route path="/console/builder/:id" element={<AgentEditorPage />} />
          <Route path="/console/model-providers" element={<ModelProvidersPage />} />
          <Route path="/console/workbench" element={<WorkbenchPage />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
