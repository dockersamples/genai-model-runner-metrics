import React from 'react';
import { SendIcon } from './Icons';

interface MessageInputProps {
  input: string;
  setInput: (input: string) => void;
  sendMessage: () => void;
  isLoading: boolean;
  error: string | null;
}

export const MessageInput: React.FC<MessageInputProps> = ({
  input,
  setInput,
  sendMessage,
  isLoading,
  error,
}) => {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        sendMessage();
      }}
      className="p-4 border-t dark:border-gray-800"
    >
      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput((e.target as HTMLInputElement).value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          className="flex-1 p-2 rounded-lg border dark:border-gray-800 bg-transparent focus:outline-none focus:ring-2 focus:ring-blue-500"
          disabled={isLoading}
        />
        <button
          type="submit"
          disabled={isLoading}
          className="p-2 rounded-lg bg-blue-500 text-white disabled:opacity-50"
        >
          <SendIcon />
        </button>
      </div>
      {error && <p className="text-red-500">{error}</p>}
    </form>
  );
};
