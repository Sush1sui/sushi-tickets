"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import useAuth from "../../lib/context/auth";
import { useServers } from "../../lib/hooks/useServers";
import LoadingScreen from "../../components/LoadingScreen";
import Navbar from "../../components/Navbar";
import Footer from "../../components/Footer";
import { ServerSummary } from "../../lib/api";
import { ChevronRight, ServerCrash, Settings, UserPlus } from "lucide-react";

function ServerCard({ server }: { server: ServerSummary }) {
  const router = useRouter();

  return (
    <button
      onClick={() => router.push(`/servers/${server.id}`)}
      className="group relative flex items-center gap-4 rounded-2xl border border-white/5 bg-zinc-900/10 p-5 text-left backdrop-blur-md hover:border-[#FF5A36]/40 hover:bg-white/5 transition-all duration-300 hover:shadow-xl hover:shadow-[#FF5A36]/2 active:scale-[0.99] w-full shadow-md"
    >
      {/* Server Icon */}
      {server.iconUrl ? (
        <img
          src={server.iconUrl}
          alt={server.name}
          className="h-12 w-12 rounded-full shrink-0 ring-2 ring-white/10 group-hover:ring-[#FF5A36]/50 transition-all"
        />
      ) : (
        <div className="h-12 w-12 rounded-full shrink-0 bg-zinc-950/60 border border-white/10 ring-2 ring-white/5 group-hover:ring-[#FF5A36]/50 flex items-center justify-center text-lg font-black text-zinc-200 transition-all select-none">
          {server.name[0]?.toUpperCase()}
        </div>
      )}

      {/* Server Info */}
      <div className="flex-1 min-w-0">
        <p className="font-extrabold text-zinc-100 text-sm truncate group-hover:text-white transition-colors tracking-tight">
          {server.name}
        </p>
      </div>

      {/* Arrow Icon */}
      <div className="flex items-center gap-2 shrink-0">
        <Settings className="h-4 w-4 text-zinc-400 group-hover:text-[#FF5A36] group-hover:rotate-45 transition-all duration-500" />
        <ChevronRight className="h-4 w-4 text-zinc-500 group-hover:text-zinc-300 group-hover:translate-x-0.5 transition-all" />
      </div>

      {/* Hover glow edge */}
      <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none ring-1 ring-inset ring-[#FF5A36]/20" />
    </button>
  );
}

export default function ServersPage() {
  const { user, authLoading } = useAuth();
  const { servers, isLoading } = useServers();

  const clientId = process.env.NEXT_PUBLIC_DISCORD_CLIENT_ID;
  const inviteUrl = clientId
    ? `https://discord.com/oauth2/authorize?client_id=${clientId}&permissions=8&scope=bot%20applications.commands`
    : "https://discord.com/oauth2/authorize";

  if (authLoading || isLoading) {
    return <LoadingScreen />;
  }

  if (!user) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-background text-zinc-100 font-sans">
        <img
          src="/assets/sushi_logo_without_bg.png"
          alt="Sushi Logo"
          className="h-14 w-14 object-contain filter drop-shadow-[0_0_12px_rgba(255,90,54,0.5)] select-none mb-2 hover:scale-105 transition-all duration-300"
        />
        <p className="text-zinc-300 text-sm font-semibold">You need to log in first.</p>
        <Link
          href="/"
          className="rounded-xl bg-[#FF5A36] px-5 py-2.5 text-sm font-bold text-white hover:bg-[#FF6B4A] shadow-md shadow-orange-950/15 transition-all duration-200 active:scale-95"
        >
          Go to Home
        </Link>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[#1E1F22] text-zinc-100 font-sans selection:bg-[#FF5A36]/30">
      <div className="relative mx-auto max-w-5xl px-6 py-8 md:py-16 flex flex-col min-h-screen justify-between gap-12">
        {/* Header */}
        <Navbar />

        {/* Page Content */}
        <main className="flex-1">
          {/* Page Title */}
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-10">
            <div>
              <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-white leading-tight uppercase text-glow-sushi/10">
                Your Servers
              </h1>
              <p className="text-zinc-400 text-sm mt-2 font-medium">
                Select a server to manage its ticket settings.
              </p>
            </div>
            {servers.length > 0 && (
              <a
                href={inviteUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="self-start sm:self-center flex items-center gap-2 rounded-xl bg-[#FF5A36] px-4.5 py-2.5 text-xs font-bold text-white hover:bg-[#FF6B4A] shadow-md shadow-orange-950/15 hover:shadow-[#FF5A36]/20 transition-all duration-200 active:scale-95"
              >
                <UserPlus className="h-4 w-4" />
                Invite Bot
              </a>
            )}
          </div>

          {/* Server Grid */}
          {servers.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-4 py-24 text-center">
              <div className="flex h-16 w-16 items-center justify-center rounded-2xl border border-white/5 bg-zinc-900/10 shadow-inner">
                <ServerCrash className="h-7 w-7 text-zinc-500" />
              </div>
              <div>
                <p className="text-zinc-200 font-bold text-sm">
                  No servers found
                </p>
                <p className="text-zinc-400 text-xs mt-1.5 leading-relaxed font-medium">
                  Make sure Sushi Tickets bot is added to at least one server.
                </p>
              </div>
              <a
                href={inviteUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-xs font-bold text-zinc-300 hover:border-white/20 hover:text-white transition-all shadow-md"
              >
                Add Bot to Server
              </a>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {servers.map((server) => (
                <ServerCard key={server.id} server={server} />
              ))}
            </div>
          )}
        </main>

        {/* Footer */}
        <Footer />
      </div>
    </div>
  );
}
