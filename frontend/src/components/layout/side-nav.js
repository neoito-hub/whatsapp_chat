/* eslint-disable react/button-has-type */
/* eslint-disable no-restricted-globals */
import React, { useRef } from 'react'
import { useNavigate } from 'react-router-dom'

const SideNav = () => {
  const navigate = useNavigate()
  const sideDrawerRef = useRef()

  const navItems = [
    { display_name: 'Chat', slug: 'chat', url: '/chat' },
    { display_name: 'Broadcast', slug: 'broadcast', url: '/broadcast' },
    { display_name: 'Template', slug: 'template', url: '/template' },
  ]
  return (
    <aside
      ref={sideDrawerRef}
      className="h-full fixed pt-16 mt-8 border left-0 !w-[10rem] slider"
    >
      <nav>
        <ul className="flex flex-col gap-4 text-primary font-semibold">
          {navItems.map((menu) => (
            <li
              key={menu.slug}
              className={`flex gap-2 p-2 whitespace-nowrap cursor-pointer pl-5 ${
                (menu.url ===
                  `/${location?.pathname?.split('/').filter(Boolean)[0]}` ||
                  (menu.url === '/chat' && location.pathname === '/')) &&
                'bg-primary/10'
              }`}
              onClick={() => navigate(menu.url)}
            >
              {menu.display_name}
            </li>
          ))}
        </ul>
      </nav>
    </aside>
  )
}

export default SideNav
