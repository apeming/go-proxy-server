import type { MetricsSnapshot, MetricsHistory } from '../types/metrics';

const API_BASE = '/api';

export async function getRealtimeMetrics(): Promise<MetricsSnapshot> {
  const response = await fetch(`${API_BASE}/metrics/realtime`);
  if (!response.ok) {
    throw new Error('Failed to fetch realtime metrics');
  }
  return response.json();
}

export async function getMetricsHistory(
  startTime?: number,
  endTime?: number,
  limit?: number
): Promise<MetricsHistory[]> {
  const params = new URLSearchParams();
  if (startTime) params.append('startTime', startTime.toString());
  if (endTime) params.append('endTime', endTime.toString());
  if (limit) params.append('limit', limit.toString());

  const response = await fetch(`${API_BASE}/metrics/history?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch metrics history');
  }
  return response.json();
}
