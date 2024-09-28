package providers

import (
	"regexp"
	"strings"
)

var patterns = map[string]struct {
	re *regexp.Regexp
}{
	"group": {
		re: regexp.MustCompile(`(?mi)fgt|evo|yts|yts?\.mx|yifi|yify|MkvCage|NoMeRcY|STRiFE|SiGMA|LucidTV|CHD|sujaidr|SAPHiRE|LEGI0N|hd4u|rarbg|ViSiON|ETRG|JYK|iFT|anoXmous|MkvCage|Ganool|TGx|klaxxon|icebane|greenbud1969|flawl3ss|metcon|proper|ntb|cm8|tbs|sva|avs|mtb|ion10|sauron|phoenix|minx|mvgroup|amiable|sadece|gooz|lite|killers|tbs|PHOENiX|memento|done|ExKinoRay|acool|starz|convoy|playnow|RedBlade|ntg|cmrg|cm|2hd|fty|haggis|Joy|dimension|0tv|fxg|kat|artsubs|horizon|axxo|diamond|asteroids|rarbg|unit3d|afg|xlf|pulsar|bamboozle|ebp|trump|bulit|pahe|lol|tjhd|DeeJayAhmed|DeeJahAhmed|HEVC|anoxmous|galaxy|aoc|flux|roen|silence|CiNEFiLE|wrd|rico|huzzah|RiSEHD|Subs-Team|iExTV|ROLLiT|CONDITION|CinemaniaHD|FraMeSToR|CtrlHD|ION265|tepes|gossip`),
	},
	"duration": {
		re: regexp.MustCompile(`(?mi)\d:\d{2}:\d{2}`),
	},
	"quality": {
		re: regexp.MustCompile(`(?i)\b(((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB-?DL(?: DVDRip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|telesync))\b`),
	},
	"resolution": {
		re: regexp.MustCompile(`\b(([0-9]{3,4}p))\b`),
	},
}

func Parse(raw string, pattern string) string {
	matches := patterns[pattern].re.FindAllStringSubmatch(strings.ToLower(raw), -1)
	if len(matches) > 0 {
		if len(matches[0]) > 0 {
			return strings.TrimSpace(matches[0][0])
		}
	}
	return ""
}
