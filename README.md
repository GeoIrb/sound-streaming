# sound-ethernet-streaming

Стриминг аудио по UDP

server - сервер для раздачи аудиосигнала

client - клиент для приема и воспроизведения аудиосигнала

## ToDo

Сервер

- [X] Из wav файла
- [X] Считывание с микрофона
- [X] Написание пакета для работы с микрофоном
- [ ] Наложение 2х дорожек

Клиент

- [X] Выбор звуковой карты
- [X] Кэш
- [ ] Регулировка громкости
- [ ] Рефакторинг

## Запуск сервера

1. Скачать проект на машину, на которой будет развернут сервер

        git clone git@github.com:GeoIrb/sound-ethernet-streaming.git
2. Поместите аудиофайл, который необходимо будет стримить в папку `audio/`

3. Собрать образ сервера

        make build-server tag=IMAGE-NAME
4. Запуск сервера

        docker run -d --rm \
        -p PORT:PORT \ 
        -e ENVIROMENTS \ 
        IMAGE-NAME

**PORT** - порт, на который будет раздача (возможно это лишнее)

**ENVIROMENTS** - переменные окружения

- FILE=/audio/`FILE`.wav - файл для стримминга
- DST_ADDRESS="IP:PORT" - на какой IP и на какой PORT будет рассылка, по умолчанию 255.255.255.255:8080 - рассылка по всей сети на порт 8080

## Запуск клиент

1. Скачать проект на машину, на которой будет развернут клиент

        git clone git@github.com:GeoIrb/sound-ethernet-streaming.git
2. Собрать образ клиент

        make build-client tag=IMAGE-NAME
3. Запуск клиента

        docker run -d --rm \
        -p 0.0.0.0:PORT:PORT -p 0.0.0.0:PORT:PORT/udp \
        --device /dev/snd \
        -e ENVIROMENTS \
        IMAGE-NAME

**PORT** - порт, на котором будет работать клиент

**ENVIROMENTS** - переменные окружения

- PORT - порт, на котором будет работать клиент
- PLAYBACK_DEVICE_NAME - устройство, на котором будет воспроизводиться принятый аудио сигнал
