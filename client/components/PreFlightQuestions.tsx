"use client";

import { useMemo } from "react";
import { ArrowDown, ArrowUp, Plus, Trash2, X } from "lucide-react";
import type { QuestionItem } from "../lib/api";
import { DarkInput, DarkSwitchTrack, SectionCard } from "./DarkFormFields";

const MAX_QUESTIONS = 10;
const DISCORD_VISIBLE_QUESTIONS = 5;
const LABEL_MAX = 45;
const PLACEHOLDER_MAX = 100;
const MODAL_TITLE_MAX = 45;
const DEFAULT_MODAL_TITLE = "Sushi Ticket Questions";

const blankQuestion = (): QuestionItem => ({
	label: "",
	isRequired: false,
	style: "short",
	placeholder: "",
});

function CharacterCount({ value, max }: { value: number; max: number }) {
	const over = value > max;
	return (
		<span
			className={`text-[12px] tabular-nums ${over ? "font-semibold text-red-400" : "text-zinc-400"}`}
			aria-live="polite"
		>
			{value}/{max}
		</span>
	);
}

function QuestionPreviewInput({ question }: { question: QuestionItem }) {
	const isParagraph = question.style === "paragraph";
	return (
		<div
			className={`rounded-[3px] border border-black/20 bg-[#1e1f22] px-3 text-[13px] leading-5 text-[#87898c] ${
				isParagraph ? "h-20 py-2" : "flex h-10 items-center"
			}`}
		>
			<span className={`block truncate ${isParagraph ? "whitespace-normal" : ""}`}>
				{question.placeholder?.trim() || "\u00a0"}
			</span>
		</div>
	);
}

export function PreFlightPreview({
	title,
	questions,
}: {
	title: string;
	questions: QuestionItem[];
}) {
	const complete = useMemo(
		() => questions.filter((question) => question.label.trim() !== ""),
		[questions],
	);
	const visible = complete.slice(0, DISCORD_VISIBLE_QUESTIONS);
	const resolvedTitle = title.trim() || DEFAULT_MODAL_TITLE;

	return (
		<div className="w-full max-w-[400px]">
			<div
				data-ui
				aria-hidden="true"
				className="overflow-hidden rounded-lg bg-[#313338] text-left"
			>
				<div className="flex items-center justify-between gap-4 border-b border-white/5 px-4 py-3">
					<p className="truncate text-[15px] font-semibold text-[#f2f3f5]">
						{resolvedTitle}
					</p>
					<X className="h-4 w-4 shrink-0 text-[#b5bac1]" />
				</div>

				<div className="space-y-3 px-4 py-4">
					{visible.length === 0 ? (
						<p className="py-8 text-center text-[13px] leading-5 text-[#949ba4]">
							No questions. The popup is skipped and tickets open directly.
						</p>
					) : (
						visible.map((question, index) => (
							<div key={`${question.label}-${index}`} className="space-y-1.5">
								<div className="flex items-center gap-1 text-[12px] font-semibold text-[#b5bac1]">
									<span className="truncate">
										{question.label.trim().slice(0, LABEL_MAX)}
									</span>
									{question.isRequired && (
										<span className="text-[#f23f43]" aria-hidden="true">
											*
										</span>
									)}
								</div>
								<QuestionPreviewInput question={question} />
							</div>
						))
					)}
				</div>

				{visible.length > 0 && (
					<div className="flex justify-end border-t border-white/5 bg-[#2b2d31] px-4 py-3">
						<span className="rounded-[3px] bg-[#5865f2] px-4 py-2 text-[13px] font-medium text-white">
							Submit
						</span>
					</div>
				)}
			</div>

			<p className="mt-2 text-center text-[12px] leading-5 text-zinc-400">
				{visible.length === 0
					? "Tickets open without a popup."
					: complete.length > DISCORD_VISIBLE_QUESTIONS
						? `Showing the first ${DISCORD_VISIBLE_QUESTIONS} of ${complete.length} questions. Discord hides the rest.`
						: "This is what users see before their ticket is created."}
			</p>
		</div>
	);
}

export default function PreFlightQuestions({
	value,
	onChange,
	modalTitle,
	onModalTitleChange,
}: {
	value: QuestionItem[];
	onChange: (next: QuestionItem[]) => void;
	modalTitle: string;
	onModalTitleChange: (next: string) => void;
}) {
	const completeCount = useMemo(
		() => value.filter((question) => question.label.trim() !== "").length,
		[value],
	);
	const incompleteCount = value.length - completeCount;
	const hiddenCount = Math.max(0, completeCount - DISCORD_VISIBLE_QUESTIONS);
	const atLimit = value.length >= MAX_QUESTIONS;

	const updateQuestion = (index: number, patch: Partial<QuestionItem>) => {
		onChange(
			value.map((question, currentIndex) =>
				currentIndex === index ? { ...question, ...patch } : question,
			),
		);
	};

	const moveQuestion = (index: number, direction: -1 | 1) => {
		const target = index + direction;
		if (target < 0 || target >= value.length) return;
		const next = [...value];
		[next[index], next[target]] = [next[target], next[index]];
		onChange(next);
	};

	const addQuestion = () => {
		if (atLimit) return;
		onChange([...value, blankQuestion()]);
		window.requestAnimationFrame(() => {
			document.getElementById(`question-label-${value.length}`)?.focus();
		});
	};

	const removeQuestion = (index: number) => {
		onChange(value.filter((_, currentIndex) => currentIndex !== index));
	};

	return (
		<SectionCard
			title="Pre-Flight Questions"
			description="Ask for the details you need before a ticket channel is created."
			action={
				<button
					type="button"
					onClick={addQuestion}
					disabled={atLimit}
					title={atLimit ? `Max ${MAX_QUESTIONS} questions` : "Add another question"}
					className="flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
				>
					<Plus className="h-3.5 w-3.5" />
					Add question
				</button>
			}
		>
			<div className="space-y-5">
				<div className="max-w-xl">
					<div className="mb-2 flex items-baseline justify-between gap-3">
						<span className="text-[13px] font-medium text-[#dbdee1]">
							Modal window title
						</span>
						<CharacterCount value={modalTitle.length} max={MODAL_TITLE_MAX} />
					</div>
					<DarkInput
						value={modalTitle}
						onChange={(event) => onModalTitleChange(event.target.value)}
						placeholder={DEFAULT_MODAL_TITLE}
						maxLength={MODAL_TITLE_MAX}
						ariaLabel="Modal window title"
					/>
					<p className="mt-1.5 text-[12px] leading-5 text-zinc-400">
						The header shown at the top of the Discord popup.
					</p>
				</div>

				<div className="flex flex-wrap items-center gap-x-2 gap-y-1 border-t border-white/5 pt-4 text-[13px]">
					<span className="font-medium text-zinc-200">
						{completeCount === 0
							? "No questions yet"
							: `${completeCount} question${completeCount === 1 ? "" : "s"}`}
					</span>
					{incompleteCount > 0 && (
						<span className="text-zinc-400">
							· {incompleteCount} without a label will be skipped
						</span>
					)}
					{hiddenCount > 0 && (
						<span className="text-[#fbbf24]">
							· only the first {DISCORD_VISIBLE_QUESTIONS} appear in Discord
						</span>
					)}
					{completeCount === 0 && incompleteCount === 0 && (
						<span className="text-zinc-400">· popup is skipped</span>
					)}
				</div>

				{hiddenCount > 0 && (
					<div
						role="status"
						className="rounded-md border border-yellow-400/20 bg-yellow-400/5 px-3.5 py-2.5 text-[13px] leading-5 text-yellow-100"
					>
						Only the first {DISCORD_VISIBLE_QUESTIONS} questions appear in Discord.
						Questions {DISCORD_VISIBLE_QUESTIONS + 1} and up are saved but hidden. Use
						the arrow buttons to choose which ones come first.
					</div>
				)}

				{value.length === 0 ? (
					<div className="rounded-lg border border-dashed border-white/10 bg-white/[0.02] px-5 py-8 text-center">
						<p className="text-sm font-medium text-zinc-200">No pre-flight questions</p>
						<p className="mx-auto mt-1 max-w-sm text-[13px] leading-5 text-zinc-400">
							Tickets open immediately. Add a question to collect details in a Discord
							popup first.
						</p>
						<button
							type="button"
							onClick={addQuestion}
							className="mt-4 inline-flex items-center gap-1.5 rounded-md bg-zinc-100 px-3.5 py-2 text-[13px] font-medium text-zinc-950 transition-colors hover:bg-white"
						>
							<Plus className="h-3.5 w-3.5" />
							Add question
						</button>
					</div>
				) : (
					<div className="space-y-2">
						{value.map((question, index) => {
							const missingLabel = question.label.trim() === "";
							return (
								<div
									key={index}
									className={`rounded-lg border p-4 transition-colors ${
										missingLabel
											? "border-yellow-400/20 bg-yellow-400/[0.03]"
											: "border-white/5 bg-white/[0.02] hover:border-white/10"
									}`}
								>
									<div className="flex items-center justify-between gap-3">
										<div className="flex min-w-0 items-center gap-2">
											<span className="grid h-6 w-6 shrink-0 place-items-center rounded-md bg-white/5 text-[12px] font-semibold tabular-nums text-zinc-300">
												{index + 1}
											</span>
											<span className="truncate text-[13px] font-medium text-zinc-200">
												{question.label.trim() || "Untitled question"}
											</span>
										</div>
										<div className="flex shrink-0 items-center gap-0.5">
											<button
												type="button"
												onClick={() => moveQuestion(index, -1)}
												disabled={index === 0}
												aria-label={`Move question ${index + 1} up`}
												className="grid h-8 w-8 place-items-center rounded-md text-zinc-400 transition-colors hover:bg-white/5 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:bg-transparent"
											>
												<ArrowUp className="h-4 w-4" />
											</button>
											<button
												type="button"
												onClick={() => moveQuestion(index, 1)}
												disabled={index === value.length - 1}
												aria-label={`Move question ${index + 1} down`}
												className="grid h-8 w-8 place-items-center rounded-md text-zinc-400 transition-colors hover:bg-white/5 hover:text-zinc-100 disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:bg-transparent"
											>
												<ArrowDown className="h-4 w-4" />
											</button>
											<button
												type="button"
												onClick={() => removeQuestion(index)}
												aria-label={`Remove question ${index + 1}`}
												className="grid h-8 w-8 place-items-center rounded-md text-zinc-400 transition-colors hover:bg-white/5 hover:text-[#ff8c72]"
											>
												<Trash2 className="h-4 w-4" />
											</button>
										</div>
									</div>

									<div className="mt-3 space-y-3">
										<div>
											<div className="mb-1.5 flex items-baseline justify-between gap-3">
												<span className="text-[13px] font-medium text-[#dbdee1]">
													Question label
												</span>
												<CharacterCount value={question.label.length} max={LABEL_MAX} />
											</div>
											<DarkInput
												id={`question-label-${index}`}
												value={question.label}
												onChange={(event) =>
													updateQuestion(index, { label: event.target.value })
												}
												placeholder={`Question ${index + 1} (e.g. Roblox username)`}
												maxLength={LABEL_MAX}
												ariaLabel={`Question ${index + 1} label`}
											/>
											<p className="mt-1.5 text-[12px] leading-5 text-zinc-400">
												{missingLabel
													? "Add a label. Questions without one are skipped when saving."
													: "Discord cuts labels after 45 characters."}
											</p>
										</div>

										<div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-3 border-t border-white/5 pt-3">
											<div
												role="group"
												aria-label={`Question ${index + 1} input style`}
												className="inline-flex items-center rounded-md border border-white/10 bg-[#1e1f22] p-0.5"
											>
												{(
													[
														{ value: "short", label: "Single line" },
														{ value: "paragraph", label: "Text area" },
													] as const
												).map((option) => {
													const selected =
														(question.style || "short") === option.value;
													return (
														<button
															key={option.value}
															type="button"
															aria-pressed={selected}
															onClick={() =>
																updateQuestion(index, { style: option.value })
															}
															className={`rounded px-2.5 py-1 text-[12px] font-medium transition-colors ${
																selected
																	? "bg-zinc-100 text-zinc-950"
																	: "text-zinc-400 hover:text-zinc-200"
															}`}
														>
															{option.label}
														</button>
													);
												})}
											</div>

											<button
												type="button"
												role="switch"
												aria-checked={question.isRequired}
												aria-label={`Question ${index + 1} required`}
												onClick={() =>
													updateQuestion(index, {
														isRequired: !question.isRequired,
													})
												}
												className="ml-auto inline-flex items-center gap-2.5 rounded-md px-1.5 py-1 text-[13px] leading-none text-zinc-300 transition-colors hover:text-zinc-100"
											>
												<DarkSwitchTrack checked={question.isRequired} />
												<span className="leading-none">
													{question.isRequired ? "Required" : "Optional"}
												</span>
											</button>
										</div>

										<div>
											<div className="mb-1.5 flex items-baseline justify-between gap-3">
												<span className="text-[13px] font-medium text-[#dbdee1]">
													Placeholder
													<span className="ml-1.5 text-[12px] font-normal text-zinc-400">
														optional
													</span>
												</span>
												<CharacterCount
													value={(question.placeholder || "").length}
													max={PLACEHOLDER_MAX}
												/>
											</div>
											<DarkInput
												value={question.placeholder || ""}
												onChange={(event) =>
													updateQuestion(index, { placeholder: event.target.value })
												}
												placeholder="Hint shown inside the Discord input (e.g. Enter your Roblox username)"
												maxLength={PLACEHOLDER_MAX}
												ariaLabel={`Question ${index + 1} placeholder`}
												className="bg-[#17181a] py-2 text-[13px]"
											/>
										</div>
									</div>
								</div>
							);
						})}
					</div>
				)}

				{value.length > 0 && (
					<button
						type="button"
						onClick={addQuestion}
						disabled={atLimit}
						className="flex w-full items-center justify-center gap-1.5 rounded-md border border-dashed border-white/10 py-2.5 text-[13px] font-medium text-zinc-400 transition-colors hover:border-white/20 hover:text-zinc-200 disabled:cursor-not-allowed disabled:opacity-40"
					>
						<Plus className="h-3.5 w-3.5" />
						{atLimit ? `Max ${MAX_QUESTIONS} questions` : "Add another question"}
					</button>
				)}
			</div>
		</SectionCard>
	);
}
