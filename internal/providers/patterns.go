package providers

import (
	"regexp"
	"strings"
)

var patterns = map[string]struct {
	re *regexp.Regexp
}{
	"group": {
		re: regexp.MustCompile(`(?mi)(fgt|evo|yts|yts?\.mx|yts\.am|yifi|yify|MkvCage|NoMeRcY|STRiFE|SiGMA|LucidTV|CHD|sujaidr|SAPHiRE|LEGI0N|hd4u|rarbg|ViSiON|ETRG|JYK|iFT|anoXmous|MkvCage|Ganool|TGx|klaxxon|icebane|greenbud1969|flawl3ss|metcon|proper|ntb|cm8|tbs|sva|avs|mtb|ion10|sauron|phoenix|minx|mvgroup|amiable|sadece|gooz|lite|killers|tbs|PHOENiX|memento|done|ExKinoRay|acool|starz|convoy|playnow|RedBlade|ntg|cmrg|cm|2hd|fty|haggis|Joy|dimension|0tv|fxg|kat|artsubs|horizon|horizon-artsub|axxo|diamond|asteroids|rarbg|rargb|unit3d|afg|xlf|pulsar|bamboozle|ebp|trump|bulit|pahe|lol|tjhd|DeeJayAhmed|DeeJahAhmed|anoxmous|galaxy|aoc|flux|roen|silence|CiNEFiLE|wrd|rico|huzzah|RiSEHD|Subs-Team|iExTV|ROLLiT|CONDITION|CinemaniaHD|FraMeSToR|CtrlHD|ion|ION265|tepes|gossip|COLLECTiVE|nate_666)`),
	},
	"duration": {
		re: regexp.MustCompile(`(?mi)(\d:\d{2}:\d{2})`),
	},
	"quality": {
		re: regexp.MustCompile(`(?i)\b(((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB-?DL(?: DVDRip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|telesync|hvec))\b`),
	},
	"resolution": {
		re: regexp.MustCompile(`\b(([0-9]{3,4}p))\b`),
	},
	"year": {
		re: regexp.MustCompile(`\((\d{4})\)`),
	},
	"season": {
		re: regexp.MustCompile(`(?i)(s?([0-9]{1,2}))(?:[exof]|$)`),
	},
	"episode": {
		re: regexp.MustCompile(`(?i)([exof]([0-9]{1,2})(?:[^0-9]|$))`),
	},
}

func Parse(raw string, pattern string) []string {
	Set := make(map[string]bool)

	for _, match := range patterns[pattern].re.FindAllStringSubmatch(strings.ToLower(raw), -1) {
		if !Set[match[0]] {
			Set[match[1]] = true
		}
	}
	var items []string
	for k := range Set {
		items = append(items, k)
	}
	return items
}
