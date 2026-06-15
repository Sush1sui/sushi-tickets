"use client";

import Link from "next/link";
import { useParams, usePathname, useRouter } from "next/navigation";
import useAuth from "../../../lib/context/auth";
import { useServers } from "../../../lib/hooks/useServers";
import LoadingScreen from "../../../components/LoadingScreen";
import {
	Settings,
	LayoutGrid,
	ScrollText,
	Users,
	ChevronLeft,
} from "lucide-react";
import Image from "next/image";

const navItems = [
	{ href: "", label: "Settings", icon: Settings, description: "Ticket configuration" },
	{ href: "/panels", label: "Ticket Panels", icon: LayoutGrid, description: "Manage panels" },
	{ href: "/transcripts", label: "Transcripts", icon: ScrollText, description: "View history" },
	{ href: "/staffs", label: "Staff Members", icon: Users, description: "Access control" },
];

export default function ServerLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const params = useParams();
	const pathname = usePathname();
	const router = useRouter();
	const { user, authLoading, logout } = useAuth();
	const { servers, isLoading: serversLoading } = useServers();

	const serverId = params.serverId as string;
	const server = servers.find((s) => s.id === serverId);

	if (authLoading || serversLoading) {
		return <LoadingScreen />;
	}

	if (!user || !server) {
		router.push("/servers");
		return null;
	}

	return (
		<div className="fixed inset-0 flex bg-[#1E1F22] text-zinc-100 font-sans selection:bg-[#FF5A36]/30 overflow-hidden">
			{/* ── Sidebar ── */}
			<aside className="relative z-10 h-full w-66 shrink-0 flex flex-col border-r border-white/5 bg-[#2B2D31] overflow-y-auto shadow-2xl shadow-zinc-950/50">
				{/* Brand */}
				<div className="px-5 py-5.5 border-b border-white/5">
					<Link href="/" className="flex items-center gap-2.5 group">
						<img
							src="/assets/sushi_logo_without_bg.png"
							alt="Sushi Logo"
							className="h-7 w-7 object-contain filter drop-shadow-[0_0_8px_rgba(255,90,54,0.5)] select-none group-hover:scale-105 group-hover:rotate-6 transition-all duration-300"
						/>
						<div className="flex flex-col">
							<span className="text-sm font-black tracking-tight bg-gradient-to-r from-white via-zinc-200 to-zinc-400 bg-clip-text text-transparent uppercase font-mono leading-none">
								Sushi Tickets
							</span>
							<span className="text-[9px] tracking-widest uppercase font-black text-[#FF5A36] text-glow-sushi leading-none mt-0.5">
								Premium Support
							</span>
						</div>
					</Link>
				</div>

				{/* Back to servers */}
				<div className="px-3.5 pt-4.5 pb-2.5">
					<Link
						href="/servers"
						className="flex items-center gap-2 px-3.5 py-2.5 rounded-xl text-xs font-bold text-zinc-400 hover:text-zinc-200 hover:bg-white/5 transition-all group shadow-sm"
					>
						<ChevronLeft className="h-4 w-4 group-hover:-translate-x-0.5 transition-transform text-zinc-500 group-hover:text-zinc-300" />
						All Servers
					</Link>
				</div>

				{/* Server identity */}
				<div className="mx-3.5 mb-5 rounded-2xl border border-white/5 bg-white/2 p-3.5 flex items-center gap-3 shadow-inner">
					{server.iconUrl ? (
						<img
							src={server.iconUrl}
							alt={server.name}
							className="h-9 w-9 rounded-full ring-2 ring-white/10 shrink-0"
						/>
					) : (
						<div className="h-9 w-9 rounded-full bg-zinc-950/60 border border-white/10 ring-2 ring-white/5 group-hover:ring-[#FF5A36]/50 flex items-center justify-center text-sm font-black text-zinc-200 select-none shrink-0">
							{server.name[0]?.toUpperCase()}
						</div>
					)}
					<div className="min-w-0">
						<p className="text-sm font-extrabold text-zinc-100 truncate tracking-tight">{server.name}</p>
					</div>
				</div>

				{/* Nav items */}
				<nav className="flex-1 px-3.5 space-y-1">
					<p className="text-[10px] font-extrabold text-zinc-400 uppercase tracking-widest px-3.5 mb-2.5">
						Dashboard
					</p>
					{navItems.map((item) => {
						const href = `/servers/${serverId}${item.href}`;
						const isActive =
							item.href === ""
								? pathname === `/servers/${serverId}`
								: pathname.startsWith(href);
						const Icon = item.icon;

						return (
							<Link
								key={href}
								href={href}
								className={`flex items-center gap-3 px-3.5 py-3 rounded-2xl text-xs font-bold transition-all duration-300 group ${isActive
									? "bg-[#FF5A36]/15 text-white border border-[#FF5A36]/30 shadow-lg shadow-orange-950/20"
									: "text-zinc-400 hover:text-zinc-100 hover:bg-white/5 border border-transparent"
									}`}
							>
								<Icon
									className={`h-4 w-4 shrink-0 ${isActive ? "text-[#FF5A36] drop-shadow-[0_0_4px_#FF5A36]" : "group-hover:text-zinc-200"
										}`}
								/>
								<div className="flex-1 min-w-0">
									<p className={isActive ? "text-white" : ""}>{item.label}</p>
									<p className={`text-[10px] font-medium mt-0.5 truncate ${isActive ? "text-[#FF5A36]" : "text-zinc-400 group-hover:text-zinc-300"}`}>
										{item.description}
									</p>
								</div>
								{isActive && (
									<div className="ml-auto h-1.5 w-1.5 rounded-full bg-[#FF5A36]/90 shadow-[0_0_6px_rgba(255,90,54,0.6)] shrink-0" />
								)}
							</Link>
						);
					})}
				</nav>

				{/* User profile at bottom */}
				<div className="p-3.5 border-t border-white/5">
					<div className="flex items-center gap-2.5 px-3.5 py-3 rounded-2xl bg-white/5 border border-white/5 shadow-inner">
						{user.image ? (
							<Image
								src={user.image}
								alt={user.name}
								width={28}
								height={28}
								className="rounded-full ring-2 ring-white/10 shrink-0"
							/>
						) : (
							<div className="h-7 w-7 rounded-full bg-zinc-900 ring-2 ring-white/5 flex items-center justify-center text-xs font-black text-zinc-300 select-none shrink-0">
								{user.name[0]?.toUpperCase()}
							</div>
						)}
						<div className="flex-1 min-w-0">
							<p className="text-xs font-bold text-zinc-200 truncate">{user.name}</p>
						</div>
						<button
							onClick={logout}
							className="text-zinc-400 hover:text-[#FF5A36] transition-colors p-1.5 rounded-lg hover:bg-white/5"
							title="Logout"
						>
							<svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
								<path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
						</button>
					</div>
				</div>
			</aside>

			{/* ── Main Content ── */}
			<div className="relative z-10 flex-1 h-full overflow-y-auto flex flex-col bg-[#1E1F22]">
				{/* Top bar */}
				<div className="sticky top-0 z-40 border-b border-white/5 bg-[#1E1F22]/85 backdrop-blur-xl px-8 py-4.5 flex items-center justify-between shrink-0">
					<div>
						{(() => {
							const current = navItems.find((item) => {
								const href = `/servers/${serverId}${item.href}`;
								return item.href === ""
									? pathname === `/servers/${serverId}`
									: pathname.startsWith(href);
							});
							return (
								<>
									<h1 className="text-base font-extrabold text-white tracking-tight uppercase text-glow-sushi/10">
										{current?.label ?? "Dashboard"}
									</h1>
									<p className="text-xs text-zinc-300 font-semibold mt-0.5">{current?.description}</p>
								</>
							);
						})()}
					</div>
				</div>

				{/* Page content */}
				<main className="flex-1 px-8 py-8">
					{children}
				</main>

				{/* Footer */}
				<footer className="px-8 py-5 border-t border-white/5 shrink-0 bg-[#1E1F22]/60">
					<p className="text-xs text-zinc-400 font-medium">
						© 2026 Sushi Tickets
					</p>
				</footer>
			</div>
		</div>
	);
}
