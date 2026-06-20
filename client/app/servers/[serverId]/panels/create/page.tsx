"use client";

import { useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";

import { api, type DiscordEmoji, type DiscordRole, type QuestionItem } from "../../../../../lib/api";
import EmojiPicker from "@/components/emoji-picker";
import DiscordMockup from "@/components/DiscordMockup";
import { useGuildEmojis, useGuildMeta } from "../../../../../lib/hooks/useGuildMeta";
import {
	DarkInput,
	DarkTextarea,
	DarkSelect,
	DarkMultiRolePicker,
	SectionCard,
	FormLabel,
} from "../../../../../components/DarkFormFields";
import { Save, Plus, Trash2 } from "lucide-react";

const buttonColorOptions = [
	{ value: "primary", label: "🔵  Blue" },
	{ value: "success", label: "🟢  Green" },
	{ value: "danger", label: "🔴  Red" },
	{ value: "secondary", label: "⚫  Gray" },
];

type PanelForm = {
	mentionRoles: string[];
	categoryId: string;
	title: string;
	content: string;
	questions: QuestionItem[];
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

export default function CreatePanelPage() {
	const router = useRouter();
	const params = useParams();
	const serverId = params.serverId as string;

	const [form, setForm] = useState<PanelForm>({
		mentionRoles: [],
		categoryId: "",
		title: "",
		content: "",
		questions: [{ label: "", isRequired: false }],
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

	const { roles, channels, categories, isLoading } = useGuildMeta(serverId);
	const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);
	const { emojis, isLoading: emojisLoading } = useGuildEmojis(
		serverId,
		form.customEmoji,
	);
	const [saving, setSaving] = useState(false);
	const [previewTab, setPreviewTab] = useState<"panel" | "welcome">("panel");

	const sortedRoles = useMemo(
		() => [...roles].sort((a, b) => b.position - a.position),
		[roles],
	);

	const toggleMentionRole = (roleId: string) => {
		setForm((prev) => {
			const next = prev.mentionRoles.includes(roleId)
				? prev.mentionRoles.filter((id) => id !== roleId)
				: [...prev.mentionRoles, roleId];
			return { ...prev, mentionRoles: next };
		});
	};

	const selectedCustomEmoji = useMemo<DiscordEmoji | undefined>(
		() => emojis.find((emoji) => emoji.id === form.customEmojiId),
		[emojis, form.customEmojiId],
	);

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

			await api.panels.create(serverId, {
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
				questions: form.questions.filter((q) => q.label.trim() !== ""),
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

	const channelOptions = channels.map((c) => ({ value: c.id, label: `#${c.name}` }));
	const categoryOptions = [
		{ value: "", label: "No category" },
		...categories.map((c) => ({ value: c.id, label: c.name })),
	];

	const customEmojiUrl = selectedCustomEmoji
		? `https://cdn.discordapp.com/emojis/${selectedCustomEmoji.id}.${selectedCustomEmoji.animated ? "gif" : "png"}`
		: undefined;

	return (
		<form onSubmit={handleSubmit} className="space-y-6 pb-6">
			{/* Header */}
			<div className="flex items-center justify-between mb-4 bg-white/2 border border-white/5 rounded-2xl p-5 shadow-sm backdrop-blur-md">
				<div>
					<h1 className="text-2xl font-black tracking-tight text-white uppercase text-glow-sushi/10">Create Panel</h1>
					<p className="text-xs text-zinc-300 font-semibold mt-1">
						Configure the ticket panel posted in your Discord channel.
					</p>
				</div>
				<button
					type="submit"
					disabled={saving || isLoading}
					className="flex items-center gap-2 rounded-xl px-5 py-2.5 text-xs font-bold bg-[#FF5A36] hover:bg-[#FF6B4A] text-white shadow-lg shadow-orange-950/20 transition-all duration-200 active:scale-95 disabled:opacity-60 shrink-0 hover:-translate-y-0.5"
				>
					<Save className="h-4 w-4" />
					{saving ? "Saving..." : "Create Panel"}
				</button>
			</div>

			{isLoading && (
				<div className="flex items-center justify-center py-12">
					<div className="flex flex-col items-center gap-3">
						<div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
						<p className="text-xs text-zinc-400 font-bold">Loading server data...</p>
					</div>
				</div>
			)}

			{/* Two-Column Responsive Split Layout */}
			<div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">

				{/* LEFT COLUMN: Clean Config Forms (7/12 width) */}
				<div className="lg:col-span-7 space-y-6">

					{/* Panel Settings */}
					<SectionCard title="Panel Info & Channel" description="Basic details of the message posted in your channel.">
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Target channel">
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
							</div>

							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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

								<div className="grid grid-cols-2 gap-2">
									<FormLabel label="Large image" hint="optional">
										<DarkInput
											value={form.largeImageUrl}
											onChange={(e) => setForm((p) => ({ ...p, largeImageUrl: e.target.value }))}
											placeholder="https://..."
										/>
									</FormLabel>
									<FormLabel label="Small image" hint="optional">
										<DarkInput
											value={form.smallImageUrl}
											onChange={(e) => setForm((p) => ({ ...p, smallImageUrl: e.target.value }))}
											placeholder="https://..."
										/>
									</FormLabel>
								</div>
							</div>

							<FormLabel label="Panel description text">
								<DarkTextarea
									value={form.content}
									onChange={(e) => setForm((p) => ({ ...p, content: e.target.value }))}
									placeholder="Click the button below to open a ticket."
									rows={3}
								/>
							</FormLabel>
						</div>
					</SectionCard>

					{/* Button & Styling */}
					<SectionCard title="Ticket Button Config" description="The actual button users click inside the channel.">
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Button text">
									<DarkInput
										value={form.buttonText}
										onChange={(e) => setForm((p) => ({ ...p, buttonText: e.target.value }))}
										placeholder="Open Ticket"
									/>
								</FormLabel>
								<FormLabel label="Button style / color">
									<DarkSelect
										value={form.buttonColor}
										onChange={(v) => setForm((p) => ({ ...p, buttonColor: v }))}
										options={buttonColorOptions}
									/>
								</FormLabel>
							</div>
							<FormLabel label="Button emoji" hint="optional">
								{emojisLoading && form.customEmoji ? (
									<p className="text-xs text-zinc-400 py-2">Loading server emojis...</p>
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

					{/* Routing Rules */}
					<SectionCard title="Routing & Assignment" description="Where channels are created and who to alert.">
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Create under Category">
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
									setForm((p) => ({ ...p, questions: [...p.questions, { label: "", isRequired: false }] }))
								}
								className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-3 py-1.5 text-xs font-bold text-zinc-300 hover:text-white hover:border-white/20 transition-all cursor-pointer shadow-sm"
							>
								<Plus className="h-3.5 w-3.5" />
								Add Question
							</button>
						}
					>
						<div className="space-y-3">
							{form.questions.map((question, index) => (
								<div key={index} className="flex items-center gap-3">
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
										placeholder={`Question ${index + 1} (e.g. Describe your issue...)`}
										className="flex-1"
									/>
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
											setForm((p) => ({ ...p, questions: next.length ? next : [{ label: "", isRequired: false }] }));
										}}
										className="p-2.5 rounded-xl text-zinc-400 hover:text-[#FF5A36] hover:bg-white/5 transition-all shrink-0 cursor-pointer"
									>
										<Trash2 className="h-4 w-4" />
									</button>
								</div>
							))}
							{form.questions.length === 0 && (
								<p className="text-xs text-zinc-400 italic text-center py-4">
									No pre-flight questions — click "Add Question" above.
								</p>
							)}
						</div>
					</SectionCard>

					{/* Welcome Message */}
					<SectionCard
						title="Channel Welcome Message"
						description="Sleek embed sent inside the ticket channel when it's opened."
					>
						<div className="space-y-4">
							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Embed Welcome Title">
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
								<FormLabel label="Embed color border">
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

							<FormLabel label="Embed Description text">
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

							<FormLabel label="Title Redirect URL" hint="optional">
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

							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Embed Large image link" hint="optional">
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
								<FormLabel label="Embed Small thumbnail" hint="optional">
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

							<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
								<FormLabel label="Embed Footer Label" hint="optional">
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
								<FormLabel label="Embed Footer Icon link" hint="optional">
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
		</form>
	);
}
