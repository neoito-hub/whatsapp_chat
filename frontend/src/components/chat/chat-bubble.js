/* eslint-disable react/no-array-index-key */
/* eslint-disable react/prop-types */
// ChatBubble.js
import React from 'react'
import * as dayjs from 'dayjs'

const ChatBubble = ({ message, isUser, chatRef }) => {
  const formattedContent = message.content.split('\n').map((line, index) => (
    <React.Fragment key={index}>
      {line}
      {index !== message.length - 1 && index !== 0 && <br />}{' '}
    </React.Fragment>
  ))
  return (
    <div
      ref={chatRef}
      className={`flex ${
        !isUser ? 'items-start justify-start' : 'items-end justify-end'
      }`}
    >
      <div
        className={`relative chat-msg-animation ${
          !isUser ? 'bg-[#E3D9FF] chatBoxAi' : 'bg-[#E5E5E5] chatBoxUser'
        }  text-black/90 rounded-lg py-2 px-4 max-w-[80%] shadow-sm`}
      >
        <p className="text-[16px] font-medium overflow-hidden text-ellipsis font-roboto">
          {formattedContent}
        </p>
        <span
          className={`text-ab-sm float-${message.role === 'owner' ? 'right' : 'left'} text-ab-black`}
        >
          {dayjs(message.created_at).format('h:mm a')}
        </span>
      </div>
    </div>
  )
}

export default ChatBubble
