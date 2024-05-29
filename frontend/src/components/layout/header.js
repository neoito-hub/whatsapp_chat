/* eslint-disable import/no-extraneous-dependencies */
import React, { useEffect, useState } from 'react'
// import { Link } from 'react-router-dom'
import useOnclickOutside from 'react-cool-onclickoutside'
import { shield } from '@appblocks/js-sdk'
import Logo from '../../assets/img/logo.png'
import LogoTxt from '../../assets/img/logo-txt.svg'
import DiscordIcon from '../../assets/img/icons/discord-icon.svg'
import apiHelper from '../common/apiHelper'

const Header = () => {
  const [responsiveView, setResponsiveView] = useState(false)
  const [userDetails, setUserDetails] = useState(null)
  const [profDropdown, setProfDropdown] = useState(false)

  const profDropContainer = useOnclickOutside(() => {
    setProfDropdown(false)
  })

  const getUserDetailsApiUrl = 'get-user-details'

  useEffect(async () => {
    ;(async () => {
      const res = await apiHelper({
        baseUrl: process.env.SHIELD_AUTH_URL,
        subUrl: getUserDetailsApiUrl,
        apiType: 'get',
      })
      setUserDetails(res)
    })()
  }, [])

  const signOut = async () => {
    setProfDropdown(false)
    localStorage.clear()
    await shield.logout()
  }

  useEffect(() => {
    function handleResize() {
      if (window.innerWidth > 768) {
        setResponsiveView(false)
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  return (
    <header className="border-ab-gray-medium fixed top-0 left-0 z-[999] w-full border-b bg-white">
      <div className="flex h-16 w-full px-4 md:items-center md:justify-between md:space-x-4 md:px-6 xl:px-12">
        <div className="flex flex-grow items-center py-2">
          <div className="flex w-full items-center">
            <span
              // to="/"
              className="flex flex-shrink-0 items-center focus:outline-none cursor-pointer"
            >
              <img className="max-w-[48px]" src={Logo} alt="" />
              <img className="lg-lt:hidden ml-3" src={LogoTxt} alt="" />
            </span>
          </div>
        </div>
        <div className="flex flex-shrink-0 items-center">
          <div
            id="navbar-responsive"
            className={`nav-menu-wrapper custom-scroll-primary md-lt:bg-white md-lt:py-1.5 md-lt:shadow-lg ${
              !responsiveView ? 'md-lt:-right-64' : 'md-lt:right-0'
            }`}
          >
            <ul className="text-ab-black text-ab-sm my-3 flex flex-col items-center md:my-0 md:flex-row md:space-x-4 md-lt:items-start md-lt:space-y-3">
              <li>
                <button
                  type="button"
                  onClick={() => {
                    window.open(process.env.DOCS_PUBLIC_PATH, '_blank')
                  }}
                  className="block hover:underline cursor-pointer font-semibold focus:outline-none"
                >
                  Docs
                </button>
              </li>
              <li>
                <button
                  type="button"
                  onClick={() => {
                    window.open(process.env.APPBLOCKS_DISCORD_URL, '_blank')
                  }}
                  className="text-primary flex items-center cursor-pointer font-semibold focus:outline-none rounded transition-all bg-primary/10 hover:bg-primary/20 px-4 py-2"
                >
                  <img className="mr-2" src={DiscordIcon} alt="Discord" />
                  Join Discord
                </button>
              </li>
            </ul>
            {userDetails && (
              <span className="text-[#D9D9D9] ml-4 md-lt:hidden">|</span>
            )}
            {userDetails && (
              <div className="my-3 flex flex-col text-sm md:my-0 md:ml-4 md:flex-row md:items-center md:space-x-4 md-lt:items-start md-lt:space-y-3">
                <div
                  className="relative float-left flex-shrink-0 md-lt:w-full modal"
                  ref={profDropContainer}
                >
                  <div
                    onClick={() =>
                      !responsiveView && setProfDropdown(!profDropdown)
                    }
                    className="flex h-8 cursor-pointer select-none items-center space-x-1.5"
                  >
                    <span className="bg-primary flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold text-white capitalize">
                      {userDetails && userDetails?.full_name
                        ? userDetails?.full_name[0]
                        : userDetails?.email[0]}
                    </span>
                    <p className="text-ab-black w-full truncate text-xs font-semibold md:hidden">
                      {userDetails && userDetails?.full_name
                        ? userDetails?.full_name
                        : userDetails?.email}
                    </p>
                  </div>
                  <div
                    className={`border-ab-medium shadow-box dropDownFade top-12 right-0 z-10 bg-white py-3 md:absolute md:min-w-[260px] md:border md:px-4 md-lt:w-full ${
                      profDropdown ? '' : 'md:hidden'
                    }`}
                  >
                    {userDetails?.email && (
                      <div className="mb-2">
                        <p>Signed in as</p>
                        <span className=" font-semibold ">
                          {userDetails && userDetails?.full_name
                            ? userDetails?.full_name
                            : userDetails?.email}
                        </span>
                      </div>
                    )}
                    <ul>
                      <li key="signout" onClick={signOut} className="py-2">
                        <span className="text-ab-red cursor-pointer text-sm hover:underline">
                          Log out
                        </span>
                      </li>
                    </ul>
                    {/* <div className="border-ab-gray-medium my-2 border-t" />
                    <ul>
                      <li className="py-2 md:hidden">
                        <button
                          type="button"
                          onClick={() => {
                            window.open(process.env.DOCS_PUBLIC_PATH, "_blank");
                          }}
                          className="text-ab-black cursor-pointer text-sm hover:underline"
                        >
                          Docs
                        </button>
                      </li>
                    </ul> */}
                  </div>
                </div>
              </div>
            )}
          </div>
          <div className="flex flex-shrink-0 items-center">
            <button
              type="button"
              onClick={() => setResponsiveView(!responsiveView)}
              className="ml-3 inline-flex h-8 w-8 flex-shrink-0 items-center justify-center focus:outline-none md:hidden"
              aria-controls="navbar-default"
              aria-expanded="false"
            >
              <span className="sr-only">Open main menu</span>
              <span
                className={`hamburger-icon ${
                  responsiveView ? 'active-hamburger' : ''
                }`}
              />
            </button>
          </div>
        </div>
      </div>
    </header>
  )
}

export default Header
