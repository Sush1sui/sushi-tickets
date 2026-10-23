"use client";

import React, { useRef } from "react";
import { AlertTriangle, Info, X } from "lucide-react";
import { Dialog } from "radix-ui";

interface DarkConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: "danger" | "warning" | "info";
}

export default function DarkConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = "Confirm",
  cancelText = "Cancel",
  type = "danger",
}: DarkConfirmModalProps) {
  const previousFocus = useRef<HTMLElement | null>(null);
  if (!isOpen) return null;

  const getIcon = () => {
    switch (type) {
      case "info":
        return <Info className="h-6 w-6 text-[#FF5A36]" />;
      case "warning":
        return <AlertTriangle className="h-6 w-6 text-[#FF5A36]" />;
      case "danger":
      default:
        return <AlertTriangle className="h-6 w-6 text-red-400" />;
    }
  };

  const getConfirmButtonClass = () => {
    switch (type) {
      case "info":
        return "bg-zinc-100 hover:bg-white text-zinc-950   active:scale-95";
      case "warning":
        return "bg-zinc-100 hover:bg-white text-zinc-950   active:scale-95";
      case "danger":
      default:
        return "bg-red-300 hover:bg-red-200 text-red-950 active:scale-95";
    }
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => { if (!open) onClose(); }}>
    <Dialog.Portal>
      {/* Backdrop */}
      <Dialog.Overlay className="fixed inset-0 z-50 bg-zinc-950/75" />

      {/* Modal Box */}
      <Dialog.Content
        onOpenAutoFocus={() => { previousFocus.current = document.activeElement as HTMLElement; }}
        onCloseAutoFocus={(event) => { event.preventDefault(); previousFocus.current?.focus(); }}
        className="fixed left-1/2 top-1/2 z-50 w-[calc(100%-32px)] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg bg-[#2B2D31] p-6 animate-in fade-in zoom-in-95 duration-150">

        {/* Decorative subtle background brand light */}

        {/* Close Button */}
        <button
          type="button"
          aria-label="Close confirmation"
          onClick={onClose}
          className="absolute top-4 right-4 rounded-lg p-1 text-zinc-400 hover:text-zinc-200 hover:bg-white/5 transition-colors cursor-pointer"
        >
          <X className="h-4 w-4" />
        </button>

        {/* Content */}
        <div className="flex gap-4 items-start mt-2">
          <div className="shrink-0 pt-1">
            {getIcon()}
          </div>
          <div className="flex-1 space-y-1.5 min-w-0">
            <Dialog.Title asChild><h3 className="text-base font-medium text-white tracking-tight leading-none">
              {title}
            </h3></Dialog.Title>
            <Dialog.Description asChild><p className="text-sm text-zinc-300 leading-relaxed">
              {message}
            </p></Dialog.Description>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-6 flex justify-end gap-2.5">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-white/5 bg-white/5 px-4 py-2.5 text-[13px] font-medium text-zinc-400 hover:bg-white/10 hover:text-zinc-200 hover:border-white/10 transition-colors active:scale-95 cursor-pointer"
          >
            {cancelText}
          </button>
          <button
            type="button"
            onClick={() => {
              onConfirm();
              onClose();
            }}
            className={`rounded-xl px-4 py-2.5 text-xs font-black transition-colors cursor-pointer ${getConfirmButtonClass()}`}
          >
            {confirmText}
          </button>
        </div>
      </Dialog.Content>
    </Dialog.Portal>
    </Dialog.Root>
  );
}
