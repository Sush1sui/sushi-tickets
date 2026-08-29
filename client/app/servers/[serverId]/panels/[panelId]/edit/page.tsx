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
} from "../../../../../../lib/api";
import EmojiPicker from "@/components/emoji-picker";
import DiscordMockup from "@/components/DiscordMockup";
import DarkConfirmModal from "@/components/DarkConfirmModal";
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
import { Save, Plus, Trash2, Trash } from "lucide-react";

const buttonColorOptions = [
	{ value: "primary", label: "🔵  Blue" },
	{ value: "success", label: "🟢  Green" },
	{ value: "danger", label: "🔴  Red" },
	{ value: "secondary", label: "⚫  Gray" },
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
		questions: [{ label: "", isRequired: false, style: "short", placeholder: "" }],
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

	const { roles, channels, categories, isLoading: metaLoading } = useGuildMeta(serverId);
	const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);
	const { emojis, isLoading: emojisLoading } = useGuildEmojis(
		serverId,
		form.customEmoji,
	);
	const [saving, setSaving] = useState(false);
	const [deleting, setDeleting] = useState(false);
	const [previewTab, setPreviewTab] = useState<"panel" | "welcome">("panel");
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
			questions: panel.questions?.length
				? panel.questions.map((q) => ({
						label: q.label,
						isRequired: q.isRequired,
						style: q.style || "short",
						placeholder: q.placeholder || "",
					}))
				: [{ label: "", isRequired: false, style: "short", placeholder: "" }],
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
			});

			router.push(`/servers/${serverId}/panels`);
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
				<p className="text-sm text-red-400 font-semibold bg-red-500/10 border border-red-500/20 px-4 py-2.5 rounded-xl">
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
			{/* Sticky Header */}
			<div className="sticky top-0 z-30 flex items-center justify-between mb-4 bg-[#1E1F22]/90 backdrop-blur-xl border border-white/5 rounded-2xl p-5 shadow-lg shadow-black/20">
				<div>
					<h1 className="text-xl font-black tracking-tight text-white">Edit Panel</h1>
					<p className="text-xs text-zinc-400 mt-0.5">
						Update the ticket panel details and embed parameters.
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
						{deleting ? "Deleting..." : "Delete Panel"}
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
						<p className="text-xs text-zinc-500">Loading panel data...</p>
					</div>
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
											<input
												type="color"
												value={form.color}
												onChange={(e) => setForm((p) => ({ ...p, color: e.target.value }))}
												className="h-10 w-12 rounded-xl border border-white/5 bg-[#1e1f22] cursor-pointer p-1"
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
										<p className="text-xs text-zinc-500 py-2">Loading server emojis...</p>
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

						{/* Questions */}
						<SectionCard
							title="Pre-Flight Questions"
							description="Users will fill these out before a ticket is created."
							action={
								<button
									type="button"
									onClick={() =>
										setForm((p) => ({
											...p,
											questions: [
												...p.questions,
												{ label: "", isRequired: false, style: "short", placeholder: "" },
											],
										}))
									}
									className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-bold text-zinc-300 hover:text-white hover:border-white/20 transition-all cursor-pointer shadow-sm"
								>
									<Plus className="h-3.5 w-3.5" />
									Add Question
								</button>
							}
						>
							<div className="space-y-4">
								<FormLabel label="Modal Window Title" hint="Max 45 chars (popup header)">
									<DarkInput
										value={form.questionsModalTitle}
										onChange={(e) =>
											setForm((p) => ({
												...p,
												questionsModalTitle: e.target.value.slice(0, 45),
											}))
										}
										placeholder="Sushi Ticket Questions"
									/>
								</FormLabel>

								<div className="pt-2 border-t border-white/5 space-y-3">
								{form.questions.map((question, index) => (
									<div
										key={index}
										className="rounded-xl border border-white/5 bg-white/[0.02] p-3.5 space-y-2.5 transition-all hover:border-white/10"
									>
										{/* Top row: number, label, type toggle, required toggle, delete */}
										<div className="flex items-center gap-2.5 flex-wrap sm:flex-nowrap">
											<span className="text-xs font-bold font-mono text-zinc-400 w-5 shrink-0 text-right">
												{index + 1}.
											</span>
											<DarkInput
												value={question.label}
												onChange={(e) => {
													const next = [...form.questions];
													next[index] = { ...next[index], label: e.target.value };
													setForm((p) => ({ ...p, questions: next }));
												}}
												placeholder={`Question ${index + 1} (e.g. Roblox Username)`}
												className="flex-1 min-w-[180px]"
											/>

											{/* Input Style Toggle */}
											<div className="flex items-center bg-white/5 border border-white/10 rounded-lg p-0.5 shrink-0">
												<button
													type="button"
													onClick={() => {
														const next = [...form.questions];
														next[index] = { ...next[index], style: "short" };
														setForm((p) => ({ ...p, questions: next }));
													}}
													className={`px-2.5 py-1 rounded text-[11px] font-bold transition-all ${
														(question.style || "short") === "short"
															? "bg-[#FF5A36] text-white shadow-sm"
															: "text-zinc-400 hover:text-zinc-200"
													}`}
													title="Single-line text input"
												>
													Single Line
												</button>
												<button
													type="button"
													onClick={() => {
														const next = [...form.questions];
														next[index] = { ...next[index], style: "paragraph" };
														setForm((p) => ({ ...p, questions: next }));
													}}
													className={`px-2.5 py-1 rounded text-[11px] font-bold transition-all ${
														question.style === "paragraph"
															? "bg-[#FF5A36] text-white shadow-sm"
															: "text-zinc-400 hover:text-zinc-200"
													}`}
													title="Multi-line textarea"
												>
													Text Area
												</button>
											</div>

											{/* Required toggle */}
											<button
												type="button"
												onClick={() => {
													const next = [...form.questions];
													next[index] = { ...next[index], isRequired: !next[index].isRequired };
													setForm((p) => ({ ...p, questions: next }));
												}}
												title={question.isRequired ? "Mark as optional" : "Mark as required"}
												className={`shrink-0 px-2.5 py-1.5 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all duration-200 cursor-pointer ${
													question.isRequired
														? "bg-[#FF5A36]/15 text-[#FF5A36] border border-[#FF5A36]/40 shadow-[0_0_8px_rgba(255,90,54,0.15)]"
														: "bg-white/5 text-zinc-500 border border-white/10 hover:text-zinc-300 hover:border-white/20"
												}`}
											>
												{question.isRequired ? "Required" : "Optional"}
											</button>

											<button
												type="button"
												onClick={() => {
													const next = form.questions.filter((_, i) => i !== index);
													setForm((p) => ({
														...p,
														questions: next.length
															? next
															: [{ label: "", isRequired: false, style: "short", placeholder: "" }],
													}));
												}}
												className="p-2 rounded-xl text-zinc-400 hover:text-[#FF5A36] hover:bg-white/5 transition-all shrink-0 cursor-pointer"
											>
												<Trash2 className="h-4 w-4" />
											</button>
										</div>

										{/* Bottom row: Optional Placeholder */}
										<div className="pl-7 pr-1">
											<DarkInput
												value={question.placeholder || ""}
												onChange={(e) => {
													const next = [...form.questions];
													next[index] = { ...next[index], placeholder: e.target.value };
													setForm((p) => ({ ...p, questions: next }));
												}}
												placeholder="Optional placeholder (e.g. Enter your Roblox username...)"
												className="text-xs py-1.5 bg-[#17181a] border-white/5"
											/>
										</div>
									</div>
								))}
								{form.questions.length === 0 && (
									<p className="text-xs text-zinc-400 italic text-center py-4">
										No pre-flight questions — click "Add Question" above.
									</p>
								)}
								</div>
							</div>
						</SectionCard>

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
											<input
												type="color"
												value={form.welcomeMessage.embedColor}
												onChange={(e) =>
													setForm((p) => ({
														...p,
														welcomeMessage: { ...p.welcomeMessage, embedColor: e.target.value },
													}))
												}
												className="h-10 w-12 rounded-xl border border-white/5 bg-[#1e1f22] cursor-pointer p-1"
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

					{/* RIGHT COLUMN: Real-Time Live Discord Preview (5/12 width) */}
					<div className="lg:col-span-5 lg:sticky lg:top-24 space-y-4">
						<div className="flex items-center justify-between px-4">
							<span className="text-xs uppercase font-extrabold tracking-widest text-[#FF5A36] text-glow-sushi">
								Live Previews
							</span>
							{/* Tab Switcher */}
							<div className="flex bg-white/5 border border-white/5 rounded-xl p-1 text-xs font-bold text-zinc-400 gap-1">
								<button
									type="button"
									onClick={() => setPreviewTab("panel")}
									className={`px-4 py-2 rounded-lg transition-all cursor-pointer font-bold text-xs ${previewTab === "panel"
										? "bg-[#FF5A36] text-white shadow-md shadow-orange-950/15"
										: "hover:text-zinc-200"
										}`}
								>
									Panel
								</button>
								<button
									type="button"
									onClick={() => setPreviewTab("welcome")}
									className={`px-4 py-2 rounded-lg transition-all cursor-pointer font-bold text-xs ${previewTab === "welcome"
										? "bg-[#FF5A36] text-white shadow-md shadow-orange-950/15"
										: "hover:text-zinc-200"
										}`}
								>
									Welcome
								</button>
							</div>
						</div>

						<div className="bg-zinc-900/10 border border-white/5 p-6 rounded-2xl shadow-xl backdrop-blur-md flex flex-col items-center justify-center min-h-[320px]">
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
							) : (
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
											emoji: "🔒",
											style: "danger"
										}
									]}
								/>
							)}
						</div>

						<div className="bg-zinc-950/20 border border-white/2 p-4 rounded-xl text-center">
							<p className="text-[11px] text-zinc-400 font-medium leading-relaxed">
								{previewTab === "panel"
									? "This preview displays a real-time simulation of exactly how your public ticket panel will render in your Discord channel."
									: "This preview displays a real-time simulation of exactly how the welcome embed will render in the opened private ticket channel."}
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
