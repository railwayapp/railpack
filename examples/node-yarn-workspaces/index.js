import cowsay from "cowsay";
import { greet as greetA } from "pkg-a";
import { greet as greetB } from "pkg-b";

console.log(cowsay.say({
  text: "Hello from yarn workspaces",
}));

console.log(greetA());
console.log(greetB());
