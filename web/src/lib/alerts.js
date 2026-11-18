import { AlertTriangle, Eye, Info, PlusCircle } from 'lucide-react';

// Niveaux d'alerte affichés (design : Urgent, À vérifier, Nouveau, Info)
export const LEVELS = {
  urgent: { label: 'Urgent', badge: 'bg-urgent-soft text-urgent-ink', fg: 'text-urgent-ink', border: 'border-urgent', icon: AlertTriangle, iconColor: '#C2410C' },
  check: { label: 'À vérifier', badge: 'bg-check-soft text-check-ink', fg: 'text-check-ink', border: 'border-check', icon: Eye, iconColor: '#7A4A00' },
  new: { label: 'Nouveau', badge: 'bg-brand-soft text-[#1E3AA8]', fg: 'text-[#1E3AA8]', border: 'border-brand', icon: PlusCircle, iconColor: '#2D5BFF' },
  info: { label: 'Info', badge: 'bg-off-soft text-body', fg: 'text-body', border: 'border-off', icon: Info, iconColor: '#3A4660' },
};

export function level(a) {
  if (a.type === 'risky_port') return a.severity === 'critical' ? 'urgent' : 'check';
  if (a.type === 'new_device') return 'new';
  return 'info';
}

export const LEVEL_ORDER = { urgent: 0, new: 1, check: 2, info: 3 };

// Explications par port risqué : ce qui se passe, pourquoi c'est important, comment corriger
const PORTS = {
  23: {
    title: (n) => `${n} accepte des connexions sans chiffrement (Telnet).`,
    what: 'L’appareil laisse ouvert un vieil accès à distance appelé Telnet. Les identifiants y passent en clair.',
    why: 'C’est l’une des portes d’entrée les plus utilisées par les virus qui s’attaquent aux objets connectés.',
    steps: (ip) => [`Ouvre la page de réglages de l’appareil depuis ton navigateur (adresse ${ip}).`, 'Dans Réseau ou Sécurité, désactive Telnet.', 'Change le mot de passe administrateur s’il n’a jamais été modifié, puis clique sur « C’est corrigé ».'],
  },
  21: {
    title: (n) => `${n} partage des fichiers sans chiffrement (FTP).`,
    what: 'Un serveur FTP est ouvert : les fichiers et les mots de passe y circulent en clair.',
    why: 'N’importe quel appareil du réseau peut lire ce qui passe, et les serveurs FTP mal réglés acceptent parfois des connexions sans mot de passe.',
    steps: (ip) => [`Ouvre l’interface de l’appareil (adresse ${ip}).`, 'Désactive le service FTP, ou remplace-le par SFTP / FTPS.', 'Vérifie qu’aucun compte « anonyme » n’est autorisé, puis clique sur « C’est corrigé ».'],
  },
  445: {
    title: (n) => `Le partage de fichiers de ${n} est ouvert à tout le réseau.`,
    what: 'Tous les appareils de la maison peuvent essayer d’accéder aux dossiers partagés.',
    why: 'Si un appareil est compromis (une caméra par exemple), il pourrait atteindre tes fichiers.',
    steps: () => ['Dans les réglages de partage de l’appareil, vérifie que SMB1 est désactivé.', 'Donne l’accès aux dossiers seulement aux comptes qui en ont besoin.', 'Si tu ne partages rien, désactive le partage de fichiers.'],
  },
  3389: {
    title: (n) => `Le Bureau à distance est activé sur ${n}.`,
    what: 'Windows permet de prendre le contrôle de cet ordinateur à distance. Cette fonction est actuellement allumée.',
    why: 'Si tu ne t’en sers pas, c’est une porte ouverte inutile. Elle est souvent ciblée quand elle est exposée.',
    steps: () => ['Sur le PC : Paramètres › Système › Bureau à distance.', 'Désactive l’option si tu ne l’utilises pas.', 'Sinon, laisse l’authentification au niveau du réseau (NLA) activée.'],
  },
  3306: database('MySQL', 'bind-address = 127.0.0.1', 'my.cnf'),
  1433: database('SQL Server', 'n’écouter que sur la machine elle-même', 'SQL Server Configuration Manager'),
  5432: database('PostgreSQL', "listen_addresses = 'localhost'", 'postgresql.conf'),
  6379: {
    title: () => 'Une base de données Redis est accessible depuis tout le réseau.',
    what: 'Le service Redis de cet appareil répond à n’importe quel appareil de la maison.',
    why: 'Redis n’a pas de mot de passe par défaut : quelqu’un sur le réseau pourrait lire ou modifier les données.',
    steps: () => ['Dans redis.conf, mets bind 127.0.0.1 ::1.', 'Ajoute une ligne requirepass avec un mot de passe long.', 'Redémarre Redis puis clique sur « C’est corrigé ».'],
  },
};

function database(name, setting, file) {
  return {
    title: () => `Une base de données ${name} est accessible depuis tout le réseau.`,
    what: `Le service ${name} de cet appareil répond à n’importe quel appareil de la maison.`,
    why: 'Une base de données ne devrait parler qu’aux programmes qui l’utilisent. Ouverte, elle peut être attaquée par mot de passe.',
    steps: () => [`Dans ${file}, règle ${setting}.`, 'Vérifie que tous les comptes ont un mot de passe solide.', 'Redémarre le service puis clique sur « C’est corrigé ».'],
  };
}

// Texte complet d'une alerte pour l'interface
export function explain(a, deviceName, device) {
  const name = deviceName || 'Un appareil';
  const ip = a.device_ip || device?.ip || '';
  if (a.type === 'risky_port') {
    const k = PORTS[a.port] || database(`(port ${a.port})`, 'l’écoute sur localhost uniquement', 'la configuration du service');
    return { title: k.title(name), what: k.what, why: k.why, steps: k.steps(ip) };
  }
  if (a.type === 'new_device') {
    const vendor = device?.vendor;
    return {
      title: 'Un appareil que tu ne connais pas encore s’est connecté.',
      what: vendor
        ? `Un nouvel appareil est apparu sur ton réseau. Son fabricant : ${vendor}.`
        : 'Un nouvel appareil est apparu sur ton réseau. Son fabricant n’a pas pu être identifié.',
      why: 'Savoir ce qui est branché chez toi, c’est la base. Un appareil inconnu peut aussi être un voisin qui utilise ton Wi-Fi.',
      steps: ['Pense à ce que tu as branché récemment (prise, ampoule, assistant vocal…).', 'Si tu le reconnais, clique sur « Oui, je le connais » et donne-lui un nom depuis la carte.', 'Sinon, change le mot de passe de ton Wi-Fi depuis l’interface de ta box.'],
    };
  }
  if (a.type === 'version_change') {
    return {
      title: `Un logiciel de ${name} a changé de version.`,
      what: 'Le programme qui répond sur un des ports de l’appareil n’a plus la même version qu’à la dernière vérification.',
      why: 'C’est généralement une mise à jour, donc une bonne nouvelle. Ça vaut juste le coup de vérifier que c’est bien toi (ou l’appareil lui-même) qui l’a faite.',
      steps: ['Si tu as fait des mises à jour récemment, rien à faire.', 'Sinon, regarde dans l’interface de l’appareil qui s’y est connecté dernièrement.'],
    };
  }
  if (a.type === 'device_offline') {
    return {
      title: `${name} n’a pas répondu à la dernière vérification.`,
      what: 'Il ne répond plus aux vérifications automatiques.',
      why: 'Souvent rien de grave : il est peut-être éteint ou hors de la maison.',
      steps: ['Allume-le et vérifie qu’il est connecté au réseau.', 'Si tu ne l’utilises plus, tu peux ignorer cette alerte.'],
    };
  }
  return { title: a.message, what: a.message, why: '', steps: [] };
}
