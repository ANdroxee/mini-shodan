// Anneau du score de sécurité (sur 100)
export default function ScoreRing({ score }) {
  const value = Math.max(0, Math.min(100, score ?? 0));
  const color = value >= 70 ? '#2D5BFF' : value >= 40 ? '#D89A1E' : '#C2410C';
  return (
    <div
      role="img"
      aria-label={`Score de sécurité : ${value} sur 100`}
      className="flex size-[168px] flex-none items-center justify-center rounded-full"
      style={{ background: `conic-gradient(${color} 0 ${value}%, #E3E8F5 ${value}% 100%)` }}
    >
      <div className="flex size-[132px] flex-col items-center justify-center rounded-full bg-white">
        <span className="font-display text-5xl font-extrabold leading-none">{value}</span>
        <span className="text-[13px] text-muted">sur 100</span>
      </div>
    </div>
  );
}
