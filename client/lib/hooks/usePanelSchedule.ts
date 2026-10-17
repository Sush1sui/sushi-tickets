import useSWR from "swr";
import { API_BASE, type ScheduleWindow } from "../api";

const fetcher = async <T>(url: string): Promise<T> => {
  const res = await fetch(`${API_BASE}${url}`, { credentials: "include" });
  if (!res.ok) throw new Error("Failed to fetch");
  return res.json();
};

export function usePanelSchedule(serverId: string, panelId: string) {
  const { data, error, isLoading, mutate } = useSWR(
    serverId && panelId
      ? `/api/servers/${serverId}/panels/${panelId}/schedule`
      : null,
    fetcher<ScheduleWindow[]>,
    {
      revalidateOnFocus: false,
      dedupingInterval: 5000,
      refreshInterval: 15000,
    },
  );

  return {
    windows: data ?? [],
    isLoading,
    isError: !!error,
    refresh: mutate,
  };
}
