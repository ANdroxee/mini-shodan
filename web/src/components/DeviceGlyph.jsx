import { ICONS, kind } from '../lib/devices.js';

// Icône de l'appareil dans une tuile arrondie
export default function DeviceGlyph({ device, size = 36, icon = 20, className = 'bg-tint', color = '#3A4660' }) {
  const Icon = ICONS[kind(device)];
  return (
    <span className={`flex flex-none items-center justify-center rounded-[10px] ${className}`} style={{ width: size, height: size }}>
      <Icon size={icon} color={color} strokeWidth={1.8} aria-hidden="true" />
    </span>
  );
}
