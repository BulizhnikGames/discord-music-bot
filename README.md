[Русский](#Как-пользоваться)

[English](#Guide)

# Как пользоваться

## Добавление моего бота на ваш сервер

**Просто перейдите по [ссылке](https://discord.com/oauth2/authorize?client_id=1332373974729887744&permissions=274881149952&integration_type=0&scope=applications.commands+bot)**

## Создание собственной копии бота

* ### Создайте своего бота на сайте [Discord](https://discord.com/developers/applications) и скопируйте оттуда application ID и token

* ### Скачайте исходный код

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
```

### Далее идут 2 возможных варианта создания копии бота

### 1. с помощью Docker (рекомендуемый способ)

* #### В папке проекта создайте .env файл

```text
BOT_TOKEN=<token вашего бота>
APP_ID=<application id вашего бота>
```

* #### В файле docker-compose.yml можно настроить параметры redis и выбрать папку, куда будут сохраняться логи, однако их можно оставить как есть

* #### Запустите докер контейнеры

```shell
docker-compose up --build
```

### 2. без использования Docker

* #### Скачайте [ffmpeg](https://ffmpeg.org/download.html) и [yt-dlp](https://github.com/yt-dlp/yt-dlp). Они должны находиться в одной папке. Рекомендую заменить содержимое папки /tools на скачанные файлы.

* #### В папке проекта создайте .env файл

```text
BOT_TOKEN=<token вашего бота>
APP_ID=<application id вашего бота>

TOOLS_PATH=<папка, где будут расположены ffmpeg и yd-dlp (оставьте пустым, если ffmpeg и yt-dlp находятся в PATH)>

LOGS_PATH=<папка, где будут храниться логи серверов (если оставить пустым, то будут выводиться в StdOut)>

#Настройка redis (если не нужен DJ mode, можно оставить пустыми)
REDIS_HOST=<redis ip>
REDIS_PORT=
REDIS_USERNAME=
REDIS_PASSWORD=
REDIS_DB_ID=
```

* #### Скомпилируйте и запустите бота

```shell
go build -o discordbot ./cmd/musicbot/main.go && ./discordbot
```

**Важно:** у меня бот не работает при использовании [zapret](https://github.com/Flowseal/zapret-discord-youtube), с впн всё работает нормально, другие способы я не проверял

# Guide

## Add my bot to your discord server

**Just follow the [link](https://discord.com/oauth2/authorize?client_id=1332373974729887744&permissions=274881149952&integration_type=0&scope=applications.commands+bot)**

## Create your own copy of bot

* ### Create your own bot on [Discord](https://discord.com/developers/applications) site and copy its application ID and token

* ### Download source code

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
```

### There are 2 possible options next

### 1. with Docker (recommended)

* #### Create .env file in project's directory

```text
BOT_TOKEN=<token of your bot>
APP_ID=<application id of your bot>
```

* #### In docker-compose.yml file you can set redis parameters and choose directory where logs will be stored, but you can leave them as they are

* #### Run docker containers

```shell
docker-compose up --build
```

### 2. without Docker

* #### Download [ffmpeg](https://ffmpeg.org/download.html) and [yt-dlp](https://github.com/yt-dlp/yt-dlp). They must be stored in the same directory. I recommend replacing executables in /tools with the ones you installed.

* #### Create .env file in project's directory

```text
BOT_TOKEN=<token of your bot>
APP_ID=<application id of your bot>

TOOLS_PATH=<path to the directory with ffmpeg and yt-dlp (leave empty if ffmpeg and yt-dlp are located in PATH)>

LOGS_PATH=<path to directory with logs from servers (if you leave this variable empty, logs will be printed in StdOut)>

#Redis configuration (if you don't need DJ mode, leave empty)
REDIS_HOST=<redis ip>
REDIS_PORT=
REDIS_USERNAME=
REDIS_PASSWORD=
REDIS_DB_ID=
```

* #### Build and run your bot

```shell
go build -o discordbot ./cmd/musicbot/main.go && ./discordbot
```