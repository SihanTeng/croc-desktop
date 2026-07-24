// The croc mark — geometric crocodile head (snout slope, brow ridge, teeth
// zigzag, knockout eye + nostril). Hand-built Tier-B SVG, 64-grid, solid
// currentColor so it takes on the surrounding ink. Source: assets/logo.svg.

export function LogoMark({ size = 26 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="currentColor"
      aria-hidden="true"
    >
      <path
        fillRule="evenodd"
        d="
          M 6 34
          L 34 34
          L 40 28
          L 46 26
          L 52 28
          L 56 32
          L 56 46
          L 40 46
          L 37 49 L 34 46
          L 31 49 L 28 46
          L 25 49 L 22 46
          L 19 49 L 16 46
          L 14 44
          L 6 40
          Z
          M 42 29 h 4.5 v 4.5 h -4.5 Z
          M 10 35.5 h 2.5 v 2.5 h -2.5 Z
        "
      />
    </svg>
  );
}

export default function Logo() {
  return (
    <div className="logo">
      <LogoMark size={30} />
      <span className="wordmark">croc</span>
    </div>
  );
}
