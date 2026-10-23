"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api } from "../../../lib/api";
import useAuth from "../../../lib/context/auth";
import { useServerConfig } from "../../../lib/hooks/useServers";
import { Save } from "lucide-react";
import {
  DarkInput,
  DarkToggle,
  DarkCheckbox,
  DarkCombobox,
  SectionCard,
  FormLabel,
} from "../../../components/DarkFormFields";

type FormData = {
  TicketNameStyle: string;
  TicketTranscripts: string;
  MaxTicketsPerUser: string | number;
  TicketPermissionsAttachFiles: boolean;
  TicketPermissionsEmbedLinks: boolean;
  TicketPermissionsAddReactions: boolean;
  AutoClose: boolean;
  AutoCloseOnUserLeave: boolean;
  AutoCloseNoResponseDays: string;
  AutoCloseNoResponseHours: string;
  AutoCloseNoResponseMins: string;
  AutoCloseSinceLastMessageDays: string;
  AutoCloseSinceLastMessageHours: string;
  AutoCloseSinceLastMessageMins: string;
};

function TimeInput({
  days,
  hours,
  mins,
  onChangeDays,
  onChangeHours,
  onChangeMins,
}: {
  days: string;
  hours: string;
  mins: string;
  onChangeDays: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onChangeHours: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onChangeMins: (e: React.ChangeEvent<HTMLInputElement>) => void;
}) {
  const blockInvalid = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (["e", "E", "-", "+", "."].includes(e.key)) e.preventDefault();
  };
  return (
    <div className="flex items-center gap-3">
      {[
        { value: days, onChange: onChangeDays, label: "days" },
        { value: hours, onChange: onChangeHours, label: "hrs" },
        { value: mins, onChange: onChangeMins, label: "min" },
      ].map(({ value, onChange, label }) => (
        <div key={label} className="flex items-center gap-2">
          <DarkInput
            type="number"
            min="0"
            value={value}
            onChange={onChange}
            onKeyDown={blockInvalid}
            className="w-16 text-center"
          />
          <span className="text-[13px] text-zinc-400 w-6">{label}</span>
        </div>
      ))}
    </div>
  );
}

export default function ServerSettingsPage() {
  const params = useParams();
  const router = useRouter();
  const { user, authLoading } = useAuth();
  const { config, channels, isLoading } = useServerConfig(
    params.serverId as string,
    true,
  );
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const serverId = params.serverId as string;

  const [formData, setFormData] = useState<FormData>({
    TicketNameStyle: "number",
    TicketTranscripts: "",
    MaxTicketsPerUser: "1",
    TicketPermissionsAttachFiles: false,
    TicketPermissionsEmbedLinks: false,
    TicketPermissionsAddReactions: false,
    AutoClose: false,
    AutoCloseOnUserLeave: false,
    AutoCloseNoResponseDays: "0",
    AutoCloseNoResponseHours: "0",
    AutoCloseNoResponseMins: "0",
    AutoCloseSinceLastMessageDays: "0",
    AutoCloseSinceLastMessageHours: "0",
    AutoCloseSinceLastMessageMins: "0",
  });

  useEffect(() => {
    if (!config) return;
    const transcriptChannel = channels.find(
      (ch) => ch.id === config.TicketTranscriptCid,
    );
    const transcriptName = transcriptChannel
      ? transcriptChannel.name
      : config.TicketTranscriptCid || "";

    setFormData((prev) => ({
      ...prev,
      TicketNameStyle: config.TicketNameStyle || "number",
      TicketTranscripts: transcriptName,
      MaxTicketsPerUser: String(config.MaxTicketsPerUser ?? 1),
      TicketPermissionsAttachFiles:
        config.TicketPermissionsAttachFiles || false,
      TicketPermissionsEmbedLinks: config.TicketPermissionsEmbedLinks || false,
      TicketPermissionsAddReactions:
        config.TicketPermissionsAddReactions || false,
      AutoClose: config.AutoClose || false,
      AutoCloseOnUserLeave: config.AutoCloseOnUserLeave || false,
      AutoCloseNoResponseDays: String(config.AutoCloseNoResponseDays ?? 0),
      AutoCloseNoResponseHours: String(config.AutoCloseNoResponseHours ?? 0),
      AutoCloseNoResponseMins: String(config.AutoCloseNoResponseMins ?? 0),
      AutoCloseSinceLastMessageDays: String(
        config.AutoCloseSinceLastMessageDays ?? 0,
      ),
      AutoCloseSinceLastMessageHours: String(
        config.AutoCloseSinceLastMessageHours ?? 0,
      ),
      AutoCloseSinceLastMessageMins: String(
        config.AutoCloseSinceLastMessageMins ?? 0,
      ),
    }));
  }, [config, channels]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const transcriptChannelId =
        channels.find((ch) => ch.name === formData.TicketTranscripts)?.id ||
        formData.TicketTranscripts;

      const toInt = (v: string) => Math.max(0, parseInt(v, 10) || 0);

      await api.config.update(serverId, {
        ...formData,
        TicketTranscripts: transcriptChannelId,
        MaxTicketsPerUser:
          typeof formData.MaxTicketsPerUser === "string"
            ? toInt(formData.MaxTicketsPerUser)
            : formData.MaxTicketsPerUser,
        AutoCloseNoResponseDays: toInt(formData.AutoCloseNoResponseDays),
        AutoCloseNoResponseHours: toInt(formData.AutoCloseNoResponseHours),
        AutoCloseNoResponseMins: toInt(formData.AutoCloseNoResponseMins),
        AutoCloseSinceLastMessageDays: toInt(
          formData.AutoCloseSinceLastMessageDays,
        ),
        AutoCloseSinceLastMessageHours: toInt(
          formData.AutoCloseSinceLastMessageHours,
        ),
        AutoCloseSinceLastMessageMins: toInt(
          formData.AutoCloseSinceLastMessageMins,
        ),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 2500);
    } catch {
      alert("Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  const handleNumberInput = (field: keyof FormData, value: string) => {
    // Allow blank (so user can clear) and non-negative integers only
    if (value === "" || /^[0-9]+$/.test(value)) {
      setFormData((prev) => ({ ...prev, [field]: value }));
    }
  };

  if (authLoading || isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="flex flex-col items-center gap-3">
          <div className="h-6 w-6 rounded-full border-2 border-zinc-700 border-t-[#FF5A36] animate-spin" />
          <p className="text-[13px] text-zinc-400">Loading settings...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    router.push("/servers");
    return null;
  }

  return (
    <div className="server-settings space-y-6 pb-6">
      {/* Header */}
      <div className="page-heading">
        <div>
          <h1 className="text-2xl font-medium tracking-tight text-white ">
            Server Settings
          </h1>
          <p className="text-[13px] text-zinc-300 font-semibold mt-1">
            Configure how tickets work in this server.
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

      {/* Modern Split Two-Column Grid */}
      <div className="grid grid-cols-1 gap-6 items-start">

        {/* LEFT COLUMN: Core Customization & Limits (7/12 width) */}
        <div className="space-y-6">

          {/* Ticket Name Style */}
          <SectionCard
            title="Ticket Name Style"
            description="How ticket channels are named when opened in Discord."
          >
            <div className="flex gap-4">
              {[
                { value: "number", label: "By Number", example: "ticket-1" },
                { value: "name", label: "By Username", example: "ticket-sush1sui" },
              ].map(({ value, label, example }) => (
                <button
                  key={value}
                  type="button"
                  aria-pressed={formData.TicketNameStyle === value}
                  onClick={() =>
                    setFormData((p) => ({ ...p, TicketNameStyle: value }))
                  }
                  className={`flex-1 rounded-2xl border p-4.5 text-left transition-colors duration-200 active:scale-98 cursor-pointer ${
                    formData.TicketNameStyle === value
                      ? "border-[#FF5A36] bg-[#FF5A36]/10  "
                      : "border-white/5 bg-zinc-950/20 hover:border-white/10 hover:bg-zinc-950/40"
                  }`}
                >
                  <div className="flex items-center gap-2.5 mb-2">
                    <div
                      className={`h-4 w-4 rounded-full border-2 flex items-center justify-center shrink-0 ${
                        formData.TicketNameStyle === value
                          ? "border-[#FF5A36]"
                          : "border-zinc-500"
                      }`}
                    >
                      {formData.TicketNameStyle === value && (
                        <div className="h-2 w-2 rounded-full bg-[#FF5A36] " />
                      )}
                    </div>
                    <span className="text-sm font-medium text-zinc-100">
                      {label}
                    </span>
                  </div>
                  <p className="text-[13px] font-medium font-mono text-zinc-400 pl-6.5">
                    #{example}
                  </p>
                </button>
              ))}
            </div>
          </SectionCard>

          {/* Routing & Limits Card (Combined) */}
          <SectionCard
            title="Routing & Limits"
            description="Configure transcripts channel and user ticket frequencies."
          >
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <FormLabel label="Transcript Channel" hint="Leave blank to disable">
                <DarkCombobox
                  value={formData.TicketTranscripts}
                  onChange={(v) =>
                    setFormData((p) => ({ ...p, TicketTranscripts: v }))
                  }
                  options={channels
                    .filter((ch) => ch.type === 0)
                    .map((ch) => ({ value: ch.id, label: ch.name }))}
                  placeholder="Search for channel..."
                  className="w-full"
                />
              </FormLabel>

              <FormLabel label="Max Open Tickets">
                <div className="flex items-center gap-3">
                  <DarkInput
                    type="number"
                    min="1"
                    value={formData.MaxTicketsPerUser}
                    onChange={(e) =>
                      handleNumberInput("MaxTicketsPerUser", e.target.value)
                    }
                    onKeyDown={(e) => {
                      if (["e", "E", "-", "+", "."].includes(e.key))
                        e.preventDefault();
                    }}
                    className="w-24 text-center font-medium text-zinc-100"
                  />
                  <p className="text-[13px] text-zinc-400 font-medium leading-none">
                    ticket{formData.MaxTicketsPerUser !== 1 ? "s" : ""} per user
                  </p>
                </div>
              </FormLabel>
            </div>
          </SectionCard>

          {/* Ticket Permissions */}
          <SectionCard
            title="Ticket Permissions"
            description="Extra permissions granted to users inside their ticket channels."
          >
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {[
                { field: "TicketPermissionsAttachFiles" as const, label: "Attach Files", desc: "Allow files upload" },
                { field: "TicketPermissionsEmbedLinks" as const, label: "Embed Links", desc: "Allow hyperlink styling" },
                { field: "TicketPermissionsAddReactions" as const, label: "Add Reactions", desc: "Allow reaction emojis" },
              ].map(({ field, label, desc }) => (
                <div key={field} className="bg-zinc-950/20 hover:bg-zinc-950/40 p-4 rounded-lg border border-white/2 hover:border-white/5 transition-colors duration-200">
                  <DarkCheckbox
                    checked={formData[field]}
                    onChange={(v) =>
                      setFormData((p) => ({ ...p, [field]: v }))
                    }
                    label={label}
                    description={desc}
                  />
                </div>
              ))}
            </div>
          </SectionCard>

        </div>

        {/* RIGHT COLUMN: Auto-Close Automation (5/12 width) */}
        <div className="lg:col-span-5">

          <SectionCard
            title="Automation Center"
            description="Manage automatic ticket closing rules."
          >
            <div className="space-y-6">

              {/* Main Toggle Pane */}
              <div className="bg-zinc-950/30 border border-white/5 p-4.5 rounded-lg ">
                <DarkToggle
                  checked={formData.AutoClose}
                  onChange={(v) => setFormData((p) => ({ ...p, AutoClose: v }))}
                  label="Enable Auto-Close"
                  description="Close idle tickets automatically based on rules."
                />
              </div>

              {formData.AutoClose ? (
                <div className="space-y-5 pt-2 animate-in fade-in slide-in-from-top-2 duration-200">

                  {/* User leaves option */}
                  <div className="bg-zinc-950/10 border border-white/2 p-4 rounded-md">
                    <DarkCheckbox
                      checked={formData.AutoCloseOnUserLeave}
                      onChange={(v) =>
                        setFormData((p) => ({ ...p, AutoCloseOnUserLeave: v }))
                      }
                      label="User Leave Safeguard"
                      description="Close ticket immediately if user leaves guild."
                    />
                  </div>

                  {/* No Response Rule */}
                  <div className="rounded-lg border border-white/5 bg-zinc-950/35 p-5 space-y-4">
                    <div>
                      <h4 className="text-[13px] font-medium text-zinc-200 flex items-center gap-1.5">
                        <span className="h-1.5 w-1.5 rounded-full bg-[#FF5A36] shrink-0" />
                        No Response Since Opening
                      </h4>
                      <p className="text-[13px] text-zinc-400 font-medium mt-1 leading-relaxed">
                        Close ticket if creator fails to send any message after opening.
                      </p>
                    </div>
                    <div className="flex items-center gap-3 bg-zinc-950/30 p-3 rounded-md border border-white/2">
                      <TimeInput
                        days={formData.AutoCloseNoResponseDays}
                        hours={formData.AutoCloseNoResponseHours}
                        mins={formData.AutoCloseNoResponseMins}
                        onChangeDays={(e) =>
                          handleNumberInput(
                            "AutoCloseNoResponseDays",
                            e.target.value,
                          )
                        }
                        onChangeHours={(e) =>
                          handleNumberInput(
                            "AutoCloseNoResponseHours",
                            e.target.value,
                          )
                        }
                        onChangeMins={(e) =>
                          handleNumberInput(
                            "AutoCloseNoResponseMins",
                            e.target.value,
                          )
                        }
                      />
                    </div>
                  </div>

                  {/* Inactivity timeout Rule */}
                  <div className="rounded-lg border border-white/5 bg-zinc-950/35 p-5 space-y-4">
                    <div>
                      <h4 className="text-[13px] font-medium text-zinc-200 flex items-center gap-1.5">
                        <span className="h-1.5 w-1.5 rounded-full bg-[#FF5A36] shrink-0" />
                        General Inactivity Timeout
                      </h4>
                      <p className="text-[13px] text-zinc-400 font-medium mt-1 leading-relaxed">
                        Close ticket if no new messages are exchanged for this duration.
                      </p>
                    </div>
                    <div className="flex items-center gap-3 bg-zinc-950/30 p-3 rounded-md border border-white/2">
                      <TimeInput
                        days={formData.AutoCloseSinceLastMessageDays}
                        hours={formData.AutoCloseSinceLastMessageHours}
                        mins={formData.AutoCloseSinceLastMessageMins}
                        onChangeDays={(e) =>
                          handleNumberInput(
                            "AutoCloseSinceLastMessageDays",
                            e.target.value,
                          )
                        }
                        onChangeHours={(e) =>
                          handleNumberInput(
                            "AutoCloseSinceLastMessageHours",
                            e.target.value,
                          )
                        }
                        onChangeMins={(e) =>
                          handleNumberInput(
                            "AutoCloseSinceLastMessageMins",
                            e.target.value,
                          )
                        }
                      />
                    </div>
                  </div>

                </div>
              ) : (
                <div className="flex flex-col items-center justify-center p-8 text-center rounded-lg border border-dashed border-white/10 bg-zinc-950/10">

                  <p className="text-[13px] text-zinc-400 font-semibold mt-2.5 leading-relaxed">
                    Auto-Close is currently disabled.
                  </p>
                </div>
              )}

            </div>
          </SectionCard>

        </div>

      </div>
    </div>
  );
}
