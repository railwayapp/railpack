import { greetFromB } from "pkg-b";

console.log("hello from Bun " + Bun.version);
console.log("Hello from bun workspaces");
console.log(greetFromB());
