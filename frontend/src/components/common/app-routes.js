import React, { Suspense, lazy } from 'react'
import { Routes, Route } from 'react-router-dom'

const AppRoute = () => {
  const Template = lazy(() => import('../template'))
  const Broadcast = lazy(() => import('../broadcast'))
  const Chat = lazy(() => import('../chat'))

  return (
    <Suspense fallback={<div>Loading...</div>}>
      <Routes>
        <Route index element={<Chat />} />
        <Route path="/chat" element={<Chat />} />
        <Route path="/broadcast" element={<Broadcast />} />
        <Route path="/template" element={<Template />} />
      </Routes>
    </Suspense>
  )
}

export default AppRoute
