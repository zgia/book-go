package conf

import (
	"net/url"
)

// ℹ️ README: This file contains static values that should only be set at initialization time.

// CustomConf returns the absolute path of custom configuration file that is used.
var CustomConf string

// ⚠️ WARNING: After changing the following section, do not forget to update template of
// "/admin/config" page as well.
var (
	// Application settings
	App struct {
		// ⚠️ WARNING: Should only be set by the main package (i.e. "main.go").
		Version string `ini:"-"`

		BrandName string
		RunMode   string
	}

	// HTTP settings
	HTTP struct {
		AccessControlAllowOrigin string
		JwtSecretKey             string
	}

	// Time settings
	Time struct {
		Format string
		Zone   string

		// Derived from other static values
		FormatLayout string `ini:"-"` // Actual layout of the Format.
	}

	// API settings
	API struct {
		ResponseItems    int
		MaxResponseItems int
	}

	// Global setting
	HasRobotsTxt bool
)

type ServerOpts struct {
	ExternalURL string `ini:"EXTERNAL_URL"`
	Domain      string
	Protocol    string
	HTTPAddr    string `ini:"HTTP_ADDR"`
	HTTPPort    string `ini:"HTTP_PORT"`

	EnableGzip bool

	EnablePprof bool

	// Derived from other static values
	URL          *url.URL `ini:"-"` // Parsed URL object of ExternalURL.
	Subpath      string   `ini:"-"` // Subpath found the ExternalURL. Should be empty when not found.
	SubpathDepth int      `ini:"-"` // The number of slashes found in the Subpath.
}

// Server settings
var Server ServerOpts

type DatabaseOpts struct {
	Type         string
	Host         string
	Name         string
	User         string
	Password     string
	MaxOpenConns int
	MaxIdleConns int
}

// Database settings
var Database DatabaseOpts

type i18nConf struct {
	Langs     []string          `delim:","`
	Names     []string          `delim:","`
}

// I18n settings
var I18n *i18nConf

// Indicates which database backend is currently being used.
var (
	UseMySQL bool
	UseSQLite bool
)
