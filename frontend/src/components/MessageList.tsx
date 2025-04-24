import React, { useEffect, useRef } from 'react';
import { Message } from '../types';
import { MessageItem } from './MessageItem';

interface MessageListProps {
  messages: Message[];
  showTokenCount?: boolean;
}

export const MessageList: React.FC<MessageListProps> = ({ messages, showTokenCount = false }) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Scroll to bottom when messages change
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  // If there are no messages, show a placeholder
  if (messages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500 p-4">
        <div className="text-center">
          <div className="mb-2 text-3xl">?</div>
          <p className="mb-1">Start a conversation</p>
          <p className="text-sm">Ask anything about Docker or other topics</p>
        </div>
      </div>
    );
  }

  return (
    <div className="p-3 space-y-3">
      {messages.map((msg) => (
        <div key={msg.id} data-testid={`message-${msg.role}`}>
          <MessageItem key={msg.id} message={msg} showTokenCount={showTokenCount} />
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
};
