// États de chargement / erreur / réseau pas encore scanné
export function Loading() {
  return <p className="py-20 text-center text-muted">Chargement…</p>;
}

export function ErrorBox({ message }) {
  return (
    <div role="alert" className="card mx-auto max-w-xl p-8 text-center">
      <h1 className="h-display text-2xl">Impossible de joindre l’API</h1>
      <p className="mt-2 text-body">{message}</p>
      <p className="mt-2 text-sm text-muted">Vérifie que les conteneurs tournent : <code>make up</code> puis <code>make logs-api</code>.</p>
    </div>
  );
}
