import { Route, Routes } from 'react-router-dom';
import Header from './components/Header.jsx';
import Alerts from './pages/Alerts.jsx';
import Home from './pages/Home.jsx';
import House from './pages/House.jsx';
import Settings from './pages/Settings.jsx';

function NotFound() {
  return (
    <>
      <Header />
      <main className="mx-auto max-w-[1200px] px-4 py-16 text-center sm:px-8">
        <h1 className="h-display text-3xl">Page introuvable</h1>
      </main>
    </>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/maison" element={<House />} />
      <Route path="/alertes" element={<Alerts />} />
      <Route path="/reglages" element={<Settings />} />
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}
