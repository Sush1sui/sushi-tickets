import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Finesse Tickets → Sushi Tickets",
  description: "Finesse Tickets has moved. You will be redirected to Sushi Tickets.",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>{children}</body>
    </html>
  );
}
