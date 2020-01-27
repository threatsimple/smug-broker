# getting a slack token

Slack has complicated adding bots just a bit in favor of slack apps.

BUT you can still create a classic slack app for now.

Visit https://api.slack.com/apps?new_classic_app=1 and then complete the "Create a Slack App (Classic)" form.

From there, click **App Home**, then click **Add Legacy Bot User**.

Click **Install app**

On the following page, click **Allow**.

Under the **Install App** settings, copy your **Bot User OAuth Access Token**.
It will look like `xoxb-...`.

Add that token to the `smug.yaml` file under the slack section, or set it in
your environment under `SMUG_SLACK_TOKEN`.

See the `config.md` file for more help from here.

