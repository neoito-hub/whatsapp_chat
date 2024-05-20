/* eslint-disable no-unused-expressions */
/* eslint-disable react/prop-types */
/* eslint-disable consistent-return */
/* eslint-disable react/no-array-index-key */
/* eslint-disable default-case */
/* eslint-disable array-callback-return */
import React, { useState, useRef, useEffect } from 'react'
import ChatBubble from './chat-bubble'
import apiHelper from '../common/apiHelper'
import Send from '../../assets/img/icons/send-icon.svg'
import Close from '../../assets/img/icons/close-button.svg'

const chatHistoryUrl = 'chat_history'
const sendMessageUrl = 'send_message'

const ChatContainer = ({ removeSelection, selectedChat }) => {
  const mainScrollArea = useRef()
  const lastMessageRef = useRef(null)
  const chatTxtarea = useRef()

  const [message, setMessage] = useState('')
  const [history, setHistory] = useState(null)
  const [loading, setLoading] = useState(false)

  const autoGrowTxtarea = (e) => {
    chatTxtarea.current.style.height = '24px'
    chatTxtarea.current.style.height = `${e.currentTarget.scrollHeight}px`
  }

  const resetTxtareaHt = () => {
    chatTxtarea.current.style.height = '24px'
  }

  const filterDataStructure = () => ({
    page: 1,
    limit: 10,
    chat_id: selectedChat?.id,
  })

  const fetchHistory = async () => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: chatHistoryUrl,
      value: filterDataStructure(),
    })
    res &&
      setHistory(
        res?.map((item) => ({
          role: item?.owner ? 'owner' : 'user',
          content: item?.messageText,
        }))
      )
  }

  useEffect(() => {
    fetchHistory()
  }, [selectedChat])

  const sendMessageFilterDataStructure = () => ({
    contact_id: selectedChat?.candidateId,
    message,
    project_id: '1',
  })

  const handleClick = async () => {
    setMessage('')
    resetTxtareaHt()
    setLoading(true)
    setTimeout(() => {
      mainScrollArea.current.scrollTop = mainScrollArea.current.scrollHeight
    }, 10)
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: sendMessageUrl,
      value: sendMessageFilterDataStructure(),
    })
    res && fetchHistory()
    setLoading(false)
  }

  return (
    <form
      className="w-full flex-grow flex flex-col max-h-full overflow-clip bg-[#FCFCFC]"
      onSubmit={(e) => {
        e.preventDefault()
        handleClick()
      }}
    >
      <div className="flex justify-between bg-primary p-4 text-white text-2xl">
        <div>{selectedChat?.chatName}</div>
        <img
          className="cursor-pointer"
          src={Close}
          width={40}
          alt="close button"
          onClick={removeSelection}
        />
      </div>
      <div className="flex-1 p-4">
        <div className="flex flex-col h-full">
          <div
            className="custom-scroll-bar overflow-y-auto flex flex-col gap-5 py-4 px-7 h-full"
            ref={mainScrollArea}
          >
            {history
              ?.slice()
              .reverse()
              .map((msg, idx) => {
                const isLastMessage = idx === history.length - 1
                switch (msg.role) {
                  case 'owner':
                    return (
                      <ChatBubble
                        key={idx}
                        message={msg.content}
                        isUser={false}
                        chatRef={isLastMessage ? lastMessageRef : null}
                      />
                    )
                  case 'user':
                    return (
                      <ChatBubble
                        key={idx}
                        message={msg.content}
                        isUser
                        chatRef={isLastMessage ? lastMessageRef : null}
                      />
                    )
                }
              })}
          </div>
          {/* {loading && (
            <div
              ref={lastMessageRef}
              className="px-6 py-2 text-black/50 text-xs flex items-center gap-x-1"
            >
              <span>Appblocks bot evaluating</span>
              <div className="typing typing-xs mt-0.5">
                <span className="typing-dot" />
                <span className="typing-dot" />
                <span className="typing-dot" />
              </div>
            </div>
          )} */}
        </div>
      </div>
      {/* input area */}
      <div className="flex sticky bottom-0 w-full px-4 pb-4">
        <div className="w-full relative">
          <div className="flex rounded w-full h-full border bg-white px-4 py-3 pr-16 text-base border-[#E5E5E5] focus-within:border-primary">
            <textarea
              ref={chatTxtarea}
              aria-label="chat input"
              value={message}
              onChange={(e) => {
                setMessage(e.target.value)
                autoGrowTxtarea(e)
              }}
              placeholder="Ask anything"
              className="w-full focus:outline-none resize-none text-[15px] max-h-[60px] overflow-y-auto"
              style={{ height: '24px' }}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  handleClick()
                }
              }}
            />
          </div>
          <button
            onClick={(e) => {
              e.preventDefault()
              // handleClick()
            }}
            className="absolute right-2.5 top-1/2 transform -translate-y-1/2 cursor-pointer hover:bg-primary/5 focus:bg-primary/10 p-2 rounded-full"
            type="submit"
            aria-label="Send"
            disabled={!message || loading}
          >
            <img className="max-w-[48px]" src={Send} alt="" />
          </button>
        </div>
      </div>
    </form>
  )
}

export default ChatContainer
