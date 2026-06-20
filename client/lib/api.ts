import { genId } from "./utils";

export const API_BASE = (() => {
  const base =
    process.env.NEXT_PUBLIC_API_BASE ??
    process.env.NEXT_PUBLIC_API_BASE_URL;
  if (!base && typeof window === "undefined") {
    // SSR/build-time: warn loudly but don't crash — Next.js needs to render
    console.error("[api] NEXT_PUBLIC_API_BASE is not set");
  }
  return base ?? "";
})();
let memoryCsrfToken: string | null = null;

export const setMemoryCsrfToken = (token: string | null) => {
  memoryCsrfToken = token;
};

const readCookie = (name: string): string | null => {
  if (typeof document === "undefined") return null;
  const value = document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${name}=`));
  if (!value) return null;
  return decodeURIComponent(value.split("=")[1] ?? "");
};

const fetchApi = async <T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> => {
  const method = (options.method ?? "GET").toUpperCase();
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (
    method !== "GET" &&
    method !== "HEAD" &&
    !headers.has("Idempotency-Key")
  ) {
    headers.set("Idempotency-Key", genId());
  }
  if (method !== "GET" && method !== "HEAD" && !headers.has("X-CSRF-Token")) {
    const token = memoryCsrfToken || readCookie("fns_csrf");
    if (token) headers.set("X-CSRF-Token", token);
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    credentials: "include",
    headers,
  });

  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }

  return res.json() as Promise<T>;
};

export const api = {
  auth: {
    me: () =>
      fetchApi<{
        user: import("./context/auth/types").User | null;
        authorized?: boolean;
        csrfToken?: string;
      }>("/api/auth/me"),
    servers: () => fetchApi<{ servers: ServerSummary[] }>("/api/auth/servers"),
    login: () => {
      window.location.href = `${API_BASE}/api/auth/login`;
    },
    logout: async () => {
      const token = memoryCsrfToken || readCookie("fns_csrf") || "";
      await fetch(`${API_BASE}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Idempotency-Key": genId(),
          "X-CSRF-Token": token,
        },
      });
      memoryCsrfToken = null;
    },
  },

  config: {
    get: (serverId: string, showChannels = false) => {
      const query = showChannels ? "?show_channels=true" : "";
      return fetchApi<ServerConfigResponse>(`/api/config/${serverId}${query}`);
    },
    update: (serverId: string, data: Partial<ServerConfig>) =>
      fetchApi<ServerConfig>(`/api/config/${serverId}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
  },

  panels: {
    list: (serverId: string) =>
      fetchApi<PanelConfig[]>(`/api/servers/${serverId}/panels`),
    listMulti: (serverId: string) =>
      fetchApi<MultiPanelConfig[]>(`/api/servers/${serverId}/multi-panels`),
    get: (serverId: string, panelId: string) =>
      fetchApi<PanelDetail>(`/api/servers/${serverId}/panels/${panelId}`),
    create: (serverId: string, payload: CreatePanelPayload) =>
      fetchApi<PanelConfig>(`/api/servers/${serverId}/panels`, {
        method: "POST",
        body: JSON.stringify(payload),
      }),
    update: (serverId: string, panelId: string, payload: CreatePanelPayload) =>
      fetchApi<PanelConfig>(`/api/servers/${serverId}/panels/${panelId}`, {
        method: "PUT",
        body: JSON.stringify(payload),
      }),
    delete: (serverId: string, panelId: string) =>
      fetchApi<{ status: string }>(
        `/api/servers/${serverId}/panels/${panelId}`,
        {
          method: "DELETE",
        },
      ),
    send: (serverId: string, panelId: string) =>
      fetchApi<{ messageId: string }>(
        `/api/servers/${serverId}/panels/${panelId}/send`,
        {
          method: "POST",
        },
      ),
  },

  multiPanels: {
    list: (serverId: string) =>
      fetchApi<MultiPanelConfig[]>(`/api/servers/${serverId}/multi-panels`),
    get: (serverId: string, multiPanelId: string) =>
      fetchApi<MultiPanelDetail>(
        `/api/servers/${serverId}/multi-panels/${multiPanelId}`,
      ),
    create: (serverId: string, payload: MultiPanelPayload) =>
      fetchApi<MultiPanelConfig>(`/api/servers/${serverId}/multi-panels`, {
        method: "POST",
        body: JSON.stringify(payload),
      }),
    update: (
      serverId: string,
      multiPanelId: string,
      payload: MultiPanelPayload,
    ) =>
      fetchApi<MultiPanelConfig>(
        `/api/servers/${serverId}/multi-panels/${multiPanelId}`,
        {
          method: "PUT",
          body: JSON.stringify(payload),
        },
      ),
    delete: (serverId: string, multiPanelId: string) =>
      fetchApi<{ status: string }>(
        `/api/servers/${serverId}/multi-panels/${multiPanelId}`,
        {
          method: "DELETE",
        },
      ),
    send: (serverId: string, multiPanelId: string) =>
      fetchApi<{ messageId: string }>(
        `/api/servers/${serverId}/multi-panels/${multiPanelId}/send`,
        {
          method: "POST",
        },
      ),
  },

  guildMeta: {
    all: (serverId: string) =>
      fetchApi<GuildMetaResponse>(`/api/servers/${serverId}/meta`),
    roles: (serverId: string) =>
      fetchApi<DiscordRole[]>(`/api/servers/${serverId}/meta/roles`),
    channels: (serverId: string) =>
      fetchApi<DiscordChannel[]>(`/api/servers/${serverId}/meta/channels`),
    categories: (serverId: string) =>
      fetchApi<DiscordChannel[]>(`/api/servers/${serverId}/meta/categories`),
    emojis: (serverId: string) =>
      fetchApi<DiscordEmoji[]>(`/api/servers/${serverId}/meta/emojis`),
  },

  transcripts: {
    list: (serverId: string, page = 1, limit = 20) =>
      fetchApi<TranscriptListResponse>(
        `/api/servers/${serverId}/transcripts?page=${page}&limit=${limit}`,
      ),
    get: (serverId: string, transcriptId: string) =>
      fetchApi<TranscriptDetailResponse>(
        `/api/servers/${serverId}/transcripts/${transcriptId}`,
      ),
  },

  staff: {
    get: (serverId: string) =>
      fetchApi<StaffConfig>(`/api/servers/${serverId}/staff`),
    update: (serverId: string, payload: StaffUpdatePayload) =>
      fetchApi<{ status: string }>(`/api/servers/${serverId}/staff`, {
        method: "PUT",
        body: JSON.stringify(payload),
      }),
  },
};

export type ServerSummary = {
  id: string;
  name: string;
  iconUrl: string;
};

export type ServerConfig = {
  ID: number;
  TicketNameStyle: string;
  TicketTranscripts: string;
  TicketTranscriptCid?: string;
  MaxTicketsPerUser: number;
  TicketPermissionsAttachFiles: boolean;
  TicketPermissionsEmbedLinks: boolean;
  TicketPermissionsAddReactions: boolean;
  AutoClose: boolean;
  AutoCloseOnUserLeave: boolean;
  AutoCloseNoResponseDays: number;
  AutoCloseNoResponseHours: number;
  AutoCloseNoResponseMins: number;
  AutoCloseSinceLastMessageDays: number;
  AutoCloseSinceLastMessageHours: number;
  AutoCloseSinceLastMessageMins: number;
};

type ServerConfigResponse =
  | ServerConfig
  | { config: ServerConfig; channels: DiscordChannel[] };

export type DiscordChannel = {
  id: string;
  name: string;
  type: number;
  parent_id: string;
};

export type DiscordRole = {
  id: string;
  name: string;
  color: number;
  position: number;
};

export type DiscordEmoji = {
  id: string;
  name: string;
  animated: boolean;
};

export type GuildMetaResponse = {
  roles: DiscordRole[];
  channels: DiscordChannel[];
  categories: DiscordChannel[];
};

export type PanelConfig = {
  ID: number;
  ServerConfigID: number;
  MentionRolesOnOpen: string[];
  CategoryID: { String: string; Valid: boolean };
  Title: string;
  Content: { String: string; Valid: boolean };
  EmbedColor: number;
  ChannelID: string;
  BtnColor: string;
  BtnTxt: string;
  BtnEmoji: { String: string; Valid: boolean };
  LargeImgUrl: { String: string; Valid: boolean };
  SmallImgUrl: { String: string; Valid: boolean };
};

export type QuestionItem = {
  label: string;
  isRequired: boolean;
};

export type CreatePanelPayload = {
  mentionRolesOnOpen: string[];
  categoryId: string;
  title: string;
  content: string;
  embedColor: number;
  channelId: string;
  btnColor: string;
  btnTxt: string;
  btnEmoji: string;
  largeImgUrl: string;
  smallImgUrl: string;
  questions: QuestionItem[];
  welcomeMessage: WelcomeMessagePayload;
};

export type PanelDetail = {
  mentionRolesOnOpen: string[];
  categoryId: string;
  title: string;
  content: string;
  embedColor: number;
  channelId: string;
  btnColor: string;
  btnTxt: string;
  btnEmoji: string;
  largeImgUrl: string;
  smallImgUrl: string;
  questions: QuestionItem[];
  welcomeMessage: WelcomeMessagePayload | null;
};

export type WelcomeMessagePayload = {
  embedColor: number;
  title: string;
  description: string;
  titleUrl: string;
  largeImgUrl: string;
  smallImgUrl: string;
  footerText: string;
  footerIconUrl: string;
};

export type MultiPanelConfig = {
  ID: number;
  ServerConfigID: number;
  Title: string;
  Content: { String: string; Valid: boolean };
  EmbedColor: number;
  ChannelID: string;
  LargeImgUrl: { String: string; Valid: boolean };
  SmallImgUrl: { String: string; Valid: boolean };
  UseDropdown: boolean;
  PanelConfigIds: number[];
  Footer: { String: string; Valid: boolean };
  FootIconUrl: { String: string; Valid: boolean };
};

export type MultiPanelPayload = {
  title: string;
  content: string;
  embedColor: number;
  channelId: string;
  largeImgUrl: string;
  smallImgUrl: string;
  useDropdown: boolean;
  panelConfigIds: number[];
  footer: string;
  footIconUrl: string;
};

export type MultiPanelDetail = {
  title: string;
  content: string;
  embedColor: number;
  channelId: string;
  largeImgUrl: string;
  smallImgUrl: string;
  useDropdown: boolean;
  panelConfigIds: number[];
  footer: string;
  footIconUrl: string;
};

export type Transcript = {
  id: number;
  ticketId: string;
  username: string;
  userId: string;
  openedAt: number;
  closedAt: number;
  closedBy: string;
  totalMessages: number;
  totalAttachments: number;
  totalEmbeds: number;
};

export type TranscriptListResponse = {
  transcripts: Transcript[];
  pagination: { page: number; limit: number; total: number; pages: number };
};

export type TranscriptDetailResponse = {
  transcript: Transcript & { storageKey: string };
  presignedUrl: string;
};

export type TranscriptMessage = {
  id: string;
  type:
    | "message"
    | "embed"
    | "attachment"
    | "voice_join"
    | "voice_leave"
    | "system";
  author: {
    id: string;
    username: string;
    discriminator: string;
    avatar: string | null;
    bot: boolean;
  };
  content: string | null;
  timestamp: string;
  embeds?: {
    title?: string;
    description?: string;
    url?: string;
    color?: number;
    fields?: { name: string; value: string; inline: boolean }[];
    image?: { url: string } | null;
    thumbnail?: { url: string } | null;
    footer?: { text: string; iconUrl: string | null } | null;
    author?: {
      name: string;
      url: string | null;
      iconUrl: string | null;
    } | null;
  }[];
  attachments?: {
    id: string;
    filename: string;
    url: string;
    proxyUrl: string;
    size: number;
    contentType: string | null;
    width: number | null;
    height: number | null;
  }[];
  edited: boolean;
  editedTimestamp: string | null;
  reactions?: { emoji: string; count: number }[];
};

export type TranscriptContent = {
  ticketId: string;
  username: string;
  userId: string;
  messages: TranscriptMessage[];
  metadata: {
    ticketOpenedAt: string;
    ticketClosedAt: string;
    closedBy: { id: string; username: string };
    totalMessages: number;
    totalAttachments: number;
    totalEmbeds: number;
    participants: { id: string; username: string; messageCount: number }[];
  };
};

export type StaffMember = {
  id: string;
  username: string;
  globalName: string;
  avatarUrl: string;
};

export type StaffRole = {
  id: string;
  name: string;
  color: number;
};

export type StaffConfig = {
  members: StaffMember[];
  roles: StaffRole[];
  authorizedMemberIds: string[];
  authorizedRoleIds: string[];
};

export type StaffUpdatePayload = {
  authorizedMemberIds: string[];
  authorizedRoleIds: string[];
};
