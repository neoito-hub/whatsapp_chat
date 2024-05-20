/* eslint-disable no-restricted-syntax */
/* eslint-disable react/no-unescaped-entities */
/* eslint-disable no-unused-expressions */
/* eslint-disable react/prop-types */
import React, { useState, useEffect } from 'react'
import Select from 'react-select'
import MultiSelectContact from './multiselect-contact'
import apiHelper from '../common/apiHelper'

const listTemplateUrl = 'list_template'
const broadcastMessageUrl = 'broadcast_message'

const CreateNewBroadcastModal = ({ isOpen, onClose }) => {
  const [formData, setFormData] = useState({
    name: '',
    templateId: null,
    recipients: [],
    template_params: [],
    project_id: '1',
  })
  const [errors, setErrors] = useState({})
  const [templates, setTemplates] = useState(null)
  const [selectedTemplate, setSelectedTemplate] = useState(null)

  const filterDataStructure = () => ({
    page: 1,
    limit: 100,
    search: '',
    project_id: '1',
  })

  useEffect(async () => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listTemplateUrl,
      value: filterDataStructure(),
    })
    res &&
      setTemplates(
        res?.templates?.map((item) => ({
          id: item?.id,
          value: item?.name,
          label: item?.name,
        })),
      )
  }, [])

  const validateForm = () => {
    let formValid = true
    const newErrors = {}

    if (formData.name.trim() === '') {
      newErrors.name = 'Name is required'
      formValid = false
    }

    if (!formData.templateId) {
      newErrors.templateId = 'Template is required'
      formValid = false
    }
    console.log(formData?.recipients?.length)
    if (!formData.recipients.length) {
      newErrors.recipients = 'Atleast one recipients required'
      formValid = false
    }

    setErrors(newErrors)
    setTimeout(() => {
      setErrors({})
    }, 3000)
    return formValid
  }

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }))
  }

  const handleReactSelectChange = (e, idName) => {
    const { id } = e
    setFormData((prevData) => ({
      ...prevData,
      [idName]: id,
    }))
  }

  const handleSubmit = async () => {
    if (validateForm()) {
      const res = await apiHelper({
        baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
        subUrl: broadcastMessageUrl,
        value: formData,
      })
      onClose()
    }
  }

  if (!isOpen) return null

  const customStyles = {
    control: (provided, state) => ({
      ...provided,
      width: '100%',
      border: state.isFocused
        ? '1px solid #63b3ed'
        : `1px solid ${state.isDisabled ? '#e2e8f0' : '#a0aec0'}`,
      borderRadius: '0.375rem',
      backgroundColor: '#ffffff', // background color
      fontSize: '0.875rem',
      color: '#1a202c', // text color
      outline: 'none',
      transition: 'border-color 0.2s ease',
      '&:hover': {
        borderColor: '#a0aec0',
      },
    }),
  }

  return (
    <div className="fixed z-10 inset-0 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <div className="fixed inset-0 transition-opacity" aria-hidden="true">
          <div className="absolute inset-0 bg-gray-500 opacity-75" />
        </div>

        <span
          className="hidden sm:inline-block sm:align-middle sm:h-screen"
          aria-hidden="true"
        >
          &#8203;
        </span>

        <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-3xl sm:w-full">
          <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div className="sm:flex sm:items-start">
              <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full flex flex-col">
                <h3 className="text-lg font-medium leading-6 text-gray-900 mb-4">
                  New Broadcast
                </h3>
                <div className="mt-2">
                  <div className="flex flex-col float-left mb-2 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Broadcast Name
                    </label>
                    <input
                      value={formData.name}
                      name="name"
                      onChange={handleChange}
                      placeholder="Enter Name"
                      type="text"
                      className="border-ab-gray-light focus:border-primary text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none"
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.name}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col float-left mb-4 w-full mt-2">
                  <label className="text-ab-sm float-left mb-2 font-medium text-black">
                    Select Template Message
                  </label>
                  <Select
                    name="templateId"
                    classNamePrefix="react-select"
                    styles={customStyles}
                    value={selectedTemplate}
                    onChange={(e) => {
                      setSelectedTemplate(e)
                      handleReactSelectChange(e, 'templateId')
                    }}
                    options={templates}
                  />
                  <p className="text-xs text-ab-red left-0 mt-0.5">
                    {errors.templateId}
                  </p>
                </div>
                <div className="flex flex-col float-left mb-4 w-full mt-2">
                  <label className="text-ab-sm float-left mb-2 font-medium text-black">
                    Select Recipients
                  </label>
                  <MultiSelectContact
                    onContactChange={(contacts) =>
                      setFormData((prevData) => ({
                        ...prevData,
                        recipients: contacts,
                      }))
                    }
                  />
                  <p className="text-xs text-ab-red left-0 mt-0.5">
                    {errors.recipients}
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-3 mb-5 sm:px-6 sm:flex sm:flex-row-reverse mt-10">
            <button
              type="button"
              onClick={handleSubmit}
              className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-primary text-base font-medium text-white hover:bg-primary/80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:ml-3 sm:w-auto sm:text-sm"
            >
              Save
            </button>
            <button
              type="button"
              onClick={onClose}
              className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default CreateNewBroadcastModal
