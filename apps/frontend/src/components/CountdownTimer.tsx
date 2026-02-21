import { useEffect, useState } from "react";

export default function CountdownTimer({ endTime, variant = "default" }: { endTime: Date; variant?: "default" | "large" }) {
  const [timeLeft, setTimeLeft] = useState(calculateTimeLeft());

  function calculateTimeLeft() {
    const difference = endTime.getTime() - new Date().getTime();
    if (difference <= 0) return { days: 0, hours: 0, minutes: 0, seconds: 0 };
    return {
      days: Math.floor(difference / (1000 * 60 * 60 * 24)),
      hours: Math.floor((difference / (1000 * 60 * 60)) % 24),
      minutes: Math.floor((difference / 1000 / 60) % 60),
      seconds: Math.floor((difference / 1000) % 60),
    };
  }

  useEffect(() => {
    const timer = window.setInterval(() => setTimeLeft(calculateTimeLeft()), 1000);
    return () => window.clearInterval(timer);
  }, [endTime]);

  const fmt = (num: number) => String(num).padStart(2, "0");
  if (variant === "large") {
    return (
      <div className="space-y-1">
        <div className="text-sm text-neutral-500">Auction ends in</div>
        <div className="bid-amount text-2xl font-heading font-semibold">
          {timeLeft.days > 0 && `${timeLeft.days}d `}
          {fmt(timeLeft.hours)}:{fmt(timeLeft.minutes)}:{fmt(timeLeft.seconds)}
        </div>
      </div>
    );
  }
  return (
    <div className="countdown bid-amount">
      Ends in {timeLeft.days > 0 && `${timeLeft.days}d `}
      {fmt(timeLeft.hours)}:{fmt(timeLeft.minutes)}:{fmt(timeLeft.seconds)}
    </div>
  );
}
