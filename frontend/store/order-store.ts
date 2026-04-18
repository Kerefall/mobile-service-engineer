import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

export type OrderStatus = 'new' | 'en_route' | 'in_progress' | 'completed' | 'sync_pending';

export interface OrderPart {
  id: string;
  name: string;
  quantity: number;
}

export interface Order {
  id: string;
  address: string;
  description: string;
  equipment: string;
  status: OrderStatus;
  photoBefore?: string; // Base64
  photoAfter?: string;  // Base64
  parts: OrderPart[];
  selectedParts?: string[]; // Выбранные запчасти из списка
  signature?: string;   // Base64 SVG/PNG
  startedAt?: Date;
  completedAt?: Date;
  synced: boolean;
}

// Моковые данные
const MOCK_ORDERS: Order[] = [
  {
    id: '039398736242',
    address: 'ул. Тверская, 18',
    description: 'Со слов жильцов - лифт застрял между 5 и 6 этажами, двери не открываются. Пассажиров внутри нет. На дисплее индикация ошибки «E34».',
    equipment: 'Сломалась кофемашина · КофеХаус',
    status: 'new',
    parts: [],
    synced: true,
  },
  {
    id: '2',
    address: 'ТЦ "Мега", пав. 203',
    description: 'Течь в системе отопления',
    equipment: 'Радиатор Rifar 500',
    status: 'new',
    parts: [{ id: 'p1', name: 'Кран шаровый 1/2"', quantity: 2 }],
    synced: true,
  },
  {
    id: '3',
    address: 'ул. Пушкина, д. 10, офис 305',
    description: 'Не включается кондиционер, индикатор мигает красным',
    equipment: 'Кондиционер LG Inverter',
    status: 'in_progress',
    parts: [],
    photoBefore: '',
    photoAfter: '',
    synced: false,
  },
];

interface OrderState {
  orders: Order[];
  currentOrderId: string | null;
  
  // Actions
  setCurrentOrder: (id: string) => void;
  updateOrder: (id: string, data: Partial<Order>) => void;
  addPhoto: (id: string, type: 'before' | 'after', photo: string) => void;
  addSignature: (id: string, signature: string) => void;
  togglePart: (id: string, partName: string) => void;
  completeOrder: (id: string) => void;
  resetToMock: () => void;
  
  // Функция имитации отправки в 1С (sync)
  syncOrder: (id: string) => Promise<void>;
}

export const useOrderStore = create<OrderState>()(
  persist(
    (set, get) => ({
      orders: MOCK_ORDERS,
      currentOrderId: null,

      setCurrentOrder: (id) => set({ currentOrderId: id }),
      
      updateOrder: (id, data) => set((state) => ({
        orders: state.orders.map((o) => 
          o.id === id ? { ...o, ...data, synced: false } : o
        ),
      })),

      addPhoto: (id, type, photo) => set((state) => ({
        orders: state.orders.map((o) => 
          o.id === id 
            ? { 
                ...o, 
                ...(type === 'before' ? { photoBefore: photo } : { photoAfter: photo }),
                synced: false 
              } 
            : o
        ),
      })),

      addSignature: (id, signature) => set((state) => ({
        orders: state.orders.map((o) => 
          o.id === id ? { ...o, signature, synced: false } : o
        ),
      })),

      togglePart: (id, partName) => set((state) => ({
        orders: state.orders.map((o) => {
          if (o.id !== id) return o;
          
          const currentSelected = o.selectedParts || [];
          const isSelected = currentSelected.includes(partName);
          
          return {
            ...o,
            selectedParts: isSelected 
              ? currentSelected.filter(p => p !== partName)
              : [...currentSelected, partName],
            synced: false,
          };
        }),
      })),

      completeOrder: (id) => set((state) => ({
        orders: state.orders.map((o) => 
          o.id === id 
            ? { 
                ...o, 
                status: 'completed', 
                completedAt: new Date(),
                synced: false 
              } 
            : o
        ),
      })),

      resetToMock: () => {
        // Очищаем localStorage и устанавливаем моковые данные
        localStorage.removeItem('mobile-engineer-storage');
        set({ 
          orders: MOCK_ORDERS, 
          currentOrderId: null 
        });
      },

      syncOrder: async (id) => {
        console.log(`📤 Отправка заказа #${id} в 1С...`);
        console.log('Данные для отправки:', get().orders.find(o => o.id === id));
        
        // Имитация задержки сети
        await new Promise(resolve => setTimeout(resolve, 1500));
        
        set((state) => ({
          orders: state.orders.map((o) => 
            o.id === id ? { ...o, synced: true } : o
          ),
        }));
        
        console.log(`✅ Заказ #${id} успешно отправлен в 1С`);
      },
    }),
    {
      name: 'mobile-engineer-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ orders: state.orders }),
      version: 1,
      migrate: (persistedState: any, version: number) => {
        // Миграция: если нет данных или старая версия, возвращаем моковые
        if (version === 0 || !persistedState) {
          return { orders: MOCK_ORDERS, currentOrderId: null };
        }
        return persistedState as OrderState;
      },
    }
  )
);

// Дополнительные селекторы для удобства
export const useCurrentOrder = () => {
  const orders = useOrderStore((state) => state.orders);
  const currentId = useOrderStore((state) => state.currentOrderId);
  return orders.find(o => o.id === currentId) || null;
};

export const useOrdersByStatus = (status: OrderStatus) => {
  return useOrderStore((state) => 
    state.orders.filter(o => o.status === status)
  );
};

export const useActiveOrdersCount = () => {
  return useOrderStore((state) => 
    state.orders.filter(o => o.status !== 'completed').length
  );
};

export const usePendingSyncCount = () => {
  return useOrderStore((state) => 
    state.orders.filter(o => !o.synced && o.status === 'completed').length
  );
};