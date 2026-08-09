import type { Metadata } from "next";
import { Inter, Geist_Mono } from "next/font/google";
import "./globals.css";
import AuthProvider from "../lib/providers/auth";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "https://sushitickets.app";

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: "Sushi Tickets - Discord Ticket Management System",
    template: "%s | Sushi Tickets",
  },
  description:
    "Manage your Discord support tickets with clean button menus, automated ticket channels, staff permissions, and secure transcript logs.",
  keywords: [
    "discord ticket bot",
    "discord support bot",
    "ticket management discord",
    "sushi tickets",
    "discord bot transcript",
    "discord panel creator",
  ],
  authors: [{ name: "Sushi Tickets Team" }],
  creator: "Sushi Tickets",
  publisher: "Sushi Tickets",
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: siteUrl,
    title: "Sushi Tickets - Discord Ticket Management System",
    description:
      "Manage your Discord support tickets with clean button menus, automated ticket channels, staff permissions, and secure transcript logs.",
    siteName: "Sushi Tickets",
    images: [
      {
        url: "/assets/sushi_logo_with_bg.png",
        width: 1200,
        height: 630,
        alt: "Sushi Tickets Banner",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Sushi Tickets - Discord Ticket Management System",
    description:
      "Manage your Discord support tickets with clean button menus, automated ticket channels, staff permissions, and secure transcript logs.",
    images: ["/assets/sushi_logo_with_bg.png"],
  },
  icons: {
    icon: "/assets/sushi_logo_without_bg.png",
    shortcut: "/assets/sushi_logo_without_bg.png",
    apple: "/assets/sushi_logo_without_bg.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${inter.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
