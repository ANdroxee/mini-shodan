import Header from '../components/Header.jsx';
import ScanButton from '../components/ScanButton.jsx';
import { next, schedule, when } from '../lib/format.js';
import { useNetwork } from '../lib/network.jsx';

function Row({ label, children }) {
  return (
    <div className="flex flex-wrap justify-between gap-2 border-b border-hairline py-3.5 last:border-0">
      <dt className="text-body">{label}</dt>
      <dd className="font-medium">{children}</dd>
    </div>
  );
}

export default function Settings() {
  const { summary } = useNetwork();
  const last = summary?.last_scan;
  return (
    <>
      <Header />
      <main className="mx-auto flex max-w-[760px] flex-col gap-5 px-4 pb-14 pt-2 sm:px-8">
        <h1 className="h-display text-[30px] sm:text-4xl">Réglages</h1>
        <section className="card p-6 sm:p-8" aria-labelledby="surv">
          <h2 id="surv" className="h-display text-xl">Surveillance</h2>
          <dl className="mt-2">
            <Row label="Réseau surveillé"><span className="font-mono">{summary?.target}</span></Row>
            <Row label="Fréquence">{schedule(summary?.schedule)}</Row>
            <Row label="Prochaine vérification">{next(summary?.next_scan) || '—'}</Row>
            <Row label="Dernière vérification">
              {last ? `${when(last.start_time, { night: false })} · ${last.hosts_found} appareils · ${Math.max(1, Math.round(last.duration_seconds / 60))} min` : 'aucune'}
            </Row>
          </dl>
          <div className="mt-5 flex flex-wrap gap-2.5"><ScanButton /></div>
        </section>
        <section className="card p-6 text-body sm:p-8" aria-labelledby="change">
          <h2 id="change" className="h-display text-xl text-ink">Changer ces réglages</h2>
          <p className="mt-2 leading-relaxed">
            Le réseau et la fréquence se règlent dans le fichier <code className="rounded bg-page px-1.5 py-0.5 text-sm">.env</code> à
            côté de <code className="rounded bg-page px-1.5 py-0.5 text-sm">docker-compose.yml</code>, puis <code className="rounded bg-page px-1.5 py-0.5 text-sm">make up</code> :
          </p>
          <pre className="mt-3 overflow-x-auto rounded-[14px] bg-ink p-4 font-mono text-[13px] leading-relaxed text-[#E3E8F5]">{`SCAN_TARGET=192.168.1.0/24
SCAN_SCHEDULE=0 3 * * *     # tous les jours à 3 h
# SCAN_SCHEDULE=@every 30m  # toutes les 30 minutes`}</pre>
        </section>
      </main>
    </>
  );
}
