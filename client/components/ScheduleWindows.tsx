"use client";

import { useEffect, useMemo, useState } from "react";
import type { ScheduleWindowDraft } from "../lib/api";
import { SectionCard, FormLabel, DarkInput } from "./DarkFormFields";
import {
	Ban,
	Clock,
	Pause,
	Pencil,
	Play,
	Plus,
	RotateCcw,
	Trash2,
} from "lucide-react";

type WindowStatus = "upcoming" | "open" | "closed_early" | "past";
type DraftValues = { start: string; end: string };
type WindowGroup = "active" | "upcoming" | "past";

const MAX_WINDOWS = 50;
const DURATION_PRESETS = [1, 2, 4, 8] as const;

const toLocalInputValue = (timestamp: number) => {
	const date = new Date(timestamp * 1000);
	const pad = (n: number) => String(n).padStart(2, "0");
	return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const fromLocalInputValue = (value: string) => {
	const ms = new Date(value).getTime();
	return Number.isFinite(ms) ? Math.floor(ms / 1000) : 0;
};

const formatWindowRange = (startTime: number, endTime: number) => {
	const nowYear = new Date().getFullYear();
	const start = new Date(startTime * 1000);
	const end = new Date(endTime * 1000);
	const opts: Intl.DateTimeFormatOptions = {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	};
	if (start.getFullYear() !== nowYear || end.getFullYear() !== nowYear) {
		opts.year = "numeric";
	}
	return `${start.toLocaleString(undefined, opts)} → ${end.toLocaleString(undefined, opts)}`;
};

const formatMoment = (timestamp: number) =>
	new Date(timestamp * 1000).toLocaleString(undefined, {
		weekday: "short",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	});

const formatDuration = (seconds: number) => {
	const totalMinutes = Math.max(1, Math.round(seconds / 60));
	if (totalMinutes < 60) return `${totalMinutes} min`;
	const hours = Math.floor(totalMinutes / 60);
	const minutes = totalMinutes % 60;
	return minutes ? `${hours} h ${minutes} min` : `${hours} h`;
};

const relativeTime = (target: number, now: number) => {
	const minutes = Math.max(1, Math.round(Math.abs(target - now) / 60));
	if (minutes < 60) return `${minutes} min`;
	const hours = Math.round(minutes / 60);
	if (hours < 48) return `${hours} h`;
	return `${Math.round(hours / 24)} days`;
};

const humanPreview = (startUnix: number, endUnix: number, now: number) => {
	if (!startUnix || !endUnix || endUnix <= startUnix) return null;
	const startStr = new Date(startUnix * 1000).toLocaleString(undefined, {
		weekday: "short",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	});
	const endStr = new Date(endUnix * 1000).toLocaleString(undefined, {
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	});
	let relative: string;
	if (startUnix > now) {
		relative = `opens in ${relativeTime(startUnix, now)}`;
	} else if (endUnix > now) {
		relative = `open now, closes in ${relativeTime(endUnix, now)}`;
	} else {
		relative = "already ended";
	}
	return `${startStr} → ${endStr} · ${formatDuration(endUnix - startUnix)} · ${relative}`;
};

const relativeWindowLabel = (
	startUnix: number,
	endUnix: number,
	now: number,
) => {
	if (startUnix > now) return `Opens in ${relativeTime(startUnix, now)}`;
	if (endUnix > now) return `Closes in ${relativeTime(endUnix, now)}`;
	return `Ended ${relativeTime(endUnix, now)} ago`;
};

const gmtLabel = () => {
	const offMin = -new Date().getTimezoneOffset();
	const sign = offMin >= 0 ? "+" : "-";
	const abs = Math.abs(offMin);
	const h = Math.floor(abs / 60);
	const m = abs % 60;
	return `GMT${sign}${h}${m ? `:${String(m).padStart(2, "0")}` : ""}`;
};

const validateDraftTimes = (
	startUnix: number,
	endUnix: number,
	others: ScheduleWindowDraft[],
): string | null => {
	if (!startUnix || !endUnix) return "Pick both an open and a close time.";
	if (endUnix <= startUnix) return "Close time must be after open time.";
	const now = Math.floor(Date.now() / 1000);
	if (endUnix <= now) return "Close time must be in the future. Remove the window instead.";
	for (const other of others) {
		if (startUnix < other.endTime && other.startTime < endUnix) {
			return `Overlaps another window (${formatWindowRange(other.startTime, other.endTime)}).`;
		}
	}
	return null;
};

const getWindowStatus = (
	win: ScheduleWindowDraft,
	nowSeconds: number,
): WindowStatus => {
	if (nowSeconds >= win.startTime && nowSeconds < win.endTime) {
		return win.closedEarly ? "closed_early" : "open";
	}
	if (nowSeconds < win.startTime) return "upcoming";
	return "past";
};

const statusStyles: Record<WindowStatus, { label: string; className: string }> = {
	upcoming: {
		label: "Upcoming",
		className: "border-[#5865f2]/40 bg-[#5865f2]/10 text-[#8ea1e1]",
	},
	open: {
		label: "Open",
		className: "border-[#57f287]/40 bg-[#57f287]/10 text-[#57f287]",
	},
	closed_early: {
		label: "Closed early",
		className: "border-[#ff5a36]/40 bg-[#ff5a36]/10 text-[#ff9a80]",
	},
	past: {
		label: "Ended",
		className: "border-white/10 bg-white/5 text-zinc-400",
	},
};

const getWindowGroup = (win: ScheduleWindowDraft, now: number): WindowGroup => {
	if (now >= win.endTime) return "past";
	if (now >= win.startTime) return "active";
	return "upcoming";
};

const groupMeta: Record<WindowGroup, string> = {
	active: "Happening now",
	upcoming: "Upcoming",
	past: "Past",
};

const newDraftDefaults = (): DraftValues => {
	const now = new Date();
	const start = new Date(now.getTime() + 60 * 60 * 1000);
	start.setMinutes(0, 0, 0);
	const end = new Date(start.getTime() + 60 * 60 * 1000);
	return {
		start: toLocalInputValue(Math.floor(start.getTime() / 1000)),
		end: toLocalInputValue(Math.floor(end.getTime() / 1000)),
	};
};

const sameLocalDay = (value: string, dayOffset: number) => {
	if (!value) return false;
	const date = new Date(value);
	const target = new Date();
	target.setDate(target.getDate() + dayOffset);
	return (
		date.getFullYear() === target.getFullYear() &&
		date.getMonth() === target.getMonth() &&
		date.getDate() === target.getDate()
	);
};

function WindowEditor({
	draft,
	now,
	onDraftChange,
	error,
	onSubmit,
	onCancel,
	submitLabel,
}: {
	draft: DraftValues;
	now: number;
	onDraftChange: (next: DraftValues) => void;
	error: string | null;
	onSubmit: () => void;
	onCancel: () => void;
	submitLabel: string;
}) {
	const startUnix = fromLocalInputValue(draft.start);
	const endUnix = fromLocalInputValue(draft.end);
	const preview = humanPreview(startUnix, endUnix, now);
	const durationSeconds = endUnix - startUnix;

	const applyDay = (dayOffset: number) => {
		if (!draft.start || !draft.end) return;
		const currentStart = new Date(draft.start);
		const currentEnd = new Date(draft.end);
		if (
			!Number.isFinite(currentStart.getTime()) ||
			!Number.isFinite(currentEnd.getTime())
		) {
			return;
		}
		const duration = Math.max(
			60 * 60 * 1000,
			currentEnd.getTime() - currentStart.getTime(),
		);
		const target = new Date();
		target.setDate(target.getDate() + dayOffset);
		const nextStart = new Date(currentStart);
		nextStart.setFullYear(
			target.getFullYear(),
			target.getMonth(),
			target.getDate(),
		);
		const nextEnd = new Date(nextStart.getTime() + duration);
		onDraftChange({
			start: toLocalInputValue(Math.floor(nextStart.getTime() / 1000)),
			end: toLocalInputValue(Math.floor(nextEnd.getTime() / 1000)),
		});
	};

	const applyDuration = (hours: number) => {
		if (!draft.start) return;
		const startMs = new Date(draft.start).getTime();
		if (!Number.isFinite(startMs)) return;
		onDraftChange({
			...draft,
			end: toLocalInputValue(
				Math.floor((startMs + hours * 60 * 60 * 1000) / 1000),
			),
		});
	};

	return (
		<div className="space-y-3.5">
			<div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<FormLabel label="Opens">
					<DarkInput
						type="datetime-local"
						value={draft.start}
						onChange={(event) =>
							onDraftChange({ ...draft, start: event.target.value })
						}
						ariaLabel="Window opens at"
					/>
				</FormLabel>
				<FormLabel label="Closes">
					<DarkInput
						type="datetime-local"
						value={draft.end}
						onChange={(event) =>
							onDraftChange({ ...draft, end: event.target.value })
						}
						ariaLabel="Window closes at"
					/>
				</FormLabel>
			</div>

			<div className="flex flex-wrap items-center gap-x-4 gap-y-2">
				<div className="flex flex-wrap items-center gap-1.5">
					<span className="text-[12px] font-medium text-zinc-400">Start</span>
					{[
						{ label: "Today", offset: 0 },
						{ label: "Tomorrow", offset: 1 },
					].map((preset) => {
						const active = sameLocalDay(draft.start, preset.offset);
						return (
							<button
								key={preset.label}
								type="button"
								aria-pressed={active}
								onClick={() => applyDay(preset.offset)}
								className={`rounded-md border px-2.5 py-1 text-[12px] font-medium transition-colors ${
									active
										? "border-[#ff5a36]/40 bg-[#ff5a36]/10 text-[#ff9a80]"
										: "border-white/10 bg-white/5 text-zinc-400 hover:border-white/20 hover:text-zinc-200"
								}`}
							>
								{preset.label}
							</button>
						);
					})}
				</div>

				<div className="flex flex-wrap items-center gap-1.5">
					<span className="text-[12px] font-medium text-zinc-400">Duration</span>
					{DURATION_PRESETS.map((hours) => {
						const active = durationSeconds === hours * 60 * 60;
						return (
							<button
								key={hours}
								type="button"
								aria-pressed={active}
								onClick={() => applyDuration(hours)}
								className={`rounded-md border px-2.5 py-1 text-[12px] font-medium tabular-nums transition-colors ${
									active
										? "border-[#ff5a36]/40 bg-[#ff5a36]/10 text-[#ff9a80]"
										: "border-white/10 bg-white/5 text-zinc-400 hover:border-white/20 hover:text-zinc-200"
								}`}
							>
								{hours}h
							</button>
						);
					})}
				</div>
			</div>

			{preview && (
				<p className="rounded-md bg-[#1e1f22] px-3 py-2 text-[13px] leading-5 text-zinc-300">
					{preview}
				</p>
			)}
			{error && (
				<p role="alert" className="text-[13px] font-medium leading-5 text-red-400">
					{error}
				</p>
			)}

			<div className="flex items-center gap-2">
				<button
					type="button"
					onClick={onSubmit}
					className="rounded-md bg-zinc-100 px-3.5 py-2 text-[13px] font-medium text-zinc-950 transition-colors hover:bg-white"
				>
					{submitLabel}
				</button>
				<button
					type="button"
					onClick={onCancel}
					className="rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 transition-colors hover:text-white"
				>
					Cancel
				</button>
			</div>
		</div>
	);
}

export default function ScheduleWindows({
	value,
	onChange,
	loading = false,
}: {
	value: ScheduleWindowDraft[];
	onChange: (next: ScheduleWindowDraft[]) => void;
	loading?: boolean;
}) {
	const [adding, setAdding] = useState<DraftValues | null>(null);
	const [addError, setAddError] = useState<string | null>(null);
	const [editingIndex, setEditingIndex] = useState<number | null>(null);
	const [editDraft, setEditDraft] = useState<DraftValues>({ start: "", end: "" });
	const [editError, setEditError] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [paused, setPaused] = useState(false);
	const [pausedNow, setPausedNow] = useState(() => Math.floor(Date.now() / 1000));
	const [liveNow, setLiveNow] = useState(() => Math.floor(Date.now() / 1000));

	useEffect(() => {
		if (paused) return;
		const timer = setInterval(
			() => setLiveNow(Math.floor(Date.now() / 1000)),
			30000,
		);
		return () => clearInterval(timer);
	}, [paused]);

	const nowSeconds = paused ? pausedNow : liveNow;

	const sortedWindows = useMemo(
		() =>
			value
				.map((win, index) => ({ win, index }))
				.sort((a, b) => a.win.startTime - b.win.startTime),
		[value],
	);

	const grouped = useMemo(() => {
		const groups: Record<WindowGroup, { win: ScheduleWindowDraft; index: number }[]> = {
			active: [],
			upcoming: [],
			past: [],
		};
		for (const item of sortedWindows) {
			groups[getWindowGroup(item.win, nowSeconds)].push(item);
		}
		return groups;
	}, [sortedWindows, nowSeconds]);

	const activeWindow = sortedWindows.find(
		({ win }) =>
			nowSeconds >= win.startTime &&
			nowSeconds < win.endTime &&
			!win.closedEarly,
	)?.win;
	const nextWindow = sortedWindows.find(
		({ win }) => win.startTime > nowSeconds && !win.closedEarly,
	)?.win;
	const pastCount = grouped.past.length;
	const atLimit = value.length >= MAX_WINDOWS;

	const update = (next: ScheduleWindowDraft[]) => {
		setError(null);
		onChange(next);
	};

	const startAdding = () => {
		if (loading) return;
		if (atLimit) {
			setError(`Max ${MAX_WINDOWS} schedule windows per panel.`);
			return;
		}
		setEditingIndex(null);
		setAddError(null);
		setAdding(newDraftDefaults());
	};

	const handleAddSubmit = () => {
		if (!adding) return;
		const start = fromLocalInputValue(adding.start);
		const end = fromLocalInputValue(adding.end);
		const validationError = validateDraftTimes(start, end, value);
		if (validationError) {
			setAddError(validationError);
			return;
		}
		update([...value, { startTime: start, endTime: end, closedEarly: false }]);
		setAdding(null);
		setAddError(null);
	};

	const startEditing = (index: number) => {
		const win = value[index];
		setAdding(null);
		setEditError(null);
		setEditingIndex(index);
		setEditDraft({
			start: toLocalInputValue(win.startTime),
			end: toLocalInputValue(win.endTime),
		});
	};

	const handleEditSubmit = () => {
		if (editingIndex == null) return;
		const current = value[editingIndex];
		const start = fromLocalInputValue(editDraft.start);
		const end = fromLocalInputValue(editDraft.end);
		const timesChanged = start !== current.startTime || end !== current.endTime;
		if (timesChanged) {
			const validationError = validateDraftTimes(
				start,
				end,
				value.filter((_, index) => index !== editingIndex),
			);
			if (validationError) {
				setEditError(validationError);
				return;
			}
		}
		update(
			value.map((win, index) =>
				index === editingIndex ? { ...win, startTime: start, endTime: end } : win,
			),
		);
		setEditingIndex(null);
		setEditError(null);
	};

	const setClosedEarly = (index: number, closedEarly: boolean) => {
		update(
			value.map((win, currentIndex) =>
				currentIndex === index ? { ...win, closedEarly } : win,
			),
		);
	};

	const removeWindow = (index: number) => {
		setEditingIndex(null);
		update(value.filter((_, currentIndex) => currentIndex !== index));
	};

	const removePastWindows = () => {
		setEditingIndex(null);
		update(value.filter((win) => win.endTime > nowSeconds));
	};

	const status = loading
		? {
				title: "Loading windows",
				detail: "Checking the schedule saved with this panel.",
				dot: "bg-zinc-500",
				warning: false,
			}
		: value.length === 0
			? {
					title: "Always open",
					detail:
						"No windows scheduled. Tickets can be created at any time until you add one.",
					dot: "bg-[#57f287]",
					warning: false,
				}
			: activeWindow
				? {
						title: "Open now",
						detail: `Closes ${formatMoment(activeWindow.endTime)}.`,
						dot: "bg-[#57f287]",
						warning: false,
					}
				: nextWindow
					? {
							title: "Closed now",
							detail: `Opens ${formatMoment(nextWindow.startTime)}.`,
							dot: "bg-[#ff5a36]",
							warning: false,
						}
					: {
							title: "Closed",
							detail:
								"No upcoming windows. Past windows keep this panel closed until you remove them.",
							dot: "bg-[#ff5a36]",
							warning: true,
						};

	return (
		<SectionCard
			title="Schedule Windows"
			description="Tickets can only be opened during these windows. No windows means the panel is always open."
			action={
				<button
					type="button"
					onClick={startAdding}
					disabled={adding !== null || loading || atLimit}
					title={atLimit ? `Max ${MAX_WINDOWS} windows` : "Add a schedule window"}
					className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
				>
					<Plus className="h-3.5 w-3.5" />
					Add window
				</button>
			}
		>
			<div className="mb-4 flex flex-wrap items-start justify-between gap-2">
				<p className="flex max-w-2xl items-start gap-1.5 text-[13px] leading-5 text-zinc-400">
					<Clock className="mt-0.5 h-3.5 w-3.5 shrink-0" />
					Times are shown in your timezone ({gmtLabel()}) and saved in UTC.
					Changes apply when you save the panel.
				</p>
				<button
					type="button"
					onClick={() => {
						if (paused) {
							setLiveNow(Math.floor(Date.now() / 1000));
							setPaused(false);
							return;
						}
						setPausedNow(Math.floor(Date.now() / 1000));
						setPaused(true);
					}}
					title={
						paused
							? "Resume live status updates"
							: "Pause the clock so statuses stop changing while you edit"
					}
					className="flex items-center gap-1.5 rounded-md px-2 py-1 text-[12px] font-medium text-zinc-400 transition-colors hover:bg-white/5 hover:text-zinc-200"
				>
					{paused ? (
						<Play className="h-3 w-3" />
					) : (
						<Pause className="h-3 w-3" />
					)}
					{paused ? "Resume clock" : "Pause clock"}
				</button>
			</div>

			<div className="mb-4 flex flex-col gap-3 rounded-lg border border-white/5 bg-white/[0.02] px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
				<div className="flex min-w-0 items-start gap-2.5">
					<span
						className={`mt-1.5 h-2 w-2 shrink-0 rounded-full ${status.dot}`}
					/>
					<div className="min-w-0">
						<p className="text-[13px] font-semibold text-zinc-100">
							{status.title}
						</p>
						<p
							className={`mt-0.5 text-[12px] leading-5 ${
								status.warning ? "text-yellow-200" : "text-zinc-400"
							}`}
						>
							{status.detail}
						</p>
					</div>
				</div>
				{pastCount > 0 && (
					<button
						type="button"
						onClick={removePastWindows}
						title="Removes on Save Changes"
						className="shrink-0 self-start rounded-md border border-white/10 px-2.5 py-1.5 text-[12px] font-medium text-zinc-400 transition-colors hover:border-white/20 hover:text-zinc-100 sm:self-auto"
					>
						Remove past windows ({pastCount})
					</button>
				)}
			</div>

			{error && (
				<p
					role="alert"
					className="mb-4 rounded-md border border-red-500/20 bg-red-500/10 px-3 py-2 text-[13px] font-medium text-red-400"
				>
					{error}
				</p>
			)}

			{adding && (
				<div className="mb-3 rounded-lg border border-[#ff5a36]/30 bg-[#ff5a36]/5 p-4">
					<div className="mb-3 flex items-center gap-2">
						<span className="rounded-md border border-[#ff5a36]/40 bg-[#ff5a36]/10 px-2 py-0.5 text-[12px] font-medium text-[#ff9a80]">
							New window
						</span>
						<span className="text-[12px] text-zinc-400">
							Applies when you save the panel
						</span>
					</div>
					<WindowEditor
						draft={adding}
						now={nowSeconds}
						onDraftChange={(next) => {
							setAddError(null);
							setAdding(next);
						}}
						error={addError}
						onSubmit={handleAddSubmit}
						onCancel={() => {
							setAdding(null);
							setAddError(null);
						}}
						submitLabel="Add window"
					/>
				</div>
			)}

			{loading ? (
				<div className="space-y-2" aria-hidden="true">
					{[0, 1].map((row) => (
						<div
							key={row}
							className="h-[68px] animate-pulse rounded-lg border border-white/5 bg-white/[0.02]"
						/>
					))}
				</div>
			) : value.length === 0 && !adding ? (
				<div className="rounded-lg border border-dashed border-white/10 bg-white/[0.02] px-5 py-8 text-center">
					<p className="text-sm font-medium text-zinc-200">Always open</p>
					<p className="mx-auto mt-1 max-w-md text-[13px] leading-5 text-zinc-400">
						No windows scheduled, so tickets can be created any time. Add a window to
						limit ticket opening hours.
					</p>
					<button
						type="button"
						onClick={startAdding}
						className="mt-4 inline-flex items-center gap-1.5 rounded-md bg-zinc-100 px-3.5 py-2 text-[13px] font-medium text-zinc-950 transition-colors hover:bg-white"
					>
						<Plus className="h-3.5 w-3.5" />
						Add first window
					</button>
				</div>
			) : (
				<div className="space-y-5">
					{(["active", "upcoming", "past"] as const).map((group) => {
						const items = grouped[group];
						if (items.length === 0) return null;
						return (
							<section key={group}>
								<div className="mb-2 flex items-center justify-between gap-3">
									<h3 className="text-[12px] font-medium text-zinc-400">
										{groupMeta[group]}
									</h3>
									<span className="text-[12px] tabular-nums text-zinc-400">
										{items.length}
									</span>
								</div>
								<div className="space-y-2">
									{items.map(({ win, index }) => {
										const status = getWindowStatus(win, nowSeconds);
										const style = statusStyles[status];
										const isEditing = editingIndex === index;
										const isSaved = typeof win.id === "number";
										return (
											<div
												key={`${win.id ?? "new"}-${index}`}
												className={`rounded-lg border px-4 py-3 transition-colors ${
													isEditing
														? "border-[#ff5a36]/30 bg-[#ff5a36]/5"
														: "border-white/5 bg-white/[0.02] hover:border-white/10"
												}`}
											>
												{isEditing ? (
													<WindowEditor
														draft={editDraft}
														now={nowSeconds}
														onDraftChange={(next) => {
															setEditError(null);
															setEditDraft(next);
														}}
														error={editError}
														onSubmit={handleEditSubmit}
														onCancel={() => {
															setEditingIndex(null);
															setEditError(null);
														}}
														submitLabel="Save window"
													/>
												) : (
													<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
														<div className="min-w-0">
															<div className="flex flex-wrap items-center gap-2">
																<span
																	className={`rounded-md border px-2 py-0.5 text-[12px] font-medium ${style.className}`}
																>
																	{style.label}
																</span>
																{!isSaved && (
																	<span
																		title="Not saved yet. Press Save Changes to persist it."
																		className="rounded-md border border-white/10 bg-white/5 px-2 py-0.5 text-[12px] font-medium text-zinc-400"
																	>
																		New
																	</span>
																)}
																<span className="text-[13px] font-medium tabular-nums text-zinc-100">
																	{formatWindowRange(
																		win.startTime,
																		win.endTime,
																	)}
																</span>
															</div>
															<p className="mt-1 text-[12px] leading-5 text-zinc-400">
																{formatDuration(win.endTime - win.startTime)} ·{" "}
																{relativeWindowLabel(
																	win.startTime,
																	win.endTime,
																	nowSeconds,
																)}
															</p>
														</div>
														<div className="flex flex-wrap items-center gap-1.5">
															{win.closedEarly && status !== "past" ? (
																<button
																	type="button"
																	onClick={() => setClosedEarly(index, false)}
																	title="Allow this window to open again"
																	className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[12px] font-medium text-zinc-400 transition-colors hover:bg-white/5 hover:text-[#57f287]"
																>
																	<RotateCcw className="h-3.5 w-3.5" />
																	Reopen
																</button>
															) : status === "open" ? (
																<button
																	type="button"
																	onClick={() => setClosedEarly(index, true)}
																	title="Stop new tickets in this window"
																	className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[12px] font-medium text-zinc-400 transition-colors hover:bg-white/5 hover:text-[#ff9a80]"
																>
																	<Ban className="h-3.5 w-3.5" />
																	Close now
																</button>
															) : null}
															<button
																type="button"
																onClick={() => startEditing(index)}
																className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[12px] font-medium text-zinc-400 transition-colors hover:bg-white/5 hover:text-zinc-100"
															>
																<Pencil className="h-3.5 w-3.5" />
																Edit
															</button>
															<button
																type="button"
																onClick={() => removeWindow(index)}
																className="flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-[12px] font-medium text-zinc-400 transition-colors hover:bg-white/5 hover:text-red-400"
															>
																<Trash2 className="h-3.5 w-3.5" />
																Remove
															</button>
														</div>
													</div>
												)}
											</div>
										);
									})}
								</div>
							</section>
						);
					})}
				</div>
			)}
		</SectionCard>
	);
}
