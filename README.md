[Русский](#Создание-собственной-копии-бота)

[English](#Create-your-own-replica-of-bot)

# Создание собственной копии бота

* ### Скачайте исходный код и скомпилируйте его

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
go build ./cmd/main.go
```

* ### Создайте своего бота на сайте [дискорда](https://discord.com/developers/applications) и скопируйте оттуда application id и token

* ### В той же папке, где находиться исполняяем файл бота создайте файл .env

* ### Скачайте [ffmpeg](https://ffmpeg.org/download.html) и [yt-dlp](https://github.com/yt-dlp/yt-dlp) (они должны быть в одной папке). Если вам нужен dj mode скачайте redis

```text
BOT_TOKEN=<token вашего бота>
APP_ID=<application id вашего бота>

TOOLS_PATH=<папка где будут расположены ffmpeg и yd-dlp (оставьте пустым, если ffmpeg и yt-dlp находяться в PATH)>

#Настройка redis (если не нужен dj mode - можно оставить пустыми)
DB_URL=
DB_ID=
DB_USERNAME=
DB_PASSWORD=
```

* ### Запустите бота
Важно: у меня бот не работает при использовании [zapret](https://github.com/Flowseal/zapret-discord-youtube), с впн всё работает нормально, другие способы не проверял

# Create your own replica of bot

* ### Install source code and build it

```shell
git clone https://github.com/BulizhnikGames/discord-music-bot && cd discord-music-bot
go build ./cmd/main.go
```

* ### Create your own bot on [discord](https://discord.com/developers/applications) site and copy from there application id and token

* ### In the same location as executable of bot create .env file

* ### Install [ffmpeg](https://ffmpeg.org/download.html) and [yt-dlp](https://github.com/yt-dlp/yt-dlp) (they must have same location). If you need dj mode also install redis

```text
BOT_TOKEN=<token of your bot>
APP_ID=<application id of your bot>

TOOLS_PATH=<path to folder with ffmpeg yt-dlp (leave empty if ffmpeg and yt-dlp are located in PATH)>

#Redis configuration (if you don't dj mode leave empty)
DB_URL=
DB_ID=
DB_USERNAME=
DB_PASSWORD=
```

* ### Run bot's executable