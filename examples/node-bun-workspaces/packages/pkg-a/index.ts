import abbrev from "abbrev";

export function greetFromA(): string {
  const result = abbrev(["test", "testing"]);
  return `pkg-a (abbrev keys: ${Object.keys(result).join(", ")})`;
}
