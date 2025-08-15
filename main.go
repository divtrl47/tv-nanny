package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Program struct {
	Code  string
	Emoji string
	Color string
}

type ScheduleItem struct {
	Code     string
	Start    string
	startMin int
	Emoji    string
	Color    string
}

type Config struct {
	Programs []Program
	Schedule []ScheduleItem
}

func parseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	cfg := &Config{}
	section := ""
	var currProg *Program
	var currItem *ScheduleItem
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "programs:" {
			section = "programs"
			currProg = nil
			currItem = nil
			continue
		}
		if line == "schedule:" {
			section = "schedule"
			currProg = nil
			currItem = nil
			continue
		}
		if strings.HasPrefix(line, "-") {
			content := strings.TrimSpace(line[1:])
			if section == "programs" {
				p := Program{}
				cfg.Programs = append(cfg.Programs, p)
				currProg = &cfg.Programs[len(cfg.Programs)-1]
				currItem = nil
				if content != "" {
					parts := strings.SplitN(content, ":", 2)
					if len(parts) == 2 && strings.TrimSpace(parts[0]) == "code" {
						currProg.Code = strings.Trim(strings.TrimSpace(parts[1]), "\"")
					}
				}
			} else if section == "schedule" {
				s := ScheduleItem{}
				cfg.Schedule = append(cfg.Schedule, s)
				currItem = &cfg.Schedule[len(cfg.Schedule)-1]
				currProg = nil
				if content != "" {
					parts := strings.SplitN(content, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
						if key == "code" {
							currItem.Code = val
						} else if key == "start" {
							currItem.Start = val
						}
					}
				}
			}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		if section == "programs" && currProg != nil {
			switch key {
			case "code":
				currProg.Code = val
			case "emoji":
				currProg.Emoji = val
			case "color":
				currProg.Color = val
			}
		} else if section == "schedule" && currItem != nil {
			switch key {
			case "code":
				currItem.Code = val
			case "start":
				currItem.Start = val
			}
		}
	}
	progMap := map[string]Program{}
	for _, p := range cfg.Programs {
		progMap[p.Code] = p
	}
	for i := range cfg.Schedule {
		s := &cfg.Schedule[i]
		if p, ok := progMap[s.Code]; ok {
			s.Color = p.Color
			s.Emoji = p.Emoji
		}
		var h, m int
		fmt.Sscanf(s.Start, "%d:%d", &h, &m)
		s.startMin = h*60 + m
	}
	sort.Slice(cfg.Schedule, func(i, j int) bool { return cfg.Schedule[i].startMin < cfg.Schedule[j].startMin })
	return cfg, nil
}

func parseHexColor(s string) color.RGBA {
	var c color.RGBA
	c.A = 0xff
	if len(s) == 7 && s[0] == '#' {
		var r, g, b uint8
		fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b)
		c.R, c.G, c.B = r, g, b
	}
	return c
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 0xff,
	}
}

func computeColors(cfg *Config) [1440]color.RGBA {
	var colors [1440]color.RGBA
	n := len(cfg.Schedule)
	for i, s := range cfg.Schedule {
		start := s.startMin
		end := cfg.Schedule[(i+1)%n].startMin
		if end <= start {
			end += 1440
		}
		col := parseHexColor(s.Color)
		for m := start; m < end; m++ {
			colors[m%1440] = col
		}
	}
	for i, s := range cfg.Schedule {
		curr := parseHexColor(s.Color)
		next := parseHexColor(cfg.Schedule[(i+1)%n].Color)
		boundary := cfg.Schedule[(i+1)%n].startMin
		for offset := -5; offset <= 5; offset++ {
			m := boundary + offset
			for m < 0 {
				m += 1440
			}
			mm := m % 1440
			t := float64(offset+5) / 10.0
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
			colors[mm] = mix(curr, next, t)
		}
	}
	return colors
}

func drawThickLine(img *image.RGBA, x1, y1, x2, y2, width int, col color.Color) {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	steps := int(math.Max(math.Abs(dx), math.Abs(dy)))
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := x1 + int(dx*t)
		y := y1 + int(dy*t)
		for ox := -width; ox <= width; ox++ {
			for oy := -width; oy <= width; oy++ {
				if ox*ox+oy*oy <= width*width {
					img.Set(x+ox, y+oy, col)
				}
			}
		}
	}
}

func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, col color.Color) {
	minX := min(x1, min(x2, x3))
	maxX := max(x1, max(x2, x3))
	minY := min(y1, min(y2, y3))
	maxY := max(y1, max(y2, y3))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if pointInTriangle(x, y, x1, y1, x2, y2, x3, y3) {
				img.Set(x, y, col)
			}
		}
	}
}

func pointInTriangle(px, py, x1, y1, x2, y2, x3, y3 int) bool {
	d1 := sign(px, py, x1, y1, x2, y2)
	d2 := sign(px, py, x2, y2, x3, y3)
	d3 := sign(px, py, x3, y3, x1, y1)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func sign(px, py, x1, y1, x2, y2 int) int {
	return (px-x2)*(y1-y2) - (x1-x2)*(py-y2)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generateImage(cfg *Config) ([]byte, error) {
	const size = 1000
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := float64(size) / 2
	radius := center * 0.9

	colors := computeColors(cfg)
	bg := color.RGBA{0x11, 0x11, 0x11, 0xff}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > radius {
				img.Set(x, y, bg)
				continue
			}
			angle := math.Atan2(dy, dx) + math.Pi/2
			if angle < 0 {
				angle += 2 * math.Pi
			}
			m := int(angle/(2*math.Pi)*1440) % 1440
			img.Set(x, y, colors[m])
		}
	}

	now := time.Now().UTC().Add(3 * time.Hour)
	minf := float64(now.Hour()*60+now.Minute()) / 1440
	angle := 2*math.Pi*minf - math.Pi/2
	shaft := radius * 0.6
	head := radius * 0.25
	width := int(radius * 0.02)
	x1 := int(center)
	y1 := int(center)
	x2 := x1 + int(math.Cos(angle)*shaft)
	y2 := y1 + int(math.Sin(angle)*shaft)
	drawThickLine(img, x1, y1, x2, y2, width, color.Black)
	tipX := x1 + int(math.Cos(angle)*(shaft+head))
	tipY := y1 + int(math.Sin(angle)*(shaft+head))
	leftX := x2 + int(math.Cos(angle+math.Pi/2)*head*0.6)
	leftY := y2 + int(math.Sin(angle+math.Pi/2)*head*0.6)
	rightX := x2 + int(math.Cos(angle-math.Pi/2)*head*0.6)
	rightY := y2 + int(math.Sin(angle-math.Pi/2)*head*0.6)
	fillTriangle(img, tipX, tipY, leftX, leftY, rightX, rightY, color.Black)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var cfg *Config

type iconPos struct {
	Emoji string
	X     float64
	Y     float64
}

func computeIconPositions(cfg *Config, size int) []iconPos {
	center := float64(size) / 2
	// push icons near the outer rim of the clock face
	radius := center * 0.9 * 0.9
	n := len(cfg.Schedule)
	out := make([]iconPos, 0, n)
	for i, s := range cfg.Schedule {
		start := s.startMin
		end := cfg.Schedule[(i+1)%n].startMin
		if end <= start {
			end += 1440
		}
		mid := (start + end) / 2
		a := 2*math.Pi*float64(mid)/1440 - math.Pi/2
		x := center + math.Cos(a)*radius
		y := center + math.Sin(a)*radius
		out = append(out, iconPos{s.Emoji, x / float64(size) * 100, y / float64(size) * 100})
	}
	return out
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method, r.RequestURI, time.Now().Format(time.DateTime))
	const size = 1000
	img, err := generateImage(cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	encoded := base64.StdEncoding.EncodeToString(img)
	icons := computeIconPositions(cfg, size)
	var sb strings.Builder
	for _, ic := range icons {
		fmt.Fprintf(&sb, `<span style="position:absolute;left:%.2f%%;top:%.2f%%;font-size:4.5vmin;transform:translate(-50%%,-50%%)">%s</span>`, ic.X, ic.Y, ic.Emoji)
	}
	html := fmt.Sprintf(`<div id="clock" style="position:relative;width:90vmin;height:90vmin;overflow:visible"><img src="data:image/png;base64,%s" style="width:100%%;height:100%%;display:block"/>%s</div>`, encoded, sb.String())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func main() {
	var err error
	cfg, err = parseConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/image", imageHandler)
	fs := http.FileServer(http.Dir("webos-app"))
	http.Handle("/", fs)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
