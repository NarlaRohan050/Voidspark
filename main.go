package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Void Spark — Prompt → World engine (GTA Jam MVP)
// Worlds saved to disk, served via /worlds/, visualized via /web/preview

type World struct {
	ID        string   `json:"id"`
	Dimension string   `json:"dimension"`
	Theme     string   `json:"theme"`
	Aesthetic string   `json:"aesthetic"`
	Rooms     []Room   `json:"rooms"`
	Seed      int64    `json:"seed"`
	Current   int      `json:"current"`
	GameState string   `json:"game_state"`
	Party     []Hero   `json:"party"`
	Log       []string `json:"log"`
}

type Room struct {
	Index int    `json:"index"`
	Type  string `json:"type"` // combat/loot/trap/rest
	Desc  string `json:"desc"`
}

type Hero struct {
	Name  string         `json:"name"`
	Role  string         `json:"role"`
	HP    int            `json:"hp"`
	MaxHP int            `json:"max_hp"`
	Stats map[string]int `json:"stats"`
}

var (
	store   = map[string]*World{}
	storeMu sync.Mutex
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Ensure worlds/ folder exists
	if _, err := os.Stat("worlds"); os.IsNotExist(err) {
		if err := os.Mkdir("worlds", 0755); err != nil {
			log.Fatalf("failed to create worlds folder: %v", err)
		}
	}

	// Core API
	http.HandleFunc("/", uiHandler)
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/party", partyHandler)
	http.HandleFunc("/explore", exploreHandler)
	http.HandleFunc("/state", stateHandler)

	// World persistence & preview support
	http.Handle("/worlds/", http.StripPrefix("/worlds/", http.FileServer(http.Dir("worlds"))))
	http.HandleFunc("/api/latest-world", latestWorldHandler)

	// Static assets (preview HTML)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	port := "8080"
	log.Printf("Void Spark running at http://localhost:%s", port)
	log.Printf("Preview: http://localhost:%s/web/preview/world_preview.html", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func uiHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!doctype html>
<html>
<head><meta charset="utf-8"><title>Void Spark</title></head>
<body style="font-family:system-ui, Arial; margin:16px">
<h2>Void Spark — Prompt → World (MVP)</h2>
<label>Prompt (world):</label><br>
<textarea id="prompt" rows="3" cols="64">a dark stone dungeon with countless treasure, candlelight, and traps</textarea><br>
<button onclick="generate()">Generate World</button>
<button onclick="createParty()">Create Party</button>
<button onclick="explore()">Go Forward (Explore)</button>
<button onclick="showState()">Show State</button>
<p><a href="/web/preview/world_preview.html" target="_blank">🌀 Open Live Preview</a></p>
<pre id="out" style="white-space:pre-wrap;border:1px solid #ddd;padding:10px;margin-top:12px;height:420px;overflow:auto"></pre>
<script>
let sessionId = ''
async function generate(){
  const prompt = document.getElementById('prompt').value
  const res = await fetch('/generate',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({prompt})})
  const js = await res.json()
  sessionId = js.id
  document.getElementById('out').innerText = JSON.stringify(js, null, 2)
}
async function createParty(){
  if(!sessionId){alert('Generate a world first');return}
  const res = await fetch('/party',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:sessionId})})
  const js = await res.json()
  document.getElementById('out').innerText = JSON.stringify(js, null, 2)
}
async function explore(){
  if(!sessionId){alert('Generate a world first');return}
  const res = await fetch('/explore',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:sessionId})})
  const js = await res.json()
  document.getElementById('out').innerText = JSON.stringify(js, null, 2)
}
async function showState(){
  if(!sessionId){alert('Generate a world first');return}
  const res = await fetch('/state?id='+sessionId)
  const js = await res.json()
  document.getElementById('out').innerText = JSON.stringify(js, null, 2)
}
</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ Prompt string `json:"prompt"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	theme, aesthetic, dim := parsePrompt(req.Prompt)
	seed := time.Now().UnixNano()
	wld := buildWorld(theme, aesthetic, dim, seed)

	storeMu.Lock()
	store[wld.ID] = wld
	storeMu.Unlock()

	// Save to disk
	data, _ := json.MarshalIndent(wld, "", "  ")
	filename := fmt.Sprintf("worlds/world_%s.json", wld.ID)
	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Printf("failed to save world json: %v", err)
	}

	writeJSON(w, wld)
}

func latestWorldHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("worlds")
	if err != nil || len(files) == 0 {
		http.Error(w, "no worlds found", http.StatusNotFound)
		return
	}

	var latestFile os.DirEntry
	var latestTime time.Time

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = file
			}
		}
	}

	if latestFile == nil {
		http.Error(w, "no valid world files found", http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{"latest": latestFile.Name()})
}

func partyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ ID string `json:"id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	storeMu.Lock()
	wld, ok := store[req.ID]
	storeMu.Unlock()
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if len(wld.Party) == 0 {
		wld.Party = generateParty()
		wld.Log = append(wld.Log, "Party assembled: "+rolesList(wld.Party))
	}
	writeJSON(w, wld)
}

func exploreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ ID string `json:"id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	storeMu.Lock()
	wld, ok := store[req.ID]
	storeMu.Unlock()
	if !ok {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if wld.GameState == "game_over" || wld.GameState == "finished" {
		writeJSON(w, wld)
		return
	}
	if wld.Current >= len(wld.Rooms) {
		wld.Log = append(wld.Log, "You have reached the dungeon's end. Victory! 🎉")
		wld.GameState = "finished"
		writeJSON(w, wld)
		return
	}
	room := wld.Rooms[wld.Current]
	wld.Log = append(wld.Log, fmt.Sprintf("Entering room %d: %s (%s)", room.Index, room.Desc, room.Type))
	switch room.Type {
	case "combat":
		clog := resolveCombat(wld, wld.Seed+int64(wld.Current))
		wld.Log = append(wld.Log, clog...)
	case "loot":
		loot := randomLoot(wld.Seed + int64(wld.Current))
		wld.Log = append(wld.Log, "Found treasure: "+loot)
	case "trap":
		effect := triggerTrap(wld, wld.Seed+int64(wld.Current))
		wld.Log = append(wld.Log, effect)
	case "rest":
		heal := restParty(wld)
		wld.Log = append(wld.Log, fmt.Sprintf("Rested: healed %d HP total", heal))
	}
	wld.Current++
	alive := 0
	for _, h := range wld.Party {
		if h.HP > 0 {
			alive++
		}
	}
	if alive == 0 {
		wld.Log = append(wld.Log, "All party members have fallen. Game over.")
		wld.GameState = "game_over"
	}
	storeMu.Lock()
	store[req.ID] = wld
	storeMu.Unlock()
	writeJSON(w, wld)
}

func stateHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	storeMu.Lock()
	wld, ok := store[id]
	storeMu.Unlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, wld)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func parsePrompt(prompt string) (theme, aesthetic, dim string) {
	p := strings.ToLower(prompt)
	dim = "2D"
	if strings.Contains(p, "3d") || strings.Contains(p, "3-d") || strings.Contains(p, "3 d") {
		dim = "3D"
	}
	theme = "generic"
	if strings.Contains(p, "dungeon") || strings.Contains(p, "treasure") {
		theme = "dungeon"
	}
	if strings.Contains(p, "city") || strings.Contains(p, "race") || strings.Contains(p, "track") {
		theme = "city"
	}
	if strings.Contains(p, "space") || strings.Contains(p, "station") {
		theme = "space"
	}
	if strings.Contains(p, "cyber") || strings.Contains(p, "neon") {
		theme = "cyberpunk"
	}
	aesthetic = "dark"
	if strings.Contains(p, "glow") || strings.Contains(p, "neon") || strings.Contains(p, "glowing") || strings.Contains(p, "bright") {
		aesthetic = "glowing"
	}
	if strings.Contains(p, "moss") || strings.Contains(p, "overgrown") {
		aesthetic = "overgrown"
	}
	return
}

func buildWorld(theme, aesthetic, dim string, seed int64) *World {
	r := rand.New(rand.NewSource(seed))
	roomCount := r.Intn(5) + 8 // 8–12 rooms
	var rooms []Room

	// Flavor pools
	enemies := []string{"goblin", "skeleton", "slime", "bandit", "warg", "orc", "shadow knight", "rat swarm"}
	treasures := []string{"ancient chest", "enchanted urn", "jeweled altar", "crystal coffer", "forgotten relic"}
	traps := []string{"collapsing floor", "poison gas nozzle", "arrow trap", "flame burst tile", "swinging axe"}
	restSpots := []string{"quiet alcove", "hidden fountain", "warm torch-lit corner", "abandoned camp", "stone bench"}

	enemyActions := []string{
		"lurking in the shadows",
		"patrolling a mossy hallway",
		"guarding a cracked archway",
		"prowling near the entrance",
		"waiting beside a flickering torch",
		"snarling behind broken bars",
		"roaming the corridor aimlessly",
	}
	treasureFlavors := []string{
		"gleaming faintly in the dark",
		"sealed with ancient runes",
		"covered in glowing dust",
		"surrounded by gold coins",
		"hidden behind a loose wall stone",
	}
	trapFlavors := []string{
		"barely visible to the eye",
		"giving off a faint mechanical hum",
		"coated in strange residue",
		"cleverly disguised as safe ground",
		"making a faint clicking sound nearby",
	}
	restFlavors := []string{
		"filled with soft candlelight",
		"surrounded by silence",
		"echoing faint dripping sounds",
		"carved with ancient runes of peace",
		"still warm from a previous traveler",
	}

	for i := 1; i <= roomCount; i++ {
		var t, desc string
		switch r.Intn(4) {
		case 0: // loot
			t = "loot"
			treasure := treasures[r.Intn(len(treasures))]
			flavor := treasureFlavors[r.Intn(len(treasureFlavors))]
			desc = fmt.Sprintf("A %s, %s.", treasure, flavor)
		case 1: // combat
			t = "combat"
			enemy := enemies[r.Intn(len(enemies))]
			action := enemyActions[r.Intn(len(enemyActions))]
			desc = fmt.Sprintf("A %s %s.", enemy, action)
		case 2: // trap
			t = "trap"
			trap := traps[r.Intn(len(traps))]
			flavor := trapFlavors[r.Intn(len(trapFlavors))]
			desc = fmt.Sprintf("A %s, %s.", trap, flavor)
		case 3: // rest
			t = "rest"
			spot := restSpots[r.Intn(len(restSpots))]
			flavor := restFlavors[r.Intn(len(restFlavors))]
			desc = fmt.Sprintf("A %s, %s.", spot, flavor)
		}
		rooms = append(rooms, Room{Index: i, Type: t, Desc: desc})
	}

	return &World{
		ID:         strconv.FormatInt(seed, 10),
		Dimension:  dim,
		Theme:      theme,
		Aesthetic:  aesthetic,
		Rooms:      rooms,
		Seed:       seed,
		Current:    0,
		GameState:  "exploring",
		Log:        []string{fmt.Sprintf("Spawned world: %s (%s) seed=%d", theme, aesthetic, seed)},
	}
}

func generateParty() []Hero {
	now := time.Now().UnixNano()
	base := int(now % 1000)
	return []Hero{
		{Name: fmt.Sprintf("Tank-%d", base+1), Role: "tank", MaxHP: 120, HP: 120, Stats: map[string]int{"str": 8, "def": 8}},
		{Name: fmt.Sprintf("Attacker-%d", base+2), Role: "attacker", MaxHP: 90, HP: 90, Stats: map[string]int{"str": 10, "def": 4}},
		{Name: fmt.Sprintf("Healer-%d", base+3), Role: "healer", MaxHP: 80, HP: 80, Stats: map[string]int{"int": 9, "def": 3}},
		{Name: fmt.Sprintf("Support-%d", base+4), Role: "support", MaxHP: 85, HP: 85, Stats: map[string]int{"dex": 7, "def": 4}},
	}
}

func rolesList(hs []Hero) string {
	roles := []string{}
	for _, h := range hs {
		roles = append(roles, fmt.Sprintf("%s(%s)", h.Name, h.Role))
	}
	return strings.Join(roles, ", ")
}

func resolveCombat(w *World, seed int64) []string {
	logs := []string{}
	if len(w.Party) == 0 {
		logs = append(logs, "No party present — combat skipped")
		return logs
	}
	r := rand.New(rand.NewSource(seed))
	enemies := 1 + r.Intn(2)
	for e := 0; e < enemies; e++ {
		enemyHP := 30 + r.Intn(30)
		logs = append(logs, fmt.Sprintf("Enemy %d appears with %d HP", e+1, enemyHP))
		for enemyHP > 0 {
			for i := range w.Party {
				if w.Party[i].HP <= 0 {
					continue
				}
				damage := 5 + r.Intn(8)
				if w.Party[i].Role == "attacker" {
					damage += 4
				}
				enemyHP -= damage
				logs = append(logs, fmt.Sprintf("%s hits enemy for %d (enemy HP %d)", w.Party[i].Name, damage, max(enemyHP, 0)))
				if enemyHP <= 0 {
					logs = append(logs, "Enemy defeated")
					break
				}
			}
			if enemyHP <= 0 {
				break
			}
			alive := []int{}
			for i := range w.Party {
				if w.Party[i].HP > 0 {
					alive = append(alive, i)
				}
			}
			if len(alive) == 0 {
				logs = append(logs, "All heroes down")
				return logs
			}
			target := alive[r.Intn(len(alive))]
			hit := 6 + r.Intn(8)
			w.Party[target].HP -= hit
			if w.Party[target].HP < 0 {
				w.Party[target].HP = 0
			}
			logs = append(logs, fmt.Sprintf("Enemy hits %s for %d (HP %d)", w.Party[target].Name, hit, w.Party[target].HP))
		}
	}
	return logs
}

func randomLoot(seed int64) string {
	r := rand.New(rand.NewSource(seed))
	items := []string{"gold coins", "sapphire amulet", "rusty sword", "potion of healing", "weird trinket"}
	return items[r.Intn(len(items))]
}

func triggerTrap(w *World, seed int64) string {
	r := rand.New(rand.NewSource(seed))
	damage := 5 + r.Intn(16)
	alive := []int{}
	for i := range w.Party {
		if w.Party[i].HP > 0 {
			alive = append(alive, i)
		}
	}
	if len(alive) == 0 {
		return "Trap triggers, but no one is alive to be affected."
	}
	target := alive[r.Intn(len(alive))]
	w.Party[target].HP -= damage
	if w.Party[target].HP < 0 {
		w.Party[target].HP = 0
	}
	return fmt.Sprintf("Trap triggers: %s takes %d damage (HP %d)", w.Party[target].Name, damage, w.Party[target].HP)
}

func restParty(w *World) int {
	healed := 0
	for i := range w.Party {
		if w.Party[i].HP <= 0 {
			continue
		}
		amount := (w.Party[i].MaxHP - w.Party[i].HP) / 2
		if amount <= 0 {
			continue
		}
		w.Party[i].HP += amount
		healed += amount
	}
	return healed
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}