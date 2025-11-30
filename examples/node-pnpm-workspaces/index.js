import { greet as greetA } from "pkg-a";
import { greet as greetB } from "pkg-b";

console.log("Node v:", process.version);
console.log("Hello from pnpm workspaces");

console.log(greetA());
console.log(greetB());
