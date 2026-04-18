export default function ChatPage() {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Чат с офисом</h1>
      <div className="space-y-4">
        <div className="bg-gray-100 p-4 rounded-lg">
          <p className="text-sm text-gray-600">Менеджер</p>
          <p>Как проходит ремонт на Ленина 15?</p>
        </div>
        <div className="bg-[#FF6F1C]/10 p-4 rounded-lg ml-auto max-w-[80%]">
          <p className="text-sm text-gray-600">Вы</p>
          <p>Заканчиваю, нужна запчасть 1/2"</p>
        </div>
      </div>
    </div>
  );
}