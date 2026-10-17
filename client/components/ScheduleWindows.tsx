"use client";

import { useEffect, useMemo, useState } from "react";
import type { ScheduleWindowDraft } from "../lib/api";
import { SectionCard, FormLabel, DarkInput } from "./DarkFormFields";
import { Plus, Trash2, Pencil, RotateCcw, Ban } from "lucide-react";

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
  const opts: Intl.DateTimeFormatOptions = {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  };
  return `${new Date(startTime * 1000).toLocaleString(undefined, opts)} → ${new Date(
    endTime * 1000,
  ).toLocaleString(undefined, opts)}`;
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

const emptyDraft = (): DraftValues => {
  const now = new Date();
  const start = new Date(now.getTime() + 60 * 60 * 1000);
  start.setMinutes(0, 0, 0);
  const end = new Date(start.getTime() + 60 * 60 * 1000);
  return {
    start: toLocalInputValue(Math.floor(start.getTime() / 1000)),
    end: toLocalInputValue(Math.floor(end.getTime() / 1000)),
  };
};

export default function ScheduleWindows({
  value,
  onChange,
}: {
  value: ScheduleWindowDraft[];
  onChange: (next: ScheduleWindowDraft[]) => void;
}) {
  const [draft, setDraft] = useState<DraftValues>(emptyDraft);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editDraft, setEditDraft] = useState<DraftValues>({ start: "", end: "" });
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

  const sortedWindowsChanged = (mapFn: (win: ScheduleWindowDraft) => ScheduleWindowDraft) => {
    return value.map((win, index) => (editingIndex === index ? mapFn(win) : win));
  };

  const handleAdd = () => {
    const start = fromLocalInputValue(draft.start);
    const end = fromLocalInputValue(draft.end);
    if (!start || !end || end <= start) {
      setError("End time must be after start time.");
      return;
    }
    if (value.length >= 50) {
      setError("Max 50 schedule windows per panel.");
      return;
    }
    update([...value, { startTime: start, endTime: end, closedEarly: false }]);
    setDraft(emptyDraft());
  };

  const handleSaveEdit = () => {
    if (editingIndex == null) return;
    const start = fromLocalInputValue(editDraft.start);
    const end = fromLocalInputValue(editDraft.end);
    if (!start || !end || end <= start) {
      setError("End time must be after start time.");
      return;
    }
    update(sortedWindowsChanged((win) => ({ ...win, startTime: start, endTime: end })));
    setEditingIndex(null);
  };

  return (
    <SectionCard
      title="Schedule Windows"
      description="Tickets can only be opened during these windows. Saved with the panel. Panels without windows are always open."
      action={
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
          className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-3 py-1.5 text-[11px] font-bold text-zinc-400 hover:text-zinc-200 hover:border-white/20 transition-all cursor-pointer"
          title={
            frozen
              ? "Resume live status updates"
              : "Freeze the clock so statuses stop changing while you edit"
          }
        >
          <RotateCcw className="h-3 w-3" />
          {frozen ? "Live" : "Frozen"}
        </button>
      }
    >
      {error && (
        <p className="mb-4 text-xs font-semibold text-red-400 bg-red-500/10 border border-red-500/20 px-3 py-2 rounded-xl">
          {error}
        </p>
      )}

      {sortedWindows.length === 0 ? (
        <p className="text-xs text-zinc-500 italic text-center py-4">
          No schedule windows — this panel is always open.
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
                className="rounded-xl border border-white/5 bg-white/[0.02] p-3.5 transition-all hover:border-white/10"
              >
                {isEditing ? (
                  <div className="space-y-3">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                      <FormLabel label="Opens">
                        <DarkInput
                          type="datetime-local"
                          value={editDraft.start}
                          onChange={(e) =>
                            setEditDraft((p) => ({ ...p, start: e.target.value }))
                          }
                        />
                      </FormLabel>
                      <FormLabel label="Closes">
                        <DarkInput
                          type="datetime-local"
                          value={editDraft.end}
                          onChange={(e) =>
                            setEditDraft((p) => ({ ...p, end: e.target.value }))
                          }
                        />
                      </FormLabel>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={handleSaveEdit}
                        className="px-3.5 py-2 rounded-xl text-xs font-bold bg-[#FF5A36] hover:bg-[#FF6B4A] text-white transition-all cursor-pointer"
                      >
                        Save Window
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingIndex(null)}
                        className="px-3.5 py-2 rounded-xl text-xs font-bold border border-white/10 bg-white/5 text-zinc-300 hover:text-white transition-all cursor-pointer"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
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
                            className="px-2 py-0.5 rounded-md border border-white/10 bg-white/5 text-[10px] font-black tracking-widest text-zinc-500"
                            title="This window is not saved yet — press Save Changes below to persist it"
                          >
                            NEW
                          </span>
                        )}
                        <span className="text-xs font-mono text-zinc-300 truncate">
                          {formatWindowRange(win.startTime, win.endTime)}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <button
                        type="button"
                        onClick={() =>
                          update(
                            sortedWindowsChanged((w) => ({ ...w, closedEarly: true })),
                          )
                        }
                        disabled={win.closedEarly || status === "past"}
                        title="Close this window early (applies when you save)"
                        className="p-2 rounded-xl text-zinc-400 hover:text-[#FF5A36] hover:bg-white/5 transition-all cursor-pointer disabled:opacity-30 disabled:hover:text-zinc-400 disabled:cursor-not-allowed"
                      >
                        <Ban className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() =>
                          update(
                            sortedWindowsChanged((w) => ({ ...w, closedEarly: false })),
                          )
                        }
                        disabled={!win.closedEarly}
                        title="Reopen window (undoes early close)"
                        className="p-2 rounded-xl text-zinc-400 hover:text-[#57f287] hover:bg-white/5 transition-all cursor-pointer disabled:opacity-30 disabled:hover:text-zinc-400 disabled:cursor-not-allowed"
                      >
                        <RotateCcw className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        onClick={() => {
                          setEditingIndex(index);
                          setEditDraft({
                            start: toLocalInputValue(win.startTime),
                            end: toLocalInputValue(win.endTime),
                          });
                        }}
                        title="Edit window"
                        className="p-2 rounded-xl text-zinc-400 hover:text-zinc-100 hover:bg-white/5 transition-all cursor-pointer"
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
                        className="p-2 rounded-xl text-zinc-400 hover:text-red-400 hover:bg-white/5 transition-all cursor-pointer"
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

      <div className="pt-2 mt-4 border-t border-white/5 space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <FormLabel label="New window — opens">
            <DarkInput
              type="datetime-local"
              value={draft.start}
              onChange={(e) => setDraft((p) => ({ ...p, start: e.target.value }))}
            />
          </FormLabel>
          <FormLabel label="New window — closes">
            <DarkInput
              type="datetime-local"
              value={draft.end}
              onChange={(e) => setDraft((p) => ({ ...p, end: e.target.value }))}
            />
          </FormLabel>
        </div>
        <button
          type="button"
          onClick={handleAdd}
          className="flex items-center gap-1.5 rounded-xl border border-[#FF5A36]/40 bg-[#FF5A36]/10 px-3.5 py-2 text-xs font-bold text-[#FF5A36] hover:bg-[#FF5A36]/20 hover:border-[#FF5A36]/60 transition-all cursor-pointer"
        >
          <Plus className="h-3.5 w-3.5" />
          Add Window
        </button>
      </div>
    </SectionCard>
  );
}
