import api from './index';
import type { UnifiedConfig, TimeoutConfig, LimiterConfig } from '../types/api';

export const getConfig = () => api.get<UnifiedConfig>('/config');

export const saveConfig = (data: {
  timeout?: TimeoutConfig;
  limiter?: LimiterConfig;
  system?: { autostartEnabled: boolean };
}) => api.post('/config', data);
