import useSWR from "swr";
import {
  API_BASE,
  type TranscriptListResponse,
  type TranscriptContent,
} from "../api";

const fetcher = async <T>(url: string): Promise<T> => {
  const res = await fetch(`${API_BASE}${url}`, { credentials: "include" });
  if (!res.ok) throw new Error("Failed to fetch");
  return res.json();
};

const contentFetcher = async <T>(url: string): Promise<T> => {
  const res = await fetch(`${API_BASE}${url}`, { credentials: "include" });
  if (!res.ok) throw new Error("Failed to fetch content");
  return res.json();
};

export function useTranscripts(serverId: string, page: number, limit = 20) {
  const { data, error, isLoading, mutate } = useSWR(
    serverId
      ? `/api/servers/${serverId}/transcripts?page=${page}&limit=${limit}`
      : null,
    fetcher<TranscriptListResponse>,
    {
      revalidateOnFocus: false,
      dedupingInterval: 30000,
    },
  );

  return {
    transcripts: data?.transcripts ?? [],
    pagination: data?.pagination ?? { page: 1, limit: 20, total: 0, pages: 0 },
    isLoading,
    isError: !!error,
    refresh: mutate,
  };
}

export function useTranscriptContent(serverId: string, transcriptId: string) {
  const { data, error, isLoading } = useSWR(
    serverId && transcriptId
      ? `/api/servers/${serverId}/transcripts/${transcriptId}/content`
      : null,
    contentFetcher<TranscriptContent>,
    {
      revalidateOnFocus: false,
      dedupingInterval: 120000,
    },
  );

  return {
    content: data ?? null,
    isLoading,
    isError: !!error,
  };
}
