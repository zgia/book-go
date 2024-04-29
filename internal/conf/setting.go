package conf

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "time/tzdata"

	"github.com/pkg/errors"
	"gopkg.in/ini.v1"

	embedConf "zgia.net/book/conf"
	log "zgia.net/book/internal/logger"
	"zgia.net/book/internal/util"
)

// IsProdMode returns true if the application is running in production mode.
func IsProdMode() bool {
	return strings.EqualFold(App.RunMode, "prod")
}

// PageSize makes sure page size is in allowed range.
func PageSize(size int) int {
	if size <= 0 {
		size = API.ResponseItems
	} else if size > API.MaxResponseItems {
		size = API.MaxResponseItems
	}
	return size
}

func init() {
	// Initialize the primary logger until logging service is up.
	log.NewConsole()
}

// File is the configuration object.
var File *ini.File

// Init initializes configuration from conf assets and given custom configuration file.
// If `customConf` is empty, it falls back to default location, i.e. "<WORK DIR>/custom".
// It is safe to call this function multiple times with desired `customConf`, but it is
// not concurrent safe.
//
// NOTE: The order of loading configuration sections matters as one may depend on another.
//
// ⚠️ WARNING: Do not print anything in this function other than warnings.
func Init(customConf string) error {
	var err error

	data, err := embedConf.Files.ReadFile("app.ini")
	if err != nil {
		return errors.Wrap(err, `read default "app.ini"`)
	}

	File, err = ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, data)
	if err != nil {
		return errors.Wrap(err, "parse 'conf/app.ini'")
	}
	File.NameMapper = ini.SnackCase

	if customConf == "" {
		customConf = filepath.Join(util.CustomDir(), "conf", "app.ini")
	} else {
		customConf, err = filepath.Abs(customConf)
		if err != nil {
			return errors.Wrap(err, "get absolute path")
		}

		os.Setenv("BOOK_WORK_DIR", strings.Replace(customConf, "/custom/conf/app.ini", "", 1))
	}

	CustomConf = customConf

	if util.IsFile(customConf) {
		if err = File.Append(customConf); err != nil {
			return errors.Wrapf(err, "append %q", customConf)
		}
	} else {
		log.Warnf("Custom config %q is not exist.\n\n", customConf)
	}

	if err = File.Section(ini.DefaultSection).MapTo(&App); err != nil {
		return errors.Wrap(err, "mapping default section")
	}

	// ***************************
	// ----- Server settings -----
	// ***************************

	if err = File.Section("server").MapTo(&Server); err != nil {
		return errors.Wrap(err, "mapping [server] section")
	}

	if !strings.HasSuffix(Server.ExternalURL, "/") {
		Server.ExternalURL += "/"
	}
	Server.URL, err = url.Parse(Server.ExternalURL)
	if err != nil {
		return errors.Wrapf(err, "parse '[server] EXTERNAL_URL' %q", err)
	}

	// Subpath should start with '/' and end without '/', i.e. '/{subpath}'.
	Server.Subpath = strings.TrimRight(Server.URL.Path, "/")
	Server.SubpathDepth = strings.Count(Server.Subpath, "/")

	// *****************************
	// ----- Database settings -----
	// *****************************

	if err = File.Section("database").MapTo(&Database); err != nil {
		return errors.Wrap(err, "mapping [database] section")
	}

	// *************************
	// ----- Time settings -----
	// *************************

	if err = File.Section("time").MapTo(&Time); err != nil {
		return errors.Wrap(err, "mapping [time] section")
	}

	// Asia/Shanghai
	loc, err := time.LoadLocation(Time.Zone)
	if err == nil {
		time.Local = loc
	} else {
		time.FixedZone("UTC", 8*3600)
	}

	Time.FormatLayout = map[string]string{
		"ANSIC":       time.ANSIC,
		"UnixDate":    time.UnixDate,
		"RubyDate":    time.RubyDate,
		"RFC822":      time.RFC822,
		"RFC822Z":     time.RFC822Z,
		"RFC850":      time.RFC850,
		"RFC1123":     time.RFC1123,
		"RFC1123Z":    time.RFC1123Z,
		"RFC3339":     time.RFC3339,
		"RFC3339Nano": time.RFC3339Nano,
		"Kitchen":     time.Kitchen,
		"Stamp":       time.Stamp,
		"StampMilli":  time.StampMilli,
		"StampMicro":  time.StampMicro,
		"StampNano":   time.StampNano,
	}[Time.Format]
	if Time.FormatLayout == "" {
		Time.FormatLayout = time.RFC3339
	}

	// *************************
	// ----- I18n settings -----
	// *************************

	I18n = new(i18nConf)
	if err = File.Section("i18n").MapTo(I18n); err != nil {
		return errors.Wrap(err, "mapping [i18n] section")
	}

	if err = File.Section("http").MapTo(&HTTP); err != nil {
		return errors.Wrap(err, "mapping [http] section")
	} else if err = File.Section("api").MapTo(&API); err != nil {
		return errors.Wrap(err, "mapping [api] section")
	}

	HasRobotsTxt = util.IsFile(filepath.Join(util.CustomDir(), "robots.txt"))
	return nil
}

var (
	logDir     string
	logDirOnce sync.Once
)

func LogDir() string {
	logDirOnce.Do(func() {
		LogDir := File.Section("log").Key("ROOT_PATH").MustString(filepath.Join(util.WorkDir(), "log"))

		logDir = util.EnsureAbs(LogDir)
	})

	return logDir
}

// MustInit panics if configuration initialization failed.
func MustInit(customConf string) {
	err := Init(customConf)
	if err != nil {
		panic(err)
	}
}
