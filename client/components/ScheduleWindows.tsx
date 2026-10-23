"use client";

import { useEffect, useMemo, useState } from "react";
import type { ScheduleWindowDraft } from "../lib/api";
import { SectionCard, FormLabel, DarkInput } from "./DarkFormFields";
import { Plus, Trash2, Pencil, RotateCcw, Ban, Clock } from "lucide-react";

type WindowStatus = "upcoming" | "open" | "closed_early" | "past";
type DraftValues = { start: string; end: string };

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

const humanPreview = (startUnix: number, endUnix: number) => {
  if (!startUnix || !endUnix || endUnix <= startUnix) return null;
  const opts: Intl.DateTimeFormatOptions = {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  };
  const startStr = new Date(startUnix * 1000).toLocaleString(undefined, opts);
  const endStr = new Date(endUnix * 1000).toLocaleString(undefined, opts);
  const now = Math.floor(Date.now() / 1000);
  let rel: string;
  if (startUnix > now) {
    const mins = Math.round((startUnix - now) / 60);
    rel = mins <= 90 ? `opens in ${mins} min` : `opens in ${Math.round(mins / 60)} h`;
  } else if (endUnix > now) {
    const mins = Math.round((endUnix - now) / 60);
    rel = mins <= 90 ? `open now. closes in ${mins} min` : `open now. closes in ${Math.round(mins / 60)} h`;
  } else {
    rel = "already over";
  }
  return `${startStr} → ${endStr} · ${rel}`;
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
  if (endUnix <= now) return "Close time must be in the future. or delete the window instead.";
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
    label: "UPCOMING",
    className: "bg-[#5865F2]/10 text-[#7289DA] border-[#5865F2]/40",
  },
  open: {
    label: "OPEN",
    className: "bg-[#57f287]/10 text-[#57f287] border-[#57f287]/40",
  },
  closed_early: {
    label: "CLOSED EARLY",
    className: "bg-[#FF5A36]/10 text-[#FF5A36] border-[#FF5A36]/40",
  },
  past: {
    label: "CLOSED",
    className: "bg-white/5 text-zinc-500 border-white/10",
  },
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

function WindowEditor({
  draft,
  onDraftChange,
  error,
  onSubmit,
  onCancel,
  submitLabel,
}: {
  draft: DraftValues;
  onDraftChange: (next: DraftValues) => void;
  error: string | null;
  onSubmit: () => void;
  onCancel: () => void;
  submitLabel: string;
}) {
  const preview = humanPreview(
    fromLocalInputValue(draft.start),
    fromLocalInputValue(draft.end),
  );
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <FormLabel label="Opens">
          <DarkInput
            type="datetime-local"
            value={draft.start}
            onChange={(e) => onDraftChange({ ...draft, start: e.target.value })}
          />
        </FormLabel>
        <FormLabel label="Closes">
          <DarkInput
            type="datetime-local"
            value={draft.end}
            onChange={(e) => onDraftChange({ ...draft, end: e.target.value })}
          />
        </FormLabel>
      </div>
      {preview && (
        <p className="text-[13px] text-zinc-400 font-medium">{preview}</p>
      )}
      {error && (
        <p className="text-[13px] font-semibold text-red-400">{error}</p>
      )}
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onSubmit}
          className="px-3.5 py-2 rounded-md text-[13px] font-medium bg-zinc-100 hover:bg-white text-zinc-950 transition-colors cursor-pointer"
        >
          {submitLabel}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-3.5 py-2 rounded-md text-[13px] font-medium border border-white/10 bg-white/5 text-zinc-300 hover:text-white transition-colors cursor-pointer"
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
}: {
  value: ScheduleWindowDraft[];
  onChange: (next: ScheduleWindowDraft[]) => void;
}) {
  const [adding, setAdding] = useState<DraftValues | null>(null);
  const [addError, setAddError] = useState<string | null>(null);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editDraft, setEditDraft] = useState<DraftValues>({ start: "", end: "" });
  const [editError, setEditError] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [frozen, setFrozen] = useState(false);
  const [frozenNow, setFrozenNow] = useState(() => Math.floor(Date.now() / 1000));
  const [liveNow, setLiveNow] = useState<number | null>(null);

  useEffect(() => {
    if (frozen) return;
    const timer = setInterval(
      () => setLiveNow(Math.floor(Date.now() / 1000)),
      30000,
    );
    return () => clearInterval(timer);
  }, [frozen]);

  const nowSeconds = frozen ? frozenNow : (liveNow ?? frozenNow);

  const sortedWindows = useMemo(
    () =>
      value
        .map((win, index) => ({ win, index }))
        .sort((a, b) => a.win.startTime - b.win.startTime),
    [value],
  );

  const update = (next: ScheduleWindowDraft[]) => {
    setError(null);
    onChange(next);
  };

  const startAdding = () => {
    if (value.length >= 50) {
      setError("Max 50 schedule windows per panel.");
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
    const timesChanged =
      start !== current.startTime || end !== current.endTime;
    if (timesChanged) {
      const validationError = validateDraftTimes(
        start,
        end,
        value.filter((_, i) => i !== editingIndex),
      );
      if (validationError) {
        setEditError(validationError);
        return;
      }
    }
    update(
      value.map((win, i) =>
        i === editingIndex ? { ...win, startTime: start, endTime: end } : win,
      ),
    );
    setEditingIndex(null);
    setEditError(null);
  };

  const changeWindows = (mapFn: (win: ScheduleWindowDraft) => ScheduleWindowDraft) => {
    update(value.map(mapFn));
  };

  return (
    <SectionCard
      title="Schedule Windows"
      description="Tickets can only be opened during these windows. Saved with the panel. Panels without windows are always open."
      action={
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => {
              if (frozen) {
                setFrozen(false);
                return;
              }
              setFrozenNow(Math.floor(Date.now() / 1000));
              setFrozen(true);
            }}
            className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-[13px] font-medium text-zinc-400 hover:text-zinc-200 hover:border-white/20 transition-colors cursor-pointer"
            title={
              frozen
                ? "Resume live status updates"
                : "Freeze the clock so statuses stop changing while you edit"
            }
          >
            <RotateCcw className="h-3 w-3" />
            {frozen ? "Live" : "Frozen"}
          </button>
          <button
            type="button"
            onClick={startAdding}
            disabled={adding !== null}
            className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-[13px] font-medium text-zinc-300 hover:text-white hover:border-white/20 transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
          >
            <Plus className="h-3.5 w-3.5" />
            Add Window
          </button>
        </div>
      }
    >
      <div className="flex items-center gap-1.5 mb-4">
        <Clock className="h-3 w-3 text-zinc-400" />
        <p className="text-[13px] text-zinc-400 font-medium">
          Times shown in your local timezone ({gmtLabel()}). Saved as UTC, so
          they&apos;re correct for everyone. Changes apply when you press Save
          Changes.
        </p>
      </div>

      {error && (
        <p className="mb-4 text-[13px] font-semibold text-red-400 bg-red-500/10 border border-red-500/20 px-3 py-2 rounded-md">
          {error}
        </p>
      )}

      {adding && (
        <div className="mb-2 rounded-md border border-dashed border-[#FF5A36]/40 bg-[#FF5A36]/5 p-3.5">
          <div className="flex items-center gap-2 mb-3">
            <span className="px-2 py-0.5 rounded-md border border-[#FF5A36]/40 bg-[#FF5A36]/10 text-[13px] font-medium text-[#FF5A36]">
              NEW
            </span>
            <span className="text-[13px] text-zinc-400 font-semibold">
              Adding a new window
            </span>
          </div>
          <WindowEditor
            draft={adding}
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
            submitLabel="Add"
          />
        </div>
      )}

      {sortedWindows.length === 0 && !adding ? (
        <p className="text-[13px] text-zinc-400 italic text-center py-4">
          No schedule windows. This panel is always open. Click &quot;Add
          Window&quot; to queue one.
        </p>
      ) : (
        <div className="space-y-2">
          {sortedWindows.map(({ win, index }) => {
            const status = getWindowStatus(win, nowSeconds);
            const style = statusStyles[status];
            const isEditing = editingIndex === index;
            const isSavedRow = typeof win.id === "number";
            return (
              <div
                key={`${win.id ?? "new"}-${index}`}
                className={`rounded-xl border p-3.5 transition-colors ${
                  isEditing
                    ? "border-[#FF5A36]/40 bg-[#FF5A36]/5"
                    : "border-white/5 bg-white/[0.02] hover:border-white/10"
                }`}
              >
                {isEditing ? (
                  <WindowEditor
                    draft={editDraft}
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
                    submitLabel="Done"
                  />
                ) : (
                  <div className="flex items-center gap-3 flex-wrap sm:flex-nowrap">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span
                          className={`px-2 py-0.5 rounded-md border text-[10px] font-black tracking-widest ${style.className}`}
                        >
                          {style.label}
                        </span>
                        {!isSavedRow && (
                          <span
                            className="px-2 py-0.5 rounded-md border border-white/10 bg-white/5 text-[13px] font-medium text-zinc-400"
                            title="This window is not saved yet. press Save Changes below to persist it"
                          >
                            NEW
                          </span>
                        )}
                        <span className="text-[13px] font-mono text-zinc-300 truncate">
                          {formatWindowRange(win.startTime, win.endTime)}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <button
                        type="button"
                        onClick={() =>
                          changeWindows((w) => ({ ...w, closedEarly: true }))
                        }
                        disabled={win.closedEarly || status === "past"}
                        title="Close this window early (applies when you save)"
                        className="p-2 rounded-md text-zinc-400 hover:text-[#FF5A36] hover:bg-white/5 transition-colors cursor-pointer disabled:opacity-30 disabled:hover:text-zinc-400 disabled:cursor-not-allowed"
                      >
                        <Ban className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          changeWindows((w) => ({ ...w, closedEarly: false }))
                        }
                        disabled={!win.closedEarly}
                        title="Reopen window (undoes early close)"
                        className="p-2 rounded-md text-zinc-400 hover:text-[#57f287] hover:bg-white/5 transition-colors cursor-pointer disabled:opacity-30 disabled:hover:text-zinc-400 disabled:cursor-not-allowed"
                      >
                        <RotateCcw className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() => startEditing(index)}
                        title="Edit window"
                        className="p-2 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-white/5 transition-colors cursor-pointer"
                      >
                        <Pencil className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          update(value.filter((_, i) => i !== index));
                          setEditingIndex(null);
                        }}
                        title="Remove window (applies when you save)"
                        className="p-2 rounded-md text-zinc-400 hover:text-red-400 hover:bg-white/5 transition-colors cursor-pointer"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </SectionCard>
  );
}
