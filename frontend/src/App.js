/* eslint-disable import/no-unresolved */
/* eslint-disable import/no-extraneous-dependencies */
/* eslint-disable react/prop-types */
/* eslint-disable react/function-component-definition */
/* eslint-disable import/extensions */
import React, { useState, useEffect } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import ClipLoader from 'react-spinners/ClipLoader'
import { shield } from '@appblocks/js-sdk'
import FallbackUI from './components/common/fallback-ui.js'
import './index.css'
// import { Inter } from "next/font/google";
import Layout from './components/layout/layout'
import './index.scss'

// const inter = Inter({ subsets: ["latin"] });

export default function RootLayout() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    ;(async () => {
      await shield.init(process.env.BLOCK_ENV_URL_CLIENT_ID)
      if (!isLoggedIn) {
        const isLoggedinn = await shield.verifyLogin()
        setIsLoggedIn(isLoggedinn)
      }
    })()
  }, [isLoggedIn])

  const handleError = (_error, errorInfo) => {
    console.log('Error occured in ', errorInfo.componentStack.split(' ')[5])
  }

  const Loader = (
    <div className="h-screen w-full flex items-center justify-center">
      <ClipLoader color="#5E5EDD" size={50} />
    </div>
  )

  return (
    <React.StrictMode>
      <ErrorBoundary
        FallbackComponent={FallbackUI}
        onError={handleError}
        onReset={() => {
          // reset the state of your app so the error doesn't happen again
        }}
      >
        <div className="App">{isLoggedIn ? <Layout /> : <>{Loader} </>}</div>
      </ErrorBoundary>
    </React.StrictMode>
  )
}
