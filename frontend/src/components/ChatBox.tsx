import { useState, useEffect } from 'react';
import { Message } from '../types';
import { MessageList } from './MessageList';
import { MessageInput } from '../MessageInput';

export default function ChatBox() {
  const [input, setInput] = useState('');
  const [isLoading, setLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);

  // Load messages from local storage on initial render
  useEffect(() => {
    const savedMessages = localStorage.getItem('chatMessages');
    if (savedMessages) {
      try {
        setMessages(JSON.parse(savedMessages));
      } catch (e) {
        console.error('Failed to parse saved messages:', e);
      }
    }
  }, []);

  // Save messages to local storage when they change
  useEffect(() => {
    localStorage.setItem('chatMessages', JSON.stringify(messages));
  }, [messages]);

  const handleSendMessage = async () => {
    if (!input.trim()) return;
    setLoading(true);
    const currentInput = input;
    setInput('');
    setError(null);

    try {
      // Add user message to the chat
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'user',
          content: currentInput,
        },
      ]);

      // Send message to the backend
      const response = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: currentInput, messages: messages }),
      });

      if (response.status !== 200) {
        setError(`Error: ${response.statusText || 'Failed to get response'}`);
        return;
      }

      await handleStreamResponse(response);
    } catch (error) {
      console.error('Error sending message:', error);
      setError('Network error. Please check your connection and try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleStreamResponse = async (response: Response) => {
    const reader = response.body?.getReader();
    const decoder = new TextDecoder();
    let done = false;

    const aiMessageId = Date.now().toString();
    const aiMessage: Message = {
      id: aiMessageId,
      role: 'assistant',
      content: '',
    };
    setMessages((prev) => [...prev, aiMessage]);

    while (!done && reader) {
      const { value, done: doneReading } = await reader.read();
      done = doneReading;
      const chunk = decoder.decode(value, { stream: true });
      setMessages((prev) =>
        prev.map((msg) =>
          msg.id === aiMessageId
            ? { ...msg, content: msg.content + chunk }
            : msg,
        ),
      );
    }
  };

  const clearConversation = () => {
    setMessages([]);
    localStorage.removeItem('chatMessages');
  };

  return (
    <div className="flex flex-col w-full max-w-3xl mx-auto h-[calc(100vh-180px)] rounded-lg shadow-lg border dark:border-gray-800 bg-white dark:bg-gray-900 overflow-hidden transition-colors duration-200">
      <div className="flex items-center justify-between p-4 border-b dark:border-gray-800">
        <h2 className="text-lg font-semibold">Chat with Llama 3.2</h2>
        {messages.length > 0 && (
          <button
            onClick={clearConversation}
            className="text-sm text-gray-500 hover:text-red-500 dark:text-gray-400 dark:hover:text-red-400 transition-colors duration-200"
          >
            Clear conversation
          </button>
        )}
      </div>
      
      <MessageList messages={messages} />
      <MessageInput
        input={input}
        setInput={setInput}
        sendMessage={handleSendMessage}
        isLoading={isLoading}
        error={error}
      />
    </div>
  );
}
