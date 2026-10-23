"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import useSWR from "swr";

import {
	api,
	type DiscordEmoji,
	type DiscordRole,
	type PanelDetail,
	type QuestionItem,
	type ScheduleWindowDraft,
	type ScheduleWindowSyncPayload,
} from "../../../../../../lib/api";
import EmojiPicker from "@/components/emoji-picker";
import DiscordMockup from "@/components/DiscordMockup";
import DarkConfirmModal from "@/components/DarkConfirmModal";
import PreFlightQuestions, {
	PreFlightPreview,
} from "@/components/PreFlightQuestions";
import {
	useGuildEmojis,
	useGuildMeta,
} from "../../../../../../lib/hooks/useGuildMeta";
import {
	DarkInput,
	DarkTextarea,
	DarkSelect,
	DarkMultiRolePicker,
	SectionCard,
	FormLabel,
} from "../../../../../../components/DarkFormFields";
import ScheduleWindows from "../../../../../../components/ScheduleWindows";
import { Save, Trash } from "lucide-react";

const buttonColorOptions = [
	{ value: "primary", label: "Blue" },
	{ value: "success", label: "Green" },
	{ value: "danger", label: "Red" },
	{ value: "secondary", label: "Gray" },
];

const normalizeButtonColor = (color?: string): string => {
	if (!color) return "primary";
	if (color === "blue") return "primary";
	if (color === "green") return "success";
	if (color === "red") return "danger";
	if (color === "gray") return "secondary";
	return color;
};

type PanelForm = {
	mentionRoles: string[];
	categoryId: string;
	title: string;
	content: string;
	questions: QuestionItem[];
	questionsModalTitle: string;
	welcomeMessage: {
		embedColor: string;
		title: string;
		description: string;
		titleUrl: string;
		largeImgUrl: string;
		smallImgUrl: string;
		footerText: string;
		footerIconUrl: string;
	};
	color: string;
	channelId: string;
	buttonColor: string;
	buttonText: string;
	emoji: string;
	customEmoji: boolean;
	customEmojiId: string;
	customEmojiToken: string;
	largeImageUrl: string;
	smallImageUrl: string;
};

const toHex = (value: number, fallback: string) => {
	if (!value) return fallback;
	return `#${value.toString(16).padStart(6, "0")}`;
};

const parseCustomEmoji = (value: string) => {
	if (!value) {
		return { useCustom: false, emoji: "", customEmojiId: "", token: "" };
	}

	const match = value.match(/<a?:(.+?):(\d+)>/);
	if (match) {
		return {
			useCustom: true,
			emoji: "",
			customEmojiId: match[2],
			token: `${match[1]}:${match[2]}`,
		};
	}

	const parts = value.split(":");
	const maybeId = parts[parts.length - 1];
	if (parts.length >= 2 && /^\d+$/.test(maybeId)) {
		return {
			useCustom: true,
			emoji: "",
			customEmojiId: maybeId,
			token: value,
		};
	}

	return { useCustom: false, emoji: value, customEmojiId: "", token: "" };
};

export default function EditPanelPage() {
	const router = useRouter();
	const params = useParams();
	const serverId = params.serverId as string;
	const panelId = params.panelId as string;

	const [form, setForm] = useState<PanelForm>({
		mentionRoles: [],
		categoryId: "",
		title: "",
		content: "",
		questions: [],
		questionsModalTitle: "Sushi Ticket Questions",
		welcomeMessage: {
			embedColor: "#57f287",
			title: "",
			description: "",
			titleUrl: "",
			largeImgUrl: "",
			smallImgUrl: "",
			footerText: "",
			footerIconUrl: "",
		},
		color: "#FF5A36",
		channelId: "",
		buttonColor: "primary",
		buttonText: "Open Ticket",
		emoji: "",
		customEmoji: false,
		customEmojiId: "",
		customEmojiToken: "",
		largeImageUrl: "",
		smallImageUrl: "",
	});

	const {
		data: panel,
		isLoading: panelLoading,
		error,
	} = useSWR<PanelDetail>(
		serverId && panelId ? [serverId, panelId] : null,
		() => api.panels.get(serverId, panelId),
		{ revalidateOnFocus: false },
	);

	// Fetch-once: no auto refetch, so a stale fetch never overwrites staged edits
	const { data: scheduleWindows } = useSWR(
		`panel-schedule:${serverId}:${panelId}`,
		() => api.panelSchedule.list(serverId, panelId),
		{
			revalidateIfStale: false,
			revalidateOnFocus: false,
			revalidateOnReconnect: false,
		},
	);

	const { roles, channels, categories, isLoading: metaLoading } = useGuildMeta(serverId);
	const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);
	const { emojis, isLoading: emojisLoading } = useGuildEmojis(
		serverId,
		form.customEmoji,
	);
	const [saving, setSaving] = useState(false);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [deleting, setDeleting] = useState(false);
	const [previewTab, setPreviewTab] = useState<"panel" | "welcome" | "questions">(
		"panel",
	);
	const [modalOpen, setModalOpen] = useState(false);

	const isLoading = panelLoading || metaLoading;

	useEffect(() => {
		if (!panel) return;

		const parsedEmoji = parseCustomEmoji(panel.btnEmoji);
		const welcome = panel.welcomeMessage;

		setForm({
			mentionRoles: panel.mentionRolesOnOpen ?? [],
			categoryId: panel.categoryId ?? "",
			title: panel.title ?? "",
			content: panel.content ?? "",
			questions:
				panel.questions?.map((q) => ({
					label: q.label,
					isRequired: q.isRequired,
					style: q.style || "short",
					placeholder: q.placeholder || "",
				})) ?? [],
			questionsModalTitle: panel.questionsModalTitle || "Sushi Ticket Questions",
			welcomeMessage: {
				embedColor: toHex(welcome?.embedColor ?? 0, "#57f287"),
				title: welcome?.title ?? "",
				description: welcome?.description ?? "",
				titleUrl: welcome?.titleUrl ?? "",
				largeImgUrl: welcome?.largeImgUrl ?? "",
				smallImgUrl: welcome?.smallImgUrl ?? "",
				footerText: welcome?.footerText ?? "",
				footerIconUrl: welcome?.footerIconUrl ?? "",
			},
			color: toHex(panel.embedColor ?? 0, "#FF5A36"),
			channelId: panel.channelId ?? "",
			buttonColor: normalizeButtonColor(panel.btnColor),
			buttonText: panel.btnTxt ?? "Open Ticket",
			emoji: parsedEmoji.useCustom ? "" : parsedEmoji.emoji,
			customEmoji: parsedEmoji.useCustom,
			customEmojiId: parsedEmoji.customEmojiId,
			customEmojiToken: parsedEmoji.token,
			largeImageUrl: panel.largeImgUrl ?? "",
			smallImageUrl: panel.smallImgUrl ?? "",
		});
	}, [panel]);

	const [scheduleDraft, setScheduleDraft] = useState<
		ScheduleWindowDraft[] | null
	>(null);

	const scheduleValue = useMemo(() => {
		if (scheduleDraft) return scheduleDraft;
		if (!Array.isArray(scheduleWindows)) return undefined;
		return scheduleWindows.map(
			(w): ScheduleWindowDraft => ({
				id: w.id,
				startTime: w.startTime,
				endTime: w.endTime,
				closedEarly: w.closedEarly,
			}),
		);
	}, [scheduleDraft, scheduleWindows]);

	const sortedRoles = useMemo(() => {
		return [...roles].sort((a, b) => b.position - a.position);
	}, [roles]);

	const toggleMentionRole = (roleId: string) => {
		setForm((prev) => {
			const next = prev.mentionRoles.includes(roleId)
				? prev.mentionRoles.filter((id) => id !== roleId)
				: [...prev.mentionRoles, roleId];
			return { ...prev, mentionRoles: next };
		});
	};

	const selectedCustomEmoji = useMemo<DiscordEmoji | undefined>(() => {
		return emojis.find((emoji) => emoji.id === form.customEmojiId);
	}, [emojis, form.customEmojiId]);

	const buildEmojiValue = (emoji: DiscordEmoji | undefined) => {
		if (form.customEmojiToken) return form.customEmojiToken;
		if (!emoji) return "";
		return `${emoji.name}:${emoji.id}`;
	};

	const handleSubmit = async (event: React.FormEvent) => {
		event.preventDefault();
		setSaving(true);
		setSaveError(null);

		try {
			const embedColor = parseInt(form.color.replace("#", ""), 16) || 0;
			const emojiValue = form.customEmoji
				? buildEmojiValue(selectedCustomEmoji)
				: form.emoji;

			await api.panels.update(serverId, panelId, {
				mentionRolesOnOpen: form.mentionRoles,
				categoryId: form.categoryId,
				title: form.title,
				content: form.content,
				embedColor,
				channelId: form.channelId,
				btnColor: form.buttonColor,
				btnTxt: form.buttonText,
				btnEmoji: emojiValue,
				largeImgUrl: form.largeImageUrl,
				smallImgUrl: form.smallImageUrl,
				questions: form.questions
					.filter((q) => q.label.trim() !== "")
					.map((q) => ({
						label: q.label.trim(),
						isRequired: q.isRequired,
						style: q.style || "short",
						placeholder: q.placeholder?.trim() || "",
					})),
			questionsModalTitle: form.questionsModalTitle.trim() || "Sushi Ticket Questions",
			welcomeMessage: {
				embedColor:
					parseInt(form.welcomeMessage.embedColor.replace("#", ""), 16) || 0,
				title: form.welcomeMessage.title,
				description: form.welcomeMessage.description,
				titleUrl: form.welcomeMessage.titleUrl,
				largeImgUrl: form.welcomeMessage.largeImgUrl,
				smallImgUrl: form.welcomeMessage.smallImgUrl,
				footerText: form.welcomeMessage.footerText,
				footerIconUrl: form.welcomeMessage.footerIconUrl,
			},
			...(scheduleValue
				? {
						scheduleWindows: scheduleValue.map(
							(w): ScheduleWindowSyncPayload => ({
								...(w.id !== undefined ? { id: w.id } : {}),
								startTime: w.startTime,
								endTime: w.endTime,
								closedEarly: w.closedEarly,
							}),
						),
					}
				: {}),
		});

			router.push(`/servers/${serverId}/panels`);
		} catch (err) {
			setSaveError(
				err instanceof Error ? err.message : "Failed to save panel.",
			);
			window.scrollTo({ top: 0, behavior: "smooth" });
		} finally {
			setSaving(false);
		}
	};

	const handleDelete = async () => {
		setDeleting(true);
		try {
			await api.panels.delete(serverId, panelId);
			router.push(`/servers/${serverId}/panels`);
		} finally {
			setDeleting(false);
		}
	};

	if (error) {
		return (
			<div className="flex items-center justify-center py-12">
				<p className="text-sm text-red-400 font-semibold bg-red-500/10 border border-red-500/20 px-4 py-2.5 rounded-md">
					Failed to load panel.
				</p>
			</div>
		);
	}

	const channelOptions = channels.map((c) => ({ value: c.id, label: `#${c.name}` }));
	const categoryOptions = [
		{ value: "", label: "No category" },
		...categories.map((c) => ({ value: c.id, label: c.name })),
	];

	const customEmojiUrl = selectedCustomEmoji
		? `https://cdn.discordapp.com/emojis/${selectedCustomEmoji.id}.${selectedCustomEmoji.animated ? "gif" : "png"}`
		: undefined;

	return (
		<>
		<form onSubmit={handleSubmit} className="space-y-5 pb-6">
			{/* Top Header */}
			<div className="page-heading">
				<div>
					<h1 className="text-xl font-medium tracking-tight text-white ">Edit Panel</h1>
					<p className="text-[13px] text-zinc-300 font-semibold mt-1">
						Update the ticket panel details and embed parameters.
					</p>
				</div>
				<div className="flex items-center gap-2.5 lg:hidden">
					<button
						type="button"
						onClick={() => setModalOpen(true)}
						disabled={deleting || isLoading}
						className="flex items-center gap-2 rounded-md px-4 py-2 text-[13px] font-medium bg-zinc-950 border border-red-900/30 text-red-400 hover:bg-red-500/10 active:scale-95 transition-colors duration-200 disabled:opacity-60 cursor-pointer shrink-0"
					>
						<Trash className="h-3.5 w-3.5" />
						{deleting ? "Deleting..." : "Delete"}
					</button>
					<button
						type="submit"
						disabled={saving || isLoading}
						className="flex items-center gap-2 rounded-md px-5 py-2.5 text-[13px] font-medium bg-zinc-100 hover:bg-white text-zinc-950 transition-colors duration-200 active:scale-95 disabled:opacity-60 shrink-0 hover:-translate-y-0.5 cursor-pointer"
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
						<p className="text-[13px] text-zinc-400">Loading panel data...</p>
					</div>
				</div>
			)}

			{!isLoading && saveError && (
				<div className="rounded-lg border border-red-500/20 bg-red-500/10 px-4 py-3 text-[13px] font-semibold text-red-400">
					{saveError}
				</div>
			)}

			{!isLoading && (
				<div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
					{/* LEFT COLUMN: Clean Config Forms (7/12 width) */}
					<div className="lg:col-span-7 space-y-6">
						{/* Panel Settings */}
						<SectionCard title="Panel Settings" description="The embed message posted in your Discord channel.">
							<div className="space-y-4">
								<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
									<FormLabel label="Send to channel">
										<DarkSelect
											value={form.channelId}
											onChange={(v) => setForm((p) => ({ ...p, channelId: v }))}
											options={channelOptions}
											placeholder="Select channel..."
										/>
									</FormLabel>
									<FormLabel label="Panel title">
										<DarkInput
											value={form.title}
											onChange={(e) => setForm((p) => ({ ...p, title: e.target.value }))}
											placeholder="Support Tickets"
										/>
									</FormLabel>
									<FormLabel label="Panel color">
										<div className="flex items-center gap-2">
											<input aria-label="Embed color" type="color"
												value={form.color}
												onChange={(e) => setForm((p) => ({ ...p, color: e.target.value }))}
												className="h-10 w-12 rounded-md border border-white/5 bg-[#1e1f22] cursor-pointer p-1"
											/>
											<DarkInput
												value={form.color}
												onChange={(e) => setForm((p) => ({ ...p, color: e.target.value }))}
												placeholder="#5865F2"
												className="flex-1"
											/>
										</div>
									</FormLabel>
								</div>

								<FormLabel label="Panel description">
									<DarkTextarea
										value={form.content}
										onChange={(e) => setForm((p) => ({ ...p, content: e.target.value }))}
										placeholder="Click the button below to open a ticket."
										rows={3}
									/>
								</FormLabel>

								<div className="grid grid-cols-2 gap-4">
									<FormLabel label="Large image URL" hint="optional">
										<DarkInput
											value={form.largeImageUrl}
											onChange={(e) => setForm((p) => ({ ...p, largeImageUrl: e.target.value }))}
											placeholder="https://..."
										/>
									</FormLabel>
									<FormLabel label="Small image URL" hint="optional">
										<DarkInput
											value={form.smallImageUrl}
											onChange={(e) => setForm((p) => ({ ...p, smallImageUrl: e.target.value }))}
											placeholder="https://..."
										/>
									</FormLabel>
								</div>
							</div>
						</SectionCard>

						{/* Button */}
						<SectionCard title="Button" description="The button users click to open a ticket.">
							<div className="space-y-4">
								<div className="grid grid-cols-2 gap-4">
									<FormLabel label="Button text">
										<DarkInput
											value={form.buttonText}
											onChange={(e) => setForm((p) => ({ ...p, buttonText: e.target.value }))}
											placeholder="Open Ticket"
										/>
									</FormLabel>
									<FormLabel label="Button color">
										<DarkSelect
											value={form.buttonColor}
											onChange={(v) => setForm((p) => ({ ...p, buttonColor: v }))}
											options={buttonColorOptions}
										/>
									</FormLabel>
								</div>
								<FormLabel label="Button emoji" hint="optional">
									{emojisLoading && form.customEmoji ? (
										<p className="text-[13px] text-zinc-400 py-2">Loading server emojis...</p>
									) : (
										<EmojiPicker
											value={form.emoji}
											onChange={(value) => setForm((p) => ({ ...p, emoji: value }))}
											customEmojis={emojis}
											customEmojiId={form.customEmojiId}
											onCustomEmojiSelect={(emojiId) => {
												const picked = emojis.find((e) => e.id === emojiId);
												setForm((p) => ({
													...p,
													customEmojiId: emojiId,
													customEmojiToken: picked
														? `${picked.name}:${picked.id}`
														: p.customEmojiToken,
												}));
											}}
											useCustom={form.customEmoji}
											onToggleCustom={(useCustom) =>
												setForm((p) => ({ ...p, customEmoji: useCustom }))
											}
											onOpenChange={setEmojiPickerOpen}
										/>
									)}
								</FormLabel>
							</div>
						</SectionCard>

						{/* Ticket Channel */}
						<SectionCard title="Ticket Channel" description="Where tickets are created.">
							<div className="space-y-4">
								<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
									<FormLabel label="Ticket category">
										<DarkSelect
											value={form.categoryId}
											onChange={(v) => setForm((p) => ({ ...p, categoryId: v }))}
											options={categoryOptions}
										/>
									</FormLabel>
									<FormLabel label="Mention roles on open" hint="optional">
										<DarkMultiRolePicker
											roles={sortedRoles.map((r: DiscordRole) => ({
												id: r.id,
												name: r.name,
												color: r.color,
											}))}
											selectedIds={form.mentionRoles}
											onToggle={toggleMentionRole}
										/>
									</FormLabel>
								</div>
							</div>
						</SectionCard>

						{/* Schedule Windows */}
						<ScheduleWindows
							value={scheduleValue ?? []}
							onChange={setScheduleDraft}
							loading={scheduleValue === undefined}
						/>

						{/* Questions */}
						<PreFlightQuestions
							value={form.questions}
							onChange={(questions) =>
								setForm((p) => ({ ...p, questions }))
							}
							modalTitle={form.questionsModalTitle}
							onModalTitleChange={(questionsModalTitle) =>
								setForm((p) => ({ ...p, questionsModalTitle }))
							}
						/>

						{/* Welcome Message */}
						<SectionCard
							title="Welcome Message"
							description="Embed sent inside the ticket channel when it's opened."
						>
							<div className="space-y-4">
								<div className="grid grid-cols-2 gap-4">
									<FormLabel label="Title">
										<DarkInput
											value={form.welcomeMessage.title}
											onChange={(e) =>
												setForm((p) => ({
													...p,
													welcomeMessage: { ...p.welcomeMessage, title: e.target.value },
												}))
											}
											placeholder="Your ticket is open!"
										/>
									</FormLabel>
									<FormLabel label="Embed color">
										<div className="flex items-center gap-2">
											<input aria-label="Embed color" type="color"
												value={form.welcomeMessage.embedColor}
												onChange={(e) =>
													setForm((p) => ({
														...p,
														welcomeMessage: { ...p.welcomeMessage, embedColor: e.target.value },
													}))
												}
												className="h-10 w-12 rounded-md border border-white/5 bg-[#1e1f22] cursor-pointer p-1"
											/>
											<DarkInput
												value={form.welcomeMessage.embedColor}
												onChange={(e) =>
													setForm((p) => ({
														...p,
														welcomeMessage: { ...p.welcomeMessage, embedColor: e.target.value },
													}))
												}
												className="flex-1"
											/>
										</div>
									</FormLabel>
								</div>

								<FormLabel label="Description">
									<DarkTextarea
										value={form.welcomeMessage.description}
										onChange={(e) =>
											setForm((p) => ({
												...p,
												welcomeMessage: { ...p.welcomeMessage, description: e.target.value },
											}))
										}
										placeholder="A staff member will be with you shortly."
										rows={3}
									/>
								</FormLabel>

								<FormLabel label="Title URL" hint="optional">
									<DarkInput
										value={form.welcomeMessage.titleUrl}
										onChange={(e) =>
											setForm((p) => ({
												...p,
												welcomeMessage: { ...p.welcomeMessage, titleUrl: e.target.value },
											}))
										}
										placeholder="https://..."
									/>
								</FormLabel>

								<div className="grid grid-cols-2 gap-4">
									<FormLabel label="Large image URL" hint="optional">
										<DarkInput
											value={form.welcomeMessage.largeImgUrl}
											onChange={(e) =>
												setForm((p) => ({
													...p,
													welcomeMessage: { ...p.welcomeMessage, largeImgUrl: e.target.value },
												}))
											}
											placeholder="https://..."
										/>
									</FormLabel>
									<FormLabel label="Small image URL" hint="optional">
										<DarkInput
											value={form.welcomeMessage.smallImgUrl}
											onChange={(e) =>
												setForm((p) => ({
													...p,
													welcomeMessage: { ...p.welcomeMessage, smallImgUrl: e.target.value },
												}))
											}
											placeholder="https://..."
										/>
									</FormLabel>
								</div>

								<div className="grid grid-cols-2 gap-4">
									<FormLabel label="Footer text" hint="optional">
										<DarkInput
											value={form.welcomeMessage.footerText}
											onChange={(e) =>
												setForm((p) => ({
													...p,
													welcomeMessage: { ...p.welcomeMessage, footerText: e.target.value },
												}))
											}
											placeholder="Sushi Tickets"
										/>
									</FormLabel>
									<FormLabel label="Footer icon URL" hint="optional">
										<DarkInput
											value={form.welcomeMessage.footerIconUrl}
											onChange={(e) =>
												setForm((p) => ({
													...p,
													welcomeMessage: { ...p.welcomeMessage, footerIconUrl: e.target.value },
												}))
											}
											placeholder="https://..."
										/>
									</FormLabel>
								</div>
							</div>
						</SectionCard>
					</div>

					{/* RIGHT COLUMN: Real-Time Live Discord Preview & Sticky Action Bar (5/12 width) */}
					<div className="lg:col-span-5 lg:sticky lg:top-24 space-y-4">
						{/* Sticky Quick Actions in Right Sidebar */}
						<div className="flex items-center justify-between gap-3 bg-white/[0.03] border border-white/5 p-3 rounded-lg ">
							<button
								type="button"
								onClick={() => setModalOpen(true)}
								disabled={deleting || isLoading || saving}
								className="flex items-center gap-1.5 rounded-md px-3.5 py-2.5 text-[13px] font-medium bg-zinc-950/80 border border-red-900/40 text-red-400 hover:bg-red-500/15 active:scale-95 transition-colors duration-200 disabled:opacity-60 cursor-pointer shrink-0"
								title="Delete Panel"
							>
								<Trash className="h-3.5 w-3.5" />
								<span>{deleting ? "Deleting..." : "Delete"}</span>
							</button>
							<button
								type="submit"
								disabled={saving || isLoading || deleting}
								className="flex items-center gap-2 rounded-md px-5 py-2.5 text-[13px] font-medium bg-zinc-100 hover:bg-white text-zinc-950 transition-colors duration-200 active:scale-95 disabled:opacity-60 hover:-translate-y-0.5 cursor-pointer shrink-0"
							>
								<Save className="h-4 w-4" />
								<span>{saving ? "Saving..." : "Save Changes"}</span>
							</button>
						</div>

						<div className="flex flex-wrap items-center justify-between gap-2 px-2">
							<span className="text-[13px] font-medium text-[#FF5A36] ">
								Live Previews
							</span>
							{/* Tab Switcher */}
							<div className="flex bg-white/5 border border-white/5 rounded-md p-1 text-[13px] font-medium text-zinc-400 gap-1">
								<button
									type="button"
									aria-pressed={previewTab === "panel"}
                onClick={() => setPreviewTab("panel")}
									className={`px-4 py-2 rounded-lg transition-colors cursor-pointer font-bold text-xs ${previewTab === "panel"
										? "bg-zinc-100 text-zinc-950  "
										: "hover:text-zinc-200"
										}`}
								>
									Panel
								</button>
								<button
									type="button"
									aria-pressed={previewTab === "welcome"}
                onClick={() => setPreviewTab("welcome")}
									className={`px-4 py-2 rounded-lg transition-colors cursor-pointer font-bold text-xs ${previewTab === "welcome"
										? "bg-zinc-100 text-zinc-950  "
										: "hover:text-zinc-200"
										}`}
								>
									Welcome
								</button>
								<button
									type="button"
									aria-pressed={previewTab === "questions"}
									onClick={() => setPreviewTab("questions")}
									className={`px-4 py-2 rounded-lg transition-colors cursor-pointer font-bold text-xs ${previewTab === "questions"
										? "bg-zinc-100 text-zinc-950"
										: "hover:text-zinc-200"
										}`}
								>
									Popup
								</button>
							</div>
						</div>

						<div className="bg-zinc-900/10 border border-white/5 p-6 rounded-lg flex flex-col items-center justify-center min-h-[320px]">
							{previewTab === "panel" ? (
								<DiscordMockup
									authorName="Sushi Tickets"
									embedTitle={form.title || "Support Tickets"}
									embedDescription={form.content || "Click the button below to open a ticket."}
									embedColor={form.color || "#FF5A36"}
									largeImageUrl={form.largeImageUrl}
									smallImageUrl={form.smallImageUrl}
									fields={[]}
									buttons={[
										{
											label: form.buttonText || "Open Ticket",
											emoji: form.customEmoji ? customEmojiUrl : form.emoji || undefined,
											style: form.buttonColor as "primary" | "secondary" | "success" | "danger"
										}
									]}
								/>
							) : previewTab === "welcome" ? (
								<DiscordMockup
									authorName="Sushi Tickets"
									embedTitle={form.welcomeMessage.title || "Ticket Opened"}
									embedDescription={form.welcomeMessage.description || "A staff member will be with you shortly.\n\nSupport staff can close this ticket by clicking the button below."}
									embedColor={form.welcomeMessage.embedColor || "#57f287"}
									largeImageUrl={form.welcomeMessage.largeImgUrl}
									smallImageUrl={form.welcomeMessage.smallImgUrl}
									fields={[]}
									buttons={[
										{
											label: "Close Ticket",
											emoji: "",
											style: "danger"
										}
									]}
								/>
							) : (
								<PreFlightPreview
									title={form.questionsModalTitle}
									questions={form.questions}
								/>
							)}
						</div>

						<div className="bg-zinc-950/20 border border-white/2 p-4 rounded-md text-center">
							<p className="text-[13px] text-zinc-400 font-medium leading-relaxed">
								{previewTab === "panel"
									? "This preview displays a real-time simulation of exactly how your public ticket panel will render in your Discord channel."
									: previewTab === "welcome"
										? "This preview displays a real-time simulation of exactly how the welcome embed will render in the opened private ticket channel."
										: "This is the popup users answer before their ticket is created. Discord shows the first five questions."}
							</p>
						</div>
					</div>
				</div>
			)}
		</form>

		<DarkConfirmModal
			isOpen={modalOpen}
			onClose={() => setModalOpen(false)}
			onConfirm={handleDelete}
			title="Delete Panel"
			message="Are you sure you want to permanently delete this panel? This action cannot be undone."
			confirmText="Delete"
			type="danger"
		/>
		</>
	);
}
