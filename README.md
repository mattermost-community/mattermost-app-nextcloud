# Nextcloud App

A Nextcloud app for Mattermost.

### First steps

#### Setting up
1. Run Mattermost server https://github.com/mattermost/mattermost-server/blob/master/README.md
2. Install/Enable Apps plugin  https://github.com/mattermost/mattermost-plugin-apps
3. Run NC server `docker run -d -p 8081:80 nextcloud:latest`
4. Inside docker container in config.php add NC ngrok url to trusted domains
5. Register a NC app - http://YOUR_NC_URL/settings/YOUR_ADMIN_USER/security
    * as a callback url use - http://YOUR_MM_SERVER/plugins/com.mattermost.apps/apps/nextcloud/oauth2/remote/complete
    * Copy the client secret for a future step
6. Run Nextcloud integration server app
    * Run `docker-compose up` in the root of the Nextcloud App repository
7. In mattermost channel run `/apps install http http://localhost:8082/manifest.json`

#### Link MM account with Nextcloud

1. Run command `/nextcloud configure` and provide Nextcloud instance url ,client id and client secret from register a Nextcloud app step http://YOUR_NC_URL/settings/YOUR_ADMIN_USER/security
2. Run command `/nextcloud connect` and open link from bot response for linking mm with Nextcloud account

### Usage

1. `/nextcloud share` - share public link for user file in MM channel
2. `/nextcloud calendars` -  show user calendars
3. Message actions - Upload file to Nextcloud


### Building aws bundle

In project root run command  `make dist-aws`

### Deploy aws bundle with appsctl
https://developers.mattermost.com/integrate/apps/deploy/deploy-aws/

`go run ./cmd/appsctl aws deploy -v bundle-aws.zip`

####Lambda configuration

Increase RAM and timeout ( Lambda -> Configuration -> General Configuration)

Add environmental variables CHUNK_FILE_SIZE_MB and GIN_MODE=release ( Lambda-> Configuration -> Environment variables)
