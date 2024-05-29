/* eslint-disable no-unused-expressions */
/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react/prop-types */
import React, { useState, useEffect } from 'react'
import Select from 'react-select'
import apiHelper from '../common/apiHelper'

const listContactUrl = 'list_contact'

const MultiSelectContacts = ({ onContactChange }) => {
  const [selectedContacts, setSelectedContacts] = useState([])
  const [contacts, setContacts] = useState(null)

  const filterDataStructure = () => ({
    search: '',
    limit: 100,
    page: 1,
    project_id: '1',
  })

  useEffect(async () => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listContactUrl,
      value: filterDataStructure(),
    })
    res && setContacts(res?.contacts)
  }, [])

  const handleChange = (selectedOptions) => {
    setSelectedContacts(selectedOptions)
    onContactChange(selectedOptions?.map((contact) => contact?.value))
  }

  const handleSelectAll = () => {
    setSelectedContacts(
      selectedContacts?.length === contacts?.length
        ? []
        : contacts?.map((contact) => ({
            value: contact.id,
            label: contact.name,
          })),
    )
    onContactChange(contacts?.map((contact) => contact.id))
  }

  const customStyles = {
    control: (provided, state) => ({
      ...provided,
      border: state.isFocused ? '2px solid #63b3ed' : '1px solid #a0aec0',
      borderRadius: '0.375rem',
      backgroundColor: '#ffffff',
      fontSize: '0.875rem',
      color: '#1a202c',
      outline: 'none',
      transition: 'border-color 0.2s ease',
      '&:hover': {
        borderColor: '#a0aec0',
      },
    }),
    // menu: (provided, state) => ({
    //   ...provided,
    //   zIndex: 9999,
    //   position: 'absolute',
    //   overflowY: 'auto',
    //   maxHeight: '100px',
    //   width: '100%',
    //   top: 'calc(100% + 10px)',
    //   left: '0',
    // }),
  }

  return (
    <div>
      <div className="flex items-center gap-2">
        <input
          type="checkbox"
          onChange={handleSelectAll}
          checked={selectedContacts?.length === contacts?.length}
        />
        <label className="text-sm">Select All</label>
      </div>
      <Select
        value={selectedContacts}
        onChange={handleChange}
        options={contacts?.map((contact) => ({
          value: contact.id,
          label: contact.name,
        }))}
        isMulti
        styles={customStyles}
        placeholder="Select contacts..."
        closeMenuOnSelect={false}
      />
    </div>
  )
}

export default MultiSelectContacts
