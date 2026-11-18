import { STATUS } from '../lib/devices.js';

export default function StatusBadge({ status, className = '' }) {
  const s = STATUS[status] || STATUS.ok;
  return <span className={`inline-block rounded-full px-2.5 py-1 text-xs font-semibold ${s.badge} ${className}`}>{s.label}</span>;
}
