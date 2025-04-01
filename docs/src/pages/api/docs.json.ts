import type { APIRoute } from "astro";
import { getCollection } from "astro:content";
// import { randomBytes } from "crypto";

// export const prerender = false;

export const GET: APIRoute = async (ctx) => {
  // const secret =
  //   // @ts-ignore import.meta is throwing type errors
  //   import.meta.env.EXPORT_ENDPOINT_PASSWORD ?? randomBytes(16).toString("hex");

  // if (ctx.request.headers.get("Authorization") !== secret) {
  //   return new Response("Unauthorized", {
  //     status: 401,
  //     headers: {
  //       "Content-Type": "text/plain",
  //     },
  //   });
  // }

  const docsEntries = await getCollection("docs");

  const docsJson = docsEntries.map((doc) => ({
    id: doc.id,
    data: doc.data,
    body: doc.body,
  }));

  return new Response(JSON.stringify(docsJson, null, 2), {
    status: 200,
    headers: {
      "Content-Type": "application/json",
    },
  });
};
