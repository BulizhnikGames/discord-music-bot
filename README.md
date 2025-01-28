[Русский](#Как-пользоваться)

[English](#Guide)

# Как пользоваться

## Добавление моего бота на ваш сервер

**Просто перейдите по [ссылке](https://discord.com/oauth2/authorize?client_id=1332373974729887744&permissions=274881149952&integration_type=0&scope=applications.commands+bot)**

## Создание собственной копии бота

* ### Скачайте исходный код и скомпилируйте его

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
go build ./cmd/main.go
```

* ### Создайте своего бота на сайте [Discord](https://discord.com/developers/applications) и скопируйте оттуда application ID и token

* ### Скачайте [ffmpeg](https://ffmpeg.org/download.html) и [yt-dlp](https://github.com/yt-dlp/yt-dlp) (они должны быть в одной папке). Если вам нужен DJ mode скачайте redis

* ### В той же папке, где находится исполняемый файл бота, создайте файл .env

```text
BOT_TOKEN=<token вашего бота>
APP_ID=<application id вашего бота>

TOOLS_PATH=<папка, где будут расположены ffmpeg и yd-dlp (оставьте пустым, если ffmpeg и yt-dlp находятся в PATH)>

LOGS_PATH=<папка, где будут храниться логи серверов (если оставить пустым, то будут выводиться в StdOut)>

#Настройка redis (если не нужен DJ mode, можно оставить пустыми)
DB_URL=
DB_ID=
DB_USERNAME=
DB_PASSWORD=
```

* ### Запустите бота
**Важно:** у меня бот не работает при использовании [zapret](https://github.com/Flowseal/zapret-discord-youtube), с впн всё работает нормально, другие способы не проверял

# Guide

## Add my bot to your discord server

**Just follow the [link](https://discord.com/oauth2/authorize?client_id=1332373974729887744&permissions=274881149952&integration_type=0&scope=applications.commands+bot)**

## Create your own copy of bot

* ### Download the source code and build it

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
go build ./cmd/main.go
```

* ### Create your own bot on [Discord](https://discord.com/developers/applications) site and copy its application ID and token

* ### Install [ffmpeg](https://ffmpeg.org/download.html) and [yt-dlp](https://github.com/yt-dlp/yt-dlp) (they must be in the same folder). If you need DJ mode also install redis

* ### In the same folder as the bot's executable, create a .env file

```text
BOT_TOKEN=<token of your bot>
APP_ID=<application id of your bot>

TOOLS_PATH=<path to the folder with ffmpeg and yt-dlp (leave empty if ffmpeg and yt-dlp are located in PATH)>

LOGS_PATH=<path to folder with logs from servers (if you leave this variable empty, logs will be printed in StdOut)>

#Redis configuration (if you don't need DJ mode, leave empty)
DB_URL=
DB_ID=
DB_USERNAME=
DB_PASSWORD=
```

* ### Run the bot's executable