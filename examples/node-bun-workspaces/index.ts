import { greetFromB } from "pkg-b";

process.stdout.write("hello from Bun " + Bun.version + "\n");
process.stdout.write("Hello from bun workspaces\n");
process.stdout.write(greetFromB() + "\n");
