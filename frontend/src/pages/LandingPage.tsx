import { Header } from "../components/Header";
import { Hero } from "../components/Hero";
import { MapSection } from "../components/MapSection";
import { Footer } from "../components/Footer";

export function LandingPage() {
  return (
    <div className="w-full overflow-x-hidden">
      <Header />
      <Hero />
      <MapSection />
      <Footer />
    </div>
  );
}
