import type { ReactNode } from "react"
import { Sidebar } from "./sidebar"
import { Header } from "./header"
import { useUIStore } from "@/stores/ui-store"
import { cn } from "@/lib/utils"

export function AppShell({ children }: { children: ReactNode }) {
  const sidebarOpen = useUIStore((s) => s.sidebarOpen)

  return (
    <div className="flex h-screen overflow-hidden bg-background">
      <Sidebar />
      <div
        className={cn(
          "flex flex-1 flex-col overflow-hidden transition-all",
          sidebarOpen ? "ml-64" : "ml-0",
        )}
      >
        <Header />
        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  )
}
