"use client";

import { useEffect, useState } from "react";

type TimeLeft = {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
};

const EMPTY_TIME_LEFT: TimeLeft = { days: 0, hours: 0, minutes: 0, seconds: 0 };

export default function CountdownTimer({
  endTime,
  variant = "default",
  showLabel = true,
  inline = false,
}: {
  endTime: Date;
  variant?: "default" | "large";
  showLabel?: boolean;
  inline?: boolean;
}) {
  const [timeLeft, setTimeLeft] = useState<TimeLeft | null>(null);

  function calculateTimeLeft(): TimeLeft {
    const difference = endTime.getTime() - new Date().getTime();
    if (difference <= 0) return EMPTY_TIME_LEFT;
    return {
      days: Math.floor(difference / (1000 * 60 * 60 * 24)),
      hours: Math.floor((difference / (1000 * 60 * 60)) % 24),
      minutes: Math.floor((difference / 1000 / 60) % 60),
      seconds: Math.floor((difference / 1000) % 60),
    };
  }

  useEffect(() => {
    setTimeLeft(calculateTimeLeft());
    const timer = window.setInterval(() => setTimeLeft(calculateTimeLeft()), 1000);
    return () => window.clearInterval(timer);
  }, [endTime]);

  const fmt = (num: number) => String(num).padStart(2, "0");
  const display = timeLeft ?? EMPTY_TIME_LEFT;
  const countdownText = timeLeft
    ? `${display.days > 0 ? `${display.days}d ` : ""}${fmt(display.hours)}:${fmt(display.minutes)}:${fmt(display.seconds)}`
    : "--:--:--";
  if (variant === "large") {
    return (
      <div className="space-y-1">
        {showLabel ? <div className="text-sm text-neutral-500">Auction ends in</div> : null}
        <div className="text-2xl font-heading font-semibold tabular-nums">{countdownText}</div>
      </div>
    );
  }
  const Tag = inline ? "span" : "div";
  return (
    <Tag className="text-sm text-neutral-500 tabular-nums">
      {showLabel ? "Ends in " : ""}
      {countdownText}
    </Tag>
  );
}
