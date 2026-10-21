import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Next.js SPA",
  description: "Next.js static export example",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
