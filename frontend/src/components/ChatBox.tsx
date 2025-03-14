import { useState } from 'react';
import { Message } from '../types';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';

export default function ChatBox() {
  const [input, setInput] = useState('');
  const [isLoading, setLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);
  const handleSendMessage = async () => {
    if (!input.trim()) return;
    setLoading(true);
    const currentInput = input;
    setInput('');

    try {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          role: 'user',
          content: currentInput,
        },
      ]);

      const response = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: currentInput, messages: messages }),
      });

      if (response.status !== 200) {
        setError(response.statusText);
        return;
      }

      await handleStreamResponse(response);
    } catch (error) {
      console.error('Error sending message:', error);
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

  return (
    <div className="flex flex-col w-full max-w-2xl mx-auto h-[80vh]">
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
