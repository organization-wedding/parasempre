import { LandingPage } from "./pages/LandingPage";
import { UnderConstruction } from "./pages/UnderConstruction";
import { GuestListPage } from "./pages/GuestListPage";
import { GuestFormPage } from "./pages/GuestFormPage";
import "./index.css";

export function App() {
  const path = window.location.pathname;

  if (path === "/") return <LandingPage />;
  if (path === "/lista-presenca") return <GuestListPage />;
  if (path === "/lista-presenca/novo") return <GuestFormPage />;

  const editMatch = path.match(/^\/lista-presenca\/(\d+)$/);
  if (editMatch) return <GuestFormPage guestId={Number(editMatch[1])} />;

  return <UnderConstruction />;
}

export default App;
