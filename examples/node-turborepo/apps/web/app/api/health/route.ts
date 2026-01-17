import { APP_NAME, API_VERSION, formatHealthCheck } from "@repo/utils";
import { NextResponse } from "next/server";

export async function GET() {
  return NextResponse.json({
    app: APP_NAME,
    version: API_VERSION,
    ...formatHealthCheck("ok"),
  });
}
