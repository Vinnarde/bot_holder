#!/bin/bash

# Проверка прав root
if [ "$EUID" -ne 0 ]; then
  echo "Этот скрипт требует привилегий root. Запустите его с sudo."
  exit 1
fi

# Параметры установки
INSTALL_DIR="/opt/redirector"
SERVICE_NAME="redirector"
SERVICE_FILE="${SERVICE_NAME}.service"


# Проверка успешной компиляции
if [ ! -f "./redirector" ]; then
  echo "Ошибка: не удалось найти исполняемый файл `redirector`."
  exit 1
fi

# Создаем директорию если не существует
mkdir -p $INSTALL_DIR

# Копируем файлы приложения
echo "Копирование файлов приложения..."
cp redirector $INSTALL_DIR/
cp config.yaml $INSTALL_DIR/
cp -r ./views $INSTALL_DIR/

# Устанавливаем права
echo "Установка прав доступа..."
chown -R www-data:www-data $INSTALL_DIR
chmod -R 755 $INSTALL_DIR
chmod +x $INSTALL_DIR/redirector

# Копируем файл сервиса в systemd
echo "Установка сервиса systemd..."
cp $SERVICE_FILE /etc/systemd/system/

# Перезагружаем systemd
systemctl daemon-reload

# Включаем и запускаем сервис
echo "Запуск сервиса..."
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo "Статус сервиса:"
systemctl status $SERVICE_NAME

echo ""
echo "Установка завершена. Сервис '$SERVICE_NAME' установлен и запущен."
echo "Используйте следующие команды для управления сервисом:"
echo "  systemctl start $SERVICE_NAME    # запустить сервис"
echo "  systemctl stop $SERVICE_NAME     # остановить сервис"
echo "  systemctl restart $SERVICE_NAME  # перезапустить сервис"
echo "  systemctl status $SERVICE_NAME   # посмотреть статус сервиса"
echo "  journalctl -u $SERVICE_NAME      # просмотреть логи сервиса" 