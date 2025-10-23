import * as o from "https://deno.land/x/cowsay/mod.ts";

let m = o.say({
  text: `Hello from Deno (deno ${Deno.version.deno})`,
});

console.log(m);
