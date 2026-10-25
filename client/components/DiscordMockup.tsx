"use client";

import Image from "next/image";
import React from "react";

export interface DiscordField {
  label: string;
  value: string;
}

export interface DiscordButton {
  label: string;
  emoji?: string;
  style?: "primary" | "secondary" | "success" | "danger";
}

interface DiscordMockupProps {
  authorName?: string;
  avatarUrl?: string;
  avatarEmoji?: string;
  avatarBg?: string;
  embedTitle?: string;
  embedDescription?: string;
  embedColor?: string;
  fields?: DiscordField[];
  buttons?: DiscordButton[];
  largeImageUrl?: string;
  smallImageUrl?: string;
}

export default function DiscordMockup({
  authorName = "Sushi Tickets",
  avatarUrl = "/assets/sushi_logo_with_bg.png",
  avatarEmoji = "🍣", // taste-skill-ignore: Real Discord message content.
  avatarBg = "#FF5A36",
  embedTitle = "Sushi Support Menu",
  embedDescription = "Need assistance? Select your plate below to open an instant secure support channel. A staff member will serve you shortly.",
  embedColor = "#FF5A36",
  fields = [
    { label: "🍣 Save base transcript", value: "Enabled" }, // taste-skill-ignore: Real Discord message content.
    { label: "🌶️ Wasabi cache speed", value: "Active" }, // taste-skill-ignore: Real Discord message content.
  ],
  buttons = [
    { label: "Open Support", emoji: "🍣", style: "primary" }, // taste-skill-ignore: Real Discord message content.
    { label: "Transcripts", emoji: "📁", style: "secondary" }, // taste-skill-ignore: Real Discord message content.
    { label: "Custom", emoji: "🌶️", style: "secondary" }, // taste-skill-ignore: Real Discord message content.
  ],
  largeImageUrl = "",
  smallImageUrl = "/assets/sushi_logo_with_bg.png",
}: DiscordMockupProps) {
  const getButtonClass = (
    style?: "primary" | "secondary" | "success" | "danger",
  ) => {
    switch (style) {
      case "primary":
        return "bg-[#5865F2] hover:bg-[#4752C4]";
      case "success":
        return "bg-[#248046] hover:bg-[#1A6535]";
      case "danger":
        return "bg-[#DA373C] hover:bg-[#A92B2F]";
      case "secondary":
      default:
        return "bg-[#4E5058] hover:bg-[#6D6F78]";
    }
  };

  return (
    <div
      data-ui
      className="w-full max-w-[480px] rounded-[8px] bg-[#313338] p-4 text-left shadow-2xl select-none font-sans"
    >
      <div className="flex gap-4">
        {/* Bot Avatar */}
        {avatarUrl ? (
          <Image
            src={avatarUrl}
            alt="Bot Avatar"
            width={40}
            height={40}
            className="h-10 w-10 shrink-0 rounded-full object-cover shadow-md select-none border border-[#FF5A36]/15"
          />
        ) : (
          <div
            className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-white font-bold shadow-md select-none text-base"
            style={{ backgroundColor: avatarBg }}
          >
            {avatarEmoji}
          </div>
        )}

        {/* Message Details */}
        <div className="flex-1 min-w-0">
          {/* Message Header */}
          <div className="flex items-center gap-1.5 flex-wrap mb-[4px]">
            <span className="font-semibold text-[#F2F3F5] text-[15px] leading-4 hover:underline cursor-pointer">
              {authorName}
            </span>
            <span className="inline-flex items-center justify-center bg-[#5865F2] text-[9px] text-white px-1 py-0.5 rounded-[3px] font-extrabold tracking-wide uppercase select-none leading-none">
              APP
            </span>
            <span className="text-[12px] text-[#949BA4] leading-none select-none">
              Today at{" "}
              {new Date().toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
              })}
            </span>
          </div>

          {/* Simulated Discord Embed */}
          <div className="max-w-[420px] rounded-[4px] bg-[#2B2D31] flex overflow-hidden">
            {/* Color Stripe */}
            <div
              className="w-[4px] shrink-0"
              style={{ backgroundColor: embedColor }}
            />

            {/* Embed Content */}
            <div className="flex-1 p-3 text-left space-y-2">
              <div className="flex justify-between items-start gap-4">
                <div className="flex-1 space-y-2">
                  {embedTitle && (
                    <div className="flex items-center gap-1.5">
                      <h3 className="font-semibold text-white text-[15px] leading-tight">
                        {embedTitle}
                      </h3>
                    </div>
                  )}
                  {embedDescription && (
                    <p className="text-[13px] text-[#DBDEE1] leading-relaxed whitespace-pre-wrap">
                      {embedDescription}
                    </p>
                  )}
                </div>
                {smallImageUrl && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={smallImageUrl}
                    alt=""
                    className="h-20 w-20 rounded-[4px] shrink-0 object-cover border border-white/5 bg-zinc-950/20"
                    onError={(e) => (e.currentTarget.style.display = "none")}
                  />
                )}
              </div>

              {/* Embed Fields */}
              {fields.length > 0 && (
                <div className="grid grid-cols-2 gap-y-2 gap-x-4 pt-1 text-[13px]">
                  {fields.map((field, idx) => (
                    <div key={idx}>
                      <div className="font-semibold text-[#949BA4] text-[12px] leading-[16px] mb-[2px] select-none">
                        {field.label}
                      </div>
                      <div className="text-[#DBDEE1] font-normal leading-[18px]">
                        {field.value}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {/* Embed Large Image */}
              {largeImageUrl && (
                <div className="mt-3 rounded-[4px] overflow-hidden max-h-[260px] w-full border border-white/5 bg-zinc-950/20 flex items-center justify-center">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={largeImageUrl}
                    alt=""
                    className="max-h-[260px] w-full object-cover"
                    onError={(e) => {
                      if (e.currentTarget.parentElement) {
                        e.currentTarget.parentElement.style.display = "none";
                      }
                    }}
                  />
                </div>
              )}
            </div>
          </div>

          {/* Simulated Discord Interaction Buttons */}
          {buttons.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-[4px]">
              {buttons.map((btn, idx) => (
                <button
                  key={idx}
                  className={`flex items-center gap-[3px] rounded-[9px] px-4 h-8 text-[14px] font-medium text-white transition-colors active:scale-98 select-none ${getButtonClass(btn.style)}`}
                >
                  {btn.emoji &&
                    (btn.emoji.startsWith("http") ? (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={btn.emoji}
                        alt=""
                        className="h-[16px] w-[16px] object-contain shrink-0 select-none mr-0.5"
                      />
                    ) : (
                      <span className="text-[15px] leading-none">
                        {btn.emoji}
                      </span>
                    ))}
                  {btn.label}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
