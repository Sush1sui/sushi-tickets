"use client";

import Link from "next/link";
import useAuth from "../lib/context/auth";
import {
  ShieldCheck,
  FileText,
  Activity,
  ExternalLink,
  LogOut,
  Terminal,
  Layout,
  Utensils,
  Award,
  Flame,
  Zap,
  CheckCircle2,
} from "lucide-react";
import DiscordMockup from "../components/DiscordMockup";
import LoadingScreen from "../components/LoadingScreen";
import Navbar from "../components/Navbar";
import Footer from "../components/Footer";

export default function Home() {
  const { user, authLoading, login, logout } = useAuth();

  if (authLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[#1E1F22] text-zinc-100 font-sans selection:bg-[#FF5A36]/30">
      {/* Main Grid Wrapper */}
      <div className="relative mx-auto max-w-7xl px-6 py-8 md:py-16 flex flex-col min-h-screen justify-between gap-12">
        {/* Top Header */}
        <Navbar />

        {/* Brand New Split Asymmetric Layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 lg:gap-8 items-center my-auto">
          {/* LEFT COLUMN: Bold Typography & Structured Onboarding (55%) */}
          <div className="lg:col-span-7 flex flex-col text-left items-start space-y-6">
            {/* Active Status Pill */}
            <div className="inline-flex items-center gap-2.5 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-4 py-1.5 text-xs font-bold text-emerald-400 backdrop-blur-md">
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-ping"></span>
              Sushi Tickets Online
            </div>

            {/* Big Bold Headline */}
            <h1 className="text-5xl sm:text-6xl xl:text-7xl font-black tracking-tight leading-none text-white uppercase">
              Roll out premium <br />
              support with <br />
              <span className="text-[#FF5A36] text-glow-sushi/25">
                Sushi Tickets.
              </span>
            </h1>

            {/* Core Description */}
            <p className="text-zinc-300 text-base sm:text-lg leading-relaxed max-w-xl font-medium">
              Manage your Discord support tickets with clean, easy-to-use button
              menus and secure, easy-to-configure conversation logs.
            </p>

            {/* Visual Onboarding Steps */}
            <div className="space-y-4 w-full max-w-md pt-3">
              <div className="flex items-start gap-4 group">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-xl bg-zinc-900/60 border border-white/10 text-xs font-black text-zinc-400 group-hover:border-[#FF5A36] group-hover:text-[#FF5A36] transition-all duration-300 shadow-inner">
                  1
                </div>
                <div>
                  <h4 className="text-sm font-bold text-zinc-100">
                    Log In with Discord
                  </h4>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Connect securely using your Discord account to see your
                    servers.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4 group">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-xl bg-zinc-900/60 border border-white/10 text-xs font-black text-zinc-400 group-hover:border-emerald-500 group-hover:text-emerald-400 transition-all duration-300 shadow-inner">
                  2
                </div>
                <div>
                  <h4 className="text-sm font-bold text-zinc-100">
                    Create Ticket Panels
                  </h4>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Design custom buttons or selections and send them to your
                    channels.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-4 group">
                <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-xl bg-zinc-900/60 border border-white/10 text-xs font-black text-zinc-400 group-hover:border-[#FF5A36] group-hover:text-[#FF5A36] transition-all duration-300 shadow-inner">
                  3
                </div>
                <div>
                  <h4 className="text-sm font-bold text-zinc-100">
                    Save Transcripts
                  </h4>
                  <p className="text-xs text-zinc-400 mt-0.5 leading-relaxed font-medium">
                    Choose a log channel, and completed conversations will be
                    saved as web links.
                  </p>
                </div>
              </div>
            </div>

            {/* Action Block */}
            <div className="pt-6 w-full sm:w-auto">
              {user ? (
                <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
                  <div className="flex items-center gap-3 rounded-xl border border-white/5 bg-white/5 px-4 py-3 backdrop-blur-md">
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
                    className="group flex items-center justify-center gap-2 rounded-xl bg-[#FF5A36] px-6 py-3.5 text-sm font-extrabold text-white shadow-lg shadow-orange-950/20 hover:bg-[#FF6B4A] hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 active:scale-95"
                  >
                    <Layout className="h-4 w-4" />
                    Manage Servers
                    <ExternalLink className="h-3.5 w-3.5 opacity-60 group-hover:translate-x-0.5 transition-transform" />
                  </Link>
                </div>
              ) : (
                <button
                  onClick={login}
                  className="group w-full sm:w-auto flex items-center justify-center gap-3 rounded-xl bg-[#FF5A36] px-8 py-4 text-sm font-extrabold text-white shadow-lg shadow-orange-950/30 hover:bg-[#FF6B4A] hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 active:scale-95"
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
          </div>

          {/* RIGHT COLUMN: Reusable Tactile Discord Message Preview (45%) */}
          <div className="lg:col-span-5 flex justify-center w-full">
            <DiscordMockup />
          </div>
        </div>

        {/* Beautiful Vertical Menu (Features Section) */}
        <section className="border-t border-white/5 pt-12">
          <div className="max-w-2xl text-left space-y-2 mb-8">
            <span className="text-xs uppercase tracking-widest font-black text-[#FF5A36] text-glow-sushi">
              Sushi Features
            </span>
            <h2 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-white uppercase">
              Everything You Need for Server Support
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="flex gap-4 items-start p-6 rounded-2xl bg-white/2 border border-white/0 hover:border-white/5 hover:bg-white/5 transition-all duration-300 shadow-sm">
              <Award className="h-7 w-7 text-[#FF5A36] shrink-0 mt-0.5 filter drop-shadow-[0_0_6px_rgba(255,90,54,0.3)]" />
              <div>
                <h3 className="font-bold text-zinc-100 text-base">
                  Clean Conversation Transcripts
                </h3>
                <p className="text-sm text-zinc-300 leading-relaxed mt-2 font-medium">
                  Configure a logging channel to automatically save completed
                  ticket chats as beautiful, secure web links.
                </p>
              </div>
            </div>

            <div className="flex gap-4 items-start p-6 rounded-2xl bg-white/2 border border-white/0 hover:border-white/5 hover:bg-white/5 transition-all duration-300 shadow-sm">
              <ShieldCheck className="h-7 w-7 text-emerald-400 shrink-0 mt-0.5 filter drop-shadow-[0_0_6px_rgba(52,211,153,0.3)]" />
              <div>
                <h3 className="font-bold text-zinc-100 text-base">
                  Role & Staff Permissions
                </h3>
                <p className="text-sm text-zinc-300 leading-relaxed mt-2 font-medium">
                  Control access easily. Assign specific Discord roles or staff
                  members to manage, view, or close active tickets.
                </p>
              </div>
            </div>

            <div className="flex gap-4 items-start p-6 rounded-2xl bg-white/2 border border-white/0 hover:border-white/5 hover:bg-white/5 transition-all duration-300 shadow-sm">
              <CheckCircle2 className="h-7 w-7 text-[#FF5A36] shrink-0 mt-0.5 filter drop-shadow-[0_0_6px_rgba(255,90,54,0.3)]" />
              <div>
                <h3 className="font-bold text-zinc-100 text-base">
                  Fully Customizable Panels
                </h3>
                <p className="text-sm text-zinc-300 leading-relaxed mt-2 font-medium">
                  Create custom buttons or selection forms in your server. Route
                  users directly to the right support staff member.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Simple Footer */}
        <Footer />
      </div>
    </div>
  );
}
