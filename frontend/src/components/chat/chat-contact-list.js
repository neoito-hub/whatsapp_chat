/* eslint-disable camelcase */
/* eslint-disable import/no-extraneous-dependencies */
/* eslint-disable no-unused-expressions */
/* eslint-disable react/prop-types */
import React, { useEffect, useState, useCallback } from 'react'
import { debounce } from 'lodash'
import AddContactModal from './add-contact-modal'
import apiHelper from '../common/apiHelper'
import Pagination from '../common/pagination'

const listChatUrl = 'chat_list'
const addContactUrl = 'add_contact'
const page_limit = 10

const ChatContactList = ({
  handleContactChange,
  selectedChat,
  showNewChatModal,
  chats,
  updateChatList,
}) => {
  const [isModalOpen, setIsModalOpen] = useState(false)
  // const [chats, setChats] = useState(null)
  const [loading, setLoading] = useState(false)
  const [flag, setFlag] = useState(false)
  const [searchText, setSearchText] = useState('')
  const [totalCount, setTotalCount] = useState(null)
  const [selectedPage, setSelectedPage] = useState(1)

  const filterDataStructure = () => ({
    search: searchText,
    limit: page_limit,
    page: selectedPage,
    project_id: '1',
  })

  useEffect(async () => {
    setLoading(true)
    updateChatList(null)
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listChatUrl,
      value: filterDataStructure(),
    })
    res && updateChatList(res || [])
    res && setTotalCount(res?.count || 0)
    setLoading(false)
  }, [flag, searchText])

  const onAddContactSubmit = async (data) => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: addContactUrl,
      value: data,
      apiType: 'post',
    })
    res && setIsModalOpen(false)
    res && setFlag((flg) => !flg)
  }

  const handler = useCallback(
    debounce((text) => {
      updateChatList(null)
      setTotalCount(0)
      setSearchText(text)
    }, 1000),
    []
  )

  const onSearchTextChange = (e) => {
    handler(e.target.value)
  }

  const handlePageChange = (event) => {
    const { selected } = event
    setSelectedPage(selected + 1)
    setFlag((flg) => !flg)
  }

  return (
    <div className="flex-none w-1/4 border-r border-gray-200">
      <div className="flex flex-col justify-between h-full">
        <div className="pt-6">
          <div className="flex justify-between items-center p-4">
            <h2 className="text-2xl font-semibold">Chats</h2>
            <button
              className="bg-primary/10 hover:bg-primary/20 shadow-sm p-2 rounded-md text-primary"
              type="button"
              onClick={() => setIsModalOpen(true)}
            >
              Add Contact
            </button>
          </div>
          <input
            placeholder="Search Chat"
            onChange={onSearchTextChange}
            className="search-input-white border-ab-gray-dark text-ab-md h-12 w-full border-t !bg-[length:14px_14px] px-2 pl-9 focus:outline-none"
          />
          <ul className="border overflow-auto">
            {!loading &&
              chats &&
              chats?.map((chat) => (
                <li
                  key={chat.id}
                  onClick={() => handleContactChange(chat)}
                  className={`flex items-center py-2 px-4 cursor-pointer transition-colors border-b ${
                    selectedChat && selectedChat.id === chat.id
                      ? 'bg-primary/20'
                      : 'hover:bg-primary/10'
                  }`}
                >
                  <span className="bg-primary flex w-10 h-10 mr-3 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold text-white capitalize">
                    {chat.chatName[0] || '-'}
                  </span>
                  <span className="text-base font-medium">{chat.chatName}</span>
                </li>
              ))}
          </ul>
        </div>
        <div className="flex items-center justify-center pb-12">
          <button
            type="button"
            onClick={showNewChatModal}
            className="p-2 px-5 hover:bg-primary/90 bg-primary text-white rounded-full "
          >
            New Chat
          </button>
        </div>
      </div>
      {totalCount > page_limit && (
        <Pagination
          Padding={0}
          marginTop={1}
          pageCount={Math.ceil(totalCount / page_limit)}
          handlePageChange={handlePageChange}
          selected={selectedPage - 1}
        />
      )}
      {isModalOpen && (
        <AddContactModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onSave={onAddContactSubmit}
        />
      )}
    </div>
  )
}

export default ChatContactList
