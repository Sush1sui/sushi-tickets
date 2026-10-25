export default function InlineError({
	message,
	onRetry,
}: {
	message: string;
	onRetry?: () => void;
}) {
	return (
		<div className="rounded-lg border border-red-500/20 bg-red-950/20 p-6 text-center">
			<p className="text-sm text-red-400 font-medium leading-relaxed">{message}</p>
			{onRetry && (
				<button
					type="button"
					onClick={onRetry}
					className="mt-4 inline-flex items-center gap-1.5 rounded-md border border-white/10 bg-white/5 px-3.5 py-2 text-[13px] font-medium text-zinc-300 transition-colors hover:border-white/20 hover:bg-white/10 hover:text-white"
				>
					Try again
				</button>
			)}
		</div>
	);
}
