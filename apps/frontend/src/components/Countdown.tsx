import CountdownTimer from "./CountdownTimer";

export default function Countdown({ endTime }: { endTime: string }) {
  return <CountdownTimer endTime={new Date(endTime)} />;
}
