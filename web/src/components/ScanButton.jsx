import { Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useNetwork } from '../lib/network.jsx';

// « Vérifier maintenant » : demande un scan au scheduler et indique qu'il tourne
export default function ScanButton({ label = 'Vérifier maintenant', className = 'btn btn-primary' }) {
  const { scanning, requestScan } = useNetwork();
  const [error, setError] = useState('');
  return (
    <>
      <button
        type="button"
        className={className}
        disabled={scanning}
        onClick={() => requestScan().catch((e) => setError(e.message))}
      >
        {scanning && <Loader2 size={18} className="animate-spin" aria-hidden="true" />}
        {scanning ? 'Vérification en cours…' : label}
      </button>
      {error && <span role="alert" className="self-center text-sm text-urgent-ink">{error}</span>}
    </>
  );
}
