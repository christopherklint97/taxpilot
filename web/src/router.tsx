import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
} from "@tanstack/react-router"
import { AppShell } from "@/components/layout/app-shell"
import { DashboardPage } from "@/routes/index"
import { ReturnPage } from "@/routes/returns/return"

const rootRoute = createRootRoute({
  component: () => (
    <AppShell>
      <Outlet />
    </AppShell>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: DashboardPage,
})

const returnRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/returns/$returnId",
  component: ReturnPage,
})

const routeTree = rootRoute.addChildren([indexRoute, returnRoute])

export const router = createRouter({ routeTree })

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router
  }
}
