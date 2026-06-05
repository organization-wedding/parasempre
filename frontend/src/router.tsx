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
import { RegisterAttendancePage } from "./pages/RegisterAttendancePage";
import { UnderConstruction } from "./pages/UnderConstruction";
import { GiftListPage } from "./pages/GiftListPage";
import { GiftDetailPage } from "./pages/GiftDetailPage";
import { GiftCheckoutPage } from "./pages/GiftCheckoutPage";
import { GiftAdminPage } from "./pages/GiftAdminPage";
import { GiftFormPage } from "./pages/GiftFormPage";
import { GiftImportPage } from "./pages/GiftImportPage";
import { MyGiftsPage } from "./pages/MyGiftsPage";
import { AdminTransactionsPage } from "./pages/AdminTransactionsPage";
import { AdminGiftMessagesPage } from "./pages/AdminGiftMessagesPage";
import { AdminLayout } from "./components/AdminLayout";
import { isAuthenticated, setAuth } from "./lib/auth";
import { devLogin } from "./lib/api";
import { IS_DEV } from "./config";
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
  validateSearch: (search: Record<string, unknown>) => ({
    autologin: typeof search.autologin === "string" ? search.autologin : undefined,
  }),
  beforeLoad: async ({ location, search }) => {
    const uracf = (search as { autologin?: string }).autologin;
    if (!uracf || !IS_DEV) return;

    const withoutAutologin = (() => {
      const params = new URLSearchParams(location.searchStr);
      params.delete("autologin");
      const qs = params.toString();
      return qs ? `${location.pathname}?${qs}` : location.pathname;
    })();

    try {
      const res = await devLogin(uracf);
      setAuth(res.token, res.role, res.uracf);
    } catch (err) {
      console.warn("autologin indisponível:", err);
      return;
    }

    throw redirect({ to: withoutAutologin });
  },
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

const adminRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/admin",
  component: AdminLayout,
  beforeLoad: requireAuth,
});

const adminIndexRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "/",
  component: GuestListPage,
});

const guestCreateRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "novo",
  component: GuestFormPage,
});

const guestEditRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "$guestId",
  component: GuestEditRoute,
});

const giftAdminRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "presentes",
  component: GiftAdminPage,
});

const giftAdminCreateRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "presentes/novo",
  component: GiftFormPage,
});

const giftAdminImportRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "presentes/importar",
  component: GiftImportPage,
});

const giftAdminEditRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "presentes/$giftId",
  component: GiftAdminEditRoute,
});

const adminTransactionsRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "pagamentos",
  component: AdminTransactionsPage,
});

const adminGiftMessagesRoute = createRoute({
  getParentRoute: () => adminRoute,
  path: "recados",
  component: AdminGiftMessagesPage,
});

const dashboardLegacyRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/$",
  component: () => null,
  beforeLoad: ({ params }) => {
    const rest = (params as { _splat?: string })._splat ?? "";
    throw redirect({ to: rest ? `/admin/${rest}` : "/admin" });
  },
});

const dashboardLegacyRootRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard",
  component: () => null,
  beforeLoad: () => {
    throw redirect({ to: "/admin" });
  },
});

const registerAttendanceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/registrar-presenca",
  component: RegisterAttendancePage,
  beforeLoad: requireAuth,
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
  component: GiftCheckoutRoute,
  beforeLoad: requireAuth,
});

function GiftCheckoutRoute() {
  const { giftId } = giftPurchaseRoute.useParams();
  return <GiftCheckoutPage giftId={Number(giftId)} />;
}

const myGiftsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/meus-presentes",
  component: MyGiftsRoute,
  beforeLoad: requireAuth,
  validateSearch: (search: Record<string, unknown>) => {
    const n = Number(search.page);
    return { page: Number.isFinite(n) && n > 1 ? n : undefined };
  },
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

function MyGiftsRoute() {
  const { page } = myGiftsRoute.useSearch();
  return <MyGiftsPage page={page ?? 1} />;
}

const routeTree = rootRoute.addChildren([
  homeRoute,
  loginRoute,
  adminRoute.addChildren([
    adminIndexRoute,
    guestCreateRoute,
    guestEditRoute,
    giftAdminRoute,
    giftAdminCreateRoute,
    giftAdminImportRoute,
    giftAdminEditRoute,
    adminTransactionsRoute,
    adminGiftMessagesRoute,
  ]),
  dashboardLegacyRootRoute,
  dashboardLegacyRoute,
  registerAttendanceRoute,
  giftListRoute,
  giftDetailRoute,
  giftPurchaseRoute,
  myGiftsRoute,
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
