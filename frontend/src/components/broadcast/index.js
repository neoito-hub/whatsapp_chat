/* eslint-disable import/no-extraneous-dependencies */
/* eslint-disable jsx-a11y/control-has-associated-label */
/* eslint-disable no-shadow */
/* eslint-disable no-unused-expressions */
/* eslint-disable camelcase */
import React, { useState, useEffect, useCallback } from 'react'
import { debounce } from 'lodash'
import * as dayjs from 'dayjs'
import apiHelper from '../common/apiHelper'
import CreateNewTemplateModal from './create-new-broadcast-modal'
import Pagination from '../common/pagination'

const listBroadcastUrl = 'list_broadcast'
const page_limit = 10

const Broadcast = () => {
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [loader, setLoader] = useState(false)
  const [flag, setFlag] = useState(false)
  const [broadcasts, setBroadcasts] = useState(null)
  const [totalCount, setTotalCount] = useState(null)
  const [selectedPage, setSelectedPage] = useState(1)
  const [searchText, setSearchText] = useState('')

  const filterDataStructure = () => ({
    page: selectedPage,
    limit: page_limit,
    search: searchText,
    project_id: '1',
  })

  useEffect(async () => {
    setLoader(true)
    setBroadcasts(null)
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listBroadcastUrl,
      value: filterDataStructure(),
    })
    res && setBroadcasts(res?.broadcasts)
    res && setTotalCount(res.count[0].total || 0)
    setLoader(false)
  }, [flag])

  const handlePageChange = (event) => {
    const { selected } = event
    setSelectedPage(selected + 1)
    setFlag((flag) => !flag)
  }

  const handler = useCallback(
    debounce((text) => {
      setBroadcasts(null)
      setSearchText(text)
      setSelectedPage(1)
      setFlag((flg) => !flg)
    }, 1000),
    [],
  )

  const onSearchTextChange = (e) => {
    handler(e.target.value)
  }

  const calculateSerialNumber = (currentPage, index, pageLimit) =>
    (currentPage - 1) * pageLimit + index + 1

  return (
    <div className="float-left w-full max-w-5xl p-6 ml-10">
      <div className="float-left mt-7 w-full">
        <div className="float-left w-full overflow-x-hidden py-6">
          <div className="mb-4 text-2xl font-medium text-ab-black">
            Broadcast List
          </div>
          <div className="float-left w-full">
            <div className="float-left flex items-center space-x-3 pb-3 w-full">
              <input
                placeholder="Search Broadcast"
                onChange={onSearchTextChange}
                className="search-input-white border-ab-gray-dark text-ab-sm h-9 w-full rounded-md border !bg-[length:14px_14px] px-2 pl-9 focus:outline-none"
              />
              <button
                type="button"
                onClick={() => setIsModalOpen(true)}
                className="btn-primary text-ab-sm flex flex-shrink-0 items-center space-x-2.5 rounded px-5 py-2.5 font-bold leading-tight text-white transition-all hover:opacity-90"
              >
                <svg
                  width="14"
                  height="14"
                  viewBox="0 0 14 14"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <path
                    d="M6.99974 0.0664062V13.9331M0.0664062 6.99974H13.9331"
                    stroke="white"
                    strokeWidth="1.5"
                  />
                </svg>
                <span>New Broadcast</span>
              </button>
            </div>
            <div className="float-left w-full py-3">
              <div className="border-ab-gray-dark custom-h-scroll-primary float-left w-full overflow-x-auto border">
                <table className="min-w-full">
                  <thead>
                    <tr className="text-ab-black bg-ab-gray-light text-left text-sm">
                      <th className="p-3 font-normal">Sl. No</th>
                      <th className="whitespace-nowrap p-3 font-normal">
                        Broadcast Name
                      </th>
                      <th className="p-3 font-normal">Broadcast Time</th>
                      <th className="p-3 font-normal">Recipients</th>
                    </tr>
                  </thead>
                  <tbody>
                    {!loader &&
                      broadcasts?.map((broadcast, index) => (
                        <tr
                          key={broadcast?.user_id || index}
                          className="border-ab-gray-dark text-ab-black border-t text-xs"
                        >
                          <td className="p-3">
                            {calculateSerialNumber(
                              selectedPage,
                              index,
                              page_limit,
                            )}
                          </td>
                          <td className="p-3">
                            <div className="flex items-center space-x-2">
                              <div className="bg-secondary float-left flex h-9 w-9 flex-shrink-0 items-center justify-center overflow-hidden rounded-full">
                                <span className="text-lg font-semibold leading-normal text-white capitalize">
                                  {broadcast?.name[0] || ' '}
                                </span>
                              </div>
                              <p className="whitespace-nowrap capitalize">
                                {broadcast.name || '-'}
                              </p>
                            </div>
                          </td>
                          <td className="whitespace-nowrap p-3">
                            {dayjs(broadcast?.broadcastTime).format(
                              'DD MMM YYYY, h:mm A',
                            )}
                          </td>
                          <td className="whitespace-nowrap p-3">
                            {broadcast?.recipients?.length}
                          </td>
                        </tr>
                      ))}
                  </tbody>
                </table>
                {!loader && !broadcasts?.length && (
                  <div className="flex justify-center items-center">
                    <span className="text-ab-black float-left w-full py-10 text-center text-sm">
                      No Broadcasts Found
                    </span>
                  </div>
                )}
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
                <CreateNewTemplateModal
                  isOpen={isModalOpen}
                  onClose={() => {
                    setIsModalOpen(false)
                    setFlag((flg) => !flg)
                  }}
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Broadcast
