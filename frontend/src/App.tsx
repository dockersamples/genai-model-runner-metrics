import { useState, useEffect } from 'react';
import './App.css';
import ChatBox from './components/ChatBox.tsx';
import { Header } from './components/Header.tsx';
import { ModelMetadata } from './types';

function App() {
  const [darkMode, setDarkMode] = useState(false);
  const [modelInfo, setModelInfo] = useState<ModelMetadata | null>(null);

  useEffect(() => {
    // Check for user preference
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    setDarkMode(isDark);
    
    // Fetch model information
    fetchModelInfo();
  }, []);

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  const fetchModelInfo = async () => {
    try {
      const response = await fetch('http://localhost:8080/health');
      if (response.ok) {
        const data = await response.json();
        if (data.model_info) {
          setModelInfo(data.model_info);
        }
      }
    } catch (e) {
      console.error('Failed to fetch model info:', e);
    }
  };

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  // Format the model name for display
  const getModelDisplayInfo = () => {
    if (!modelInfo || !modelInfo.model) {
      return "AI Model";  // Default fallback
    }
    
    let modelName = modelInfo.model;
    let modelSize = "";
    
    // Try to extract size information (like 1B, 7B, etc.)
    const sizeMatch = modelName.match(/[:\-](\d+[bB])/);
    if (sizeMatch && sizeMatch[1]) {
      modelSize = `(${sizeMatch[1].toUpperCase()})`;
    }
    
    // Extract base model name
    if (modelName.includes('/')) {
      modelName = modelName.split('/').pop() || modelName;
    }
    
    if (modelName.includes(':')) {
      modelName = modelName.split(':')[0];
    }
    
    // Clean up common model name formats
    modelName = modelName
      .replace(/\.(\d)/g, ' $1')  // Add space before version numbers
      .replace(/([a-z])(\d)/gi, '$1 $2')  // Add space between letters and numbers
      .replace(/llama/i, 'Llama')  // Capitalize model names
      .replace(/smollm/i, 'SmolLM');
    
    return `${modelName} ${modelSize}`;
  };

  return (
    <div className="min-h-screen flex flex-col bg-white dark:bg-gray-900 dark:text-white transition-colors duration-200">
      <Header toggleDarkMode={toggleDarkMode} darkMode={darkMode} />
      <div className="flex-1 p-4">
        <ChatBox />
      </div>
      <footer className="text-center p-4 text-sm text-gray-500 dark:text-gray-400">
        <p>Powered by <span className="font-semibold">Docker Model Runner</span> running <span className="font-semibold">{getModelDisplayInfo()}</span> in a Docker container</p>
      </footer>
    </div>
  );
}

export default App;
