"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";

import { useTranscripts } from "../../../../lib/hooks/useTranscripts";
import { ScrollText, ChevronLeft, ChevronRight, ExternalLink } from "lucide-react";
import InlineError from "../../../../components/InlineError";

function formatTimestamp(ts: number): string {
	if (!ts) return "—";
	return new Date(ts * 1000).toLocaleDateString("en-US", {
		month: "short",
		day: "numeric",
		year: "numeric",
		hour: "numeric",
		minute: "2-digit",
		hour12: true,
	});
}

export default function TranscriptsPage() {
	const params = useParams();
	const serverId = params.serverId as string;
	const [page, setPage] = useState(1);
	const limit = 20;

	const { transcripts, pagination, isLoading, isError, refresh } = useTranscripts(serverId, page, limit);

	if (isError) {
		return (
			<InlineError
				message="We couldn't load this server's transcripts. You may no longer have access, or it may be temporarily unavailable."
				onRetry={refresh}
			/>
		);
	}

	return (
		<div className="space-y-5 pb-6">
			{/* Header */}
			<div className="mb-2">
				<h1 className="text-xl font-medium tracking-tight text-white ">Transcripts</h1>
				<p className="text-[13px] text-zinc-300 font-semibold mt-0.5">
					Saved conversation history from closed tickets.
				</p>
			</div>

			<div className="transcript-table">
				{/* Table header */}
				<div className="transcript-row transcript-head" aria-hidden="true">
					<span className="text-[13px] font-medium text-zinc-400">Ticket</span>
					<span className="text-[13px] font-medium text-zinc-400">User ID</span>
					<span className="text-[13px] font-medium text-zinc-400">Messages</span>
					<span className="text-[13px] font-medium text-zinc-400">Closed At</span>
					<span className="text-[13px] font-medium text-zinc-400">Closed By</span>
					<span className="text-[13px] font-medium text-zinc-400 text-center">Action</span>
				</div>

				{/* Rows */}
				{isLoading ? (
					<div className="flex items-center justify-center py-20">
						<div className="flex flex-col items-center gap-3">
							<div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
							<p className="text-[13px] text-zinc-400 font-medium ">Loading transcripts...</p>
						</div>
					</div>
				) : transcripts.length === 0 ? (
					<div className="flex flex-col items-center justify-center gap-3 py-20 text-center">
						<div className="flex h-12 w-12 items-center justify-center rounded-md border border-white/5 bg-zinc-950/40">
							<ScrollText className="h-5 w-5 text-zinc-400" />
						</div>
						<div>
							<p className="text-sm font-medium text-zinc-300">No transcripts yet</p>
							<p className="text-[13px] text-zinc-400 mt-1 font-medium leading-relaxed">
								Transcripts appear here when tickets are closed.
							</p>
						</div>
					</div>
				) : (
					<div className="divide-y divide-white/5">
						{transcripts.map((t) => (
							<div
								key={t.id}
								className="transcript-row"
							>
								<span className="font-mono text-[13px] text-zinc-100 font-medium tracking-tight">
									{t.ticketId || `#${t.id}`}
								</span>
								<span className="font-mono text-[13px] text-zinc-400 font-medium truncate">
									<span className="transcript-mobile-label">User </span>{t.userId || "Unknown"}
								</span>
								<span className="text-[13px] text-zinc-200 font-medium pl-3">{t.totalMessages}<span className="transcript-mobile-label"> messages</span></span>
								<span className="text-[13px] text-zinc-300 font-medium">
									{formatTimestamp(t.closedAt)}
								</span>
								<span className="text-[13px] text-zinc-300 font-semibold truncate"><span className="transcript-mobile-label">Closed by </span>{t.closedBy}</span>
								<Link
									href={`/servers/${serverId}/transcripts/${t.id}`}
									className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white hover:bg-white/10 hover:border-white/20 transition-colors shrink-0"
								>
									<ExternalLink className="h-3.5 w-3.5" />
									View
								</Link>
							</div>
						))}
					</div>
				)}

				{/* Footer: count + pagination */}
				{!isLoading && (
					<div className="flex items-center justify-between px-6 py-4.5 border-t border-white/5 bg-white/2">
						<span className="text-[13px] text-zinc-400 font-medium">
							{pagination.total > 0
								? `${pagination.total} transcript${pagination.total > 1 ? "s" : ""}`
								: "No transcripts"}
						</span>

						{pagination.pages > 1 && (
							<div className="flex items-center gap-3.5">
								<button
									onClick={() => setPage((p) => Math.max(1, p - 1))}
									disabled={page === 1}
									className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white hover:bg-white/10 hover:border-white/20 transition-colors disabled:opacity-30 disabled:cursor-not-allowed "
								>
									<ChevronLeft className="h-4 w-4" />
									Prev
								</button>
								<span className="text-[13px] text-zinc-300 font-medium font-mono">
									{page} / {pagination.pages}
								</span>
								<button
									onClick={() => setPage((p) => Math.min(pagination.pages, p + 1))}
									disabled={page === pagination.pages}
									className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 hover:text-white hover:bg-white/10 hover:border-white/20 transition-colors disabled:opacity-30 disabled:cursor-not-allowed "
								>
									Next
									<ChevronRight className="h-4 w-4" />
								</button>
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
}
