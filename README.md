# Nextcloud App

A Nextcloud app for Mattermost.

### First steps

#### Setting up
1. Run Mattermost server https://github.com/mattermost/mattermost-server/blob/master/README.md
2. Install/Enable Apps plugin  https://github.com/mattermost/mattermost-plugin-apps
3. Run NC server `docker run -d -p 8081:80 nextcloud:latest`
4. Crete NC admin account via opening NC link http://localhost:8081
   <br /> `Note`: After user creation install recommended apps 
5. Inside docker container find config.php and add NC ngrok url to trusted domains
> docker exec -it %NEXTCLOUD_SERVER_DOCKER_CONTAINER_ID% /bin/sh  <br />
   cd config <br />
   apt-get update <br />
   apt-get install nano <br />
   nano config.php <br />
 

> 'trusted_domains' => <br />
array ( <br />
0 => 'localhost:8081', <br />
1 => '%YOUR_NC_DOMAIN%.ngrok.io',<br />
),

`Note`: For HTTPS connection add to config.php
> 'overwriteprotocol' => 'https', <br />
6. Register a NC app - http(s)://YOUR_NC_URL/settings/admin/security
    * as a callback url use - http(s)://YOUR_MM_SERVER/plugins/com.mattermost.apps/apps/nextcloud/oauth2/remote/complete
    * Copy the client secret for a future step
7. Run Nextcloud integration server app
    * Run `docker-compose up` in the root of the Nextcloud App repository
8. In mattermost channel run `/apps install http http://localhost:8082/manifest.json`
   <br /> `Note`: For http deploy type JWT_SECRET env variable contains value for Outgoing JWT Secret field 

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

#### Lambda configuration

Increase RAM and timeout ( Lambda -> Configuration -> General Configuration)

Add environmental variables:   <br /> 
( Lambda-> Configuration -> Environment variables) <br />
CHUNK_FILE_SIZE_MB <br />
MAX_FILE_SIZE_MB <br />
MAX_FILES_SIZE_MB <br />
GIN_MODE=release <br />
MAX_REQUEST_RETRIES=3 <br />

#### HTTP configuration
Add environmental variables:   <br />
JWT_SECRET=secret <br />
STATIC_FOLDER=static <br />
APP_TYPE=HTTP <br />
CHUNK_FILE_SIZE_MB=15 <br />
MAX_FILE_SIZE_MB=25 <br />
MAX_FILES_SIZE_MB=50 <br />
APP_URL=http://localhost:8082 <br />
PORT=8082 <br />
GIN_MODE=release <br />
MAX_REQUEST_RETRIES=3 <br />
