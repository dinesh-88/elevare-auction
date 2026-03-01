import CountdownTimer from "./CountdownTimer";

export default function Countdown({ endTime, emphasizeUrgency = false }: { endTime: string; emphasizeUrgency?: boolean }) {
  return <CountdownTimer endTime={new Date(endTime)} showLabel={false} inline emphasizeUrgency={emphasizeUrgency} />;
}
