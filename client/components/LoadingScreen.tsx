"use client";

import React from "react";

interface LoadingScreenProps {
  text?: string;
}

export default function LoadingScreen({
  text = "Loading...",
}: LoadingScreenProps) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-[#1E1F22] text-zinc-100 font-sans">
      <div className="relative flex h-16 w-16 items-center justify-center">
        <div className="absolute h-full w-full rounded-full bg-[#FF5A36]/20"></div>
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#FF5A36] border-t-transparent"></div>
      </div>
      <p className="mt-4 text-[13px] font-semibold text-[#FF5A36]/80 ">
        {text}
      </p>
    </div>
  );
}
