import { Router } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import DeviceGlyph from '../components/DeviceGlyph.jsx';
import { ErrorBox, Loading } from '../components/Empty.jsx';
import Header from '../components/Header.jsx';
import StatusBadge from '../components/StatusBadge.jsx';
import { explain } from '../lib/alerts.js';
import { ICONS, STATUS, byPriority, displayName, ipNum, kind, portLabel, summaryText } from '../lib/devices.js';
import { plural } from '../lib/format.js';
import { useNetwork } from '../lib/network.jsx';

const W = 900;
const H = 680;
const CX = W / 2;
const CY = H / 2;

// Place les appareils en ellipse autour de la box (deux anneaux au-delà de 14 appareils)
function layout(list) {
  const rings = list.length > 14 ? [Math.ceil(list.length / 2), Math.floor(list.length / 2)] : [list.length];
  const out = [];
  let i = 0;
  rings.forEach((count, r) => {
    const rx = r === 0 ? 350 : 215;
    const ry = r === 0 ? 262 : 158;
    for (let k = 0; k < count; k++, i++) {
      const a = -Math.PI / 2 + (k * 2 * Math.PI) / count + (r ? Math.PI / count : 0);
      out.push({ device: list[i], x: CX + rx * Math.cos(a), y: CY + ry * Math.sin(a) });
    }
  });
  return out;
}

function Legend() {
  return (
    <div className="flex flex-wrap gap-3.5 text-[13px] text-body">
      {['bad', 'warn', 'new', 'ok'].map((s) => (
        <span key={s} className="inline-flex items-center gap-1.5">
          <span className={`box-border size-3.5 rounded-full ${STATUS[s].ring}`} />
          {STATUS[s].label}
        </span>
      ))}
    </div>
  );
}

function Rename({ device, onDone }) {
  const { updateDevice } = useNetwork();
  const [value, setValue] = useState(device.name || '');
  const [busy, setBusy] = useState(false);
  return (
    <form
      className="flex flex-wrap gap-2"
      onSubmit={async (e) => {
        e.preventDefault();
        setBusy(true);
        try {
          await updateDevice(device.ip, { name: value.trim() });
          onDone();
        } finally {
          setBusy(false);
        }
      }}
    >
      <label htmlFor="rename" className="sr-only">Nouveau nom</label>
      <input
        id="rename"
        autoFocus
        maxLength={100}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder={displayName({ ...device, name: '' })}
        className="h-11 min-w-0 flex-1 rounded-full border border-line px-4 focus:border-brand focus:outline-none"
      />
      <button className="btn btn-sm btn-dark" disabled={busy}>Enregistrer</button>
      <button type="button" className="btn btn-sm btn-ghost" onClick={onDone}>Annuler</button>
    </form>
  );
}

export default function House() {
  const { devices, alerts, loading, error, summary, updateDevice, requestScan, scanning } = useNetwork();
  const [params, setParams] = useSearchParams();
  const [tech, setTech] = useState(true);
  const [renaming, setRenaming] = useState(false);

  // La box au centre ; les autres autour, triés par adresse
  const { gateway, others } = useMemo(() => {
    const gw = devices.find((d) => kind(d) === 'box');
    return { gateway: gw, others: devices.filter((d) => d !== gw).sort((a, b) => ipNum(a.ip) - ipNum(b.ip)) };
  }, [devices]);
  const nodes = useMemo(() => layout(others), [others]);

  const selectedIp = params.get('ip') || [...devices].sort(byPriority)[0]?.ip;
  const sel = devices.find((d) => d.ip === selectedIp);
  useEffect(() => setRenaming(false), [selectedIp]);

  if (loading) return (<><Header wide /><Loading /></>);
  if (error && !summary) return (<><Header wide /><main className="px-4"><ErrorBox message={error} /></main></>);

  const pick = (ip) => setParams({ ip }, { replace: true });
  const online = devices.filter((d) => d.online).length;
  const offline = devices.length - online;
  const selAlerts = sel ? alerts.filter((a) => a.device_id === sel.id) : [];
  const fixable = selAlerts.find((a) => a.type === 'risky_port' && a.severity === 'critical') || selAlerts.find((a) => a.type === 'risky_port') || selAlerts.find((a) => a.type === 'new_device');
  const summaryLine = sel && (summaryText(sel, alerts) ?? explain(fixable, displayName(sel), sel).what);
  const GatewayIcon = gateway ? ICONS.box : Router;

  return (
    <>
      <Header wide />
      <main className="mx-auto flex max-w-[1360px] flex-col gap-5 px-4 pb-14 pt-2 sm:px-8">
        <div className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
          <div className="flex flex-col gap-1.5">
            <h1 className="h-display text-[30px] sm:text-4xl">Tout ce qui est connecté chez toi</h1>
            <p className="text-body">
              {online} {plural(online, 'appareil', 'appareils')} en ligne{offline ? `, ${offline} hors ligne` : ''} · clique sur un appareil pour en savoir plus
            </p>
          </div>
          <Legend />
        </div>

        {devices.length === 0 ? (
          <div className="card p-8 text-body">Aucun appareil pour l’instant. La première vérification va les faire apparaître ici.</div>
        ) : (
          <div className="flex flex-wrap items-start gap-5">
            <section aria-label="Carte de la maison" className="card min-w-0 flex-[999_1_640px] p-4">
              <div className="overflow-x-auto">
                <div className="relative mx-auto" style={{ width: W, height: H }}>
                  <svg className="absolute inset-0" width={W} height={H} aria-hidden="true">
                    {nodes.map(({ device, x, y }) => {
                      const l = STATUS[device.status]?.line || STATUS.ok.line;
                      return <line key={device.id} x1={CX} y1={CY} x2={x} y2={y} stroke={l.stroke} strokeWidth={l.width} strokeDasharray={l.dash} />;
                    })}
                  </svg>

                  <button
                    type="button"
                    onClick={() => gateway && pick(gateway.ip)}
                    aria-pressed={gateway ? sel?.ip === gateway.ip : undefined}
                    disabled={!gateway}
                    className={`absolute flex size-[140px] flex-col items-center justify-center gap-1.5 rounded-full bg-ink text-white ${gateway && sel?.ip === gateway.ip ? 'ring-[5px] ring-brand' : ''}`}
                    style={{ left: CX - 70, top: CY - 70 }}
                  >
                    <GatewayIcon size={30} strokeWidth={1.8} aria-hidden="true" />
                    <span className="font-semibold">{gateway ? displayName(gateway) : 'Ton réseau'}</span>
                    <span className="text-xs text-[#C3CBE0]">{gateway ? STATUS[gateway.status]?.label : summary?.target}</span>
                  </button>

                  {nodes.map(({ device, x, y }) => {
                    const on = device.ip === sel?.ip;
                    return (
                      <button
                        key={device.id}
                        type="button"
                        onClick={() => pick(device.ip)}
                        aria-pressed={on}
                        className={`group absolute flex w-[120px] flex-col items-center gap-1.5 text-center ${device.status === 'off' ? 'opacity-60' : ''}`}
                        style={{ left: x - 60, top: y - 34 }}
                      >
                        <span
                          className={`box-border flex size-[68px] items-center justify-center rounded-full transition group-hover:scale-105 ${STATUS[device.status]?.ring} ${on ? 'bg-brand-soft shadow-[0_0_0_5px_#0F1B33]' : 'bg-white'}`}
                        >
                          <DeviceGlyph device={device} size={30} icon={26} className="" color="#0F1B33" />
                        </span>
                        <span className="line-clamp-2 max-w-full text-[13px] font-semibold leading-tight text-ink [overflow-wrap:anywhere]">{displayName(device)}</span>
                      </button>
                    );
                  })}
                </div>
              </div>
            </section>

            {sel && (
              <aside aria-label="Appareil sélectionné" className="flex min-w-0 flex-[1_1_360px] flex-col gap-4">
                <div className="card flex flex-col gap-4 p-7">
                  <div className="flex items-center gap-3.5">
                    <DeviceGlyph device={sel} size={56} icon={28} className="rounded-2xl bg-tint" color="#0F1B33" />
                    <div className="flex min-w-0 flex-col gap-1">
                      <h2 className="h-display break-words text-[26px] leading-[1.1]">{displayName(sel)}</h2>
                      <StatusBadge status={sel.status} className="self-start" />
                    </div>
                  </div>
                  <p className="leading-relaxed text-body">{summaryLine}</p>

                  {fixable?.type === 'risky_port' && (
                    <Link to={`/alertes?id=${fixable.id}`} className="btn btn-dark w-full">Voir comment corriger</Link>
                  )}
                  {sel.status === 'new' && (
                    <button type="button" className="btn btn-dark w-full" onClick={() => updateDevice(sel.ip, { known: true })}>Oui, je le connais</button>
                  )}

                  {renaming ? (
                    <Rename device={sel} onDone={() => setRenaming(false)} />
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      <button type="button" className="btn btn-sm btn-outline" onClick={() => setRenaming(true)}>Renommer</button>
                      <button type="button" className="btn btn-sm btn-outline" disabled={scanning} onClick={() => requestScan()}>
                        {scanning ? 'Vérification…' : 'Revérifier'}
                      </button>
                    </div>
                  )}

                  <div className="flex flex-col gap-3 border-t border-hairline pt-3.5">
                    <button
                      type="button"
                      onClick={() => setTech((t) => !t)}
                      aria-expanded={tech}
                      className="flex min-h-11 w-full items-center justify-between font-semibold text-ink"
                    >
                      Détails techniques <span className="text-sm font-medium text-brand-ink">{tech ? 'Masquer' : 'Afficher'}</span>
                    </button>
                    {tech && (
                      <>
                        <dl className="grid grid-cols-2 gap-3 text-sm">
                          <div><dt className="text-xs text-muted">Adresse IP</dt><dd className="mt-0.5 font-mono">{sel.ip}</dd></div>
                          <div><dt className="text-xs text-muted">Adresse MAC</dt><dd className="mt-0.5 break-all font-mono">{sel.mac || 'inconnue'}</dd></div>
                          <div><dt className="text-xs text-muted">Fabricant</dt><dd className="mt-0.5">{sel.vendor || 'inconnu'}</dd></div>
                          <div><dt className="text-xs text-muted">Système</dt><dd className="mt-0.5">{sel.os || 'inconnu'}</dd></div>
                        </dl>
                        <div className="flex flex-col gap-1.5">
                          <span className="text-xs text-muted">Ports ouverts</span>
                          {(sel.ports || []).length === 0 ? (
                            <div className="flex justify-between gap-2 rounded-[10px] bg-page px-3 py-2 text-sm"><span className="font-mono">—</span><span className="text-body">Aucun</span></div>
                          ) : (
                            sel.ports.map((p) => (
                              <div key={p.id} className={`flex justify-between gap-2 rounded-[10px] px-3 py-2 text-sm ${p.risk === 'critical' ? 'bg-urgent-tint' : p.risk ? 'bg-check-soft' : 'bg-page'}`}>
                                <span className="font-mono">{p.port}</span>
                                <span className="text-right text-body">{portLabel(p)}{p.version ? <span className="block text-xs text-muted">{p.version}</span> : null}</span>
                              </div>
                            ))
                          )}
                        </div>
                      </>
                    )}
                  </div>
                </div>
              </aside>
            )}
          </div>
        )}
      </main>
    </>
  );
}
