import { useState, useEffect } from 'react';
import './App.css';
import ChatBox from './components/ChatBox.tsx';
import { Header } from './components/Header.tsx';

function App() {
  const [darkMode, setDarkMode] = useState(false);

  useEffect(() => {
    // Check for user preference
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    setDarkMode(isDark);
  }, []);

  useEffect(() => {
    if (darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [darkMode]);

  const toggleDarkMode = () => {
    setDarkMode(!darkMode);
  };

  return (
    <div className="min-h-screen flex flex-col bg-white dark:bg-gray-900 dark:text-white transition-colors duration-200">
      <Header toggleDarkMode={toggleDarkMode} darkMode={darkMode} />
      <div className="flex-1 p-4">
        <ChatBox />
      </div>
      <footer className="text-center p-4 text-sm text-gray-500 dark:text-gray-400">
        <p>Powered by <span className="font-semibold">Docker Model Runner</span> running <span className="font-semibold">Llama 3.2 (1B)</span> in a Docker container</p>
      </footer>
    </div>
  );
}

export default App;
