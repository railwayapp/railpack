import { greetFromA } from "pkg-a";

export function greetFromB(): string {
  return `pkg-b depends on: ${greetFromA()}`;
}
