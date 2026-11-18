import { Cpu, HardDrive, Laptop, Monitor, Network, Printer, Router, Server, Smartphone, Tablet, Tv, Video } from 'lucide-react';

// Statuts d'appareil (calculés par l'API) → libellés et couleurs du design
export const STATUS = {
  bad: { label: 'Urgent', badge: 'bg-urgent-soft text-urgent-ink', ring: 'border-[3px] border-urgent', line: { stroke: '#E9A27F', width: 2 } },
  warn: { label: 'À vérifier', badge: 'bg-check-soft text-check-ink', ring: 'border-[3px] border-check', line: { stroke: '#EBCB85', width: 2 } },
  new: { label: 'Nouveau', badge: 'bg-brand-soft text-[#1E3AA8]', ring: 'border-[3px] border-dashed border-brand', line: { stroke: '#8EA6FF', width: 2, dash: '6 5' } },
  ok: { label: 'RAS', badge: 'bg-fine-soft text-fine-ink', ring: 'border-2 border-line', line: { stroke: '#D5DBE8', width: 1 } },
  off: { label: 'Hors ligne', badge: 'bg-off-soft text-body', ring: 'border-2 border-dashed border-off', line: { stroke: '#C8D0E2', width: 1, dash: '4 4' } },
};

// Ordre d'affichage : ce qui demande de l'attention d'abord
export const STATUS_ORDER = { bad: 0, new: 1, warn: 2, off: 4, ok: 3 };
export const byPriority = (a, b) => (STATUS_ORDER[a.status] ?? 9) - (STATUS_ORDER[b.status] ?? 9) || ipNum(a.ip) - ipNum(b.ip);

export function ipNum(ip = '') {
  return ip.split('.').reduce((n, part) => n * 256 + (Number(part) || 0), 0);
}

const has = (d, ...ports) => (d.ports || []).some((p) => ports.includes(p.port));
const match = (text, re) => re.test(text || '');

// Type d'appareil deviné à partir du fabricant, du nom, du système et des ports ouverts
export function kind(d) {
  const all = `${d.vendor} ${d.hostname} ${d.name} ${d.os}`;
  if (/\.(1|254)$/.test(d.ip) || match(all, /\b(box|livebox|freebox|bbox|router|gateway|routeur|sagemcom|fritz)/i)) return 'box';
  if (has(d, 554, 8554) || match(all, /hikvision|dahua|reolink|axis comm|ezviz|foscam|amcrest|camera|cam\b/i)) return 'cam';
  if (match(all, /synology|qnap|western digital|nas\b|truenas|unraid/i) || has(d, 5000, 5001)) return 'nas';
  if (has(d, 631, 9100, 515) || match(all, /brother|epson|canon|hewlett|\bhp\b|lexmark|xerox|kyocera|printer|imprimante/i)) return 'print';
  if (match(all, /raspberry/i)) return 'pi';
  if (has(d, 8008, 8009, 8001, 8002) || match(all, /samsung tv|lg electronics|tizen|webos|roku|chromecast|android tv|\btv\b|bravia/i)) return 'tv';
  if (match(all, /ipad|\btab\b|tablet|tablette/i)) return 'tablet';
  if (has(d, 62078) || match(all, /iphone|android|pixel|galaxy|oneplus|xiaomi|redmi|huawei|\bios\b/i)) return 'phone';
  if (match(all, /macbook|mac os|macos|laptop|portable/i)) return 'laptop';
  if (has(d, 3389, 135, 139) || match(all, /windows|micro-star|msi|dell|lenovo|asus|desktop|\bpc\b/i)) return 'pc';
  if (match(all, /switch|tp-link|netgear|ubiquiti|unifi|mikrotik/i)) return 'switch';
  if (has(d, 22) && match(all, /linux|debian|ubuntu/i)) return 'server';
  return 'chip';
}

export const ICONS = {
  box: Router, cam: Video, nas: HardDrive, print: Printer, pi: Cpu, tv: Tv, tablet: Tablet,
  phone: Smartphone, laptop: Laptop, pc: Monitor, switch: Network, server: Server, chip: Cpu,
};

const KIND_NAMES = {
  box: 'Box internet', cam: 'Caméra', nas: 'NAS', print: 'Imprimante', pi: 'Raspberry Pi', tv: 'Télévision',
  tablet: 'Tablette', phone: 'Téléphone', laptop: 'Ordinateur portable', pc: 'Ordinateur', switch: 'Switch réseau', server: 'Serveur',
};

// Nom affiché : celui donné par l'utilisateur, sinon le nom réseau, sinon le type deviné
export function displayName(d) {
  if (d.name) return d.name;
  const hostname = d.hostname || '';
  const host = /^\d+\.\d+\.\d+\.\d+$/.test(hostname) ? '' : hostname.split('.')[0];
  if (host) return host;
  const k = kind(d);
  if (KIND_NAMES[k]) return d.vendor && !['box', 'pi'].includes(k) ? `${KIND_NAMES[k]} ${d.vendor}` : KIND_NAMES[k];
  return d.known ? `Appareil ${d.vendor || d.ip}` : 'Appareil inconnu';
}

// Libellés des ports en français courant
const PORT_LABELS = {
  21: 'Transfert de fichiers (FTP) · risqué', 22: 'Accès SSH', 23: 'Telnet · risqué', 25: 'Courrier (SMTP)', 53: 'Annuaire réseau (DNS)',
  80: 'Page web', 110: 'Courrier (POP3)', 135: 'Service Windows', 139: 'Partage Windows', 143: 'Courrier (IMAP)',
  443: 'Page web sécurisée', 445: 'Partage de fichiers', 515: 'Impression', 548: 'Partage Apple', 554: 'Flux vidéo',
  631: 'Impression', 1433: 'Base SQL Server', 1883: 'Objets connectés (MQTT)', 3306: 'Base MySQL', 3389: 'Bureau à distance',
  5000: 'Interface web', 5001: 'Interface web sécurisée', 5432: 'Base PostgreSQL', 5900: 'Partage d’écran (VNC)',
  6379: 'Base Redis', 8001: 'Télécommande', 8002: 'Télécommande (sécurisée)', 8008: 'Diffusion (Cast)', 8009: 'Diffusion (Cast)',
  8080: 'Page web', 8443: 'Page web sécurisée', 9100: 'Impression', 62078: 'Synchronisation',
};

export function portLabel(p) {
  return PORT_LABELS[p.port] || (p.service ? p.service.charAt(0).toUpperCase() + p.service.slice(1) : 'Service inconnu');
}

// Petit résumé d'une ligne pour les cartes de l'accueil
export function detailLine(d) {
  if (d.status === 'off') return 'Hors ligne pour l’instant';
  const risky = (d.ports || []).find((p) => p.risk);
  const brand = d.vendor || d.os || '';
  if (risky) return [brand, `${portLabel(risky).replace(' · risqué', '')} ouvert`].filter(Boolean).join(' · ');
  if (d.status === 'new') return [brand, 'apparu récemment'].filter(Boolean).join(' · ');
  return 'Tout est normal';
}

// Résumé de l'appareil sélectionné sur la carte
export function summaryText(d, alerts) {
  if (d.status === 'off') return 'Pas vu lors de la dernière vérification. Normal s’il est éteint ou sorti de la maison.';
  const mine = alerts.filter((a) => a.device_id === d.id);
  const top = mine.find((a) => a.type === 'risky_port' && a.severity === 'critical') || mine.find((a) => a.type === 'risky_port');
  if (top) return null; // le texte vient de l'alerte (voir alerts.js)
  if (d.status === 'new') return 'Apparu récemment sur ton réseau. Tu le reconnais ?';
  if (!(d.ports || []).length) return 'Aucun port ouvert, c’est parfait.';
  return 'Rien à signaler.';
}
