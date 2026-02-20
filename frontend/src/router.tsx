import {
  Outlet,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { LandingPage } from "./pages/LandingPage";
import { GuestListPage } from "./pages/GuestListPage";
import { GuestFormPage } from "./pages/GuestFormPage";
import { UnderConstruction } from "./pages/UnderConstruction";
import { ImpersonationModal } from "./components/ImpersonationModal";
import "./index.css";

function RootLayout() {
  return (
    <>
      <Outlet />
      <ImpersonationModal />
    </>
  );
}

const rootRoute = createRootRoute({
  component: RootLayout,
});

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: LandingPage,
});

const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard",
  component: GuestListPage,
});

const guestCreateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/novo",
  component: GuestFormPage,
});

const guestEditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/dashboard/$guestId",
  component: GuestEditRoute,
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
