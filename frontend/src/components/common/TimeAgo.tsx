import { formatDistanceToNow, parseISO } from 'date-fns';

interface TimeAgoProps {
  date: string;
  className?: string;
}

export function TimeAgo({ date, className }: TimeAgoProps) {
  let display: string;
  try {
    display = formatDistanceToNow(parseISO(date), { addSuffix: true });
  } catch {
    display = date;
  }

  return (
    <time dateTime={date} className={className} title={date}>
      {display}
    </time>
  );
}
