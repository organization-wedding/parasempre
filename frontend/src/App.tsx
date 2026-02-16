import { LandingPage } from "./pages/LandingPage";
import { UnderConstruction } from "./pages/UnderConstruction";
import "./index.css";

export function App() {
  if (window.location.pathname !== "/") {
    return <UnderConstruction />;
  }

  return <LandingPage />;
}

export default App;
