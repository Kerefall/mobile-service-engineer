import type { Metadata, Viewport } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { Toaster } from '@/components/ui/sonner';
import { MobileFooter } from '@/components/mobile-footer';

const inter = Inter({ subsets: ['latin', 'cyrillic'] });

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
  viewportFit: 'cover',
};

export const metadata: Metadata = {
  title: 'Мобильный Инженер',
  manifest: '/manifest.json', // Для PWA
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ru">
      <body className={inter.className}>
        <div className="mx-auto max-w-md bg-background min-h-screen shadow-xl relative border-x border-border">
          <main className="pb-20"> {/* Отступ для футера */}
            {children}
          </main>
          <MobileFooter />
          <Toaster position="top-center" richColors />
        </div>
      </body>
    </html>
  );
}