package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

//go:embed templates/*.html static/*.css
var embeddedFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"pad2": func(n int) string { return fmt.Sprintf("%02d", n) },
}).ParseFS(embeddedFS, "templates/*.html"))

// ---------- storage ----------

// DayEntry holds one day's raw start/end time strings in 24h "HH:MM" format.
type DayEntry struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	data map[string]DayEntry // key "YYYY-MM-DD"
}

func NewStore(path string) *Store {
	s := &Store{path: path, data: map[string]DayEntry{}}
	s.load()
	return s
}

func (s *Store) load() {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return // file doesn't exist yet — start empty
	}
	_ = json.Unmarshal(b, &s.data)
}

func (s *Store) save() error {
	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Get(key string) DayEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *Store) Set(key string, e DayEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.Start == "" && e.End == "" {
		delete(s.data, key)
	} else {
		s.data[key] = e
	}
	return s.save()
}

func (s *Store) MonthEntries(year int, month int) map[int]DayEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[int]DayEntry{}
	prefix := fmt.Sprintf("%04d-%02d-", year, month)
	for k, v := range s.data {
		if len(k) == 10 && k[:8] == prefix {
			day, err := strconv.Atoi(k[8:10])
			if err == nil {
				out[day] = v
			}
		}
	}
	return out
}

// ---------- duration logic ----------

// duration computes worked hours given 24h "HH:MM" start/end strings.
// If end is earlier than or equal to start, it's treated as crossing
// midnight into the next day (matches the Excel template's rule).
func duration(start, end string) (time.Duration, bool) {
	if start == "" || end == "" {
		return 0, false
	}
	layout := "15:04"
	st, err1 := time.Parse(layout, start)
	en, err2 := time.Parse(layout, end)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	if !en.After(st) {
		en = en.Add(24 * time.Hour)
	}
	return en.Sub(st), true
}

func fmtHM(d time.Duration) string {
	total := int(d.Minutes())
	h := total / 60
	m := total % 60
	return fmt.Sprintf("%dh %02dm", h, m)
}

// ---------- view models ----------

type DayRow struct {
	Day     int
	DateKey string
	Label   string
	Start   string
	End     string
	HasData bool
	DurStr  string
}

type MonthPage struct {
	Year        int
	Month       int
	MonthName   string
	Rows        []DayRow
	TotalStr    string
	DaysFilled  int
	PrevY, PrevM int
	NextY, NextM int
}

var monthNames = []string{"", "January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func buildMonthPage(store *Store, year, month int) MonthPage {
	entries := store.MonthEntries(year, month)
	nDays := daysInMonth(year, month)
	rows := make([]DayRow, 0, nDays)
	var total time.Duration
	filled := 0
	for d := 1; d <= nDays; d++ {
		key := fmt.Sprintf("%04d-%02d-%02d", year, month, d)
		e := entries[d]
		row := DayRow{
			Day:     d,
			DateKey: key,
			Label:   fmt.Sprintf("%02d %s", d, monthNames[month][:3]),
			Start:   e.Start,
			End:     e.End,
		}
		if dur, ok := duration(e.Start, e.End); ok {
			row.HasData = true
			row.DurStr = fmtHM(dur)
			total += dur
			filled++
		}
		rows = append(rows, row)
	}
	pm, py := month-1, year
	if pm == 0 {
		pm, py = 12, year-1
	}
	nm, ny := month+1, year
	if nm == 13 {
		nm, ny = 1, year+1
	}
	return MonthPage{
		Year: year, Month: month, MonthName: monthNames[month],
		Rows: rows, TotalStr: fmtHM(total), DaysFilled: filled,
		PrevY: py, PrevM: pm, NextY: ny, NextM: nm,
	}
}

type MonthSummary struct {
	Month    int
	Name     string
	TotalStr string
	HasData  bool
}

type YearPage struct {
	Year       int
	Months     []MonthSummary
	GrandTotal string
	PrevY, NextY int
}

func buildYearPage(store *Store, year int) YearPage {
	var grand time.Duration
	months := make([]MonthSummary, 0, 12)
	for m := 1; m <= 12; m++ {
		entries := store.MonthEntries(year, m)
		var mt time.Duration
		has := false
		for _, e := range entries {
			if dur, ok := duration(e.Start, e.End); ok {
				mt += dur
				has = true
			}
		}
		grand += mt
		months = append(months, MonthSummary{Month: m, Name: monthNames[m], TotalStr: fmtHM(mt), HasData: has})
	}
	return YearPage{Year: year, Months: months, GrandTotal: fmtHM(grand), PrevY: year - 1, NextY: year + 1}
}

// ---------- handlers ----------

func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "workhours-data.json"
	}
	store := NewStore(dataFile)

	mux := http.NewServeMux()

	staticFS, _ := embeddedFS.ReadFile("static/style.css")
	mux.HandleFunc("/static/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.Write(staticFS)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		http.Redirect(w, r, fmt.Sprintf("/month?y=%d&m=%d", now.Year(), int(now.Month())), http.StatusFound)
	})

	mux.HandleFunc("/month", func(w http.ResponseWriter, r *http.Request) {
		year, month := parseYM(r)
		page := buildMonthPage(store, year, month)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "month.html", page); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	mux.HandleFunc("/summary", func(w http.ResponseWriter, r *http.Request) {
		year := time.Now().Year()
		if y := r.URL.Query().Get("y"); y != "" {
			if v, err := strconv.Atoi(y); err == nil {
				year = v
			}
		}
		page := buildYearPage(store, year)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "summary.html", page); err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	// AJAX endpoint: save one day's start/end, return computed duration + new month total instantly.
	mux.HandleFunc("/api/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var body struct {
			Date  string `json:"date"` // YYYY-MM-DD
			Start string `json:"start"`
			End   string `json:"end"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		if len(body.Date) != 10 {
			http.Error(w, "bad date", 400)
			return
		}
		if err := store.Set(body.Date, DayEntry{Start: body.Start, End: body.End}); err != nil {
			http.Error(w, "save failed", 500)
			return
		}
		year, _ := strconv.Atoi(body.Date[0:4])
		month, _ := strconv.Atoi(body.Date[5:7])

		resp := struct {
			DurStr     string `json:"durStr"`
			HasData    bool   `json:"hasData"`
			MonthTotal string `json:"monthTotal"`
			DaysFilled int    `json:"daysFilled"`
		}{}
		if dur, ok := duration(body.Start, body.End); ok {
			resp.DurStr = fmtHM(dur)
			resp.HasData = true
		}
		mp := buildMonthPage(store, year, month)
		resp.MonthTotal = mp.TotalStr
		resp.DaysFilled = mp.DaysFilled

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Work Hours app listening on %s (data file: %s)", addr, dataFile)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func parseYM(r *http.Request) (int, int) {
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if y := r.URL.Query().Get("y"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := r.URL.Query().Get("m"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
			month = v
		}
	}
	return year, month
}
