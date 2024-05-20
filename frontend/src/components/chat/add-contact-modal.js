/* eslint-disable react/prop-types */
import React, { useState } from 'react'

const AddContactModal = ({ isOpen, onClose, onSave }) => {
  const [formData, setFormData] = useState({
    name: '',
    countryCode: '+91',
    phoneNumber: '',
    email: '',
    address: '',
    projectId: '1',
  })
  const [errors, setErrors] = useState({})

  const validateForm = () => {
    let formValid = true
    const newErrors = {}

    if (!formData.name) {
      newErrors.name = 'Name is required'
      formValid = false
    }

    if (!formData.phoneNumber) {
      newErrors.phoneNumber = 'Phone number is required'
      formValid = false
    } else if (!/^\d+$/.test(formData.phoneNumber)) {
      newErrors.phoneNumber = 'Phone number should contain only digits'
      formValid = false
    }

    if (!formData.email) {
      newErrors.email = 'Email is required'
      formValid = false
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid'
      formValid = false
    }

    if (!formData.address) {
      newErrors.address = 'Address is required'
      formValid = false
    }

    setErrors(newErrors)
    return formValid
  }

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }))
  }

  const handleSubmit = () => {
    if (validateForm()) {
      onSave(formData)
      // onClose()
    }
  }

  if (!isOpen) return null

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

        <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div className="sm:flex sm:items-start">
              <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                <h3 className="text-lg font-medium leading-6 text-gray-900">
                  Add Contact
                </h3>
                <div className="mt-2">
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Name
                    </label>
                    <input
                      value={formData.name}
                      name="name"
                      onChange={handleChange}
                      placeholder="Enter Name"
                      type="text"
                      className={`${
                        errors.name
                          ? 'border-ab-red'
                          : 'border-ab-gray-light focus:border-primary'
                      } text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none`}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.name}
                    </p>
                  </div>
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Phone Numer
                    </label>
                    <div className="flex items-center justify-betweeen">
                      <span className="mr-2">+91</span>
                      <input
                        value={formData.phoneNumber}
                        name="phoneNumber"
                        onChange={handleChange}
                        placeholder="Enter Phone Number"
                        type="text"
                        className={`${
                          errors.name
                            ? 'border-ab-red'
                            : 'border-ab-gray-light focus:border-primary'
                        } text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none`}
                      />
                    </div>
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.phoneNumber}
                    </p>
                  </div>
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Email
                    </label>
                    <input
                      value={formData.email}
                      name="email"
                      onChange={handleChange}
                      placeholder="Enter Email"
                      type="text"
                      className={`${
                        errors.email
                          ? 'border-ab-red'
                          : 'border-ab-gray-light focus:border-primary'
                      } text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none`}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.email}
                    </p>
                  </div>
                  <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Address
                    </label>
                    <input
                      value={formData.address}
                      name="address"
                      onChange={handleChange}
                      placeholder="Enter Address"
                      type="text"
                      className={`${
                        errors.address
                          ? 'border-ab-red'
                          : 'border-ab-gray-light focus:border-primary'
                      } text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none`}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.address}
                    </p>
                  </div>
                  {/* <div className="flex flex-col float-left mb-4 w-full">
                    <label className="text-ab-sm float-left mb-2 font-medium text-black">
                      Name
                    </label>
                    <input
                      value={formData.name}
                      name="name"
                      onChange={handleChange}
                      placeholder="Enter Name"
                      type="text"
                      className={`${
                        errors.name
                          ? 'border-ab-red'
                          : 'border-ab-gray-light focus:border-primary'
                      } text-ab-sm bg-ab-gray-light float-left w-full rounded-md border py-3 px-4 focus:outline-none`}
                    />
                    <p className="text-xs text-ab-red left-0 mt-0.5">
                      {errors.name}
                    </p>
                  </div> */}
                </div>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
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

export default AddContactModal
