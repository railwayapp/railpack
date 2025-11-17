import abbrev from "abbrev";

console.log("pkg-a loaded");

export function greet() {
  const words = abbrev(["package", "packed"]);
  return `pkg-a says: ${Object.keys(words).join(", ")}`;
}
