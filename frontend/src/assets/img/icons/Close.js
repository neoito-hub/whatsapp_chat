/* eslint-disable react/prop-types */
import React from 'react'

const Close = ({ width = '24', height = '24' }) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={width}
    height={height}
    fill="none"
    viewBox="0 0 24 24"
  >
    <path
      stroke="#363636"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.8"
      strokeWidth="1.5"
      d="M6 6l12 12M6 18L18 6"
    />
  </svg>
)

export default Close
