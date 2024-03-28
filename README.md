# Introduction

This Discord bot will expose a command to protect messages from automated systems. It will require from the users
to pass the challenge (manual interaction + optionally a captcha if configured).
This Discord bot was created as a learning experience, although I try to provide it at least as a template project to
use,
It is not meant to be a sophisticated distributed enterprise application with complex features,
just a simple, yet usable Bot that is run from one single server.

# How to

1. Clone this repository and change into the directory

```bash
$ git clone https://github.com/melardev/discord-message-protect
$ cd discord-message-protect
```

2. Copy the sample config into a config.json file:

```
$ cp config_sample.json config.json
```

3. Replace the database password with a password you would like to give to the database. Password must not contain `@`

```
"database": {
    "password": "<DB_PASSWORD>"
}
```

4. Make sure also to update the `docker-compose.yaml` with that same password you generated previously:

```yaml
    environment:
      - MYSQL_DATABASE=discord_protect
      - MYSQL_ROOT_PASSWORD=<DB_PASSWORD>
```

5. Create a Discord bot and add it to your server, as show in the following Youtube video, You are interested up to 4:
    12.
   https://www.youtube.com/watch?v=hoDLj0IzZMU&ab_channel=Indently

6. The bot token is confidential, never share it with anyone,
   copy the bot token and paste it in the config.json file in the `bot_discord` field:

```
"bot_token": "<YOUR_BOT_TOKEN>",
```

7. You will also need to take the Application ID of your bot and fill the `app_id` in the config.json, this id
   is not secret, but there is no point in sharing it either.
   ![discord_app_id.png](excluded%2Fimages%2Fdiscord_app_id.png)

```
"app_id": "<YOUR_APP_ID>"
```

8. Now we need a server to run our bot 24/7, You can choose any cloud provider you want, Amazon AWS, Microsoft Azure,
   Google Cloud platform or any other, personally, for small projects where I want simplicity and cheaper prices
   I always pick DigitalOcean, I have been using it for years and can say it is a quality service, blasting fast network
   with cheaper prices than other cloud providers offering you the same. If you choose DigitalOcean,
   You must create an account on https://www.digitalocean.com and create a Virtual Private Servcer(VPS) or as they call
   it: Droplet.
   This video shows you how. How much RAM and CPU you need is up to your application needs, my guess is this bot
   does not need much at all, the cheaper server should do the job well, but I don't have a discord of
   hundreds/thousands of users
   to stress test it, so it is just a guess.
   If you want to be sure, the 2GB Ram / 2 CPUs should be more than enough. Now they introduced the NVMe SSD servers,
   they offer
   greater filesystem speed (faster database access for our application), but I don't think you really need it.
   https://www.youtube.com/watch?v=uXDlnEUow0A&ab_channel=OdooMates
9. Access our server via SSH as the video shows.
10. This bot is shipped as Docker container app, deployed through docker-compose, so we must install `docker`
    and `docker-compose` as
    explained
    on:
    - https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-20-04
    - https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-compose-on-ubuntu-20-04
11. Create the following two directories, one will be for the database, the other for the app logs itself:

```bash
$ mkdir -p /opt/discord-message-protect/app
$ mkdir -p /opt/discord-message-protect/db
```

12. build the containers:

```
$ docker-compose build --no-cache
```

13. Launch the application

```bash
docker-compose up
```

14. You must see the application starting without any errors, it is important to see this message indicating a
    successful
    connection to the database:
    ![db_success.png](excluded%2Fimages%2Fdb_success.png)

If it fails, it will show `Database - retrying again in 3 seconds`, and it will retry a
couple more times,
if it fails, it will finally show: `Database - Falling back to SQLite"`. You don't want to see this message, if it
does, stop
the application with `Ctrl+C`, then rerun again `docker-compose up`, if the error persists please let me know creating
a GitHub issue.

# Features

- Protect messages requiring user to click to reveal the message.

https://user-images.githubusercontent.com/18094815/211221108-ba6c33e3-74cc-4471-9e7b-024501a8fe03.mp4
  
- Protect with captcha.
 
https://user-images.githubusercontent.com/18094815/211221103-9ec50fbc-698d-4913-a817-ae89c05ca423.mp4

- Pollute messages with unique identifiers.
  
https://user-images.githubusercontent.com/18094815/211221100-dee2ae25-3cf0-426e-ba50-656b76a06e44.mp4

- Protected messages persisted across applications restarts.

# Flow

1. Author creates secret
2. Bot "memorizes" the secret
3. Creates interact button
4. User Interacts
5. User gets link (Optional)
6. User visits link + resolves challenge (Optional)
7. User gets secret

# Pollution

The bot can "tint" or "pollute" messages with a unique identifier per user.
This can uncover which user is exfiltrating messages to other servers.
It is possible because the protected messages are sent to each user privately, so it is possible
to send each one a message slightly different. It is a well known technique used in companies to
detect rogue employees, for example Tesla used it to uncover the employee selling the news to the
media ([Full story](https://www.ndtv.com/world-news/elon-musk-explains-how-tesla-caught-employee-leaking-data-3433802)).
The draw-down is rogue users can program their bots to detect these ids and remove them before leaking them, this is
why multiple pollution strategies are implemented, some strategies are easier to detect and neutralise, some are harder
and will force the rogue user's bot to remove a big chunk of the message leaving it meaningless.

The pollution procedure is, legitimate user creates a protected message, the pollution mechanism adds some unique
identifiers per user basis,
the rogue user's bot exfiltrates the polluted message with the pollution indicators, you go to the server where the
messages are
being exfiltrated,
retrieve the pollution indicators, those identify a single and unique message tied to a specific user, you check the
logs
to see which user was given those indicators, once you find it, you know who is the rogue user.

To check the pollution logs to see which user was assigned a specific set of pollution idicators you check the
file `/opt/discord-message-protect/app/pollution.log`:

```
$ cat /opt/discord-message-protect/app/pollution.log
[...]
06/01/2023 16:02:24 Info - Applied random_string strategy, User: melardev#8373, Id: 832d58a5-8a10-4574-a1dd-5bcbf9ba6e6b, Indicators: [fgDsc WD8 2qNfH 5a84j]
06/01/2023 16:02:55 Info - Applied random_string strategy, User: melardev#8373, Id: f0c58a0b-9fa1-41c8-8e47-b0da62e904c6, Indicators: [kwzD h9h fhfUV S9jZ]
06/01/2023 16:03:04 Info - Applied random_string strategy, User: melardev#8373, Id: f0c58a0b-9fa1-41c8-8e47-b0da62e904c6, Indicators: [bhV3 C5A X39]
[...]
```

The `Id` is the secret id, the indicators are the pollution indicators assigned to the user for that secret.

# TODO

- Graceful exit, we must have a mean to exit all app components(secret manager, session manager, etc.) gracefully
  without
  corrupting the work they are engaged in at the time.
- Protect by role, ability to use the protect command may be restricted by role.
- Edit the secrets
- Rate Limit related features need to be implemented/improved.
- Implement the Random IndicatorPosition for all pollution strategies.
- provide ability to choose pollution method via command argument, overriding the current's pollution config settings
- Clean up the architecture, there are some relations or assumptions that should not exist, like http server knowing
  about the captcha html code, that should be left to the specific captcha impl as captcha usage differs.
- More validation code
- cleaner logs
- HTTPs support
