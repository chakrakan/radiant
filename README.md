# radiant

Grab your valorant rank and display it in a gist

Shows up like this [gist here](https://gist.github.com/chakrakan/f83781805df37a7ee3ee1186330d67c8)

In order to get this working:
1. Clone the repo
2. Setup a `.env` file at the root of the project with the following env vars:

```sh
GITHUB_TOKEN=<Your Github Token with Gist edit access>
GIST_ID=<ID of the Gist you created to host the stats> # this project does not setup a gist for you
TRACKER_PROFILE_ID=<Your tracker profile ID> # e.g. mine is adorn#1625
```

3. Run the scraper with `go run radiant.go`
4. ...
5. Profit???

### To-do

- [] CI config so repo can be forked and leveraged directly instead of local cloning
