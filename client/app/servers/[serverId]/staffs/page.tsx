"use client";

import { useParams, useRouter } from "next/navigation";
import { useState, useMemo, useCallback } from "react";
import { api, type StaffMember, type StaffRole } from "../../../../lib/api";
import useAuth from "../../../../lib/context/auth";
import { useStaff } from "../../../../lib/hooks/useStaff";
import { Save, Users, Shield } from "lucide-react";
import InlineError from "../../../../components/InlineError";
import {
	DarkSearchDropdown,
	DarkChipList,
	SectionCard,
	type DropdownItem,
} from "../../../../components/DarkFormFields";

export default function StaffPage() {
	const params = useParams();
	const router = useRouter();
	const { user, authLoading } = useAuth();
	const serverId = params.serverId as string;

	const {
		members: allMembers,
		roles: allRoles,
		authorizedMemberIds,
		authorizedRoleIds,
		isLoading,
		isError,
		refresh,
	} = useStaff(serverId);

	const [saving, setSaving] = useState(false);
	const [saved, setSaved] = useState(false);
	const [selectedMemberIds, setSelectedMemberIds] = useState<string[]>([]);
	const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);
	const [initialized, setInitialized] = useState(false);

	const initSelections = useCallback(() => {
		if (!initialized) {
			setSelectedMemberIds([...authorizedMemberIds]);
			setSelectedRoleIds([...authorizedRoleIds]);
			setInitialized(true);
		}
	}, [initialized, authorizedMemberIds, authorizedRoleIds]);

	if (!initialized && !isLoading) {
		initSelections();
	}

	const authorizedMembers = useMemo(
		() =>
			selectedMemberIds
				.map((id) => allMembers.find((m: StaffMember) => m.id === id))
				.filter((m): m is StaffMember => m !== undefined)
				.map((m: StaffMember): DropdownItem => ({
					id: m.id,
					label: m.globalName || m.username,
					sublabel: m.globalName ? m.username : undefined,
					avatarUrl: m.avatarUrl || undefined,
				})),
		[selectedMemberIds, allMembers],
	);

	const authorizedRoles = useMemo(
		() =>
			selectedRoleIds
				.map((id) => allRoles.find((r: StaffRole) => r.id === id))
				.filter((r): r is StaffRole => r !== undefined)
				.map((r: StaffRole): DropdownItem => ({
					id: r.id,
					label: r.name,
					color: r.color,
				})),
		[selectedRoleIds, allRoles],
	);

	const allMemberItems = useMemo(
		() =>
			allMembers.map((m: StaffMember): DropdownItem => ({
				id: m.id,
				label: m.globalName || m.username,
				sublabel: m.globalName ? m.username : undefined,
				avatarUrl: m.avatarUrl || undefined,
			})),
		[allMembers],
	);

	const allRoleItems = useMemo(
		() =>
			allRoles.map((r: StaffRole): DropdownItem => ({
				id: r.id,
				label: r.name,
				color: r.color,
			})),
		[allRoles],
	);

	const selectedMemberIdSet = useMemo(() => new Set(selectedMemberIds), [selectedMemberIds]);
	const selectedRoleIdSet = useMemo(() => new Set(selectedRoleIds), [selectedRoleIds]);

	const addMember = useCallback(
		(id: string) => setSelectedMemberIds((prev) => (prev.includes(id) ? prev : [...prev, id])),
		[],
	);
	const removeMember = useCallback(
		(id: string) => setSelectedMemberIds((prev) => prev.filter((x) => x !== id)),
		[],
	);
	const addRole = useCallback(
		(id: string) => setSelectedRoleIds((prev) => (prev.includes(id) ? prev : [...prev, id])),
		[],
	);
	const removeRole = useCallback(
		(id: string) => setSelectedRoleIds((prev) => prev.filter((x) => x !== id)),
		[],
	);

	const handleSave = async () => {
		setSaving(true);
		try {
			await api.staff.update(serverId, {
				authorizedMemberIds: selectedMemberIds,
				authorizedRoleIds: selectedRoleIds,
			});
			refresh();
			setSaved(true);
			setTimeout(() => setSaved(false), 2500);
		} catch {
			alert("Failed to save staff settings");
		} finally {
			setSaving(false);
		}
	};

	if (authLoading || isLoading) {
		return (
			<div className="flex items-center justify-center py-20">
				<div className="flex flex-col items-center gap-3">
					<div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
					<p className="text-[13px] text-zinc-400">Loading staff...</p>
				</div>
			</div>
		);
	}

	if (isError) {
		return (
			<InlineError
				message="We couldn't load this server's staff settings. You may no longer have access, or it may be temporarily unavailable."
				onRetry={refresh}
			/>
		);
	}

	if (!user) {
		router.push("/servers");
		return null;
	}

	return (
		<div className="space-y-5 pb-6">
			{/* Header */}
			<div className="page-heading">
				<div>
					<h1 className="text-xl font-medium tracking-tight text-white ">Staff Members</h1>
					<p className="text-[13px] text-zinc-300 font-semibold mt-1">
						Control who can manage tickets in this server.
					</p>
				</div>
				<button
					onClick={handleSave}
					disabled={saving}
					className={`flex items-center gap-2 rounded-xl px-5 py-2.5 text-xs font-bold transition-colors duration-200 active:scale-95 disabled:opacity-60 shrink-0 cursor-pointer ${
						saved
							? "bg-emerald-500/20 border border-emerald-500/40 text-emerald-400"
							: "bg-zinc-100 hover:bg-white text-zinc-950   hover:-translate-y-0.5"
					}`}
				>
					<Save className="h-4 w-4" />
					{saving ? "Saving..." : saved ? "Saved!" : "Save Changes"}
				</button>
			</div>

			{/* Grid Layout for Members & Roles to prevent wasted horizontal space */}
			<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
				{/* Members */}
				<SectionCard
					title="Authorized Members"
					description="Individual users who can manage tickets."
					action={
						<span className="rounded-full border border-zinc-800/60 bg-zinc-900/40 px-2.5 py-1 text-[13px] font-semibold text-zinc-400">
							<Users className="inline h-3 w-3 mr-1 -mt-0.5" />
							{selectedMemberIds.length}
						</span>
					}
				>
					<div className="space-y-3">
						<DarkSearchDropdown
							items={allMemberItems}
							selectedIds={selectedMemberIdSet}
							onSelect={addMember}
							placeholder="Search and add members..."
						/>
						<DarkChipList
							items={authorizedMembers}
							onRemove={removeMember}
							emptyMessage="No authorized members. Add one above."
						/>
					</div>
				</SectionCard>

				{/* Roles */}
				<SectionCard
					title="Authorized Roles"
					description="Anyone with these roles can manage tickets."
					action={
						<span className="rounded-full border border-zinc-800/60 bg-zinc-900/40 px-2.5 py-1 text-[13px] font-semibold text-zinc-400">
							<Shield className="inline h-3 w-3 mr-1 -mt-0.5" />
							{selectedRoleIds.length}
						</span>
					}
				>
					<div className="space-y-3">
						<DarkSearchDropdown
							items={allRoleItems}
							selectedIds={selectedRoleIdSet}
							onSelect={addRole}
							placeholder="Search and add roles..."
						/>
						<DarkChipList
							items={authorizedRoles}
							onRemove={removeRole}
							emptyMessage="No authorized roles. Add one above."
						/>
					</div>
				</SectionCard>
			</div>
		</div>
	);
}
