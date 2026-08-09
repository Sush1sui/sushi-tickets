"use client";

import React from "react";

export default function Footer() {
  const supportUrl =
    process.env.NEXT_PUBLIC_SUPPORT_DISCORD_URL ||
    "https://discord.gg/CvyTbqtGA3";

  return (
    <footer className="border-t border-zinc-900 pt-6 pb-2 text-xs text-zinc-500 flex flex-col sm:flex-row justify-between items-center gap-4">
      <div className="flex items-center gap-2">
        <img
          src="/assets/sushi_logo_without_bg.png"
          alt="Sushi Tickets Logo"
          className="h-5 w-5 object-contain select-none opacity-80 hover:opacity-100 hover:scale-105 transition-all duration-300"
        />
        <span className="font-black text-zinc-400 uppercase tracking-widest">
          Sushi Tickets
        </span>
      </div>

      {/* Official Support Server Button */}
      <a
        href={supportUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 rounded-full border border-[#5865F2]/30 bg-[#5865F2]/10 px-3.5 py-1.5 font-semibold text-[#5865F2] hover:bg-[#5865F2] hover:text-white hover:shadow-[0_0_12px_rgba(88,101,242,0.3)] transition-all duration-300"
      >
        <svg
          className="h-3.5 w-3.5 fill-current"
          viewBox="0 0 127.14 96.36"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path d="M107.7,8.07A105.15,105.15,0,0,0,77.26,0a77.19,77.19,0,0,0-3.3,6.83A96.67,96.67,0,0,0,53.22,6.83,77.19,77.19,0,0,0,49.88,0,105.15,105.15,0,0,0,19.44,8.07C3.66,31.58-1.86,54.65,1,77.53A105.73,105.73,0,0,0,32,96.36a77.7,77.7,0,0,0,6.63-10.85,68.43,68.43,0,0,1-10.4-5c.87-.64,1.71-1.32,2.51-2a75.46,75.46,0,0,0,92.59,0c.8.71,1.64,1.39,2.51,2a68.43,68.43,0,0,1-10.4,5,77.7,77.7,0,0,0,6.63,10.85,105.73,105.73,0,0,0,32.61-18.83C129.82,48.51,123.63,25.64,107.7,8.07ZM42.45,65.69C36.18,65.69,31,60,31,53S36.18,40.36,42.45,40.36,53.83,46,53.83,53,48.72,65.69,42.45,65.69Zm42.24,0C78.41,65.69,73.24,60,73.24,53S78.41,40.36,84.69,40.36,96.07,46,96.07,53,91,65.69,84.69,65.69Z" />
        </svg>
        <span>Support Server</span>
      </a>

      <div className="text-zinc-500">
        © {new Date().getFullYear()} Sushi Tickets. Roll support out with sushi!
      </div>
    </footer>
  );
}
