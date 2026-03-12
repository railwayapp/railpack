import { format } from "date-fns";

export function getCurrentTimestamp(): string {
  return format(new Date(), "yyyy-MM-dd HH:mm:ss");
}

export function formatHealthCheck(status: string): object {
  return {
    status,
    timestamp: getCurrentTimestamp(),
  };
}
