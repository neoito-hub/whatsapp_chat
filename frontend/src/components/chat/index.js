import React, { useEffect, useState } from 'react'
import ChatContactList from './chat-contact-list'
import ChatContainer from './chat-container'
import CreateNewChatModal from './create-new-chat-modal'

const Chat = () => {
  const [selectedChat, setSelectedChat] = useState(null)
  const [showChat, setShowChat] = useState(false)
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [chats, setChats] = useState(null)
  const [flag, setFlag] = useState(false)

  const handleContactChange = (contact) => {
    setShowChat(true)
    setSelectedChat(contact)
  }

  const onAddNewChatSuccess = () => {
    setSelectedChat(null)
    setShowChat(null)
    setFlag((flg) => !flg)
  }

  return (
    <div className="flex h-full">
      <ChatContactList
        handleContactChange={handleContactChange}
        selectedChat={selectedChat}
        showNewChatModal={() => setIsModalOpen(true)}
        chats={chats}
        updateChatList={(list) => setChats(list)}
        flag1={flag}
      />
      {showChat && (
        <ChatContainer
          removeSelection={() => {
            setSelectedChat(null)
            setShowChat(false)
          }}
          selectedChat={selectedChat}
        />
      )}
      <CreateNewChatModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSuccess={onAddNewChatSuccess}
      />
    </div>
  )
}

export default Chat
