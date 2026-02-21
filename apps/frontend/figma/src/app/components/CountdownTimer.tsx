import { useEffect, useState } from 'react';

interface CountdownTimerProps {
  endTime: Date;
  variant?: 'default' | 'large';
}

export function CountdownTimer({ endTime, variant = 'default' }: CountdownTimerProps) {
  const [timeLeft, setTimeLeft] = useState(calculateTimeLeft());

  function calculateTimeLeft() {
    const difference = endTime.getTime() - new Date().getTime();
    
    if (difference <= 0) {
      return { days: 0, hours: 0, minutes: 0, seconds: 0 };
    }

    return {
      days: Math.floor(difference / (1000 * 60 * 60 * 24)),
      hours: Math.floor((difference / (1000 * 60 * 60)) % 24),
      minutes: Math.floor((difference / 1000 / 60) % 60),
      seconds: Math.floor((difference / 1000) % 60),
    };
  }

  useEffect(() => {
    const timer = setInterval(() => {
      setTimeLeft(calculateTimeLeft());
    }, 1000);

    return () => clearInterval(timer);
  }, [endTime]);

  const formatTime = (num: number) => String(num).padStart(2, '0');

  if (variant === 'large') {
    return (
      <div className="space-y-1">
        <div className="text-sm text-muted-foreground">Auction ends in</div>
        <div className="text-2xl font-serif font-semibold tabular-nums">
          {timeLeft.days > 0 && `${timeLeft.days}d `}
          {formatTime(timeLeft.hours)}:{formatTime(timeLeft.minutes)}:{formatTime(timeLeft.seconds)}
        </div>
      </div>
    );
  } else {

  return (
    <div className="text-sm text-muted-foreground tabular-nums">
      Ends in {timeLeft.days > 0 && `${timeLeft.days}d `}
      {formatTime(timeLeft.hours)}:{formatTime(timeLeft.minutes)}:{formatTime(timeLeft.seconds)}
    </div>
  );
}
}
