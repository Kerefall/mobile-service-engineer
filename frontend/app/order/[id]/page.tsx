'use client';

import { useOrderStore, type Order } from '@/store/order-store';
import { useParams, useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { 
  MapPin, Calendar, Wrench, Package, FileText,
  Camera, PenTool, CheckCircle, Navigation, Clock,
  ChevronRight, Upload, Paperclip,
} from 'lucide-react';
import { useState, useRef } from 'react';
import { toast } from 'sonner';
import SignatureCanvas from 'react-signature-canvas';
import Image from 'next/image';
import { format } from 'date-fns';
import { ru } from 'date-fns/locale';

// Статусы и их цвета
const statusConfig = {
  new: { label: 'Новое', variant: 'default' as const, color: 'bg-blue-500' },
  in_progress: { label: 'В работе', variant: 'secondary' as const, color: 'bg-orange-500' },
  en_route: { label: 'В пути', variant: 'secondary' as const, color: 'bg-purple-500' },
  completed: { label: 'Завершён', variant: 'outline' as const, color: 'bg-green-500' },
};

export default function OrderDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;
  
  const order = useOrderStore((state) => state.orders.find(o => o.id === id));
  const updateOrder = useOrderStore((state) => state.updateOrder);
  const addPhoto = useOrderStore((state) => state.addPhoto);
  const addSignature = useOrderStore((state) => state.addSignature);
  const completeOrder = useOrderStore((state) => state.completeOrder);
  const syncOrder = useOrderStore((state) => state.syncOrder);

  const [isSigning, setIsSigning] = useState(false);
  const sigPadRef = useRef<SignatureCanvas>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [photoType, setPhotoType] = useState<'before' | 'after'>('before');
  
  // Состояние для отслеживания времени в пути (для демо)
  const [eta, setEta] = useState({ minutes: 2, meters: 177 });

  if (!order) return <div className="p-4">Заказ не найден</div>;

  const handleTakePhoto = (type: 'before' | 'after') => {
    setPhotoType(type);
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        addPhoto(order.id, photoType, reader.result as string);
        toast.success(`Фото "${photoType === 'before' ? 'До' : 'После'}" добавлено`);
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSaveSignature = () => {
    if (sigPadRef.current && !sigPadRef.current.isEmpty()) {
      const signatureData = sigPadRef.current.toDataURL();
      addSignature(order.id, signatureData);
      setIsSigning(false);
      toast.success('Подпись клиента сохранена');
    } else {
      toast.error('Подпись не может быть пустой');
    }
  };

  const handleStartWork = () => {
    // @ts-ignore - добавим статус en_route
    updateOrder(order.id, { status: 'en_route', startedAt: new Date() });
    toast.success('Вы начали выполнение задания');
  };

  const handleArrived = () => {
    updateOrder(order.id, { status: 'in_progress' });
    toast.success('Вы на месте');
  };

  const handleComplete = async () => {
    if (!order.photoBefore || !order.photoAfter) {
      toast.error('Необходимо сделать фото До и После');
      return;
    }
    if (!order.signature) {
      toast.error('Необходимо получить подпись клиента');
      return;
    }
    
    completeOrder(order.id);
    
    toast.promise(syncOrder(order.id), {
      loading: 'Отправка данных в 1С...',
      success: 'Заказ закрыт и выгружен в 1С',
      error: 'Ошибка отправки. Данные сохранены локально',
    });
    
    router.push('/');
  };

  const openMap = () => {
    const query = encodeURIComponent(order.address);
    window.open(`https://yandex.ru/maps/?text=${query}`, '_blank');
  };

  // Моковые запчасти
  const mockParts = [
    'Болтик', 'Гайка', 'Шуруп', 'Новые двери лифта', 'Рулевое колесо', 'Новый самовар'
  ];

  // Рендер в зависимости от статуса
  const renderContent = () => {
    const status = order.status as 'new' | 'en_route' | 'in_progress' | 'completed';
    
    return (
      <>
        {/* Общая информация о заказе */}
        <div className="space-y-3">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-muted-foreground">#{order.id.slice(0, 12)}</p>
              <h1 className="text-xl font-semibold">{order.equipment}</h1>
              <p className="text-muted-foreground">{order.address.split(',').pop()}</p>
            </div>
            <Badge className={statusConfig[status]?.color || 'bg-gray-500'}>
              {statusConfig[status]?.label || status}
            </Badge>
          </div>

          <Card>
            <CardContent className="p-4 space-y-4">
              {/* Описание */}
              <div>
                <p className="text-sm font-medium mb-1">Описание</p>
                <p className="text-sm text-muted-foreground">{order.description}</p>
              </div>

              <Separator />

              {/* Адрес */}
              <div>
                <p className="text-sm font-medium mb-1">Адрес</p>
                <div className="flex items-start gap-2">
                  <MapPin className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
                  <p className="text-sm">{order.address}</p>
                </div>
              </div>

              <Separator />

              {/* Дата и время */}
              <div>
                <p className="text-sm font-medium mb-1">Дата и время</p>
                <div className="flex items-center gap-2">
                  <Calendar className="w-4 h-4 text-muted-foreground" />
                  <p className="text-sm">
                    {format(new Date(), 'd MMMM, HH:mm', { locale: ru })}
                  </p>
                </div>
              </div>

              <Separator />

              {/* Запчасти (список) */}
              <div>
                <p className="text-sm font-medium mb-2">Запчасти</p>
                <div className="flex flex-wrap gap-2">
                  {mockParts.map((part, idx) => (
                    <span 
                      key={idx}
                      className="px-3 py-1 bg-gray-100 rounded-full text-sm"
                    >
                      {part}
                    </span>
                  ))}
                </div>
              </div>

              {/* Акт на выданные запчасти */}
              <Button variant="outline" className="w-full h-10 justify-start" size="sm">
                <Paperclip className="w-4 h-4 mr-2" />
                Акт на выданные запчасти
              </Button>
            </CardContent>
          </Card>
        </div>

        {/* Контент в зависимости от статуса */}
        
        {/* СТАТУС 1: НОВОЕ */}
        {status === 'new' && (
          <div className="mt-4">
            <Button 
              className="w-full h-14 text-base bg-[#FF6F1C] hover:bg-[#FF6F1C]/90"
              onClick={handleStartWork}
            >
              Приступить к выполнению
            </Button>
          </div>
        )}

        {/* СТАТУС 2: В ПУТИ */}
        {status === 'en_route' && (
          <div className="mt-4 space-y-4">
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <div className="flex items-center gap-2 text-green-700 mb-2">
                <CheckCircle className="w-5 h-5" />
                <span className="font-medium">Маршрут построен</span>
              </div>
              <Button 
                className="w-full h-14 text-base bg-[#FF6F1C] hover:bg-[#FF6F1C]/90"
                onClick={handleArrived}
              >
                Я на месте
              </Button>
              <div className="flex items-center justify-center gap-2 mt-3 text-sm text-muted-foreground">
                <Clock className="w-4 h-4" />
                <span>Еще {eta.minutes} мин · {eta.meters}м</span>
              </div>
            </div>
            <Button 
              variant="outline" 
              className="w-full"
              onClick={openMap}
            >
              <Navigation className="w-4 h-4 mr-2" />
              Открыть карту
            </Button>
          </div>
        )}

        {/* СТАТУС 3: В РАБОТЕ (форма выполнения) */}
        {status === 'in_progress' && (
          <div className="mt-4 space-y-6 pb-4">
            {/* Фото до/после */}
            <div>
              <div className="flex items-center gap-2 mb-3">
                <Camera className="w-5 h-5 text-[#FF6F1C]" />
                <h3 className="font-semibold">Выполнение заказа</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-4">
                Сделайте фото до и после работ. Сначала сфотографируйте объект до начала работ 
                (проблему, неисправность), затем после завершения — результат.
              </p>
              
              <div className="space-y-3">
                {/* Фото ДО */}
                <div>
                  <p className="text-sm font-medium mb-2">
                    Фото до начала работы <span className="text-red-500">*</span>
                  </p>
                  {order.photoBefore ? (
                    <div className="relative h-40 rounded-lg overflow-hidden border">
                      <Image src={order.photoBefore} alt="До" fill className="object-cover" />
                    </div>
                  ) : (
                    <Button 
                      variant="outline" 
                      className="w-full h-20 border-dashed"
                      onClick={() => handleTakePhoto('before')}
                    >
                      <Upload className="w-5 h-5 mr-2" />
                      Загрузить
                    </Button>
                  )}
                </div>

                {/* Фото ПОСЛЕ */}
                <div>
                  <p className="text-sm font-medium mb-2">
                    Фото по окончании работы <span className="text-red-500">*</span>
                  </p>
                  {order.photoAfter ? (
                    <div className="relative h-40 rounded-lg overflow-hidden border">
                      <Image src={order.photoAfter} alt="После" fill className="object-cover" />
                    </div>
                  ) : (
                    <Button 
                      variant="outline" 
                      className="w-full h-20 border-dashed"
                      onClick={() => handleTakePhoto('after')}
                    >
                      <Upload className="w-5 h-5 mr-2" />
                      Загрузить
                    </Button>
                  )}
                </div>
              </div>
            </div>

            <Separator />

            {/* Подпись клиента */}
            <div>
              <div className="flex items-center gap-2 mb-3">
                <PenTool className="w-5 h-5 text-[#FF6F1C]" />
                <h3 className="font-semibold">Подпись клиента</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-4">
                Получите подпись заказчика. После того как работа выполнена, передайте планшет 
                или телефон клиенту. Он расписывается на экране.
              </p>
              <Button variant="outline" className="w-full h-10 justify-start my-3" size="sm">
                <Paperclip className="w-4 h-4 mr-2" />
                Акт выполненных работ
              </Button>

              <div>
                <p className="text-sm font-medium mb-2">
                  Подпись клиента <span className="text-red-500">*</span>
                </p>
                
                {order.signature ? (
                  <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                    <div className="flex items-center gap-2 text-green-700">
                      <CheckCircle className="w-5 h-5" />
                      <span>Подпись получена</span>
                    </div>
                  </div>
                ) : isSigning ? (
                  <div className="space-y-3">
                    <div className="border rounded-lg overflow-hidden bg-white">
                      <SignatureCanvas
                        ref={sigPadRef}
                        canvasProps={{
                          className: 'w-full h-40',
                        }}
                      />
                    </div>
                    <div className="flex gap-2">
                      <Button variant="outline" onClick={() => sigPadRef.current?.clear()}>
                        Очистить
                      </Button>
                      <Button className="flex-1" onClick={handleSaveSignature}>
                        Сохранить
                      </Button>
                    </div>
                  </div>
                ) : (
                  <Button 
                    variant="outline" 
                    className="w-full"
                    onClick={() => setIsSigning(true)}
                  >
                    <PenTool className="w-4 h-4 mr-2" />
                    Подписать
                  </Button>
                )}
              </div>
            </div>

            <Separator />

            {/* Запчасти */}
            <div>
              <div className="flex items-center gap-2 mb-3">
                <Package className="w-5 h-5 text-[#FF6F1C]" />
                <h3 className="font-semibold">Запчасти</h3>
              </div>
              <p className="text-sm text-muted-foreground mb-4">
                Отметьте потраченные запчасти. Выберите из списка детали и расходные материалы, 
                которые вы использовали при ремонте.
              </p>

              <div>
                <p className="text-sm font-medium mb-2">
                  Выберите потраченные запчасти <span className="text-red-500">*</span>
                </p>
                <div className="space-y-2">
                  {mockParts.map((part, idx) => (
                    <label key={idx} className="flex items-center gap-3 p-3 border rounded-lg">
                      <input type="checkbox" className="w-5 h-5 accent-[#FF6F1C]" />
                      <span>{part}</span>
                    </label>
                  ))}
                </div>
              </div>
            </div>

            {/* Кнопка Завершить */}
            <Button 
              className="w-full h-14 text-base bg-[#FF6F1C] hover:bg-[#FF6F1C]/90"
              onClick={handleComplete}
            >
              Завершить
            </Button>
          </div>
        )}
      </>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Хедер с заголовком в зависимости от статуса */}
      <div className="sticky top-0 z-10 bg-white border-b">
        <div className="px-4 py-3">
          <div className="flex items-center gap-3">
            <Button 
              variant="ghost" 
              size="icon" 
              onClick={() => router.back()}
              className="shrink-0"
            >
              ←
            </Button>
            <div className="flex-1">
              {order.status === 'new' && (
                <p className="text-sm">Задание #{order.id.slice(0, 12)}</p>
              )}
              {order.status === 'en_route' && (
                <p className="text-sm text-green-600">
                  Вы начали выполнение задания #{order.id.slice(0, 12)}
                </p>
              )}
              {order.status === 'in_progress' && (
                <p className="text-sm text-green-600">
                  Вы на месте. Задание #{order.id.slice(0, 12)}
                </p>
              )}
            </div>
          </div>
        </div>
        {order.status === 'en_route' && (
          <div className="h-0.5 bg-gray-100">
            <div className="h-full w-1/3 bg-green-500" />
          </div>
        )}
      </div>

      {/* Основной контент */}
      <div className="p-4">
        {renderContent()}
      </div>

      {/* Скрытый инпут для фото */}
      <input 
        type="file" 
        ref={fileInputRef} 
        className="hidden" 
        accept="image/*" 
        capture="environment"
        onChange={handleFileChange}
      />
    </div>
  );
}