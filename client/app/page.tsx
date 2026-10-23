"use client";

import Link from "next/link";
import useAuth from "../lib/context/auth";
import { ExternalLink, Layout } from "lucide-react";
import DiscordMockup from "../components/DiscordMockup";
import LoadingScreen from "../components/LoadingScreen";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";

export default function Home() {
  const { user, authLoading, login } = useAuth();

  if (authLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="site-page">
      {/* Main Grid Wrapper */}
      <div className="site-frame">
        {/* Top Header */}
        <Navbar />

        <main>
        {/* Brand New Split Asymmetric Layout */}
        <div className="home-hero">
          {/* LEFT COLUMN: Bold Typography & Structured Onboarding (55%) */}
          <div className="home-intro">
            {/* Active Status Pill */}


            {/* Big Bold Headline */}
            <h1>Support, served<br />with Sushi Tickets.</h1>

            {/* Core Description */}
            <p className="home-description">
              Manage your Discord support tickets with clean, easy-to-use button
              menus and secure, easy-to-configure conversation logs.
            </p>

            {/* Action Block */}
            <div className="home-action">
              {user ? (
                <div className="flex flex-wrap items-center gap-4">
                  <div className="home-user">
                    {user.image ? (
                      <img
                        src={user.image}
                        alt={user.name}
                        className="h-6 w-6 rounded-full border border-white/10 ring-1 ring-white/5"
                      />
                    ) : (
                      <div className="h-6 w-6 rounded-full bg-rose-500/20 flex items-center justify-center text-xs font-bold text-rose-400 border border-rose-500/10">
                        {user.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <span className="text-sm font-bold text-zinc-200">
                      Welcome, {user.name}
                    </span>
                  </div>
                  <Link
                    href="/servers"
                    className="home-primary-action"
                  >
                    <Layout className="h-4 w-4" />
                    Manage Servers
                    <ExternalLink className="h-3.5 w-3.5 opacity-60 group-hover:translate-x-0.5 transition-transform" />
                  </Link>
                </div>
              ) : (
                <button
                  onClick={login}
                  className="home-primary-action"
                >
                  <svg
                    className="h-5 w-5 fill-current"
                    viewBox="0 0 127.14 96.36"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path d="M107.7,8.07A105.15,105.15,0,0,0,77.26,0a77.19,77.19,0,0,0-3.3,6.83A96.67,96.67,0,0,0,53.22,6.83,77.19,77.19,0,0,0,49.88,0,105.15,105.15,0,0,0,19.44,8.07C3.66,31.58-1.86,54.65,1,77.53A105.73,105.73,0,0,0,32,96.36a77.7,77.7,0,0,0,6.63-10.85,68.43,68.43,0,0,1-10.4-5c.87-.64,1.71-1.32,2.51-2a75.46,75.46,0,0,0,92.59,0c.8.71,1.64,1.39,2.51,2a68.43,68.43,0,0,1-10.4,5,77.7,77.7,0,0,0,6.63,10.85,105.73,105.73,0,0,0,32.61-18.83C129.82,48.51,123.63,25.64,107.7,8.07ZM42.45,65.69C36.18,65.69,31,60,31,53S36.18,40.36,42.45,40.36,53.83,46,53.83,53,48.72,65.69,42.45,65.69Zm42.24,0C78.41,65.69,73.24,60,73.24,53S78.41,40.36,84.69,40.36,96.07,46,96.07,53,91,65.69,84.69,65.69Z" />
                  </svg>
                  Connect Discord Server
                </button>
              )}
            </div>

            {/* Visual Onboarding Steps */}
            <div className="home-steps">
              <div className="flex items-start gap-4 group">
                <div className="home-step-number">
                  1
                </div>
                <div>
                  <h3 className="text-sm font-bold text-zinc-100">
                    Log In with Discord
                  </h3>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Connect securely using your Discord account to see your
                    servers.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4 group">
                <div className="home-step-number">
                  2
                </div>
                <div>
                  <h3 className="text-sm font-bold text-zinc-100">
                    Create Ticket Panels
                  </h3>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Design custom buttons or selections and send them to your
                    channels.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4 group">
                <div className="home-step-number">
                  3
                </div>
                <div>
                  <h3 className="text-sm font-bold text-zinc-100">
                    Save Transcripts
                  </h3>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Choose a log channel, and completed conversations will be
                    saved as web links.
                  </p>
                </div>
              </div>
            </div>


          </div>

          {/* RIGHT COLUMN: Reusable Tactile Discord Message Preview (45%) */}
          <div className="home-preview">
            <div className="home-preview-content"><DiscordMockup /></div>
          </div>
        </div>

        <section className="home-features" aria-labelledby="home-features-title">
          <div className="home-feature-heading">
            <h2 className="home-section-title" id="home-features-title">
              Built for server support
            </h2>
            <p className="home-feature-lead">
              One ticket carries the panel people click, the staff who can
              answer, and the conversation you keep.
            </p>
          </div>

          <div className="home-feature-layout">
            <article className="home-feature home-feature--transcripts">
              <div className="home-feature-copy">
                <h3>Clean Conversation Transcripts</h3>
                <p>
                  Configure a logging channel to automatically save completed
                  ticket chats as beautiful, secure web links.
                </p>
              </div>

              <div className="home-transcript" data-ui>
                <div className="home-transcript-head">
                  <span className="home-transcript-title">#ticket-0142</span>
                  <span className="home-transcript-state">
                    Saved to #transcripts
                  </span>
                </div>

                <div className="home-transcript-thread">
                  <div className="home-transcript-message">
                    <span className="home-transcript-avatar">AR</span>
                    <div>
                      <p className="home-transcript-meta">
                        <span>Ari</span>
                        <span>10:42</span>
                      </p>
                      <p className="home-transcript-text">
                        I was charged twice for the same order.
                      </p>
                    </div>
                  </div>

                  <div className="home-transcript-message">
                    <span className="home-transcript-avatar">MI</span>
                    <div>
                      <p className="home-transcript-meta">
                        <span>Mika</span>
                        <span>Staff</span>
                        <span>10:44</span>
                      </p>
                      <p className="home-transcript-text">
                        Refund issued. It should land in 3-5 days.
                      </p>
                    </div>
                  </div>

                  <div className="home-transcript-message">
                    <span className="home-transcript-avatar">AR</span>
                    <div>
                      <p className="home-transcript-meta">
                        <span>Ari</span>
                        <span>10:46</span>
                      </p>
                      <p className="home-transcript-text">
                        Got it, thanks for the quick help.
                      </p>
                    </div>
                  </div>
                </div>

                <div className="home-transcript-foot">
                  <span>12 messages</span>
                  <span>3 attachments</span>
                  <span>Closed by Mika</span>
                </div>
              </div>
            </article>

            <article className="home-feature home-feature--permissions">
              <div className="home-feature-copy">
                <h3>Role & Staff Permissions</h3>
                <p>
                  Choose which Discord roles or staff members can manage tickets
                  in this server.
                </p>
              </div>

              <div className="home-access" data-ui>
                <div className="home-access-column">
                  <div className="home-access-head">
                    <h4>Authorized Members</h4>
                    <span>1</span>
                  </div>
                  <p className="home-access-description">
                    Individual users who can manage tickets.
                  </p>
                  <div className="home-access-search">Search and add members...</div>
                  <div className="home-access-chip">
                    <span className="home-access-avatar">MI</span>
                    <span className="home-access-name">Mika</span>
                    <span className="home-access-handle">mika</span>
                    <span className="home-access-remove" aria-hidden="true">×</span>
                  </div>
                </div>

                <div className="home-access-column">
                  <div className="home-access-head">
                    <h4>Authorized Roles</h4>
                    <span>2</span>
                  </div>
                  <p className="home-access-description">
                    Anyone with these roles can manage tickets.
                  </p>
                  <div className="home-access-search">Search and add roles...</div>
                  <div className="home-access-roles">
                    <div className="home-access-chip">
                      <span className="home-access-dot home-access-dot--support" />
                      <span className="home-access-name">Support</span>
                      <span className="home-access-remove" aria-hidden="true">×</span>
                    </div>
                    <div className="home-access-chip">
                      <span className="home-access-dot home-access-dot--billing" />
                      <span className="home-access-name">Billing</span>
                      <span className="home-access-remove" aria-hidden="true">×</span>
                    </div>
                  </div>
                </div>
              </div>
            </article>

            <article className="home-feature home-feature--panels">
              <div className="home-feature-copy">
                <h3>Fully Customizable Panels</h3>
                <p>
                  Create custom buttons or selection forms in your server. Route
                  users directly to the right support staff member.
                </p>
              </div>

              <div className="home-panel-preview">
                <DiscordMockup
                  authorName="Sushi Tickets"
                  embedTitle="How can we help?"
                  embedDescription="Choose a counter and we'll open a private ticket for you."
                  embedColor="#FF5A36"
                  fields={[]}
                  buttons={[
                    { label: "Billing", style: "primary" },
                    { label: "Technical", style: "secondary" },
                    { label: "Report a user", style: "secondary" },
                  ]}
                />
              </div>
            </article>
          </div>
        </section>

        </main>

        {/* Simple Footer */}
        <Footer />
      </div>
    </div>
  );
}
