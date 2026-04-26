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
import { GiftListPage } from "./pages/GiftListPage";
import { GiftDetailPage } from "./pages/GiftDetailPage";
import { GiftAdminPage } from "./pages/GiftAdminPage";
import { GiftFormPage } from "./pages/GiftFormPage";
import { GiftImportPage } from "./pages/GiftImportPage";
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

const giftListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presentes",
  component: GiftListRoute,
  validateSearch: (search: Record<string, unknown>) => {
    const n = Number(search.page);
    return { page: Number.isFinite(n) && n > 1 ? n : undefined };
  },
});

const giftDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presentes/$giftId",
  component: GiftDetailRoute,
});

const giftPurchaseRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presentes/$giftId/comprar",
  component: UnderConstruction,
});

const giftAdminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/presentes",
  component: GiftAdminPage,
  beforeLoad: requireAuth,
});

const giftAdminCreateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/presentes/novo",
  component: GiftFormPage,
  beforeLoad: requireAuth,
});

const giftAdminImportRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/presentes/importar",
  component: GiftImportPage,
  beforeLoad: requireAuth,
});

const giftAdminEditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/presentes/$giftId",
  component: GiftAdminEditRoute,
  beforeLoad: requireAuth,
});

function GuestEditRoute() {
  const { guestId } = guestEditRoute.useParams();
  return <GuestFormPage guestId={Number(guestId)} />;
}

function GiftAdminEditRoute() {
  const { giftId } = giftAdminEditRoute.useParams();
  return <GiftFormPage giftId={Number(giftId)} />;
}

function GiftListRoute() {
  const { page } = giftListRoute.useSearch();
  return <GiftListPage page={page ?? 1} />;
}

function GiftDetailRoute() {
  const { giftId } = giftDetailRoute.useParams();
  return <GiftDetailPage giftId={Number(giftId)} />;
}

const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  dashboardRoute,
  guestCreateRoute,
  guestEditRoute,
  guestListRoute,
  giftListRoute,
  giftDetailRoute,
  giftPurchaseRoute,
  giftAdminRoute,
  giftAdminCreateRoute,
  giftAdminImportRoute,
  giftAdminEditRoute,
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
