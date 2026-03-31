import { Link } from "@tanstack/react-router"
import { useUIStore } from "@/stores/ui-store"
import { cn } from "@/lib/utils"
import { FileText, Home, Plus } from "lucide-react"

const navItems = [
  { to: "/" as const, label: "Dashboard", icon: Home },
]

export function Sidebar() {
  const open = useUIStore((s) => s.sidebarOpen)

  return (
    <aside
      className={cn(
        "fixed inset-y-0 left-0 z-30 flex w-64 flex-col border-r border-sidebar-border bg-sidebar-background transition-transform",
        open ? "translate-x-0" : "-translate-x-full",
      )}
    >
      <div className="flex h-14 items-center gap-2 border-b border-sidebar-border px-4">
        <FileText className="h-6 w-6 text-sidebar-primary" />
        <span className="text-lg font-bold text-sidebar-primary">TaxPilot</span>
      </div>

      <nav className="flex-1 space-y-1 p-3">
        {navItems.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
            activeProps={{
              className:
                "bg-sidebar-accent text-sidebar-accent-foreground font-medium",
            }}
          >
            <item.icon className="h-4 w-4" />
            {item.label}
          </Link>
        ))}
      </nav>

      <div className="border-t border-sidebar-border p-3">
        <Link
          to="/"
          className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-sidebar-foreground hover:bg-sidebar-accent"
        >
          <Plus className="h-4 w-4" />
          New Return
        </Link>
      </div>
    </aside>
  )
}
