// Appels à l'API Go. En production nginx relaie /api ; en dev, Vite fait de même.
const BASE = '/api/v1';

async function request(path, options = {}) {
  const res = await fetch(BASE + path, {
    ...options,
    headers: options.body ? { 'Content-Type': 'application/json' } : undefined,
  });
  if (!res.ok) {
    let msg = `Erreur ${res.status}`;
    try {
      msg = (await res.json()).error || msg;
    } catch {}
    throw new Error(msg);
  }
  return res.json();
}

export const api = {
  summary: () => request('/summary'),
  devices: () => request('/devices').then((r) => r.devices || []),
  alerts: () => request('/alerts').then((r) => r.alerts || []),
  updateDevice: (ip, patch) => request(`/devices/${encodeURIComponent(ip)}`, { method: 'PATCH', body: JSON.stringify(patch) }),
  setAlertRead: (id, read) => request(`/alerts/${id}/read`, { method: 'POST', body: JSON.stringify({ read }) }),
  markAllRead: () => request('/alerts/read-all', { method: 'POST' }),
  requestScan: () => request('/scans', { method: 'POST' }),
};
