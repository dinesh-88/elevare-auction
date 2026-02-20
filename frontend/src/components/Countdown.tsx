import { useEffect, useMemo, useState } from "react";

function format(secondsTotal: number) {
  const h = Math.floor(secondsTotal / 3600)
    .toString()
    .padStart(2, "0");
  const m = Math.floor((secondsTotal % 3600) / 60)
    .toString()
    .padStart(2, "0");
  const s = Math.floor(secondsTotal % 60)
    .toString()
    .padStart(2, "0");
  return `${h}:${m}:${s}`;
}

export default function Countdown({ endTime }: { endTime: string }) {
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, []);

  const remaining = useMemo(() => {
    const end = new Date(endTime).getTime();
    return Math.max(0, Math.floor((end - now) / 1000));
  }, [endTime, now]);

  return <span className={remaining === 0 ? "countdown ended" : "countdown"}>{format(remaining)}</span>;
}
