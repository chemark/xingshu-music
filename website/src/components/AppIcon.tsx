// The app's icon, drawn from the layers in icon/AppIcon.icon: a cream star
// between two green leaves, on a dark card.
export function AppIcon(props: { size?: number; class?: string }) {
  return (
    <svg
      class={props.class}
      width={props.size ?? 32}
      height={props.size ?? 32}
      viewBox="0 0 1024 1024"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="icon-card" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stop-color="#12151C" />
          <stop offset="1" stop-color="#0E1116" />
        </linearGradient>
      </defs>
      <rect width="1024" height="1024" rx="230" fill="url(#icon-card)" />
      <path
        transform="translate(497 503) skewX(-3)"
        d="M-240.6 0C-120.3 -15.57 -9.62 -116.78 0 -259.5C9.62 -116.78 120.3 -15.57 240.6 0C120.3 15.57 9.62 116.78 0 259.5C-9.62 116.78 -120.3 15.57 -240.6 0Z"
        fill="#F8EFE4"
      />
      <g fill="#69984F">
        <path d="M518 246C565 199 684 179 764 204C823 224 845 278 845 338L845 477C808 432 674 423 599 378C550 348 525 298 518 246Z" />
        <path d="M488 773C445 820 336 840 276 815C232 795 207 746 207 686L207 532C239 577 326 592 396 621C450 646 480 696 488 773Z" />
      </g>
    </svg>
  )
}
