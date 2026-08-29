"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import useSWR from "swr";

import { api, type MultiPanelDetail } from "../../../../../../lib/api";
import { useGuildMeta } from "../../../../../../lib/hooks/useGuildMeta";
import { usePanels } from "../../../../../../lib/hooks/usePanels";
import DarkConfirmModal from "@/components/DarkConfirmModal";
import {
	DarkInput,
	DarkTextarea,
	DarkSelect,
	DarkCheckbox,
	SectionCard,
	FormLabel,
} from "../../../../../../components/DarkFormFields";
import { Save, Plus, Trash2, Trash } from "lucide-react";

type MultiPanelForm = {
	embedColor: string;
	title: string;
	content: string;
	authorName: string;
	authorIconUrl: string;
	authorUrl: string;
	channelId: string;
	useDropdown: boolean;
	panelConfigIds: string[];
	largeImgUrl: string;
	smallImgUrl: string;
	footer: string;
	footIconUrl: string;
};

const toHex = (value: number, fallback: string) => {
	if (!value) return fallback;
	return `#${value.toString(16).padStart(6, "0")}`;
};

export default function EditMultiPanelPage() {
	const router = useRouter();
	const params = useParams();
	const serverId = params.serverId as string;
	const multiPanelId = params.multiPanelId as string;

	const [form, setForm] = useState<MultiPanelForm>({
		embedColor: "#FF5A36",
		title: "",
		content: "",
		authorName: "",
		authorIconUrl: "",
		authorUrl: "",
		channelId: "",
		useDropdown: false,
		panelConfigIds: ["", ""],
		largeImgUrl: "",
		smallImgUrl: "",
		footer: "",
		footIconUrl: "",
	});

	const {
		data: multiPanel,
		isLoading: panelLoading,
		error,
	} = useSWR<MultiPanelDetail>(
		serverId && multiPanelId ? [serverId, multiPanelId] : null,
		() => api.multiPanels.get(serverId, multiPanelId),
		{ revalidateOnFocus: false },
	);

	const { channels, isLoading: metaLoading } = useGuildMeta(serverId);
	const { panels, isLoading: panelsLoading } = usePanels(serverId);
	const [saving, setSaving] = useState(false);
	const [deleting, setDeleting] = useState(false);
	const [modalOpen, setModalOpen] = useState(false);

	const isLoading = panelLoading || metaLoading || panelsLoading;

	useEffect(() => {
		if (!multiPanel) return;

		const panelIds = (multiPanel.panelConfigIds ?? []).map(String);
		while (panelIds.length < 2) {
			panelIds.push("");
		}

		setForm({
			title: multiPanel.title ?? "",
			content: multiPanel.content ?? "",
			embedColor: toHex(multiPanel.embedColor ?? 0, "#FF5A36"),
			authorName: "",
			authorIconUrl: "",
			authorUrl: "",
			channelId: multiPanel.channelId ?? "",
			useDropdown: multiPanel.useDropdown ?? false,
			panelConfigIds: panelIds,
			largeImgUrl: multiPanel.largeImgUrl ?? "",
			smallImgUrl: multiPanel.smallImgUrl ?? "",
			footer: multiPanel.footer ?? "",
			footIconUrl: multiPanel.footIconUrl ?? "",
		});
	}, [multiPanel]);

	const sortedPanels = useMemo(() => {
		return [...panels].sort((a, b) => a.Title.localeCompare(b.Title));
	}, [panels]);

	const handlePanelSelect = (index: number, value: string) => {
		setForm((prev) => {
			const next = [...prev.panelConfigIds];
			next[index] = value;
			return { ...prev, panelConfigIds: next };
		});
	};

	const handleAddPanel = () => {
		setForm((prev) => ({
			...prev,
			panelConfigIds: [...prev.panelConfigIds, ""],
		}));
	};

	const handleRemovePanel = (index: number) => {
		setForm((prev) => {
			if (prev.panelConfigIds.length <= 2) return prev;
			const next = prev.panelConfigIds.filter((_, idx) => idx !== index);
			return { ...prev, panelConfigIds: next };
		});
	};

	const handleSubmit = async (event: React.FormEvent) => {
		event.preventDefault();
		setSaving(true);

		try {
			const embedColor = parseInt(form.embedColor.replace("#", ""), 16) || 0;

			const panelConfigIds = form.panelConfigIds
				.filter((value) => value !== "")
				.map((value) => Number(value));

			await api.multiPanels.update(serverId, multiPanelId, {
				title: form.title,
				content: form.content,
				embedColor,
				channelId: form.channelId,
				largeImgUrl: form.largeImgUrl,
				smallImgUrl: form.smallImgUrl,
				useDropdown: form.useDropdown,
				panelConfigIds,
				footer: form.footer,
				footIconUrl: form.footIconUrl,
			});

			router.push(`/servers/${serverId}/panels`);
		} finally {
			setSaving(false);
		}
	};

	const handleDelete = async () => {
		setDeleting(true);
		try {
			await api.multiPanels.delete(serverId, multiPanelId);
			router.push(`/servers/${serverId}/panels`);
		} finally {
			setDeleting(false);
		}
	};

	if (error) {
		return (
			<div className="flex items-center justify-center py-12">
				<p className="text-sm text-red-400 font-semibold bg-red-500/10 border border-red-500/20 px-4 py-2.5 rounded-xl">
					Failed to load multi panel.
				</p>
			</div>
		);
	}

	const channelOptions = channels.map((c) => ({ value: c.id, label: `#${c.name}` }));
	const panelOptions = sortedPanels.map((p) => ({ value: String(p.ID), label: p.Title }));

	return (
		<>
		<form onSubmit={handleSubmit} className="space-y-5 pb-6">
			{/* Sticky Header */}
			<div className="sticky top-0 z-30 flex items-center justify-between mb-4 bg-[#1E1F22]/90 backdrop-blur-xl border border-white/5 rounded-2xl p-5 shadow-lg shadow-black/20">
				<div>
					<h1 className="text-xl font-black tracking-tight text-white">Edit Multi Panel</h1>
					<p className="text-xs text-zinc-400 mt-0.5">
						Update details and configuration of this multi panel embed.
					</p>
				</div>
				<div className="flex items-center gap-2.5">
					<button
						type="button"
						onClick={() => setModalOpen(true)}
						disabled={deleting || isLoading}
						className="flex items-center gap-2 rounded-xl px-4 py-2 text-xs font-bold bg-zinc-950 border border-red-900/30 text-red-400 hover:bg-red-500/10 active:scale-95 transition-all duration-200 disabled:opacity-60 cursor-pointer"
					>
						<Trash className="h-3.5 w-3.5" />
						{deleting ? "Deleting..." : "Delete Multi Panel"}
					</button>
					<button
						type="submit"
						disabled={saving || isLoading}
						className="flex items-center gap-2 rounded-xl px-5 py-2.5 text-xs font-bold bg-[#FF5A36] hover:bg-[#FF6B4A] text-white shadow-lg shadow-orange-950/20 transition-all duration-200 active:scale-95 disabled:opacity-60 shrink-0 hover:-translate-y-0.5 cursor-pointer"
					>
						<Save className="h-4 w-4" />
						{saving ? "Saving..." : "Save Changes"}
					</button>
				</div>
			</div>

			{isLoading && (
				<div className="flex items-center justify-center py-12">
					<div className="flex flex-col items-center gap-3">
						<div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
						<p className="text-xs text-zinc-500">Loading multi panel data...</p>
					</div>
				</div>
			)}

			{!isLoading && (
				<>
					{/* Multi Panel Settings */}
					<SectionCard title="Multi Panel Settings" description="General configuration of the multi panel.">
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Send to channel">
									<DarkSelect
										value={form.channelId}
										onChange={(v) => setForm((p) => ({ ...p, channelId: v }))}
										options={channelOptions}
										placeholder="Select channel..."
									/>
								</FormLabel>
								<div className="flex items-end pb-3">
									<DarkCheckbox
										label="Use Dropdown Menu"
										description="Render a select dropdown instead of side-by-side buttons."
										checked={form.useDropdown}
										onChange={(checked) =>
											setForm((p) => ({ ...p, useDropdown: checked }))
										}
									/>
								</div>
							</div>
						</div>
					</SectionCard>

					{/* Panels List */}
					<SectionCard
						title="Panels List"
						description="Assign panels to include in this layout. (Min. 2)"
						action={
							<button
								type="button"
								onClick={handleAddPanel}
								className="flex items-center gap-1.5 rounded-lg border border-zinc-800/60 bg-zinc-900/40 px-3 py-1.5 text-xs font-semibold text-zinc-400 hover:text-zinc-200 hover:border-zinc-700 transition-all"
							>
								<Plus className="h-3.5 w-3.5" />
								Add Panel Link
							</button>
						}
					>
						<div className="space-y-2.5">
							{form.panelConfigIds.map((panelId, index) => (
								<div key={index} className="flex items-center gap-2">
									<span className="text-xs font-mono text-zinc-600 w-5 shrink-0 text-right">
										{index + 1}.
									</span>
									<div className="flex-1">
										<DarkSelect
											value={panelId}
											onChange={(v) => handlePanelSelect(index, v)}
											options={panelOptions}
											placeholder="Select a panel to link..."
										/>
									</div>
									<button
										type="button"
										disabled={form.panelConfigIds.length <= 2}
										onClick={() => handleRemovePanel(index)}
										className="p-2 rounded-lg text-zinc-600 hover:text-zinc-300 hover:bg-zinc-800/60 disabled:opacity-30 disabled:hover:bg-transparent transition-all"
									>
										<Trash2 className="h-3.5 w-3.5" />
									</button>
								</div>
							))}
						</div>
					</SectionCard>

					{/* Message Details */}
					<SectionCard title="Message Details" description="Customise the Discord message embed description.">
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Embed title">
									<DarkInput
										value={form.title}
										onChange={(e) => setForm((p) => ({ ...p, title: e.target.value }))}
										placeholder="Support Tickets"
									/>
								</FormLabel>
								<FormLabel label="Embed color">
									<div className="flex items-center gap-2">
										<input
											type="color"
											value={form.embedColor}
											onChange={(e) => setForm((p) => ({ ...p, embedColor: e.target.value }))}
											className="h-10 w-12 rounded-lg border border-zinc-800/80 bg-zinc-900/60 cursor-pointer p-1"
										/>
										<DarkInput
											value={form.embedColor}
											onChange={(e) => setForm((p) => ({ ...p, embedColor: e.target.value }))}
											placeholder="#FF5A36"
											className="flex-1"
										/>
									</div>
								</FormLabel>
							</div>

							<FormLabel label="Description">
								<DarkTextarea
									value={form.content}
									onChange={(e) => setForm((p) => ({ ...p, content: e.target.value }))}
									placeholder="Choose an option below to initiate a support request."
									rows={3}
								/>
							</FormLabel>

							<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
								<FormLabel label="Author name" hint="optional">
									<DarkInput
										value={form.authorName}
										onChange={(e) => setForm((p) => ({ ...p, authorName: e.target.value }))}
										placeholder="Sushi Support"
									/>
								</FormLabel>
								<FormLabel label="Author icon URL" hint="optional">
									<DarkInput
										value={form.authorIconUrl}
										onChange={(e) => setForm((p) => ({ ...p, authorIconUrl: e.target.value }))}
										placeholder="https://..."
									/>
								</FormLabel>
								<FormLabel label="Author URL" hint="optional">
									<DarkInput
										value={form.authorUrl}
										onChange={(e) => setForm((p) => ({ ...p, authorUrl: e.target.value }))}
										placeholder="https://..."
									/>
								</FormLabel>
							</div>

							<div className="grid grid-cols-2 gap-4">
								<FormLabel label="Large image URL" hint="optional">
									<DarkInput
										value={form.largeImgUrl}
										onChange={(e) => setForm((p) => ({ ...p, largeImgUrl: e.target.value }))}
										placeholder="https://..."
									/>
								</FormLabel>
								<FormLabel label="Small image URL" hint="optional">
									<DarkInput
										value={form.smallImgUrl}
										onChange={(e) => setForm((p) => ({ ...p, smallImgUrl: e.target.value }))}
										placeholder="https://..."
									/>
								</FormLabel>
							</div>

							<div className="grid grid-cols-2 gap-4">
								<FormLabel label="Footer text" hint="optional">
									<DarkInput
										value={form.footer}
										onChange={(e) => setForm((p) => ({ ...p, footer: e.target.value }))}
										placeholder="Sushi Tickets"
									/>
								</FormLabel>
								<FormLabel label="Footer icon URL" hint="optional">
									<DarkInput
										value={form.footIconUrl}
										onChange={(e) => setForm((p) => ({ ...p, footIconUrl: e.target.value }))}
										placeholder="https://..."
									/>
								</FormLabel>
							</div>
						</div>
					</SectionCard>
				</>
			)}
		</form>

		<DarkConfirmModal
			isOpen={modalOpen}
			onClose={() => setModalOpen(false)}
			onConfirm={handleDelete}
			title="Delete Multi Panel"
			message="Are you sure you want to permanently delete this multi panel? This action cannot be undone."
			confirmText="Delete"
			type="danger"
		/>
		</>
	);
}
