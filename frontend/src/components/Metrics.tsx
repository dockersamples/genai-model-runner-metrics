import { useState, useEffect } from 'react';
import { MetricsData } from '../types';

interface MetricsProps {
  isVisible: boolean;
}

export function Metrics({ isVisible }: MetricsProps) {
  const [metrics, setMetrics] = useState<MetricsData>({
    totalRequests: 0,
    averageResponseTime: 0,
    tokensGenerated: 0,
    tokensProcessed: 0,
    activeUsers: 0,
    errorRate: 0,
  });

  useEffect(() => {
    // Skip fetching if the metrics panel is not visible
    if (!isVisible) return;

    const fetchMetrics = async () => {
      try {
        const response = await fetch('http://localhost:8080/metrics/summary');
        if (response.ok) {
          const data = await response.json();
          setMetrics(data);
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

  return (
    <div className="bg-gray-100 dark:bg-gray-800 p-4 rounded-lg shadow mb-4">
      <h3 className="text-lg font-semibold mb-2">System Metrics</h3>
      <div className="grid grid-cols-2 md:grid-cols-6 gap-4">
        <MetricCard title="Total Requests" value={metrics.totalRequests} />
        <MetricCard 
          title="Avg Response Time" 
          value={`${metrics.averageResponseTime.toFixed(2)}s`} 
        />
        <MetricCard 
          title="Input Tokens" 
          value={metrics.tokensProcessed?.toLocaleString() || '0'} 
        />
        <MetricCard 
          title="Output Tokens" 
          value={metrics.tokensGenerated.toLocaleString()} 
        />
        <MetricCard title="Active Users" value={metrics.activeUsers} />
        <MetricCard 
          title="Error Rate" 
          value={`${(metrics.errorRate * 100).toFixed(2)}%`} 
          isError={metrics.errorRate > 0.05}
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
}

function MetricCard({ title, value, isError = false }: MetricCardProps) {
  return (
    <div className="bg-white dark:bg-gray-700 p-3 rounded shadow">
      <div className="text-sm text-gray-500 dark:text-gray-400">{title}</div>
      <div className={`text-lg font-semibold ${isError ? 'text-red-500' : ''}`}>
        {value}
      </div>
    </div>
  );
}
