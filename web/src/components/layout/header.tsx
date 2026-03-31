import { useUIStore } from "@/stores/ui-store"
import { Menu } from "lucide-react"

export function Header() {
  const toggleSidebar = useUIStore((s) => s.toggleSidebar)

  return (
    <header className="flex h-14 items-center border-b border-border px-4">
      <button
        onClick={toggleSidebar}
        className="mr-4 rounded-md p-2 hover:bg-accent"
        aria-label="Toggle sidebar"
      >
        <Menu className="h-5 w-5" />
      </button>
      <h1 className="text-lg font-semibold">TaxPilot</h1>
    </header>
  )
}
