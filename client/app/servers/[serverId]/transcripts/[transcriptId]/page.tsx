"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { ChevronLeft, ScrollText, Calendar, ShieldAlert, MessagesSquare, Paperclip, AppWindow } from "lucide-react";

import {
  useTranscript,
  useTranscriptContent,
} from "../../../../../lib/hooks/useTranscripts";
import type { TranscriptMessage } from "../../../../../lib/api";

function formatTimestamp(ts: string | number): string {
  if (!ts) return "—";
  const date = typeof ts === "number" ? new Date(ts * 1000) : new Date(ts);
  return date.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
}

function getAvatarUrl(author: TranscriptMessage["author"]): string {
  if (author.avatar) {
    return `https://cdn.discordapp.com/avatars/${author.id}/${author.avatar}.png`;
  }
  const discrim = parseInt(author.discriminator || "0", 10);
  return `https://cdn.discordapp.com/embed/avatars/${discrim % 5}.png`;
}

function MessageContent({ message }: { message: TranscriptMessage }) {
  if (message.type === "system") {
    return (
      <div className="italic text-zinc-400 text-sm font-semibold flex items-center gap-2">
        <span className="h-1.5 w-1.5 rounded-full bg-zinc-400" />
        {message.content}
      </div>
    );
  }

  if (message.type === "voice_join") {
    return (
      <div className="text-sm text-emerald-400 font-semibold flex items-center gap-2">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 " />
        {message.author.username} joined the voice channel
      </div>
    );
  }

  if (message.type === "voice_leave") {
    return (
      <div className="text-sm text-zinc-400 font-semibold flex items-center gap-2">
        <span className="h-1.5 w-1.5 rounded-full bg-zinc-500" />
        {message.author.username} left the voice channel
      </div>
    );
  }

  return (
    <>
      {message.content && (
        <div className="text-sm leading-relaxed whitespace-pre-wrap break-words text-zinc-100 font-medium">
          {message.content}
        </div>
      )}

      {message.embeds && message.embeds.length > 0 && (
        <div className="mt-2.5 space-y-3">
          {message.embeds.map((embed, idx) => (
            <div
              key={idx}
              className="rounded-lg border border-white/5 border-l-4 bg-zinc-950/40 p-4 text-sm max-w-[520px] w-full"
              style={{
                borderLeftColor: embed.color
                  ? `#${embed.color.toString(16).padStart(6, "0")}`
                  : "#FF5A36",
              }}
            >
              {embed.author && (
                <div className="mb-2 flex items-center gap-1.5 text-[13px] font-medium text-zinc-300">
                  {embed.author.iconUrl && (
                    <img
                      src={embed.author.iconUrl}
                      alt=""
                      className="h-4 w-4 rounded-full"
                    />
                  )}
                  {embed.author.url ? (
                    <a
                      href={embed.author.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-[#FF5A36] hover:underline"
                    >
                      {embed.author.name}
                    </a>
                  ) : (
                    <span>{embed.author.name}</span>
                  )}
                </div>
              )}
              {embed.title && (
                <div className="mb-1.5 font-medium text-zinc-100 tracking-tight">
                  {embed.url ? (
                    <a
                      href={embed.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-[#FF5A36] hover:underline"
                    >
                      {embed.title}
                    </a>
                  ) : (
                    embed.title
                  )}
                </div>
              )}
              {embed.description && (
                <div className="mb-3 whitespace-pre-wrap text-zinc-300 font-medium leading-relaxed">
                  {embed.description}
                </div>
              )}
              {embed.fields && embed.fields.length > 0 && (
                <div
                  className="mb-3 grid gap-3"
                  style={{
                    gridTemplateColumns: embed.fields.some((f) => f.inline)
                      ? "repeat(auto-fill, minmax(180px, 1fr))"
                      : "1fr",
                  }}
                >
                  {embed.fields.map((field, fIdx) => (
                    <div key={fIdx} className="bg-zinc-950/20 p-2.5 rounded-md border border-white/2">
                      <div className="text-[13px] font-medium text-zinc-300">
                        {field.name}
                      </div>
                      <div className="whitespace-pre-wrap text-zinc-400 font-medium mt-1 leading-relaxed text-[13px]">
                        {field.value}
                      </div>
                    </div>
                  ))}
                </div>
              )}
              {embed.image && (
                <img
                  src={embed.image.url}
                  alt=""
                  className="mt-2.5 max-h-72 max-w-full rounded-md object-contain border border-white/5"
                />
              )}
              {embed.footer && (
                <div className="mt-3 flex items-center gap-1.5 text-[13px] text-zinc-400 font-semibold border-t border-white/2 pt-2">
                  {embed.footer.iconUrl && (
                    <img
                      src={embed.footer.iconUrl}
                      alt=""
                      className="h-4 w-4 rounded-full"
                    />
                  )}
                  <span>{embed.footer.text}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {message.attachments && message.attachments.length > 0 && (
        <div className="mt-2.5 space-y-2.5">
          {message.attachments.map((att) => (
            <div key={att.id}>
              {att.contentType?.startsWith("image/") ? (
                <img
                  src={att.url}
                  alt={att.filename}
                  className="max-h-72 max-w-full rounded-md object-contain border border-white/5"
                />
              ) : (
                <a
                  href={att.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 rounded-md border border-white/10 bg-zinc-950/40 px-3.5 py-2 text-[13px] font-medium text-[#FF5A36] hover:bg-zinc-950/60 hover:text-white transition-colors "
                >
                  <Paperclip className="h-4 w-4" aria-hidden="true" />
                  {att.filename}
                </a>
              )}
            </div>
          ))}
        </div>
      )}

      {message.reactions && message.reactions.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1.5">
          {message.reactions.map((reaction, idx) => (
            <span
              key={idx}
              className="inline-flex items-center gap-1 rounded-md border border-white/10 bg-zinc-950/40 px-2.5 py-1 text-[13px] font-medium text-zinc-300"
            >
              <span>{reaction.emoji}</span>
              <span className="text-zinc-400">{reaction.count}</span>
            </span>
          ))}
        </div>
      )}
    </>
  );
}

export default function TranscriptViewerPage() {
  const params = useParams();
  const serverId = params.serverId as string;
  const transcriptId = params.transcriptId as string;

  const {
    transcript,
    isLoading: metaLoading,
    isError: metaError,
  } = useTranscript(serverId, transcriptId);

  const {
    content,
    isLoading: contentLoading,
    isError: contentError,
  } = useTranscriptContent(serverId, transcriptId);

  if (metaLoading) {
    return (
      <div className="space-y-6">
        <div className="transcript-heading">
          <Link
            href={`/servers/${serverId}/transcripts`}
            className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white transition-colors "
          >
            <ChevronLeft className="h-4 w-4" />
            Back
          </Link>
          <h1 className="text-base font-medium text-white tracking-tight">Transcript</h1>
        </div>
        <div className="flex items-center justify-center py-20">
          <div className="flex flex-col items-center gap-3">
            <div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
            <p className="text-[13px] text-zinc-400 font-semibold ">Loading transcript...</p>
          </div>
        </div>
      </div>
    );
  }

  if (metaError || !transcript) {
    return (
      <div className="space-y-6">
        <div className="transcript-heading">
          <Link
            href={`/servers/${serverId}/transcripts`}
            className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white transition-colors "
          >
            <ChevronLeft className="h-4 w-4" />
            Back
          </Link>
          <h1 className="text-base font-medium text-white tracking-tight">Transcript</h1>
        </div>
        <div className="rounded-lg border border-red-500/20 bg-red-950/20 p-6 text-center ">
          <p className="text-sm text-red-400 font-medium">
            Failed to load transcript. It may not exist.
          </p>
        </div>
      </div>
    );
  }

  const ticketLabel = transcript.ticketId || `#${transcript.id}`;

  return (
    <div className="space-y-6">
      <div className="transcript-heading">
        <Link
          href={`/servers/${serverId}/transcripts`}
          className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white transition-colors shrink-0"
        >
          <ChevronLeft className="h-4 w-4" />
          Back
        </Link>
        <h1 className="text-lg font-medium text-white tracking-tight ">
          Ticket {ticketLabel}
        </h1>
      </div>

      {/* Metadata */}
      <section className="rounded-lg border border-white/5 bg-zinc-900/10 p-5 ">
        <h2 className="mb-4 text-[13px] font-medium text-[#FF5A36] ">
          Ticket Information
        </h2>
        <div className="transcript-metadata">
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <Calendar className="h-3 w-3" /> Opened
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px] break-words">
              {formatTimestamp(transcript.openedAt)}
            </div>
          </div>
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <Calendar className="h-3 w-3" /> Closed
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px] break-words">
              {formatTimestamp(transcript.closedAt)}
            </div>
          </div>
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <ShieldAlert className="h-3 w-3" /> Closed By
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px] break-words">
              {transcript.closedBy}
            </div>
          </div>
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <MessagesSquare className="h-3 w-3" /> Messages
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px]">
              {transcript.totalMessages}
            </div>
          </div>
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <Paperclip className="h-3 w-3" /> Attachments
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px]">
              {transcript.totalAttachments}
            </div>
          </div>
          <div className="bg-zinc-950/20 p-3 rounded-md border border-white/2 ">
            <div className="text-[13px] font-medium text-zinc-400 flex items-center gap-1">
              <AppWindow className="h-3 w-3" /> Embeds
            </div>
            <div className="font-medium text-zinc-100 mt-1.5 text-[13px]">
              {transcript.totalEmbeds}
            </div>
          </div>
        </div>
      </section>

      {/* Conversation */}
      <section className="rounded-lg border border-white/5 bg-zinc-900/10 p-6 ">
        <h2 className="mb-4 text-[13px] font-medium text-[#FF5A36] ">
          Conversation Log
        </h2>

        {contentLoading && (
          <div className="flex items-center justify-center py-20">
            <div className="flex flex-col items-center gap-3">
              <div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
              <p className="text-[13px] text-zinc-400 font-semibold ">Loading messages...</p>
            </div>
          </div>
        )}

        {contentError && (
          <div className="rounded-lg border border-red-500/20 bg-red-950/20 p-6 text-center ">
            <p className="text-sm text-red-400 font-medium leading-relaxed">
              Failed to load message content. The stored transcript may be
              unavailable.
            </p>
          </div>
        )}

        {content && (
          <div className="space-y-4">
            {content.messages.map((msg, idx) => (
              <div
                key={`${msg.id}-${idx}`}
                className="flex gap-4 rounded-lg border border-white/5 bg-zinc-950/20 px-4 py-4 transition-colors duration-200 hover:bg-zinc-950/40 hover:border-white/10"
              >
                <img
                  src={getAvatarUrl(msg.author)}
                  alt={msg.author.username}
                  className="h-10 w-10 shrink-0 rounded-full border border-white/10 ring-2 ring-white/5 "
                />
                <div className="min-w-0 flex-1">
                  <div className="mb-2 flex flex-wrap items-center gap-2">
                    <span className="font-medium text-zinc-100 tracking-tight">
                      {msg.author.username}
                    </span>
                    {msg.author.bot && (
                      <span className="rounded-md bg-[#5865F2] px-1.5 py-0.5 text-[9px] font-black uppercase text-white tracking-wider">
                        Bot
                      </span>
                    )}
                    <span className="text-[13px] font-semibold text-zinc-400">
                      {formatTimestamp(msg.timestamp)}
                    </span>
                    {msg.edited && (
                      <span className="text-[13px] font-medium italic text-zinc-400">
                        (edited)
                      </span>
                    )}
                  </div>
                  <MessageContent message={msg} />
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
