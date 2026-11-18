import { Link, NavLink } from 'react-router-dom';
import { useNetwork } from '../lib/network.jsx';

const LINKS = [
  { to: '/', label: 'Accueil', end: true },
  { to: '/maison', label: 'Ma maison' },
  { to: '/alertes', label: 'Alertes', badge: true },
  { to: '/reglages', label: 'Réglages' },
];

export default function Header({ wide = false }) {
  const { alerts } = useNetwork();
  const unread = alerts.filter((a) => !a.read).length;
  return (
    <header className={`mx-auto flex flex-wrap items-center justify-between gap-3 px-4 py-6 sm:px-8 ${wide ? 'max-w-[1360px]' : 'max-w-[1200px]'}`}>
      <Link to="/" className="font-display text-[22px] font-extrabold tracking-[-0.01em] text-ink no-underline hover:text-ink">
        Shodan<span className="text-brand">@</span>Home
      </Link>
      <nav aria-label="Navigation" className="flex flex-wrap gap-1">
        {LINKS.map((l) => (
          <NavLink
            key={l.to}
            to={l.to}
            end={l.end}
            className={({ isActive }) =>
              `inline-flex items-center gap-2 rounded-full px-4 py-2.5 font-medium no-underline transition ${
                isActive ? 'bg-ink text-white hover:text-white' : 'text-body hover:bg-white hover:text-ink'
              }`
            }
          >
            {l.label}
            {l.badge && unread > 0 && (
              <span className="inline-flex h-[22px] min-w-[22px] items-center justify-center rounded-full bg-urgent px-1.5 text-xs font-semibold text-white">
                {unread}
              </span>
            )}
          </NavLink>
        ))}
      </nav>
    </header>
  );
}
