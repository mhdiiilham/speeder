interface Props {
  game: string;
  size?: number;
}

export default function GameLogo({ game, size = 28 }: Props) {
  if (game === "cs2" || game === "cs") {
    return (
      <img
        src="/cs2.png"
        alt="CS2"
        width={size}
        height={size}
        className="rounded"
        style={{ objectFit: "contain" }}
      />
    );
  }

  if (game === "dota2" || game === "dota") {
    return (
      <img
        src="/dota2.png"
        alt="Dota 2"
        width={size}
        height={size}
        className="rounded"
        style={{ objectFit: "contain" }}
      />
    );
  }

  return null;
}
