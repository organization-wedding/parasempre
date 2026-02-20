import brasaoImg from "../assets/brasao-sem-fundo-padrao.png";

export function CoatOfArms({ className = "" }: { className?: string }) {
  return (
    <img
      src={brasaoImg}
      alt="Brasão Para Sempre"
      className={className}
    />
  );
}
