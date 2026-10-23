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
import { useEffect, useState } from "react";
import { Menu, X, ChevronRight } from "lucide-react";

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
	const [menuOpen, setMenuOpen] = useState(false);
	useEffect(() => {
		if (!menuOpen) return;
		document.querySelector<HTMLAnchorElement>(".workspace-sidebar nav a")?.focus();
		const closeOnEscape = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				setMenuOpen(false);
				document.querySelector<HTMLButtonElement>(".workspace-menu-toggle")?.focus();
			}
		};
		window.addEventListener("keydown", closeOnEscape);
		return () => window.removeEventListener("keydown", closeOnEscape);
	}, [menuOpen]);

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
		<div className="workspace">
			{/* ── Sidebar ── */}
			<aside className={`workspace-sidebar ${menuOpen ? "is-open" : ""}`}>
				{/* Brand */}
				<div className="px-5 py-5.5 border-b border-white/5">
					<Link href="/" onClick={() => setMenuOpen(false)} className="flex items-center gap-2.5 group">
						<img
							src="/assets/sushi_logo_without_bg.png"
							alt="Sushi Logo"
							className="h-7 w-7 object-contain filter drop-shadow-[0_0_8px_rgba(255,90,54,0.5)] select-none group-hover:scale-105 group-hover:rotate-6 transition-all duration-300"
						/>
						<div className="flex flex-col">
							<span className={"text-sm font-black tracking-tight bg-gradient-to-r from-white via-zinc-200 to-zinc-400 bg-clip-text text-transparent uppercase font-mono leading-none" /* taste-skill-ignore: Existing brand wordmark treatment is protected. */}>
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
						onClick={() => setMenuOpen(false)}
						className="workspace-back"
					>
						<ChevronLeft className="h-4 w-4 group-hover:-translate-x-0.5 transition-transform text-zinc-400 group-hover:text-zinc-300" />
						All Servers
					</Link>
				</div>

				{/* Server identity */}
				<div className="workspace-server">
					{server.iconUrl ? (
						<img
							src={server.iconUrl}
							alt={server.name}
							className="h-9 w-9 rounded-full ring-2 ring-white/10 shrink-0"
						/>
					) : (
						<div className="h-9 w-9 rounded-full bg-zinc-950/60 border border-white/10 ring-2 ring-white/5 group-hover:ring-[#FF5A36]/50 flex items-center justify-center text-sm font-medium text-zinc-200 select-none shrink-0">
							{server.name[0]?.toUpperCase()}
						</div>
					)}
					<div className="min-w-0">
						<p className="text-sm font-medium text-zinc-100 truncate tracking-tight">{server.name}</p>
					</div>
				</div>

				{/* Nav items */}
				<nav aria-label="Server navigation" className="workspace-nav">
					<p className="text-sm text-zinc-400 px-3 mb-3">
						Dashboard
					</p>
					{navItems.map((item) => {
						const href = `/servers/${serverId}${item.href}`;
						const isActive =
							item.href === ""
								? pathname === `/servers/${serverId}`
								: pathname.startsWith(href) || (item.href === "/panels" && pathname.includes("/multi-panels"));
						const Icon = item.icon;

						return (
							<Link
								key={href}
								href={href}
								onClick={() => setMenuOpen(false)}
								aria-current={isActive ? "page" : undefined}
								className={`workspace-nav-link ${isActive ? "is-active" : ""}`}
							>
								<Icon
									className="h-4 w-4 shrink-0"
								/>
								<div className="flex-1 min-w-0">
									<p className={isActive ? "text-white" : ""}>{item.label}</p>
									<p className="text-[13px] text-zinc-400 mt-1 truncate">
										{item.description}
									</p>
								</div>
								{isActive && (
									<ChevronRight className="h-3.5 w-3.5 shrink-0 text-zinc-400" />
								)}
							</Link>
						);
					})}
				</nav>

				{/* User profile at bottom */}
				<div className="p-3.5 border-t border-white/5">
					<div className="flex items-center gap-2.5 px-2 py-2">
						{user.image ? (
							<Image
								src={user.image}
								alt={user.name}
								width={28}
								height={28}
								className="rounded-full ring-2 ring-white/10 shrink-0"
							/>
						) : (
							<div className="h-7 w-7 rounded-full bg-zinc-900 ring-2 ring-white/5 flex items-center justify-center text-[13px] font-medium text-zinc-300 select-none shrink-0">
								{user.name[0]?.toUpperCase()}
							</div>
						)}
						<div className="flex-1 min-w-0">
							<p className="text-[13px] font-medium text-zinc-200 truncate">{user.name}</p>
						</div>
						<button
							onClick={logout}
							className="text-zinc-400 hover:text-[#FF5A36] transition-colors p-1.5 rounded-lg hover:bg-white/5"
							title="Logout"
							aria-label="Logout"
						>
							<svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
								<path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
							</svg>
						</button>
					</div>
				</div>
			</aside>

			{/* ── Main Content ── */}
			<div className="workspace-content">
				{/* Top bar */}
				<div className="workspace-topbar">
					<button className="workspace-menu-toggle" type="button" aria-expanded={menuOpen} aria-label={menuOpen ? "Close navigation" : "Open navigation"} onClick={() => setMenuOpen(!menuOpen)}>
						{menuOpen ? <X size={20} /> : <Menu size={20} />}
					</button>
					<div className="flex min-w-0 items-center gap-3 text-sm">
						<span className="truncate text-zinc-400">{server.name}</span>
						<ChevronRight className="h-3.5 w-3.5 shrink-0 text-zinc-400" />
						{(() => {
							const current = navItems.find((item) => {
								const href = `/servers/${serverId}${item.href}`;
								return item.href === ""
									? pathname === `/servers/${serverId}`
									: pathname.startsWith(href) || (item.href === "/panels" && pathname.includes("/multi-panels"));
							});
							return (
								<>
									<p className="shrink-0 font-medium text-zinc-100">
										{current?.label ?? "Dashboard"}
									</p>
								</>
							);
						})()}
					</div>
				</div>

				{/* Page content */}
				<main className="workspace-main" inert={menuOpen ? true : undefined}>
					{children}
				</main>

				{/* Footer */}
				<footer className="workspace-footer" inert={menuOpen ? true : undefined}>
					<p className="text-[13px] text-zinc-400 font-medium">
						© 2026 Sushi Tickets
					</p>
				</footer>
			</div>
		</div>
	);
}
