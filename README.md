# Working Hours Tracker (Go)

Ekta chhoto web app — Start Time ar End Time bosale automatic hours calculate
hoy (midnight cross korle o thik moto), protidin auto-save hoy, ar mash/bochor
er total dekha jay. Excel template-er logic-i eikhane Go-te implement kora.

## Kivabe kaj kore (architecture)

- **Language:** Pure Go (standard library only — no external package lagbe na,
  tai `go build` korle kono internet/proxy issue hoy na).
- **Storage:** Ekta simple **JSON file** (`workhours-data.json`) — protita
  diner date, start time, end time save thake. Database server lagbe na, tai
  hosting shohoj ar free.
- **Frontend:** Server-render kora HTML + ektu vanilla JavaScript (AJAX) —
  time input change korle sathe sathe server-e save hoy ar duration/total
  update hoye jay, page reload lage na.

## Local-e run kora (test korar jonno)

```bash
go build -o workhours .
./workhours
```

Tarpor browser-e `http://localhost:8080` khulle current month-er page ashbe.

Port change korte chaile: `PORT=3000 ./workhours`
Data file location change korte chaile: `DATA_FILE=/path/to/file.json ./workhours`

## Free-e Live Deploy Kora (Fly.io — recommended)

Fly.io-te ekta free allowance ache jeta diye **persistent storage shoho**
(mane app restart hole o data thakbe) ekta chhoto app free-e host kora jay.
Domain kinte hobe na — Fly nijer subdomain dey (`https://your-app.fly.dev`).

1. Fly CLI install koro: https://fly.io/docs/hands-on/install-flyctl/
2. Account banao ar login koro:
   ```bash
   fly auth signup
   ```
3. Ei project folder-e giye:
   ```bash
   fly launch
   ```
   - App-er নাম দিতে বলবে — jekono ekta name dao.
   - "Would you like to set up a Postgres database?" → **No**
   - "Would you like to deploy now?" → **No** (age volume banabo)
4. Persistent volume banao (data jate delete na hoy):
   ```bash
   fly volumes create data --size 1
   ```
5. `fly.toml` file-e ei mounts section add koro (na thakle):
   ```toml
   [mounts]
     source = "data"
     destination = "/app/data"
   ```
6. Deploy koro:
   ```bash
   fly deploy
   ```
7. Deploy hoye gele ekta link pabe (jemon `https://your-app.fly.dev`) — oi
   link-i tomar bondhu ke pathiye dao. Oi link e dhukei time bosate parbe.

## Alternative: Render.com (free, kintu ekta caution ache)

Render-e o "Web Service" free-e banano jay (GitHub repo connect kore, Docker
auto-detect hoye jabe), kintu Render-er **free tier-e local disk ephemeral**
— mane app kichukhon idle thakle ghumiye jay, ar abar uthle **data file reset
hoye jete pare**. Tai Render use korle regularly `workhours-data.json`
download kore rekhe dio backup hisebe, othoba Fly.io use koro (upore
deya steps) jekhane real persistent volume free-e paoa jay.

## Data backup

Jekono somoy `workhours-data.json` file-ta copy kore rekhe dile shob data
backup thakbe. Server-e SSH kore (`fly ssh console` diye Fly-te) ba download
kore dekhte paro.

## Notun bochor / month automatic

Kono setup lagbe na — je kono year/month-e URL-e `?y=2027&m=1` dile oi
month-er faka table automatic show hobe.
