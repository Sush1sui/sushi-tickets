"use client";

import React from "react";
import Link from "next/link";

export default function Footer() {
  const supportUrl =
    process.env.NEXT_PUBLIC_SUPPORT_DISCORD_URL ||
    "https://discord.gg/CvyTbqtGA3";

  return (
    <footer className="site-footer">
      <div className="site-footer-identity">
        <Link href="/" className="site-footer-brand">
          <img
            src="/assets/sushi_logo_without_bg.png"
            alt=""
            width={24}
            height={24}
          />
          <span>Sushi Tickets</span>
        </Link>
        <p className="site-footer-copyright">© {new Date().getFullYear()} Sushi Tickets</p>
      </div>

      <nav aria-label="Footer" className="site-footer-nav">
        {/* Official Support Server Button */}
        <a
          href={supportUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="site-footer-link"
        >
          Support Server
        </a>

        <Link href="/terms" className="site-footer-link">Terms of Service</Link>
        <Link href="/privacy" className="site-footer-link">Privacy Policy</Link>
      </nav>
    </footer>
  );
}
