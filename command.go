package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type punchLine struct {
	slug  string
	id    int
	stats int
}

type punchLines []punchLine

func (p punchLines) Len() int           { return len(p) }
func (p punchLines) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p punchLines) Less(i, j int) bool { return p[i].stats < p[j].stats }
func (p punchLines) idByCmd(cmd string) (int, bool) {
	for i, l := range p {
		if l.slug == cmd || strconv.Itoa(l.id) == cmd {
			return i, true
		}
	}
	return 0, false
}

var knownPunchLines = punchLines{
	{id: 0, slug: "ani-za-kokot-vole", stats: 0},
	{id: 1, slug: "do-pice", stats: 0},
	{id: 2, slug: "hajzli-jedni", stats: 0},
	{id: 3, slug: "hosi-to-je-neuveritelne", stats: 0},
	{id: 4, slug: "ja-se-z-toho-musim-pojebat", stats: 0},
	{id: 5, slug: "ja-to-mrdam", stats: 0},
	{id: 6, slug: "jedinou-picovinku", stats: 0},
	{id: 7, slug: "jedu-do-pici-stadyma", stats: 0},
	{id: 8, slug: "kurva", stats: 0},
	{id: 9, slug: "kurva-do-pice-to-neni-mozne", stats: 0},
	{id: 10, slug: "nebudu-to-delat", stats: 0},
	{id: 11, slug: "past-vedle-pasti-pico", stats: 0},
	{id: 12, slug: "to-je-pico-nemozne", stats: 0},
	{id: 13, slug: "to-neni-normalni-kurva", stats: 0},
	{id: 14, slug: "to-sou-nervy-ty-pico", stats: 0},
	{id: 15, slug: "abych-mohl-toto", stats: 0},
	{id: 16, slug: "ani-ocko-nenasadis", stats: 0},
	{id: 17, slug: "banalni-vec", stats: 0},
	{id: 18, slug: "ja-to-nejdu-delat", stats: 0},
	{id: 19, slug: "kurva-uz", stats: 0},
	{id: 20, slug: "ne-nenasadis-ho", stats: 0},
	{id: 21, slug: "nejvetsi-blbec-na-zemekouli", stats: 0},
	{id: 22, slug: "nenasadim", stats: 0},
	{id: 23, slug: "neresitelny-problem-hosi", stats: 0},
	{id: 24, slug: "nevim-jak", stats: 0},
	{id: 25, slug: "okamzite-zabit-ty-kurvy", stats: 0},
	{id: 26, slug: "pockej-kamo", stats: 0},
	{id: 27, slug: "tady-musis-vsechno-rozdelat", stats: 0},
	{id: 28, slug: "tuto-picu-potrebuju-utahnout", stats: 0},
	{id: 29, slug: "zasrane-zamrdane", stats: 0},
}

func NewCommand(triggerWord string) *command {
	return &command{
		triggerWord: triggerWord,
		punchLines:  knownPunchLines,
		start:       time.Now(),
	}
}

type command struct {
	triggerWord string
	punchLines  punchLines
	start       time.Time
}

func (c *command) Describe(_ chan<- *prometheus.Desc) {
}

func (c *command) Collect(metrics chan<- prometheus.Metric) {
	for _, l := range c.punchLines {
		metric, err := prometheus.NewConstMetric(
			prometheus.NewDesc("milujipraci_stats", "How many times has the punch line has been called.", []string{"punch_line"}, map[string]string{}),
			prometheus.CounterValue,
			float64(l.stats),
			l.slug,
		)
		if err != nil {
			log.Errorf("failed to collect stats metric for punch line %s: %v", l.slug, err)
			continue
		}
		metrics <- metric
	}
}

func (c *command) slash(w http.ResponseWriter, r *http.Request) {
	command := strings.TrimSpace(r.PostFormValue("text"))
	switch command {
	case "":
		c.help(w)
	case "help":
		c.help(w)
	case "stats":
		c.stats(w)
	default:
		c.punchLine(command, w)
	}
}

func (c *command) invalidInput(input string, w http.ResponseWriter) {
	writeJsonResponse(w, model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         fmt.Sprintf("Jakoby toho nebylo už tak dost, `%s` neznám!\nZkus se podívat do nápovědy `/%s help`", input, c.triggerWord),
	})
}

func (c *command) punchLine(command string, w http.ResponseWriter) {
	plId, ok := c.punchLines.idByCmd(command)
	if !ok {
		c.invalidInput(command, w)
		return
	}
	audioUrl := fmt.Sprintf("http://milujipraci.cz/#%02d", c.punchLines[plId].id)
	ImageUrl := fmt.Sprintf("https://fusakla.github.io/slash-milujipraci/assets/punch-lines/%s.png", c.punchLines[plId].slug)
	c.punchLines[plId].stats++
	writeJsonResponse(w, model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text:         fmt.Sprintf(`[![%s](%s)](%s)`, command, ImageUrl, audioUrl),
		GotoLocation: audioUrl,
	})
}

func (c *command) help(w http.ResponseWriter) {
	var help strings.Builder
	help.WriteString(fmt.Sprintf("Vyber si něco od srdce a zavolej  `/%s ani-za-kokot-vole` nebo pokud už se nezmůžeš ani na, stačí jen číslo hlášky `/%s 0`\nNa výběr máš:\n", c.triggerWord, c.triggerWord))
	for _, p := range c.punchLines {
		help.WriteString(fmt.Sprintf("  - %d %s\n", p.id, p.slug))
	}
	help.WriteString(fmt.Sprintf("\nVybírej moudře!\n\nPokud tě zajímají statistiky volání jednotlivých hlášek, zavolej `/%s stats`", c.triggerWord))
	writeJsonResponse(w, model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         help.String(),
	})
}

func (c *command) stats(w http.ResponseWriter) {
	var help strings.Builder
	help.WriteString("Statistiky volání jednotlivých hlášek:\n")
	sortedPunchLines := make(punchLines, len(c.punchLines))
	for i, l := range c.punchLines {
		sortedPunchLines[i] = punchLine{
			slug:  l.slug,
			id:    l.id,
			stats: l.stats,
		}
	}
	sort.Sort(sortedPunchLines)
	for _, l := range sortedPunchLines {
		help.WriteString(fmt.Sprintf("  - %s: %d\n", l.slug, l.stats))
	}
	writeJsonResponse(w, model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         help.String(),
	})
}
