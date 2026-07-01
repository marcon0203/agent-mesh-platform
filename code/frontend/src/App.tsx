import { Navigate, Route, Routes } from "react-router-dom"
import AppLayout from "@/layouts/AppLayout"
import MarketplacePage from "@/pages/MarketplacePage"
import AgentBuilderPage from "@/pages/AgentBuilderPage"
import ModelProvidersPage from "@/pages/ModelProvidersPage"
import WorkbenchPage from "@/pages/WorkbenchPage"
import DashboardPage from "@/pages/DashboardPage"

export default function App() {
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<Navigate to="/marketplace" replace />} />
        <Route path="/marketplace" element={<MarketplacePage />} />
        <Route path="/builder" element={<AgentBuilderPage />} />
        <Route path="/model-providers" element={<ModelProvidersPage />} />
        <Route path="/workbench" element={<WorkbenchPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
      </Route>
    </Routes>
  )
}
