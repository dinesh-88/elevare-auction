"use client";

import { useEffect, useState } from "react";

type Props = {
  startTime: string;
  status: "SCHEDULED" | "LIVE" | "ENDED" | string;
};

function getCountdown(startTime: string) {
  const difference = new Date(startTime).getTime() - Date.now();
  if (difference <= 0) {
    return { days: 0, hours: 0, minutes: 0 };
  }
  return {
    days: Math.floor(difference / (1000 * 60 * 60 * 24)),
    hours: Math.floor((difference / (1000 * 60 * 60)) % 24),
    minutes: Math.floor((difference / (1000 * 60)) % 60),
  };
}

export default function Countdown({ startTime, status }: Props) {
  const [timeLeft, setTimeLeft] = useState(() => getCountdown(startTime));

  useEffect(() => {
    setTimeLeft(getCountdown(startTime));
    const timer = window.setInterval(() => setTimeLeft(getCountdown(startTime)), 1000 * 30);
    return () => window.clearInterval(timer);
  }, [startTime]);

  if (status === "LIVE") {
    return <span className="text-sm font-medium text-success-500">Live</span>;
  }

  if (new Date(startTime).getTime() <= Date.now()) {
    return <span className="text-sm font-medium text-brand-700">Starting soon</span>;
  }

  return (
    <span className="text-sm font-medium tabular-nums text-brand-700">
      {timeLeft.days}d {String(timeLeft.hours).padStart(2, "0")}h {String(timeLeft.minutes).padStart(2, "0")}m
    </span>
  );
}
