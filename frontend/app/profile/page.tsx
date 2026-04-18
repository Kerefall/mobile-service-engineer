'use client';

import { useOrderStore } from '@/store/order-store';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { 
  User, FileText, Settings, ChevronRight, 
  CircleCheck, Clock, Calendar, TrendingUp,
  Wifi, WifiOff
} from 'lucide-react';
import { useEffect, useState } from 'react';

export default function ProfilePage() {
  const orders = useOrderStore((state) => state.orders);
  const [currentTime, setCurrentTime] = useState('9:41');
  const [isOnline, setIsOnline] = useState(true);
  const [lastSync, setLastSync] = useState('Только что');

  // Статистика
  const activeOrders = orders.filter(o => o.status !== 'completed');
  const completedOrders = orders.filter(o => o.status === 'completed');
  const inProgressOrders = orders.filter(o => o.status === 'in_progress');
  
  // Моковые данные статистики
  const monthlyCompleted = 47;
  const shiftCompleted = 3;
  const monthlyEarnings = '345 678 ₽';

  // Обновление времени
  useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      const hours = now.getHours().toString().padStart(2, '0');
      const minutes = now.getMinutes().toString().padStart(2, '0');
      setCurrentTime(`${hours}:${minutes}`);
    };
    
    updateTime();
    const interval = setInterval(updateTime, 1000);
    
    // Проверка онлайн статуса
    setIsOnline(navigator.onLine);
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);
    
    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);
    
    return () => {
      clearInterval(interval);
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  const menuItems = [
    {
      icon: User,
      label: 'Личные данные',
      value: 'Алексей Смирнов',
      href: '/profile/personal',
    },
    {
      icon: FileText,
      label: 'Документы',
      value: 'Паспорт, ИНН, СНИЛС',
      href: '/profile/documents',
    },
    {
      icon: Settings,
      label: 'Настройки',
      value: 'Уведомления, тема, язык',
      href: '/profile/settings',
    },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Хедер с временем */}
      <div className="sticky top-0 z-10 bg-white border-b px-4 py-3">
        <div className="flex justify-between items-center">
          <span className="text-lg font-medium">{currentTime}</span>
          <div className="flex items-center gap-2">
            {isOnline ? (
              <>
                <Wifi className="w-4 h-4 text-green-500" />
                <span className="text-xs text-green-600">Онлайн</span>
                <span className="text-xs text-gray-400">/</span>
                <span className="text-xs text-gray-500">Синхр. {lastSync}</span>
              </>
            ) : (
              <>
                <WifiOff className="w-4 h-4 text-amber-500" />
                <span className="text-xs text-amber-600">Офлайн</span>
              </>
            )}
          </div>
        </div>
      </div>

      <div className="p-4 space-y-4 pb-24">
        {/* Приветствие и задачи на сегодня */}
        <div className="bg-gradient-to-r from-[#FF6F1C] to-[#FF8F4C] rounded-2xl p-5 text-white">
          <p className="text-white/90 text-sm mb-1">Здравствуйте, Алексей</p>
          <div className="flex items-baseline gap-2">
            <span className="text-3xl font-bold">{activeOrders.length}</span>
            <span className="text-white/80">
              {activeOrders.length === 1 ? 'задача' : 
               activeOrders.length >= 2 && activeOrders.length <= 4 ? 'задачи' : 
               'задач'} на сегодня
            </span>
          </div>
        </div>

        {/* Статистика */}
        <div className="grid grid-cols-2 gap-3">
          {/* Выполнено за месяц */}
          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                  <Calendar className="w-4 h-4 text-blue-600" />
                </div>
                <span className="text-xs text-gray-500">Выполнено за месяц</span>
              </div>
              <p className="text-2xl font-bold">{monthlyCompleted}</p>
            </CardContent>
          </Card>

          {/* Доход */}
          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                  <TrendingUp className="w-4 h-4 text-green-600" />
                </div>
                <span className="text-xs text-gray-500">Доход за месяц</span>
              </div>
              <p className="text-2xl font-bold text-green-600">{monthlyEarnings}</p>
            </CardContent>
          </Card>

          {/* Выполнено за смену */}
          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-purple-100 rounded-full flex items-center justify-center">
                  <CircleCheck className="w-4 h-4 text-purple-600" />
                </div>
                <span className="text-xs text-gray-500">Выполнено за смену</span>
              </div>
              <p className="text-2xl font-bold">{shiftCompleted}</p>
            </CardContent>
          </Card>

          {/* В работе */}
          <Card className="border-0 shadow-sm">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 mb-2">
                <div className="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
                  <Clock className="w-4 h-4 text-orange-600" />
                </div>
                <span className="text-xs text-gray-500">В работе</span>
              </div>
              <p className="text-2xl font-bold">{inProgressOrders.length}</p>
            </CardContent>
          </Card>
        </div>

        {/* Меню */}
        <div className="mt-4">
          <h3 className="text-sm font-medium text-gray-500 mb-3 px-1">МЕНЮ</h3>
          <Card className="border-0 shadow-sm overflow-hidden">
            <CardContent className="p-0">
              {menuItems.map((item, index) => (
                <div key={item.label}>
                  <div className="flex items-center justify-between p-4 hover:bg-gray-50 cursor-pointer transition-colors">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                        <item.icon className="w-4 h-4 text-gray-600" />
                      </div>
                      <div>
                        <p className="font-medium text-sm">{item.label}</p>
                        <p className="text-xs text-gray-400">{item.value}</p>
                      </div>
                    </div>
                    <ChevronRight className="w-4 h-4 text-gray-400" />
                  </div>
                  {index < menuItems.length - 1 && <Separator />}
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}