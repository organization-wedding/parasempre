import {
  Outlet,
  createRootRoute,
  createRoute,
  createRouter,
  redirect,
} from "@tanstack/react-router";
import { LandingPage } from "./pages/LandingPage";
import { GuestListPage } from "./pages/GuestListPage";
import { GuestFormPage } from "./pages/GuestFormPage";
import { LoginPage } from "./pages/LoginPage";
import { UnderConstruction } from "./pages/UnderConstruction";
import { isAuthenticated } from "./lib/auth";
import "./index.css";

function requireAuth({ location }: { location: { pathname: string } }) {
  if (!isAuthenticated()) {
    throw redirect({
      to: "/login",
      search: { redirect: location.pathname },
    });
  }
}

function RootLayout() {
  return <Outlet />;
}

const rootRoute = createRootRoute({
  component: RootLayout,
});

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: LandingPage,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
  validateSearch: (search: Record<string, unknown>) => ({
    redirect: (search.redirect as string) || undefined,
  }),
});

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard",
  component: GuestListPage,
  beforeLoad: requireAuth,
});

const guestCreateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/novo",
  component: GuestFormPage,
  beforeLoad: requireAuth,
});

const guestEditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/$guestId",
  component: GuestEditRoute,
  beforeLoad: requireAuth,
});

const guestListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presenca",
  component: UnderConstruction,
});

function GuestEditRoute() {
  const { guestId } = guestEditRoute.useParams();
  return <GuestFormPage guestId={Number(guestId)} />;
}

const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  dashboardRoute,
  guestCreateRoute,
  guestEditRoute,
  guestListRoute,
]);

export const router = createRouter({
  routeTree,
  defaultNotFoundComponent: UnderConstruction,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
