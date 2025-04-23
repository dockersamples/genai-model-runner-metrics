import { useState, useEffect } from 'react';
import { MetricsData } from '../types';

interface MetricsProps {
  isVisible: boolean;
  messages?: { id: string; role: string; content: string; metrics?: { tokensIn?: number; tokensOut?: number } }[];
}

export function Metrics({ isVisible, messages = [] }: MetricsProps) {
  const [serverMetrics, setServerMetrics] = useState<MetricsData>({
    totalRequests: 0,
    averageResponseTime: 0,
    tokensGenerated: 0,
    tokensProcessed: 0,
    activeUsers: 0,
    errorRate: 0,
  });

  // Calculate local metrics from messages
  const calculateLocalMetrics = () => {
    let inputTokens = 0;
    let outputTokens = 0;

    messages.forEach(message => {
      if (message.role === 'user' && message.metrics?.tokensIn) {
        inputTokens += message.metrics.tokensIn;
      }
      if (message.role === 'assistant' && message.metrics?.tokensOut) {
        outputTokens += message.metrics.tokensOut;
      }
    });

    return { inputTokens, outputTokens };
  };

  const { inputTokens, outputTokens } = calculateLocalMetrics();

  useEffect(() => {
    // Skip fetching if the metrics panel is not visible
    if (!isVisible) return;

    const fetchMetrics = async () => {
      try {
        const response = await fetch('http://localhost:8080/metrics/summary');
        if (response.ok) {
          const data = await response.json();
          setServerMetrics(data);
        }
      } catch (error) {
        console.error('Failed to fetch metrics:', error);
      }
    };

    // Fetch metrics immediately and then every 10 seconds
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 10000);

    return () => clearInterval(interval);
  }, [isVisible]);

  if (!isVisible) return null;
  
  // Debug logging
  console.log('Current message metrics:', { inputTokens, outputTokens });
  console.log('Current messages:', messages);

  return (
    <div className="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg shadow mb-4">
      <h3 className="text-lg font-semibold mb-2">System Metrics</h3>
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
        <MetricCard title="Total Requests" value={serverMetrics.totalRequests} />
        <MetricCard 
          title="Avg Response Time" 
          value={`${serverMetrics.averageResponseTime.toFixed(2)}s`} 
        />
        <MetricCard 
          title="Input Tokens" 
          value={inputTokens || 0} 
          highlight={true}
        />
        <MetricCard 
          title="Output Tokens" 
          value={outputTokens || 0} 
          highlight={true}
        />
        <MetricCard title="Active Users" value={serverMetrics.activeUsers} />
        <MetricCard 
          title="Error Rate" 
          value={`${(serverMetrics.errorRate * 100).toFixed(2)}%`} 
          isError={serverMetrics.errorRate > 0.05}
        />
      </div>
      <div className="mt-3 text-xs text-gray-500 dark:text-gray-400">
        <p className="mb-1">Tokens are approximately calculated based on the text length. Actual token count may vary based on the model's tokenizer.</p>
        <p>
          <span className="font-medium">Error Rate</span> is the percentage of requests that failed out of the total requests.
        </p>
        <a 
          href="http://localhost:3001" 
          target="_blank" 
          rel="noopener noreferrer"
          className="mt-2 inline-block underline hover:text-blue-500"
        >
          View detailed metrics dashboard
        </a>
      </div>
    </div>
  );
}

interface MetricCardProps {
  title: string;
  value: string | number;
  isError?: boolean;
  highlight?: boolean;
}

function MetricCard({ title, value, isError = false, highlight = false }: MetricCardProps) {
  return (
    <div className={`${highlight ? 'bg-blue-50 dark:bg-blue-900/20' : 'bg-white dark:bg-gray-700'} p-3 rounded shadow`}>
      <div className="text-sm text-gray-500 dark:text-gray-400">{title}</div>
      <div className={`text-lg font-semibold ${isError ? 'text-red-500' : (highlight ? 'text-blue-600 dark:text-blue-400' : '')}`}>
        {value}
      </div>
    </div>
  );
}
