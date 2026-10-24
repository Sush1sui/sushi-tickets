import Link from "next/link";
import type { ReactNode } from "react";
import Navbar from "./Navbar";
import Footer from "./Footer";
import "./legal.css";

type LegalSection = { id: string; title: string; content: ReactNode };

export default function LegalPage({ title, introduction, sections }: {
  title: string;
  introduction: string;
  sections: LegalSection[];
}) {
  const supportUrl = process.env.NEXT_PUBLIC_SUPPORT_DISCORD_URL || "https://discord.gg/CvyTbqtGA3";

  return (
    <div className="site-page legal-page">
      <div className="site-frame">
        <a className="legal-skip" href="#policy">Skip to policy</a>
        <Navbar />
        <main id="policy" className="legal-main">
          <div className="legal-heading">
            <h1>{title}</h1>
            <p>{introduction}</p>
            <p className="legal-date">Effective October 24, 2026 · Last updated October 24, 2026</p>
            <nav className="legal-documents" aria-label="Legal documents">
              <Link href="/terms" aria-current={title === "Terms of Service" ? "page" : undefined}>Terms of Service</Link>
              <Link href="/privacy" aria-current={title === "Privacy Policy" ? "page" : undefined}>Privacy Policy</Link>
            </nav>
          </div>
          <div className="legal-layout">
            <aside>
              <nav className="legal-contents" aria-label="On this page">
                <h2>On this page</h2>
                {sections.map(section => <a key={section.id} href={`#${section.id}`}>{section.title}</a>)}
                <a href="#contact">Contact</a>
              </nav>
            </aside>
            <article className="legal-copy" aria-label={title}>
              {sections.map(section => (
                <section key={section.id} id={section.id}>
                  <h2>{section.title}</h2>
                  {section.content}
                </section>
              ))}
              <section id="contact">
                <h2>Contact</h2>
                <p>Sushi Tickets is operated by the Sushi Tickets developer team, maintained by Sush1sui. For questions, privacy requests, or reports of misuse, contact the developer team through our <a href={supportUrl} target="_blank" rel="noopener noreferrer">support server</a>. Ask to speak privately with a team member about personal data. Never post access tokens, passwords, or private ticket contents in a public channel.</p>
                <p>If you cannot use Discord, you can reach the maintainer through the contact options on <a href="https://github.com/Sush1sui" target="_blank" rel="noopener noreferrer">Sush1sui&apos;s GitHub profile</a>. Include only enough information to arrange a private conversation.</p>
              </section>
            </article>
          </div>
        </main>
        <Footer />
      </div>
    </div>
  );
}
