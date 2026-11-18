// Dates en langage courant : « cette nuit à 3 h », « hier à 18 h 40 », « 15 nov. »
const sameDay = (a, b) => a.toDateString() === b.toDateString();

function hour(d) {
  const h = d.getHours();
  const m = d.getMinutes();
  return m ? `${h} h ${String(m).padStart(2, '0')}` : `${h} h`;
}

export function when(iso, { night = true } = {}) {
  if (!iso) return '';
  const d = new Date(iso);
  const now = new Date();
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (sameDay(d, now)) {
    if (night && d.getHours() < 6) return `cette nuit à ${hour(d)}`;
    return `aujourd’hui à ${hour(d)}`;
  }
  if (sameDay(d, yesterday)) return `hier à ${hour(d)}`;
  return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' });
}

export function next(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  const now = new Date();
  const tomorrow = new Date(now);
  tomorrow.setDate(now.getDate() + 1);
  if (sameDay(d, now)) return `aujourd’hui à ${hour(d)}`;
  if (sameDay(d, tomorrow)) return `demain à ${hour(d)}`;
  return `le ${d.toLocaleDateString('fr-FR', { weekday: 'long', day: 'numeric', month: 'long' })} à ${hour(d)}`;
}

// Expression cron → phrase. Les cas courants seulement ; sinon on affiche l'expression.
export function schedule(expr) {
  if (!expr) return '';
  const every = expr.match(/^@every\s+(\d+)([mh])$/);
  if (every) return `Toutes les ${every[1]} ${every[2] === 'h' ? 'heures' : 'minutes'}`;
  const p = expr.trim().split(/\s+/);
  if (p.length === 5 && p[2] === '*' && p[3] === '*' && p[4] === '*') {
    if (/^\d+$/.test(p[0]) && /^\d+$/.test(p[1])) return `Tous les jours à ${hour(new Date(2000, 0, 1, +p[1], +p[0]))}`;
    const h = p[1].match(/^\*\/(\d+)$/);
    if (/^\d+$/.test(p[0]) && h) return `Toutes les ${h[1]} heures`;
    const m = p[0].match(/^\*\/(\d+)$/);
    if (m && p[1] === '*') return `Toutes les ${m[1]} minutes`;
  }
  return `Planning : ${expr}`;
}

export const plural = (n, one, many) => (n > 1 ? many : one);
