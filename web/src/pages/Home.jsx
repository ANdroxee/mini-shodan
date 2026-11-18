import { AlertTriangle, HelpCircle } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import DeviceGlyph from '../components/DeviceGlyph.jsx';
import { ErrorBox, Loading } from '../components/Empty.jsx';
import Header from '../components/Header.jsx';
import ScanButton from '../components/ScanButton.jsx';
import ScoreRing from '../components/ScoreRing.jsx';
import StatusBadge from '../components/StatusBadge.jsx';
import { LEVEL_ORDER, explain, level } from '../lib/alerts.js';
import { byPriority, detailLine, displayName } from '../lib/devices.js';
import { next, plural, when } from '../lib/format.js';
import { useNetwork } from '../lib/network.jsx';

function headline(score, todo) {
  if (todo === 0) return ['Ton réseau va bien.', 'Rien à régler.'];
  const things = `${todo} ${plural(todo, 'chose', 'choses')} à régler.`;
  if (score >= 60) return ['Ton réseau va plutôt bien.', things];
  return ['Ton réseau a besoin d’attention.', things];
}

function story(summary) {
  const { urgent, new: fresh, check } = summary.issues;
  const parts = [];
  if (urgent) parts.push(urgent === 1 ? 'un appareil laisse une porte grande ouverte' : `${urgent} appareils laissent une porte grande ouverte`);
  if (fresh) parts.push(fresh === 1 ? 'un appareil inconnu est apparu' : `${fresh} appareils inconnus sont apparus`);
  if (check) parts.push(check === 1 ? 'un réglage est à vérifier' : `${check} réglages sont à vérifier`);
  if (!parts.length) return 'Aucune porte ouverte, aucun inconnu. Tout est en ordre.';
  const text = parts.length > 1 ? `${parts.slice(0, -1).join(', ')} et ${parts.at(-1)}` : parts[0];
  return `${text.charAt(0).toUpperCase()}${text.slice(1)}.${check ? '' : ' Le reste est en ordre.'}`;
}

export default function Home() {
  const { summary, devices, alerts, loading, error, updateDevice } = useNetwork();
  const navigate = useNavigate();

  if (loading) return (<><Header /><Loading /></>);
  if (error && !summary) return (<><Header /><main className="px-4"><ErrorBox message={error} /></main></>);

  const byId = Object.fromEntries(devices.map((d) => [d.id, d]));
  const todoAlerts = alerts.filter((a) => a.type === 'risky_port' || a.type === 'new_device');
  const top = [...todoAlerts].sort((a, b) => LEVEL_ORDER[level(a)] - LEVEL_ORDER[level(b)]).slice(0, 2);
  const shown = [...devices].sort(byPriority).slice(0, 8);
  const last = summary.last_scan;
  const [l1, l2] = headline(summary.score, todoAlerts.length);

  return (
    <>
      <Header />
      <main className="mx-auto flex max-w-[1200px] flex-col gap-7 px-4 pb-14 pt-2 sm:px-8">
        <section aria-labelledby="etat" className="card flex flex-wrap items-center gap-8 p-6 sm:p-9">
          <ScoreRing score={last ? summary.score : 0} />
          <div className="flex min-w-0 flex-[1_1_360px] flex-col gap-2.5">
            {last ? (
              <>
                <h1 id="etat" className="h-display text-[32px] leading-[1.1] sm:text-[40px]">
                  {l1}
                  <br />
                  {l2}
                </h1>
                <p className="max-w-[560px] leading-relaxed text-body">
                  J’ai vérifié {summary.devices.online} {plural(summary.devices.online, 'appareil', 'appareils')} {when(last.start_time)}. {story(summary)}
                </p>
              </>
            ) : (
              <>
                <h1 id="etat" className="h-display text-[32px] leading-[1.1] sm:text-[40px]">
                  {summary.scanning ? 'Première vérification en cours…' : 'Pas encore de vérification.'}
                </h1>
                <p className="max-w-[560px] leading-relaxed text-body">
                  Je fais le tour de tout ce qui est connecté sur le réseau {summary.target}. Ça prend en général quelques minutes.
                </p>
              </>
            )}
            <div className="mt-1.5 flex flex-wrap gap-2.5">
              <ScanButton />
              <Link to="/maison" className="btn btn-outline">Voir la carte de ma maison</Link>
            </div>
          </div>
        </section>

        {last && (
          <section aria-labelledby="todo" className="flex flex-col gap-3.5">
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <h2 id="todo" className="h-display text-2xl">À régler</h2>
              {alerts.length > 0 && <Link to="/alertes" className="font-medium">Toutes les alertes ({alerts.length})</Link>}
            </div>
            {top.length === 0 ? (
              <div className="card rounded-[20px] p-6 text-body">Rien à régler pour l’instant. Je continue de surveiller.</div>
            ) : (
              <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(340px,100%),1fr))]">
                {top.map((a) => {
                  const d = byId[a.device_id];
                  const name = d ? displayName(d) : a.device_name || a.device_ip;
                  const text = explain(a, name, d);
                  const lv = level(a);
                  const urgent = lv === 'urgent' || lv === 'check';
                  return (
                    <article key={a.id} className={`flex flex-col gap-3 rounded-[20px] border-2 bg-white p-6 ${urgent ? (lv === 'urgent' ? 'border-urgent' : 'border-check') : 'border-brand'}`}>
                      <div className="flex items-center gap-2.5">
                        <span className={`flex size-10 items-center justify-center rounded-xl ${lv === 'new' ? 'bg-brand-soft' : lv === 'urgent' ? 'bg-urgent-tint' : 'bg-check-soft'}`}>
                          {lv === 'new'
                            ? <HelpCircle size={22} color="#2D5BFF" strokeWidth={2} aria-hidden="true" />
                            : <AlertTriangle size={22} color={lv === 'urgent' ? '#C2410C' : '#7A4A00'} strokeWidth={2} aria-hidden="true" />}
                        </span>
                        <span className={`text-[13px] font-semibold ${lv === 'new' ? 'text-brand-ink' : lv === 'urgent' ? 'text-urgent-ink' : 'text-check-ink'}`}>
                          {lv === 'new' ? `Nouveau · ${when(a.created_at)}` : `${lv === 'urgent' ? 'Urgent' : 'À vérifier'} · ${name}`}
                        </span>
                      </div>
                      <h3 className="text-[19px] font-semibold leading-snug">{lv === 'new' ? 'Un appareil inconnu s’est connecté. C’est à toi ?' : text.title}</h3>
                      <p className="leading-relaxed text-body">{lv === 'new' ? text.what : `${text.what} ${text.why}`}</p>
                      <div className="flex flex-wrap gap-2.5">
                        {lv === 'new' ? (
                          <>
                            <button type="button" className="btn btn-sm btn-dark" onClick={() => updateDevice(a.device_ip, { known: true })}>Oui, je le connais</button>
                            <Link to={`/alertes?id=${a.id}`} className="btn btn-sm btn-outline">Non, que faire ?</Link>
                          </>
                        ) : (
                          <>
                            <Link to={`/alertes?id=${a.id}`} className="btn btn-sm btn-dark">Comment corriger</Link>
                            <button type="button" className="btn btn-sm btn-ghost" onClick={() => navigate(`/maison?ip=${a.device_ip}`)}>Voir sur la carte</button>
                          </>
                        )}
                      </div>
                    </article>
                  );
                })}
              </div>
            )}
          </section>
        )}

        {devices.length > 0 && (
          <section aria-labelledby="apps" className="flex flex-col gap-3.5">
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <h2 id="apps" className="h-display text-2xl">Chez toi en ce moment</h2>
              <Link to="/maison" className="font-medium">Voir les {devices.length} appareils sur la carte</Link>
            </div>
            <div className="grid gap-3 [grid-template-columns:repeat(auto-fill,minmax(200px,1fr))]">
              {shown.map((d) => (
                <Link key={d.id} to={`/maison?ip=${d.ip}`} className="flex flex-col gap-2.5 rounded-2xl bg-white p-[18px] text-ink no-underline transition hover:-translate-y-0.5 hover:text-ink hover:shadow-[0_6px_20px_rgba(15,27,51,0.06)]">
                  <span className="flex items-center justify-between">
                    <DeviceGlyph device={d} />
                    <StatusBadge status={d.status} />
                  </span>
                  <span className="truncate font-semibold">{displayName(d)}</span>
                  <span className="truncate text-[13px] text-muted">{detailLine(d)}</span>
                </Link>
              ))}
            </div>
          </section>
        )}

        <p className="text-[13px] text-muted">
          {summary.next_scan && `Prochaine vérification automatique ${next(summary.next_scan)} · `}réseau {summary.target}
        </p>
      </main>
    </>
  );
}
