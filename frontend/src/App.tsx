import { useState, useEffect } from 'react';
import './App.css';
import ChatBox from './components/ChatBox.tsx';
import { Header } from './components/Header.tsx';
import { ModelMetadata } from './types';

function App() {
  // Initialize darkMode from localStorage or system preference
  const [darkMode, setDarkMode] = useState(() => {
    const savedDarkMode = localStorage.getItem('darkMode');
    
    // If we have a saved preference, use it
    if (savedDarkMode !== null) {
      return savedDarkMode === 'true';
    }
    
    // Otherwise, check system preference
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  });
  
  const [modelInfo, setModelInfo] = useState<ModelMetadata | null>(null);

  // Listen for changes to system preference
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    
    const handleChange = (e: MediaQueryListEvent) => {
      // Only update if user hasn't explicitly set a preference
      if (localStorage.getItem('darkMode') === null) {
        setDarkMode(e.matches);
      }
    };
    
    // Some browsers use addEventListener, some use addListener
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    } else {
      // @ts-ignore - For older browsers
      mediaQuery.addListener(handleChange);
      return () => {
        // @ts-ignore - For older browsers
        mediaQuery.removeListener(handleChange);
      };
    }
  }, []);

  // Fetch model information on component mount
  useEffect(() => {
    fetchModelInfo();
  }, []);

  // Apply dark mode class when darkMode state changes
  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('darkMode', 'true');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('darkMode', 'false');
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
    setDarkMode(prevMode => !prevMode);
  };

  // Format the model name for display
  const getModelDisplayInfo = () => {
    if (!modelInfo || !modelInfo.model) {
      return "AI Model";  // Default fallback
    }
    
    let modelName = modelInfo.model;
    let modelSize = "";
    
    // Try to extract size information (like 1B, 7B, etc.)
    const sizeMatch = modelName.match(/[:\-](\\d+[bB])/);
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
      .replace(/\\.(\\d)/g, ' $1')  // Add space before version numbers
      .replace(/([a-z])(\\d)/gi, '$1 $2')  // Add space between letters and numbers
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
