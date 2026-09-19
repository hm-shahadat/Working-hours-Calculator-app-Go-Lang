# Working Hours Tracker (Go)

A small web app — enter Start Time and End Time, and it automatically
calculates the hours worked (handles midnight-crossing shifts correctly
too), auto-saves every day, and shows monthly/yearly totals. The same
logic from the Excel template has been implemented here in Go.

## Live App

**https://working-hours-calculator-app-go-lang.onrender.com**

Just open the link above (or send it to anyone who needs to log their
hours) — no installation needed. You can also jump straight to a
specific month with `?y=2026&m=9` in the URL, e.g.:
`https://working-hours-calculator-app-go-lang.onrender.com/month?y=2026&m=9`

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

## Deployed on Render.com

This app is hosted for free on **Render.com** as a Web Service, built
directly from the `Dockerfile` in this repo, connected to the GitHub
repo:
`https://github.com/hm-shahadat/Working-hours-Calculator-app-Go-Lang`

Steps used:

1. Pushed this project to a GitHub repo.
2. On Render.com -> **New +** -> **Web Service** -> connected the
   GitHub repo above.
3. Runtime: **Docker** (auto-detected from the `Dockerfile`).
4. Instance Type: **Free**.
5. Clicked **Create Web Service** — Render built and deployed it
   automatically, giving the live link above.

### Important caveat: free tier disk is ephemeral

Render's free tier does **not** provide persistent disk storage. This
means:

- After ~15 minutes of no visits, the app goes to sleep. The next
  visit takes 30-50 seconds to wake back up (then it's fast again).
- If the service **restarts or redeploys**, the `workhours-data.json`
  file resets and previously entered hours can be lost.

**To avoid losing data:** periodically back up the data (see below).

## Data backup

At any time, you can copy the `workhours-data.json` file to keep all
your data backed up. On Render, you can view/download it via the
Shell tab in the service dashboard.

## New year / month — automatic

No extra setup is needed — for any year/month, just visit the URL with
`?y=2027&m=1` and the blank table for that month will show up
automatically.