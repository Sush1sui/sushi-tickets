"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import Link from "next/link";

import { api } from "../../../../lib/api";
import { useMultiPanels, usePanels } from "../../../../lib/hooks/usePanels";
import { Plus, Send, Pencil, Trash2 } from "lucide-react";
import DarkConfirmModal from "../../../../components/DarkConfirmModal";
import InlineError from "../../../../components/InlineError";

function PanelRow({
	title,
	onEdit,
	onSend,
	onDelete,
	sending,
	deleting,
}: {
	title: string;
	onEdit: () => void;
	onSend: () => void;
	onDelete: () => void;
	sending: boolean;
	deleting: boolean;
}) {
	return (
		<div className="panel-row">
			<div className="flex-1 min-w-0">
				<p className="text-sm font-semibold text-zinc-100 truncate">{title}</p>
			</div>
			<div className="panel-row-actions">
				<button
					onClick={onEdit}
					className="flex min-h-9 items-center gap-1.5 rounded-md border border-white/10 bg-white/[0.03] px-2.5 py-1.5 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:bg-white/5 hover:text-white"
				>
					<Pencil className="h-3 w-3" />
					Edit
				</button>
				<button
					onClick={onSend}
					disabled={sending}
					className="flex min-h-9 items-center gap-1.5 rounded-md border border-white/10 bg-white/[0.03] px-2.5 py-1.5 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:bg-white/5 hover:text-white disabled:opacity-50"
				>
					<Send className="h-3 w-3" />
					{sending ? "Sending..." : "Send"}
				</button>
				<button
					onClick={onDelete}
					disabled={deleting}
					className="flex min-h-9 items-center gap-1.5 rounded-md border border-red-400/15 bg-red-400/[0.04] px-2.5 py-1.5 text-[13px] font-medium text-red-300 transition-colors hover:border-red-400/30 hover:bg-red-400/10 disabled:opacity-50"
				>
					<Trash2 className="h-3 w-3" />
					{deleting ? "Deleting..." : "Delete"}
				</button>
			</div>
		</div>
	);
}

export default function PanelsPage() {
	const params = useParams();
	const serverId = params.serverId as string;
	const { panels, isLoading: panelsLoading, isError: panelsError, refresh } = usePanels(serverId);
	const {
		multiPanels,
		isLoading: multiPanelsLoading,
		isError: multiPanelsError,
		refresh: refreshMultiPanels,
	} = useMultiPanels(serverId);
	const [deletingId, setDeletingId] = useState<number | null>(null);
	const [deletingMultiId, setDeletingMultiId] = useState<number | null>(null);
	const [sendingId, setSendingId] = useState<number | null>(null);
	const [sendingMultiId, setSendingMultiId] = useState<number | null>(null);
	const [panelType, setPanelType] = useState<"single" | "multi">("single");

	// State for custom modal
	const [modalOpen, setModalOpen] = useState(false);
	const [modalConfig, setModalConfig] = useState<{
		title: string;
		message: string;
		type: "danger" | "warning" | "info";
		confirmText?: string;
		onConfirm: () => void;
	}>({
		title: "",
		message: "",
		type: "danger",
		onConfirm: () => {},
	});

	const showConfirm = (config: {
		title: string;
		message: string;
		type: "danger" | "warning" | "info";
		confirmText?: string;
		onConfirm: () => void;
	}) => {
		setModalConfig(config);
		setModalOpen(true);
	};

	const handleDelete = (panelId: number) => {
		showConfirm({
			title: "Delete Panel",
			message: "Are you sure you want to permanently delete this panel? This action cannot be undone.",
			type: "danger",
			confirmText: "Delete",
			onConfirm: async () => {
				setDeletingId(panelId);
				try {
					await api.panels.delete(serverId, panelId.toString());
					await refresh();
				} finally {
					setDeletingId(null);
				}
			}
		});
	};

	const handleDeleteMulti = (panelId: number) => {
		showConfirm({
			title: "Delete Multi Panel",
			message: "Are you sure you want to permanently delete this multi panel? This action cannot be undone.",
			type: "danger",
			confirmText: "Delete",
			onConfirm: async () => {
				setDeletingMultiId(panelId);
				try {
					await api.multiPanels.delete(serverId, panelId.toString());
					await refreshMultiPanels();
				} finally {
					setDeletingMultiId(null);
				}
			}
		});
	};

	const handleSend = async (panelId: number) => {
		setSendingId(panelId);
		try {
			await api.panels.send(serverId, panelId.toString());
		} catch (err) {
			console.error(err);
			showConfirm({
				title: "Send Failed",
				message: "Failed to send panel to Discord. Please check your bot configurations or server permissions.",
				type: "warning",
				confirmText: "OK",
				onConfirm: () => {}
			});
		} finally {
			setSendingId(null);
		}
	};

	const handleSendMulti = async (panelId: number) => {
		setSendingMultiId(panelId);
		try {
			await api.multiPanels.send(serverId, panelId.toString());
		} catch (err) {
			console.error(err);
			showConfirm({
				title: "Send Failed",
				message: "Failed to send multi panel to Discord. Please check your bot configurations or server permissions.",
				type: "warning",
				confirmText: "OK",
				onConfirm: () => {}
			});
		} finally {
			setSendingMultiId(null);
		}
	};

	const isSingle = panelType === "single";
	const items = isSingle ? panels : multiPanels;
	const listLoading = isSingle ? panelsLoading : multiPanelsLoading;
	const listError = isSingle ? panelsError : multiPanelsError;
	const listRefresh = isSingle ? refresh : refreshMultiPanels;
	const createHref = isSingle
		? `/servers/${serverId}/panels/create`
		: `/servers/${serverId}/multi-panels/create`;

	if (listError) {
		return (
			<InlineError
				message="We couldn't load this server's panels. You may no longer have access, or it may be temporarily unavailable."
				onRetry={listRefresh}
			/>
		);
	}

	const tabClass = (active: boolean) =>
		`-mb-px border-b-2 pt-1 pb-3 text-sm font-medium transition-colors ${
			active
				? "border-zinc-100 text-zinc-100"
				: "border-transparent text-zinc-400 hover:text-zinc-200"
		}`;

	return (
		<div className="space-y-5 pb-6">
			{/* Page header */}
			<div className="mb-2">
				<h1 className="text-xl font-medium tracking-tight text-white">Ticket Panels</h1>
				<p className="text-[13px] text-zinc-400 mt-0.5">
					Post one ticket button, or a menu of ticket types, in your Discord channels.
				</p>
			</div>

			<div className="flex gap-6 border-b border-white/10">
				<button
					type="button"
					aria-pressed={isSingle}
					onClick={() => setPanelType("single")}
					className={tabClass(isSingle)}
				>
					Single panels
				</button>
				<button
					type="button"
					aria-pressed={!isSingle}
					onClick={() => setPanelType("multi")}
					className={tabClass(!isSingle)}
				>
					Multi panels
				</button>
			</div>

			<section className="pt-1">
				<div className="flex flex-wrap items-start justify-between gap-4">
					<div className="max-w-2xl">
						<h2 className="text-[17px] font-medium text-white">
							{isSingle ? "Single panels" : "Multi panels"}
						</h2>
						<p className="mt-1 text-[13px] leading-5 text-zinc-400">
							{isSingle
								? "One ticket button in each message. Use this when every ticket starts the same way."
								: "A menu of ticket types in one message. Each button opens its own queue."}
						</p>
					</div>
					<Link
						href={createHref}
						className="flex min-h-10 items-center gap-2 rounded-md bg-zinc-100 px-4 py-2 text-sm font-medium text-zinc-950 transition-colors hover:bg-white"
					>
						<Plus className="h-3.5 w-3.5" />
						{isSingle ? "New single panel" : "New multi panel"}
					</Link>
				</div>

				<div className="mt-5">
					{listLoading ? (
						<div className="flex items-center justify-center border-t border-white/10 py-10">
							<div className="h-5 w-5 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
						</div>
					) : items.length === 0 ? (
						<div className="border-y border-white/10 py-12 text-center">
							<p className="text-sm font-medium text-zinc-200">
								{isSingle ? "No single panels yet" : "No multi panels yet"}
							</p>
							<p className="mt-1 text-[13px] text-zinc-400">
								{isSingle
									? "Create one with a single ticket button."
									: "Create one to group several ticket types."}
							</p>
						</div>
					) : (
						<div className="border-t border-white/10">
							{items.map((panel) => (
								<PanelRow
									key={`${panelType}-${panel.ID}`}
									title={panel.Title}
									onEdit={() =>
										(window.location.href = `/servers/${serverId}/${
											isSingle ? "panels" : "multi-panels"
										}/${panel.ID}/edit`)
									}
									onSend={() =>
										isSingle ? handleSend(panel.ID) : handleSendMulti(panel.ID)
									}
									onDelete={() =>
										isSingle ? handleDelete(panel.ID) : handleDeleteMulti(panel.ID)
									}
									sending={
										isSingle
											? sendingId === panel.ID
											: sendingMultiId === panel.ID
									}
									deleting={
										isSingle
											? deletingId === panel.ID
											: deletingMultiId === panel.ID
									}
								/>
							))}
						</div>
					)}
				</div>
			</section>

			<DarkConfirmModal
				isOpen={modalOpen}
				onClose={() => setModalOpen(false)}
				onConfirm={modalConfig.onConfirm}
				title={modalConfig.title}
				message={modalConfig.message}
				confirmText={modalConfig.confirmText}
				type={modalConfig.type}
			/>
		</div>
	);
}
