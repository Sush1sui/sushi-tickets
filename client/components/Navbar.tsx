"use client";

import React from "react";
import useAuth from "../lib/context/auth";
import { LogOut } from "lucide-react";
import Link from "next/link";

export default function Navbar() {
  const { user, logout } = useAuth();

  return (
    <header className="flex items-center justify-between border-b border-zinc-800/40 pb-5">
      <Link href="/">
        <div className="flex items-center gap-2.5">
          <img
            src="/assets/sushi_logo_without_bg.png"
            alt="Sushi Logo"
            className="h-9 w-9 object-contain filter drop-shadow-[0_0_8px_rgba(255,90,54,0.45)] select-none hover:rotate-6 transition-all duration-300"
          />
          <div className="flex flex-col">
            <span className={"text-lg font-black tracking-tight bg-gradient-to-r from-white via-zinc-200 to-zinc-400 bg-clip-text text-transparent uppercase font-mono" /* taste-skill-ignore: Existing brand wordmark treatment is protected. */}>
              Sushi Tickets
            </span>
            <span className="text-[10px] tracking-widest uppercase font-semibold text-[#FF5A36] -mt-1">
              Premium Support
            </span>
          </div>
        </div>
      </Link>
      <div>
        {user && (
          <div className="flex items-center gap-3">
            {/* Glassmorphic User Profile Capsule matching website theme */}
            <div className="flex items-center gap-2.5 px-2 py-2">
              {user.image ? (
                <img
                  src={user.image}
                  alt={user.name}
                  className="h-5.5 w-5.5 rounded-full shrink-0 ring-2 ring-white/10 group-hover:ring-[#FF5A36]/50 transition-all duration-300 group-hover:scale-105"
                />
              ) : (
                <div className="h-5.5 w-5.5 rounded-full shrink-0 bg-zinc-800 ring-2 ring-white/10 group-hover:ring-[#FF5A36]/50 flex items-center justify-center text-[10px] font-black text-zinc-300 select-none transition-all duration-300">
                  {user.name[0]?.toUpperCase()}
                </div>
              )}
              <span className="hidden sm:inline text-sm text-zinc-200">
                {user.name}
              </span>
            </div>

            {/* Subtle divider */}
            <div className="h-5 w-[1px] bg-white/5" />

            {/* Premium brand-colored Logout Button matching homepage buttons */}
            <button
              onClick={logout}
              className="flex min-h-11 items-center gap-2 px-2 py-2 text-sm text-zinc-300 hover:text-white transition-colors"
            >
              <LogOut className="h-3.5 w-3.5 text-zinc-500 group-hover:text-[#FF5A36] group-hover:translate-x-0.5 transition-all" />
              <span>Logout</span>
            </button>
          </div>
        )}
      </div>
    </header>
  );
}
