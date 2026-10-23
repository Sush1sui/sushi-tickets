"use client";

import { createContext, useContext, useState, useMemo } from "react";
import { ChevronDown, X, Search, Check } from "lucide-react";

// ── Shared base classes ─────────────────────────────────────────────────────
const FieldLabelContext = createContext<string | undefined>(undefined);
const inputBase =
  "w-full min-w-0 bg-[#1e1f22] border border-[#45474f] rounded-md px-3 py-2.5 text-sm text-zinc-100 placeholder-zinc-400 focus:border-[#FF5A36] transition-colors duration-150";

// ── DarkInput ───────────────────────────────────────────────────────────────
export function DarkInput({
  value,
  onChange,
  placeholder,
  list,
  type = "text",
  min,
  maxLength,
  onKeyDown,
  className = "",
  readOnly,
  onClick,
  ariaLabel,
  id,
}: {
  value: string | number;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  list?: string;
  type?: string;
  min?: string;
  maxLength?: number;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  className?: string;
  readOnly?: boolean;
  onClick?: () => void;
  ariaLabel?: string;
  id?: string;
}) {
  const fieldLabel = useContext(FieldLabelContext);
  return (
    <input
      id={id}
      aria-label={ariaLabel || fieldLabel || placeholder}
      type={type}
      list={list}
      value={value}
      onChange={onChange}
      onKeyDown={onKeyDown}
      placeholder={placeholder}
      min={min}
      maxLength={maxLength}
      readOnly={readOnly}
      onClick={onClick}
      className={`${inputBase} ${readOnly ? "cursor-pointer" : ""} ${className}`}
    />
  );
}

// ── DarkTextarea ────────────────────────────────────────────────────────────
export function DarkTextarea({
  value,
  onChange,
  placeholder,
  rows = 3,
  className = "",
}: {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  placeholder?: string;
  rows?: number;
  className?: string;
}) {
  const fieldLabel = useContext(FieldLabelContext);
  return (
    <textarea
      aria-label={fieldLabel || placeholder}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      rows={rows}
      className={`${inputBase} resize-none leading-relaxed ${className}`}
    />
  );
}

// ── DarkSelect ──────────────────────────────────────────────────────────────
export function DarkSelect({
  value,
  onChange,
  options,
  placeholder = "Select...",
}: {
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
  placeholder?: string;
}) {
  const [open, setOpen] = useState(false);
  const fieldLabel = useContext(FieldLabelContext);
  const selected = options.find((o) => o.value === value);

  return (
    <div className={`relative ${open ? "z-30" : ""}`}>
      <button
        type="button"
        aria-label={fieldLabel}
        aria-expanded={open}
        onKeyDown={(event) => { if (event.key === "Escape") setOpen(false); }}
        onClick={() => setOpen((p) => !p)}
        className="w-full flex items-center justify-between gap-2 bg-[#1e1f22] border border-white/5 rounded-md px-4 py-2.5 text-sm text-left transition-colors duration-200 hover:border-white/10 focus:outline-none focus:border-[#FF5A36]/60 focus:ring-2 focus:ring-[#FF5A36]/15 "
      >
        <span className={selected ? "text-zinc-100 font-medium" : "text-zinc-400"}>
          {selected?.label ?? placeholder}
        </span>
        <div className="flex items-center gap-1.5 shrink-0">
          {value && (
            <span
              onClick={(e) => {
                e.stopPropagation();
                onChange("");
              }}
              className="text-zinc-400 hover:text-zinc-200 transition-colors p-1 hover:bg-white/5 rounded-lg cursor-pointer"
            >
              <X className="h-4 w-4" />
            </span>
          )}
          <ChevronDown
            className={`h-4 w-4 text-zinc-400 shrink-0 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          />
        </div>
      </button>

      {open && (
        <>
          {/* Backdrop */}
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute z-20 mt-2 w-full rounded-lg border border-white/10 bg-[#1e1f22] overflow-hidden animate-in fade-in slide-in-from-top-1 duration-150">
            <div className="max-h-52 overflow-y-auto py-1.5">
              {options.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => {
                    onChange(opt.value);
                    setOpen(false);
                  }}
                  className={`flex w-full items-center justify-between gap-2 px-4 py-2.5 text-sm text-left transition-colors ${
                    opt.value === value
                      ? "text-zinc-950 bg-zinc-100/20 font-bold"
                      : "text-zinc-300 hover:bg-white/10 hover:text-white"
                  }`}
                >
                  {opt.label}
                  {opt.value === value && (
                    <Check className="h-4 w-4 text-[#FF5A36] shrink-0" strokeWidth={2.5} />
                  )}
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

// ── DarkCombobox ─────────────────────────────────────────────────────────────
export function DarkCombobox({
  value,
  onChange,
  options,
  placeholder = "Search...",
  className = "",
}: {
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
  placeholder?: string;
  className?: string;
}) {
  const [open, setOpen] = useState(false);
  const fieldLabel = useContext(FieldLabelContext);
  const [query, setQuery] = useState("");

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    if (!q) return options.slice(0, 20);
    return options
      .filter(
        (o) =>
          o.label.toLowerCase().includes(q) ||
          o.value.toLowerCase().includes(q),
      )
      .slice(0, 20);
  }, [options, query]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    onChange(e.target.value);
    setOpen(true);
  };

  const handleSelect = (opt: { value: string; label: string }) => {
    onChange(opt.label);
    setQuery("");
    setOpen(false);
  };

  return (
    <div className={`relative ${open ? "z-30" : ""} ${className}`}>
      <div className="flex items-center gap-2 bg-[#1e1f22] border border-white/10 rounded-md px-4 py-2.5 transition-colors duration-200 focus-within:border-[#FF5A36] focus-within:ring-2 focus-within:ring-[#FF5A36]/20 hover:border-white/20 ">
        <Search className="h-4 w-4 text-zinc-400 shrink-0" />
        <input
          type="text"
          aria-label={fieldLabel || placeholder}
          aria-expanded={open}
          onKeyDown={(event) => { if (event.key === "Escape") setOpen(false); }}
          value={query || value}
          onChange={handleChange}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          placeholder={placeholder}
          className="min-w-0 flex-1 bg-transparent text-sm text-zinc-100 placeholder-zinc-400 focus:outline-none"
        />
        {(query || value) && (
          <button
            type="button"
            aria-label="Clear channel"
            onClick={() => {
              onChange("");
              setQuery("");
            }}
            className="text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>

      {open && filtered.length > 0 && (
        <div className="absolute z-20 mt-2 w-full rounded-lg border border-white/10 bg-[#1e1f22] overflow-hidden ">
          <div className="max-h-52 overflow-y-auto py-1.5">
            {filtered.map((opt) => (
              <button
                key={opt.value}
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => handleSelect(opt)}
                className={`flex w-full items-center justify-between gap-2 px-4 py-2.5 text-sm text-left transition-colors ${
                  opt.label === value
                    ? "text-zinc-950 bg-zinc-100/20 font-bold"
                    : "text-zinc-300 hover:bg-white/10 hover:text-white"
                }`}
              >
                <span className="flex items-center gap-2">
                  <span className="text-zinc-400 font-mono">#</span>
                  {opt.label}
                </span>
                {opt.label === value && (
                  <Check className="h-4 w-4 text-[#FF5A36] shrink-0" strokeWidth={2.5} />
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// ── DarkMultiRolePicker ─────────────────────────────────────────────────────
export function DarkMultiRolePicker({
  roles,
  selectedIds,
  onToggle,
}: {
  roles: { id: string; name: string; color?: number }[];
  selectedIds: string[];
  onToggle: (id: string) => void;
}) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const fieldLabel = useContext(FieldLabelContext);

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    return roles.filter((r) => r.name.toLowerCase().includes(q));
  }, [roles, query]);

  const selectedRoles = roles.filter((r) => selectedIds.includes(r.id));

  return (
    <div className="space-y-2.5">
      {/* Selected chips */}
      {selectedRoles.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {selectedRoles.map((role) => {
            const hex = role.color
              ? `#${role.color.toString(16).padStart(6, "0")}`
              : "#64748b";
            return (
              <span
                key={role.id}
                className="inline-flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-[13px] font-medium transition-colors hover:bg-white/10"
                style={{ color: "#dbdee1" }}
              >
                <span
                  className="h-2 w-2 rounded-full shrink-0"
                  style={{ backgroundColor: hex }}
                />
                @{role.name}
                <button
                  type="button"
                  aria-label={`Remove ${role.name}`}
                  onClick={() => onToggle(role.id)}
                  className="ml-1 text-zinc-400 hover:text-zinc-100 transition-colors"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </span>
            );
          })}
        </div>
      )}

      {/* Trigger */}
      <div className={`relative ${open ? "z-30" : ""}`}>
        <button
          type="button"
          aria-label={fieldLabel || "Select roles"}
          aria-expanded={open}
          onKeyDown={(event) => { if (event.key === "Escape") setOpen(false); }}
          onClick={() => setOpen((p) => !p)}
          className="w-full flex items-center justify-between gap-2 bg-[#1e1f22] border border-white/5 rounded-md px-4 py-2.5 text-sm text-left transition-colors duration-200 hover:border-white/10 focus:outline-none focus:border-[#FF5A36]/60 focus:ring-2 focus:ring-[#FF5A36]/15 "
        >
          <span className={selectedIds.length > 0 ? "text-zinc-100 font-medium" : "text-zinc-400"}>
            {selectedIds.length === 0
              ? "Select roles..."
              : `${selectedIds.length} role${selectedIds.length > 1 ? "s" : ""} selected`}
          </span>
          <div className="flex items-center gap-1.5 shrink-0">
            {selectedIds.length > 0 && (
              <span
                onClick={(e) => {
                  e.stopPropagation();
                  selectedIds.forEach((id) => onToggle(id));
                }}
                className="text-zinc-400 hover:text-zinc-200 transition-colors p-1 hover:bg-white/5 rounded-lg cursor-pointer"
              >
                <X className="h-4 w-4" />
              </span>
            )}
            <ChevronDown
              className={`h-4 w-4 text-zinc-400 shrink-0 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
            />
          </div>
        </button>

        {open && (
          <>
            <div
              className="fixed inset-0 z-10"
              onClick={() => setOpen(false)}
            />
            <div className="absolute z-20 mt-2 w-full rounded-lg border border-white/10 bg-[#1e1f22] overflow-hidden ">
              {/* Search */}
              <div className="flex items-center gap-2 border-b border-white/5 px-4 py-2.5">
                <Search className="h-4 w-4 text-zinc-400 shrink-0" />
                <input
                  type="text"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  aria-label="Search roles"
                  placeholder="Search roles..."
                  className="min-w-0 flex-1 bg-transparent text-sm text-zinc-100 placeholder-zinc-400 focus:outline-none"
                  autoFocus
                />
              </div>
              <div className="max-h-52 overflow-y-auto py-1.5">
                {filtered.length === 0 ? (
                  <p className="px-4 py-3 text-[13px] text-zinc-400">
                    No roles found.
                  </p>
                ) : (
                  filtered.map((role) => {
                    const checked = selectedIds.includes(role.id);
                    const hex = role.color
                      ? `#${role.color.toString(16).padStart(6, "0")}`
                      : "#64748b";
                    return (
                      <button
                        key={role.id}
                        type="button"
                        onClick={() => onToggle(role.id)}
                        className="flex w-full items-center gap-2.5 px-4 py-2.5 text-sm text-left transition-colors hover:bg-white/10"
                      >
                        <span
                          className="h-2.5 w-2.5 rounded-full shrink-0"
                          style={{ backgroundColor: hex }}
                        />
                        <span className="flex-1 text-left text-zinc-200 font-medium">
                          @{role.name}
                        </span>
                        {checked && (
                          <Check className="h-4 w-4 text-[#FF5A36] shrink-0" strokeWidth={2.5} />
                        )}
                      </button>
                    );
                  })
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

// ── DarkSearchDropdown ──────────────────────────────────────────────────────
export type DropdownItem = {
  id: string;
  label: string;
  sublabel?: string;
  color?: number;
  avatarUrl?: string;
};

export function DarkSearchDropdown({
  items,
  selectedIds,
  onSelect,
  placeholder,
}: {
  items: DropdownItem[];
  selectedIds: Set<string>;
  onSelect: (id: string) => void;
  placeholder: string;
}) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);

  const filtered = useMemo(() => {
    const q = query.toLowerCase();
    return items
      .filter((item) => !selectedIds.has(item.id))
      .filter((item) => {
        if (!q) return true;
        return (
          item.label.toLowerCase().includes(q) ||
          item.sublabel?.toLowerCase().includes(q)
        );
      })
      .slice(0, 12);
  }, [items, query, selectedIds]);

  return (
    <div className={`relative ${open ? "z-30" : ""}`}>
      <div className="flex items-center gap-2 bg-[#1e1f22] border border-white/10 rounded-md px-4 py-2.5 transition-colors duration-200 focus-within:border-[#FF5A36] focus-within:ring-2 focus-within:ring-[#FF5A36]/20 hover:border-white/20 ">
         <Search className="h-4 w-4 text-zinc-400 shrink-0" />
        <input
          type="text"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          aria-label={placeholder}
          placeholder={placeholder}
          className="min-w-0 flex-1 bg-transparent text-sm text-zinc-100 placeholder-zinc-400 focus:outline-none"
        />
      </div>

      {open && filtered.length > 0 && (
        <div className="absolute z-20 mt-2 w-full rounded-lg border border-white/10 bg-[#1e1f22] overflow-hidden ">
          <div className="max-h-56 overflow-y-auto py-1.5">
            {filtered.map((item) => {
              const hex = item.color
                ? `#${item.color.toString(16).padStart(6, "0")}`
                : undefined;
              return (
                <button
                  key={item.id}
                  type="button"
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => {
                    onSelect(item.id);
                    setQuery("");
                    setOpen(false);
                  }}
                  className="flex w-full items-center gap-2.5 px-4 py-2.5 text-sm text-left transition-colors hover:bg-white/10"
                >
                  {item.avatarUrl ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={item.avatarUrl}
                      alt=""
                      className="h-6 w-6 shrink-0 rounded-full ring-1 ring-white/10"
                    />
                  ) : hex ? (
                    <span
                      className="h-3 w-3 shrink-0 rounded-full border border-white/10"
                      style={{ backgroundColor: hex }}
                    />
                  ) : null}
                  <span className="flex-1 truncate text-zinc-200 font-medium">
                    {item.label}
                  </span>
                  {item.sublabel && (
                    <span className="truncate text-[13px] text-zinc-400">
                      {item.sublabel}
                    </span>
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

// ── DarkChipList ────────────────────────────────────────────────────────────
export function DarkChipList({
  items,
  onRemove,
  emptyMessage,
}: {
  items: DropdownItem[];
  onRemove: (id: string) => void;
  emptyMessage: string;
}) {
  if (items.length === 0) {
    return (
      <p className="py-8 text-center text-[13px] text-zinc-400 italic">{emptyMessage}</p>
    );
  }

  return (
    <div className="space-y-2">
      {items.map((item) => {
        const hex = item.color
          ? `#${item.color.toString(16).padStart(6, "0")}`
          : undefined;
        return (
          <div
            key={item.id}
            className="flex items-center gap-3 rounded-md border border-white/5 bg-white/5 px-4 py-3 hover:border-white/10 transition-colors duration-200"
          >
            {/* Avatar or color swatch */}
            {item.avatarUrl ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={item.avatarUrl}
                alt=""
                className="h-6 w-6 shrink-0 rounded-full ring-1 ring-white/10"
              />
            ) : hex ? (
              <span
                className="h-3 w-3 shrink-0 rounded-full border border-white/10"
                style={{ backgroundColor: hex }}
              />
            ) : null}
            {/* Label + sublabel */}
            <span className="flex-1 text-sm font-semibold text-zinc-100 truncate">
              {item.label}
            </span>
            {item.sublabel && (
              <span className="shrink-0 text-[13px] text-zinc-400 font-mono">
                {item.sublabel}
              </span>
            )}
            {/* Remove button */}
            <button
              type="button"
              aria-label={`Remove ${item.label}`}
              onClick={() => onRemove(item.id)}
              className="shrink-0 rounded-lg p-1.5 text-zinc-400 hover:bg-white/10 hover:text-zinc-100 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        );
      })}
    </div>
  );
}

// ── DarkSwitchTrack ─────────────────────────────────────────────────────────
export function DarkSwitchTrack({ checked }: { checked: boolean }) {
  return (
    <span
      aria-hidden="true"
      className={`relative h-5 w-9 shrink-0 rounded-full border transition-colors duration-200 ${
        checked
          ? "border-[#FF5A36] bg-[#FF5A36]"
          : "border-white/5 bg-white/10"
      }`}
    >
      <span
        className={`absolute top-1/2 h-4 w-4 -translate-y-1/2 rounded-full bg-white transition-[left] duration-200 ${
          checked ? "left-[18px]" : "left-0.5"
        }`}
      />
    </span>
  );
}

// ── DarkToggle ──────────────────────────────────────────────────────────────
export function DarkToggle({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label: string;
  description?: string;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-label={label}
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className="group flex w-full items-start gap-3.5 rounded-md text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#FF5A36]/40 focus-visible:ring-offset-2 focus-visible:ring-offset-[#1e1f22]"
    >
      <DarkSwitchTrack checked={checked} />
      <span className="min-w-0">
        <span className="block text-sm font-semibold text-zinc-200 transition-colors group-hover:text-white">
          {label}
        </span>
        {description && (
          <span className="mt-1 block text-[13px] leading-relaxed text-zinc-400">
            {description}
          </span>
        )}
      </span>
    </button>
  );
}

// ── DarkCheckbox ────────────────────────────────────────────────────────────
export function DarkCheckbox({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label: string;
  description?: string;
}) {
  return (
    <label className="flex items-start gap-3 cursor-pointer group">
      <button
        type="button"
        role="checkbox"
        aria-label={label}
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={`mt-0.5 shrink-0 rounded-lg border transition-colors duration-200 flex items-center justify-center ${
          checked
            ? "bg-[#FF5A36] border-[#FF5A36]  "
            : "bg-[#1e1f22] border-white/5 group-hover:border-white/15"
        }`}
        style={{ height: "1.25rem", width: "1.25rem" }}
      >
        {checked && <Check className="h-3.5 w-3.5 text-white" strokeWidth={3.5} />}
      </button>
      <div>
        <span className="text-sm text-zinc-200 group-hover:text-white font-medium transition-colors">
          {label}
        </span>
        {description && (
          <p className="text-[13px] text-zinc-400 mt-1 leading-relaxed">{description}</p>
        )}
      </div>
    </label>
  );
}

// ── SectionCard ─────────────────────────────────────────────────────────────
export function SectionCard({
  title,
  description,
  children,
  action,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <section className="section-card">
      <div className="section-card-header">
        <div>
          <h2>{title}</h2>
          {description && (
            <p>{description}</p>
          )}
        </div>
        {action && <div className="shrink-0">{action}</div>}
      </div>
      <div className="min-w-0">{children}</div>
    </section>
  );
}

// ── FormLabel ───────────────────────────────────────────────────────────────
export function FormLabel({
  label,
  hint,
  children,
}: {
  label: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <FieldLabelContext.Provider value={label}>
    <div className="form-label" role="group" aria-label={label}>
      <div className="form-label-title">
        <span>
          {label}
        </span>
        {hint && <span className="form-label-hint">{hint}</span>}
      </div>
      {children}
    </div>
    </FieldLabelContext.Provider>
  );
}
