package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tuimaps "github.com/branden-thompson/go-tuimaps"
)

// defaultSource is where the app fetches tiles from unless it is told to
// stay offline. It is OpenFreeMap, whose planet file the tiles built into
// this program were cut from, so the map looks the same either way.
const defaultSource = "https://tiles.openfreemap.org/planet/20260913_164504_pt/"

// cacheCap is how much of the disk the app's tile cache may use.
const cacheCap = 256 << 20

// languageMax is the longest label-language code the app takes, as the
// library takes it: a code such as en or pt-BR, never a sentence.
const languageMax = 8

// settings is everything the command line said.
type settings struct {
	headless bool
	describe bool
	purge    bool
	verify   bool

	offline   bool
	noCache   bool
	cacheRoot string

	safeRamps    bool
	reduceMotion bool
	noColour     bool

	language  string
	stylePath string
	cols      int
	rows      int
	scenario  int
	places    []tuimaps.Place

	help bool
}

// m1Scenarios are the M1 scenarios this release can draw. Scenario 5 is
// wind, which arrives with the wind overlay (D-44).
func m1Scenarios() []int { return []int{1, 2, 3, 4, 6, 7} }

// switchable is one accessibility setting, with the two ways to reach it:
// NFR-15 asks that no state be reachable by one of them only.
type switchable struct {
	what string
	flag string
	key  string
	env  string // an environment variable, where that is the way in
}

// accessibility is every switch the help must list (PL-AX-5).
func accessibility() []switchable {
	return []switchable{
		{what: "safe ramps", flag: "--safe-ramps", key: "s"},
		{what: "reduce motion", flag: "--reduce-motion", key: "r"},
		{what: "no colour", flag: "--no-colour", key: "C"},
		{what: "describe", flag: "--describe", key: "d"},
		{what: "NO_COLOR", env: "NO_COLOR"},
	}
}

// parse reads the command line. It reaches nothing and opens nothing: a
// mistake is answered before the terminal is touched.
func parse(args []string) (settings, error) {
	s := settings{language: "en", cacheRoot: defaultCacheRoot()}
	fs := flag.NewFlagSet("tuimaps", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var size, places, language string
	var noColor, noColour, help bool
	scenario := ""
	fs.BoolVar(&s.headless, "headless", false, "")
	fs.BoolVar(&s.describe, "describe", false, "")
	fs.BoolVar(&s.purge, "purge", false, "")
	fs.BoolVar(&s.verify, "verify", false, "")
	fs.BoolVar(&s.offline, "offline", false, "")
	fs.BoolVar(&s.noCache, "no-cache", false, "")
	fs.BoolVar(&s.safeRamps, "safe-ramps", false, "")
	fs.BoolVar(&s.reduceMotion, "reduce-motion", false, "")
	fs.BoolVar(&noColour, "no-colour", false, "")
	fs.BoolVar(&noColor, "no-color", false, "")
	fs.BoolVar(&help, "help", false, "")
	fs.BoolVar(&help, "h", false, "")
	fs.StringVar(&size, "size", "", "")
	fs.StringVar(&language, "lang", "en", "")
	fs.StringVar(&s.stylePath, "style", "", "")
	fs.StringVar(&scenario, "scenario", "", "")
	named := &namedPlaces{}
	fs.Var(named, "place", "")
	_ = places

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			s.help = true
			return s, nil
		}
		// The flag package writes what was on the command line into its own
		// message, so the message is made printable before it goes anywhere
		// near a terminal.
		return settings{}, fmt.Errorf("%s; run tuimaps --help to see what the app takes", printable(err.Error(), wholeComplaint))
	}
	if rest := fs.Args(); len(rest) > 0 {
		return settings{}, fmt.Errorf("tuimaps takes no arguments of its own, and %s is one; run tuimaps --help", quoted(rest[0]))
	}
	s.help, s.noColour = help, noColour || noColor
	if s.noCache {
		s.cacheRoot = ""
	}
	if err := s.readPlaces(named.raw); err != nil {
		return settings{}, err
	}
	if err := s.readSize(size); err != nil {
		return settings{}, err
	}
	if err := s.readLanguage(language); err != nil {
		return settings{}, err
	}
	if err := s.readScenario(scenario); err != nil {
		return settings{}, err
	}
	return s, nil
}

// readSize takes a size written as COLSxROWS, which is how M1's two sizes
// are named.
func (s *settings) readSize(text string) error {
	if text == "" {
		return nil
	}
	cols, rows, found := strings.Cut(strings.ToLower(text), "x")
	if !found {
		return fmt.Errorf("--size %s is not a size; write it as columns by rows, such as --size 149x38", quoted(text))
	}
	across, err := strconv.Atoi(cols)
	if err != nil {
		return fmt.Errorf("--size %s: %s is not a number of columns; write it as --size 149x38", quoted(text), quoted(cols))
	}
	down, err := strconv.Atoi(rows)
	if err != nil {
		return fmt.Errorf("--size %s: %s is not a number of rows; write it as --size 149x38", quoted(text), quoted(rows))
	}
	if across < 1 || down < 1 {
		return fmt.Errorf("--size %s: a map is at least 1 column by 1 row", quoted(text))
	}
	s.cols, s.rows = across, down
	return nil
}

// readLanguage takes the label language, which is a code and not a name.
func (s *settings) readLanguage(code string) error {
	if code == "" {
		return errors.New("--lang was given nothing; it takes a language code, such as en, de or pt-BR")
	}
	if len(code) > languageMax || strings.TrimFunc(code, codeRune) != "" {
		return fmt.Errorf("--lang %s is not a language code; pass a short code such as en, de or pt-BR", quoted(code))
	}
	s.language = code
	return nil
}

// codeRune reports whether a character may stand in a language code.
func codeRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-'
}

// readScenario takes the number of an M1 scenario to load.
func (s *settings) readScenario(text string) error {
	if text == "" {
		return nil
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		return fmt.Errorf("--scenario %s is not a number; the M1 scenarios are numbered 1 to 7", quoted(text))
	}
	if n < 1 || n > 7 {
		return fmt.Errorf("--scenario %d: the M1 scenarios are numbered 1 to 7", n)
	}
	if n == 5 {
		return errors.New("--scenario 5 is the wind scenario, and this release does not draw wind; the scenarios it can draw are 1, 2, 3, 4, 6 and 7")
	}
	s.scenario = n
	return nil
}

// namedPlaces collects the --place flag, which may be written more than
// once. It keeps what was written and checks none of it: a complaint about
// a place is the app's own wording, and the flag package would bury it
// inside a sentence of its own.
type namedPlaces struct {
	raw []string
}

// String is what the flag package prints for a default; the app prints its
// own help, so it is empty.
func (n *namedPlaces) String() string { return "" }

// Set keeps one place as it was written.
func (n *namedPlaces) Set(text string) error {
	if n == nil {
		return errors.New("--place was given nowhere to put a place")
	}
	n.raw = append(n.raw, text)
	return nil
}

// readPlaces turns what was written into places: NAME@LON,LAT, or LON,LAT
// alone, which is then named by its own position - a description that says
// "the place" and not which one is no answer.
func (s *settings) readPlaces(written []string) error {
	for _, text := range written {
		if strings.TrimSpace(text) == "" {
			return errors.New("--place was given nothing; write it as --place Home@-84.51,33.82, or as --place lon,lat")
		}
		name, position, found := strings.Cut(text, "@")
		if !found {
			name, position = "", text
		}
		lon, lat, err := readPosition(position, text)
		if err != nil {
			return err
		}
		if strings.TrimSpace(name) == "" {
			name = strings.TrimSpace(position)
		}
		s.places = append(s.places, tuimaps.Place{Name: strings.TrimSpace(name), At: tuimaps.LonLat{Lon: lon, Lat: lat}})
	}
	return nil
}

// readPosition reads a longitude and a latitude written with a comma
// between them, in that order: the order they are written in a file of
// places, and the order the library takes them in.
func readPosition(position, whole string) (float64, float64, error) {
	east, north, found := strings.Cut(position, ",")
	if !found {
		return 0, 0, fmt.Errorf("--place %s has no position in it; write it as --place Home@lon,lat", quoted(whole))
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(east), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("--place %s: %s is not a longitude", quoted(whole), quoted(east))
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(north), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("--place %s: %s is not a latitude", quoted(whole), quoted(north))
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("--place %s: a longitude is between -180 and 180", quoted(whole))
	}
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("--place %s: a latitude is between -90 and 90", quoted(whole))
	}
	return lon, lat, nil
}

// defaultCacheRoot is where the app keeps tiles between runs. The cache is
// the app's, not the library's: the library keeps nothing on disk unless it
// is given a directory (FR-21b).
func defaultCacheRoot() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "tuimaps")
}
