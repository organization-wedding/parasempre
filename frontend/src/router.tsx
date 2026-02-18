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
import "./index.css";

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

const guestListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presenca",
  component: GuestListPage,
});

const guestCreateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presenca/novo",
  component: GuestFormPage,
});

const guestEditRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lista-presenca/$guestId",
  component: GuestEditRoute,
});

const fallbackRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "*",
  component: UnderConstruction,
});

function GuestEditRoute() {
  const { guestId } = guestEditRoute.useParams();
  return <GuestFormPage guestId={Number(guestId)} />;
}

const routeTree = rootRoute.addChildren([
  homeRoute,
  guestListRoute,
  guestCreateRoute,
  guestEditRoute,
  fallbackRoute,
]);

export const router = createRouter({
  routeTree,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
