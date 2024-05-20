/* eslint-disable no-restricted-syntax */
/* eslint-disable react/no-unescaped-entities */
/* eslint-disable no-unused-expressions */
/* eslint-disable react/prop-types */
import React, { useState, useEffect } from 'react'
import Select from 'react-select'
import apiHelper from '../common/apiHelper'

const listCategoryUrl = 'list_category'
const listLanguageUrl = 'list_language'
const createTemplateUrl = 'create_template'

const CreateNewTemplateModal = ({ isOpen, onClose }) => {
  const [formData, setFormData] = useState({
    name: '',
    projectId: '1',
    category: null,
    categoryId: null,
    languageId: null,
    language: null,
    type: 'standard',
    components: [
      {
        type: 'BODY',
        text: '',
      },
    ],
  })
  const [errors, setErrors] = useState({})
  const [categories, setCategories] = useState([])
  const [selectedCategory, setSelectedCategory] = useState(null)
  const [languages, setLanguages] = useState([])
  const [selectedLanguage, setSelectedLanguage] = useState(null)
  const [templateTypes] = useState([
    { value: 'standard', label: 'Standard (Text Only)' },
    { value: 'interactive', label: 'Media & Interactive' },
  ])
  const [selectedTemplateType, setSelectedTemplateType] = useState({
    value: 'standard',
    label: 'Standard (Text Only)',
  })

  useEffect(async () => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listCategoryUrl,
    })
    res &&
      setCategories(
        res.map((item) => ({
          id: item?.id,
          value: item?.name,
          label: item?.name,
        }))
      )
  }, [])

  useEffect(async () => {
    const res = await apiHelper({
      baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
      subUrl: listLanguageUrl,
    })
    res &&
      setLanguages(
        res.map((item) => ({
          id: item?.id,
          value: item?.code,
          label: item?.name,
        }))
      )
  }, [])

  const isTextEmpty = (type) => {
    const bodyComponents = formData.components.filter(
      (component) => component.type === type
    )
    for (const component of bodyComponents) {
      if (component.text.trim() === '') {
        return true
      }
    }
    return false
  }

  const validateForm = () => {
    let formValid = true
    const newErrors = {}

    if (formData.name.trim() === '') {
      newErrors.name = 'Name is required'
      formValid = false
    }

    if (!formData.category) {
      newErrors.category = 'Category is required'
      formValid = false
    }

    if (!formData.language) {
      newErrors.language = 'Language is required'
      formValid = false
    }

    if (!formData.type) {
      newErrors.templateType = 'Template Type is required'
      formValid = false
    }

    if (isTextEmpty('BODY')) {
      newErrors.bodyText = 'Required'
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

  const handleReactSelectChange = (e, name, idName) => {
    const { value, id } = e
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
      [idName]: id,
    }))
  }

  const updateComponent = (e, type) => {
    const { value } = e.target
    setFormData((prevFormData) => {
      const updatedFormData = { ...prevFormData } // Make a copy of the previous state
      // Iterate over components array
      updatedFormData.components.forEach((component, index) => {
        if (component.type === type) {
          // Update text
          updatedFormData.components[index] = { ...component, text: value }
        }
      })
      return updatedFormData // Return the updated state
    })
  }

  const handleSubmit = async () => {
    if (validateForm()) {
      const res = await apiHelper({
        baseUrl: process.env.BLOCK_ENV_URL_API_BASE_URL,
        subUrl: createTemplateUrl,
        value: formData,
      })
      res && onClose()
    }
  }

  if (!isOpen) return null

  const customStyles = {
    control: (provided, state) => ({
      ...provided,
      width: '100%',
      //   padding: '0.375rem 0.75rem',
      border: state.isFocused
        ? '1px solid #63b3ed'
        : `1px solid ${state.isDisabled ? '#e2e8f0' : '#a0aec0'}`, // conditional border color
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
                  New Template
                </h3>
                <div className="mt-2">
                  <div className="flex flex-col float-left mb-2 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Name
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
                <div className="flex gap-3 mt-2">
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Category
                    </label>
                    <Select
                      name="category"
                      classNamePrefix="react-select"
                      styles={customStyles}
                      value={selectedCategory}
                      onChange={(e) => {
                        setSelectedCategory(e)
                        handleReactSelectChange(e, 'category', 'categoryId')
                      }}
                      options={categories}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.category}
                    </p>
                  </div>
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Language
                    </label>
                    <Select
                      name="language"
                      classNamePrefix="react-select"
                      styles={customStyles}
                      value={selectedLanguage}
                      onChange={(e) => {
                        setSelectedLanguage(e)
                        handleReactSelectChange(e, 'language', 'languageId')
                      }}
                      options={languages}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.language}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col float-left mb-4 w-full mt-2">
                  <label className="text-ab-sm float-left mb-2 font-medium text-black">
                    Template Type
                  </label>
                  <Select
                    name="tempalate-type"
                    classNamePrefix="react-select"
                    isDisabled
                    styles={customStyles}
                    value={selectedTemplateType}
                    // onChange={(e) => setSelectedTemplateType(e)}
                    options={templateTypes}
                  />
                  <p className="text-xs text-ab-red left-0 mt-0.5">
                    {errors.templateType}
                  </p>
                </div>
                <div className="text-sm mb-4">
                  <p>
                    <strong>Note -</strong> Enter the parameters that you want
                    to dynamically inject into the template using the parameter
                    format (inside of double curly brackets ). The corresponding
                    keyword will be replaced with{' '}
                    <strong>{'{{parameter name}}'}</strong>. Empty parameters
                    won't be accepted by meta. Example : ' Hi {'{{name}}'},
                    welcome to our business ' will be replaced with ' Hi ABC,
                    welcome to our business '.
                  </p>
                </div>
                <div className="mt-2">
                  <div className="flex flex-col">
                    <label
                      htmlFor="textarea"
                      className="block text-sm font-medium text-gray-700 border border-ab-gray px-4 py-2"
                    >
                      <p className="p-2 bg-ab-gray-dark w-fit text-md font-semibold rounded-sm">
                        Body
                      </p>
                    </label>
                    <textarea
                      id="body"
                      name="body"
                      placeholder="Type your template body text here..."
                      style={{ fontSize: '1.2rem' }}
                      rows="10"
                      className="border-ab-gray-light text-ab-sm bg-ab-gray-light float-left w-full border py-3 px-4 focus:outline-none"
                      onChange={(e) => updateComponent(e, 'BODY')}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.bodyText}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-3 mb-5 sm:px-6 sm:flex sm:flex-row-reverse">
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

export default CreateNewTemplateModal
