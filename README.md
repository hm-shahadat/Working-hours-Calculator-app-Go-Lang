# Working Hours Tracker (Go)

A small web app — enter Start Time and End Time, and it automatically
calculates the hours worked (handles midnight-crossing shifts correctly
too), auto-saves every day, and shows monthly/yearly totals. The same
logic from the Excel template has been implemented here in Go.

## How it works (architecture)

- **Language:** Pure Go (standard library only — no external packages
  needed, so `go build` works without any internet/proxy issues).
- **Storage:** A simple **JSON file** (`workhours-data.json`) — each
  day's date, start time, and end time are stored there. No database
  server needed, which makes hosting simple and free.
- **Frontend:** Server-rendered HTML plus a bit of vanilla JavaScript
  (AJAX) — when you change a time input, it saves to the server right
  away and updates the duration/total instantly, without reloading
  the page.

## Running locally (for testing)

```bash
go build -o workhours .
./workhours
```

Then open `http://localhost:8080` in your browser to see the current
month's page.

To change the port: `PORT=3000 ./workhours`
To change the data file location: `DATA_FILE=/path/to/file.json ./workhours`

## Deploying live for free (Fly.io — recommended)

Fly.io has a free allowance that lets you host a small app for free
**with persistent storage** (meaning your data survives even if the
app restarts). You don't need to buy a domain — Fly gives you its own
subdomain (`https://your-app.fly.dev`).

1. Install the Fly CLI: https://fly.io/docs/hands-on/install-flyctl/
2. Create an account and log in:
   ```bash
   fly auth signup
   ```
3. From this project folder, run:
   ```bash
   fly launch
   ```
   - It will ask for an app name — give it any name you like.
   - "Would you like to set up a Postgres database?" -> **No**
   - "Would you like to deploy now?" -> **No** (we'll create the volume first)
4. Create a persistent volume (so data isn't lost):
   ```bash
   fly volumes create data --size 1
   ```
5. Add this mounts section to your `fly.toml` file (if it isn't there already):
   ```toml
   [mounts]
     source = "data"
     destination = "/app/data"
   ```
6. Deploy:
   ```bash
   fly deploy
   ```
7. Once deployed, you'll get a link (like `https://your-app.fly.dev`) —
   send that link to your friend. They can open it and start entering
   their time right away.

## Alternative: Render.com (free, but with a caveat)

You can also create a free "Web Service" on Render (connect your GitHub
repo, and it will auto-detect the Docker setup), but Render's **free
tier has an ephemeral local disk** — meaning if the app goes idle and
then spins back up, the **data file may get reset**. So if you use
Render, regularly download and back up `workhours-data.json`, or use
Fly.io instead (steps above), where you get a real persistent volume
for free.

## Data backup

At any time, you can copy the `workhours-data.json` file to keep all
your data backed up. You can access it by SSH-ing into the server
(`fly ssh console` on Fly) or by downloading it.

## New year / month — automatic

No extra setup is needed — for any year/month, just visit the URL with
`?y=2027&m=1` and the blank table for that month will show up
automatically.
