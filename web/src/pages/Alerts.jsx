import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { ErrorBox, Loading } from '../components/Empty.jsx';
import Header from '../components/Header.jsx';
import { LEVELS, LEVEL_ORDER, explain, level } from '../lib/alerts.js';
import { displayName } from '../lib/devices.js';
import { when } from '../lib/format.js';
import { useNetwork } from '../lib/network.jsx';

const FILTERS = [['all', 'Toutes'], ['urgent', 'Urgent'], ['check', 'À vérifier'], ['new', 'Nouveau'], ['info', 'Info']];

export default function Alerts() {
  const { alerts, devices, loading, error, summary, setAlertRead, markAllRead, requestScan, scanning, updateDevice } = useNetwork();
  const [params, setParams] = useSearchParams();
  const [filter, setFilter] = useState('all');
  const [tech, setTech] = useState(false);
  const [notice, setNotice] = useState('');

  const byId = useMemo(() => Object.fromEntries(devices.map((d) => [d.id, d])), [devices]);
  const sorted = useMemo(
    () => [...alerts].sort((a, b) => LEVEL_ORDER[level(a)] - LEVEL_ORDER[level(b)] || new Date(b.created_at) - new Date(a.created_at)),
    [alerts]
  );
  const visible = sorted.filter((a) => filter === 'all' || level(a) === filter);
  const selected = alerts.find((a) => String(a.id) === params.get('id')) || visible[0];

  // Ouvrir une alerte la marque comme lue
  useEffect(() => {
    if (selected && !selected.read && params.get('id')) setAlertRead(selected.id, true);
  }, [selected?.id]); // eslint-disable-line react-hooks/exhaustive-deps

  if (loading) return (<><Header wide /><Loading /></>);
  if (error && !summary) return (<><Header wide /><main className="px-4"><ErrorBox message={error} /></main></>);

  const counts = { all: alerts.length, urgent: 0, check: 0, new: 0, info: 0 };
  alerts.forEach((a) => counts[level(a)]++);
  const unread = alerts.filter((a) => !a.read).length;

  const device = selected && byId[selected.device_id];
  const name = selected && (device ? displayName(device) : selected.device_name || selected.device_ip);
  const text = selected && explain(selected, name, device);
  const lv = selected && LEVELS[level(selected)];

  return (
    <>
      <Header wide />
      <main className="mx-auto flex max-w-[1360px] flex-col gap-5 px-4 pb-14 pt-2 sm:px-8">
        <div className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
          <div className="flex flex-col gap-1.5">
            <h1 className="h-display text-[30px] sm:text-4xl">Tes alertes</h1>
            <p className="text-body">
              {unread} à lire{summary?.last_scan ? ` · dernière vérification ${when(summary.last_scan.start_time)}` : ''}
            </p>
          </div>
          {unread > 0 && <button type="button" className="btn btn-sm btn-outline" onClick={markAllRead}>Tout marquer comme lu</button>}
        </div>

        <div role="group" aria-label="Filtrer" className="flex flex-wrap gap-2">
          {FILTERS.map(([key, label]) => (
            <button
              key={key}
              type="button"
              aria-pressed={filter === key}
              onClick={() => { setFilter(key); setParams({}, { replace: true }); }}
              className={`h-11 rounded-full border px-[18px] font-medium transition ${filter === key ? 'border-ink bg-ink text-white' : 'border-line bg-white text-ink hover:border-ink'}`}
            >
              {label} · {counts[key]}
            </button>
          ))}
        </div>

        {alerts.length === 0 ? (
          <div className="card p-8 text-body">Aucune alerte. Tout est calme : je te préviens ici dès que quelque chose change.</div>
        ) : (
          <div className="flex flex-wrap items-start gap-5">
            <section aria-label="Liste des alertes" className="flex min-w-0 flex-[1_1_400px] flex-col gap-2.5">
              {visible.length === 0 && <p className="card p-5 text-body">Aucune alerte dans cette catégorie.</p>}
              {visible.map((a) => {
                const L = LEVELS[level(a)];
                const d = byId[a.device_id];
                const on = a.id === selected?.id;
                const Icon = L.icon;
                return (
                  <button
                    key={a.id}
                    type="button"
                    aria-pressed={on}
                    onClick={() => { setParams({ id: a.id }, { replace: true }); setTech(false); setNotice(''); }}
                    className={`flex w-full items-start gap-3.5 rounded-[18px] border-2 bg-white p-4 text-left transition ${on ? L.border : 'border-transparent hover:border-hairline'}`}
                  >
                    <span className={`flex size-11 flex-none items-center justify-center rounded-xl ${L.badge.split(' ')[0]}`}>
                      <Icon size={22} color={L.iconColor} strokeWidth={2} aria-hidden="true" />
                    </span>
                    <span className="flex min-w-0 flex-1 flex-col gap-1">
                      <span className="flex justify-between gap-2 text-[13px]">
                        <span className={`truncate font-semibold ${L.fg}`}>{L.label} · {d ? displayName(d) : a.device_name || a.device_ip}</span>
                        <span className="flex-none text-muted">{when(a.created_at)}</span>
                      </span>
                      <span className={`leading-snug text-ink ${a.read ? 'font-normal' : 'font-semibold'}`}>{explain(a, d ? displayName(d) : a.device_name, d).title}</span>
                    </span>
                    <span className={`mt-1.5 size-2.5 flex-none rounded-full ${a.read ? 'bg-transparent' : 'bg-brand'}`} aria-label={a.read ? undefined : 'non lue'} />
                  </button>
                );
              })}
            </section>

            {selected && (
              <section aria-label="Détail de l’alerte" className="card flex min-w-0 flex-[999_1_560px] flex-col gap-6 p-6 sm:p-8">
                <div className="flex flex-col gap-2.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className={`rounded-full px-3 py-1 text-[13px] font-semibold ${lv.badge}`}>{lv.label}</span>
                    <span className="text-[13px] text-muted">{name} · {when(selected.created_at)}</span>
                  </div>
                  <h2 className="h-display text-[26px] leading-[1.15] tracking-[-0.01em] sm:text-[30px]">{text.title}</h2>
                </div>

                <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(min(240px,100%),1fr))]">
                  <div className="flex flex-col gap-1.5 rounded-2xl bg-page p-[18px]">
                    <h3 className="text-[15px] font-semibold">Ce qui se passe</h3>
                    <p className="leading-relaxed text-body">{text.what}</p>
                  </div>
                  <div className="flex flex-col gap-1.5 rounded-2xl bg-page p-[18px]">
                    <h3 className="text-[15px] font-semibold">Pourquoi c’est important</h3>
                    <p className="leading-relaxed text-body">{text.why}</p>
                  </div>
                </div>

                {text.steps.length > 0 && (
                  <div className="flex flex-col gap-3">
                    <h3 className="h-display text-xl">Comment corriger</h3>
                    <ol className="flex flex-col gap-2.5">
                      {text.steps.map((s, i) => (
                        <li key={i} className="flex items-start gap-3.5">
                          <span className="flex size-[30px] flex-none items-center justify-center rounded-full bg-brand text-sm font-semibold text-white">{i + 1}</span>
                          <span className="pt-[3px] leading-relaxed">{s}</span>
                        </li>
                      ))}
                    </ol>
                  </div>
                )}

                <div className="flex flex-wrap gap-2.5">
                  {selected.type === 'new_device' ? (
                    <button type="button" className="btn btn-primary" onClick={() => updateDevice(selected.device_ip, { known: true })}>Oui, je le connais</button>
                  ) : (
                    <button
                      type="button"
                      className="btn btn-primary"
                      disabled={scanning}
                      onClick={() => requestScan().then(() => setNotice('Vérification lancée. L’alerte disparaîtra d’elle-même si le problème est réglé.'))}
                    >
                      {scanning ? 'Vérification en cours…' : 'C’est corrigé, revérifier'}
                    </button>
                  )}
                  {selected.device_ip && <Link to={`/maison?ip=${selected.device_ip}`} className="btn btn-outline">Voir l’appareil sur la carte</Link>}
                  <button type="button" className="btn btn-ghost" onClick={() => setAlertRead(selected.id, !selected.read)}>
                    {selected.read ? 'Marquer comme non lue' : 'Marquer comme lue'}
                  </button>
                </div>
                {notice && <p role="status" className="-mt-3 text-sm text-body">{notice}</p>}

                {selected.details && (
                  <div className="flex flex-col gap-3 border-t border-hairline pt-4">
                    <button
                      type="button"
                      onClick={() => setTech((t) => !t)}
                      aria-expanded={tech}
                      className="flex min-h-11 w-full items-center justify-between font-semibold text-ink"
                    >
                      Détails techniques <span className="text-sm font-medium text-brand-ink">{tech ? 'Masquer' : 'Afficher'}</span>
                    </button>
                    {tech && (
                      <pre className="overflow-x-auto whitespace-pre-wrap rounded-[14px] bg-ink p-4 font-mono text-[13px] leading-relaxed text-[#E3E8F5]">{selected.details}</pre>
                    )}
                  </div>
                )}
              </section>
            )}
          </div>
        )}
      </main>
    </>
  );
}
