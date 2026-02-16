import brazaoImg from "./brasao-sem-fundo.svg";

export function CoatOfArms({ className = "" }: { className?: string }) {
  return (
    <img
      src={brazaoImg}
      alt="Brasão Para Sempre"
      className={className}
      width={120}
      height={155}
    />
  );
}
