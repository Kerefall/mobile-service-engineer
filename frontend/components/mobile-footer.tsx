'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { ReactNode } from 'react';

// SVG иконки
const TasksIcon = ({ active }: { active: boolean }) => (
  <svg 
    width="24" 
    height="24" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <path 
      d="M4 4H20V20H4V4Z" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round"
      fill="none"
    />
    <path 
      d="M8 8H16" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round"
    />
    <path 
      d="M8 12H16" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round"
    />
    <path 
      d="M8 16H13" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round"
    />
  </svg>
);

const ChatIcon = ({ active }: { active: boolean }) => (
  <svg 
    width="24" 
    height="24" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <path 
      d="M20 2H4C2.9 2 2 2.9 2 4V22L6 18H20C21.1 18 22 17.1 22 16V4C22 2.9 21.1 2 20 2Z" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round"
      fill="none"
    />
    <path 
      d="M7 9H17" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round"
    />
    <path 
      d="M7 13H14" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round"
    />
  </svg>
);

const ProfileIcon = ({ active }: { active: boolean }) => (
  <svg 
    width="24" 
    height="24" 
    viewBox="0 0 24 24" 
    fill="none" 
    xmlns="http://www.w3.org/2000/svg"
  >
    <path 
      d="M12 12C14.2091 12 16 10.2091 16 8C16 5.79086 14.2091 4 12 4C9.79086 4 8 5.79086 8 8C8 10.2091 9.79086 12 12 12Z" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round"
      fill="none"
    />
    <path 
      d="M20 21V19C20 16.7909 18.2091 15 16 15H8C5.79086 15 4 16.7909 4 19V21" 
      stroke={active ? '#FFFFFF' : '#A0A0A0'} 
      strokeWidth="2" 
      strokeLinecap="round" 
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);

interface NavItemProps {
  href: string;
  icon: (props: { active: boolean }) => ReactNode;
  label: string;
  active: boolean;
}

const NavItem = ({ href, icon: Icon, label, active }: NavItemProps) => {
  return (
    <Link href={href} className="flex flex-col items-center gap-1">
      <div 
        className={`
          w-14 h-14 rounded-full flex items-center justify-center transition-all
          ${active 
            ? 'bg-[#FF6F1C] shadow-lg shadow-[#FF6F1C]/30' 
            : 'bg-transparent hover:bg-gray-100'
          }
        `}
      >
        <Icon active={active} />
      </div>
      <span 
        className={`
          text-xs font-medium transition-colors
          ${active ? 'text-[#FF6F1C]' : 'text-gray-400'}
        `}
      >
        {label}
      </span>
    </Link>
  );
};

export function MobileFooter() {
  const pathname = usePathname();

  const navItems = [
    {
      href: '/',
      icon: TasksIcon,
      label: 'Задания',
    },
    {
      href: '/chat',
      icon: ChatIcon,
      label: 'Чат',
    },
    {
      href: '/profile',
      icon: ProfileIcon,
      label: 'Профиль',
    },
  ];

  return (
    <footer className="fixed bottom-0 left-0 right-0 max-w-md mx-auto bg-white border-t border-gray-200 px-4 py-3 z-100">
      <nav className="flex justify-around items-center">
        {navItems.map((item) => (
          <NavItem
            key={item.href}
            href={item.href}
            icon={item.icon}
            label={item.label}
            active={pathname === item.href}
          />
        ))}
      </nav>
    </footer>
  );
}