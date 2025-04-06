export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

// Metrics-related types
export interface MetricsData {
  totalRequests: number;
  averageResponseTime: number;
  tokensGenerated: number;
  activeUsers: number;
  errorRate: number;
}

export interface MessageMetrics {
  requestTime: number;
  responseTime: number;
  tokensIn: number;
  tokensOut: number;
  firstTokenTime: number;
}

export interface ModelMetadata {
  model: string;
  contextWindow: number;
  parameters?: string;
}
