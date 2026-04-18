'use client';

import { useOrderStore } from '@/store/order-store';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { MapPin, Wrench, CircleCheck, AlertCircle, WifiOff, Navigation, List, Map } from 'lucide-react';
import Link from 'next/link';
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { useState } from 'react';

// Обновленный statusConfig с поддержкой en_route
const statusConfig: Record<string, { label: string; variant: 'default' | 'secondary' | 'outline' | 'destructive'; color: string }> = {
  new: { 
    label: 'Новое', 
    variant: 'default', 
    color: 'bg-green-500'
  },
  en_route: { 
    label: 'В пути', 
    variant: 'secondary', 
    color: 'bg-purple-500'
  },
  in_progress: { 
    label: 'В работе', 
    variant: 'secondary', 
    color: 'bg-orange-500'
  },
  completed: { 
    label: 'Завершён', 
    variant: 'outline', 
    color: 'bg-green-500'
  },
  sync_pending: { 
    label: 'Ожидает сети', 
    variant: 'destructive', 
    color: 'bg-yellow-500'
  },
};

export default function HomePage() {
  const orders = useOrderStore((state) => state.orders);
  const activeOrders = orders.filter(o => o.status !== 'completed');
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState('list');

  // Фильтрация заказов по поиску
  const filteredOrders = activeOrders.filter(order => 
    order.equipment.toLowerCase().includes(searchQuery.toLowerCase()) ||
    order.address.toLowerCase().includes(searchQuery.toLowerCase()) ||
    order.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    order.id.includes(searchQuery)
  );

  return ( 
    <div className="min-h-screen bg-gray-50">
      {/* Фиксированный хедер */}
      <div className="sticky top-0 z-20 bg-white border-b">
        <div className="p-4">
          <header className="flex justify-between items-center mb-4">
            <h1 className="text-2xl font-bold tracking-tight">
              <span className="text-[#FF6F1C]">ЗАДАНИЯ</span>
            </h1>
          </header>

          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            {/* Табы */}
            <TabsList className="bg-[#F5F5F5] h-9 rounded-[12px] p-[2px] w-full mb-4">
              <TabsTrigger 
                value="list" 
                className="flex-1 data-[state=active]:bg-[#FF6F1C] data-[state=active]:text-white h-8 rounded-[10px]"
              >
                <List className="w-4 h-4 mr-2" />
                Список
              </TabsTrigger>
              <TabsTrigger 
                value="map" 
                className="flex-1 data-[state=active]:bg-[#FF6F1C] data-[state=active]:text-white h-8 rounded-[10px]"
              >
                <Map className="w-4 h-4 mr-2" />
                На карте
              </TabsTrigger>
            </TabsList>

            {/* Поиск */}
            <div className="relative mb-4">
              <Input 
                placeholder="Поиск" 
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full h-10 pl-10 bg-white border-gray-200"
              />
              <svg 
                className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>

            {/* Контент табов */}
            <TabsContent value="list" className="mt-0">
              <div className="space-y-3 pb-20">
                {filteredOrders.length === 0 && (
                  <Card className="border-dashed">
                    <CardContent className="py-10 text-center text-muted-foreground">
                      {searchQuery ? 'Ничего не найдено' : 'Нет активных заказов'}
                    </CardContent>
                  </Card>
                )}

                {filteredOrders.map((order) => {
                  const config = statusConfig[order.status] || statusConfig.new;
                  
                  return (
                    <Link key={order.id} href={`/order/${order.id}`}>
                      <Card className="active:scale-[0.99] transition-transform cursor-pointer hover:shadow-md bg-white border-0 shadow-sm mb-5">
                        <CardHeader className="pb-2">
                          <div className="flex justify-between items-start">
                            <div className="flex-1">
                              <p className="text-xs text-gray-400 mb-1">#{order.id.slice(0, 12)}</p>
                              <CardTitle className="text-base font-medium line-clamp-2">
                                {order.equipment}
                              </CardTitle>
                            </div>
                            <Badge className={`ml-2 shrink-0 ${config.color} text-white border-0`}>
                              {config.label}
                            </Badge>
                          </div>
                        </CardHeader>
                        <CardContent>
                          <div className="space-y-1 text-sm">
                            <div className="flex items-start gap-2 text-gray-500">
                              <MapPin className="w-4 h-4 mt-0.5 shrink-0" />
                              <span className="line-clamp-2">{order.address}</span>
                            </div>
                            <p className="text-sm mt-2 text-gray-600 line-clamp-2">{order.description}</p>
                            
                            {/* Индикаторы прогресса */}
                            <div className="flex gap-4 mt-3 pt-2 border-t border-gray-100">
                              <div className="flex items-center gap-1">
                                <div className={`w-2 h-2 rounded-full ${order.photoBefore ? 'bg-green-500' : 'bg-gray-300'}`} />
                                <span className="text-xs text-gray-500">До</span>
                              </div>
                              <div className="flex items-center gap-1">
                                <div className={`w-2 h-2 rounded-full ${order.photoAfter ? 'bg-green-500' : 'bg-gray-300'}`} />
                                <span className="text-xs text-gray-500">После</span>
                              </div>
                              <div className="flex items-center gap-1">
                                <div className={`w-2 h-2 rounded-full ${order.signature ? 'bg-green-500' : 'bg-gray-300'}`} />
                                <span className="text-xs text-gray-500">Подпись</span>
                              </div>
                            </div>

                            {!order.synced && order.status === 'completed' && (
                              <div className="flex items-center gap-1 text-amber-500 mt-2">
                                <WifiOff className="w-3 h-3" />
                                <span className="text-xs">Не отправлен в 1С</span>
                              </div>
                            )}
                          </div>
                        </CardContent>
                      </Card>
                    </Link>
                  );
                })}
              </div>
            </TabsContent>

            <TabsContent value="map" className="mt-0">
              <div className="h-[calc(100vh-280px)] w-full rounded-lg overflow-hidden border bg-white ">
                <img src="/image 39.jpg" className=""/>
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </div>

      
    </div>
  );
}