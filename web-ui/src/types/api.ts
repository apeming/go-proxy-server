export interface ApiResponse<T = any> {
  status?: string;
  data?: T;
  error?: string;
}

export interface SystemSettings {
  autostartEnabled: boolean;
  registryEnabled: boolean;
  autostartSupported: boolean;
}

export interface TimeoutConfig {
  connect: number;
  idleRead: number;
  idleWrite: number;
}
