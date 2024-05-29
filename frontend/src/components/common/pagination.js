/* eslint-disable import/no-extraneous-dependencies */
/* eslint-disable react/prop-types */
import React from 'react'
import ReactPaginate from 'react-paginate'
import PaginationNextIcon from '../../assets/img/icons/pagination-next.svg'
import PaginationPrevIcon from '../../assets/img/icons/pagination-prev.svg'

const Pagination = (props) => {
  const { pageCount, handlePageChange, selected, Padding, marginTop } = props

  return (
    <div
      className={`float-left mt-${marginTop} flex w-full flex-wrap items-center justify-center p-${Padding} md:justify-between`}
    >
      <div className="list-pagination-wrapper mt-4 flex items-center">
        <ReactPaginate
          className="flex items-center"
          breakLabel="..."
          pageLinkClassName="flex items-center justify-center text-gray-dark text-sm rounded-md px-2 py-1 min-w-[32px] h-8 border border-gray-dark text-center hover:bg-ab-gray-light cursor-pointer focus:outline-none select-none"
          pageClassName="mx-0.5 rounded-md"
          nextLinkClassName="flex items-center justify-center rounded-md px-2 py-1 min-w-[32px] h-8 border border-ab-gray-dark text-center hover:bg-ab-gray-light cursor-pointer ml-0.5 focus:outline-none select-none"
          previousLinkClassName="flex items-center justify-center rounded-md px-2 py-1 min-w-[32px] h-8 border border-ab-gray-dark text-center hover:bg-ab-gray-light cursor-pointer mr-0.5 focus:outline-none select-none"
          disabledLinkClassName="opacity-40 hover:bg-transparent cursor-default"
          breakLinkClassName="flex items-center justify-center rounded-md px-2 py-1 min-w-[32px] h-8 border border-ab-gray-dark text-center hover:bg-ab-gray-light cursor-pointer mr-0.5 focus:outline-none select-none"
          nextLabel={<img src={PaginationNextIcon} alt="" />}
          previousLabel={<img src={PaginationPrevIcon} alt="" />}
          activeLinkClassName="!bg-[#F5F0FF]"
          pageRangeDisplayed={3}
          pageCount={pageCount || 1}
          renderOnZeroPageCount={null}
          onPageChange={handlePageChange}
          forcePage={selected || 0}
        />
      </div>
    </div>
  )
}

export default Pagination
