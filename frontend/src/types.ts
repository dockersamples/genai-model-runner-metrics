export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  metrics?: {
    tokensIn?: number;
    tokensOut?: number;
  };
}

// Metrics-related types
export interface MetricsData {
  totalRequests: number;
  averageResponseTime: number;
  tokensGenerated: number;
  tokensProcessed?: number; // New field for input tokens
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
