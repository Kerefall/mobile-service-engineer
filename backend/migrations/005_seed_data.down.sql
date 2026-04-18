-- Очищаем тестовые данные
DELETE FROM order_parts WHERE order_id IN (1,2,3,4,5);
DELETE FROM orders WHERE id IN (1,2,3,4,5);
DELETE FROM parts WHERE id IN (1,2,3,4,5,6,7,8,9,10);
DELETE FROM engineers WHERE id IN (1,2,3);