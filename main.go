package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

type NginxCachePlugin struct {
	ProxyCachePath string
	ProxyCacheSize uint64
	Tempfile       string
}

var graphdef map[string](mp.Graphs) = map[string](mp.Graphs){
	"nginx-cache.usage": mp.Graphs{
		Label: "nginx cache usage byte",
		Unit:  "integer",
		Metrics: [](mp.Metrics){
			mp.Metrics{Name: "usage", Label: "Usage", Diff: false, Type: "uint64"},
			mp.Metrics{Name: "size", Label: "Size", Diff: false, Type: "uint64"},
		},
	},
}

var (
	duResultPat  *regexp.Regexp
	usageUnitPat *regexp.Regexp
)

func init() {
	duResultPat = regexp.MustCompile("^(\\d+)")
	usageUnitPat = regexp.MustCompile("m$")
}

func buildTempfilePath(path string) string {
	return fmt.Sprintf("/tmp/mackerel-plugin-nginx-cache-%s", strings.Replace(path, "/", "-", -1))
}

func (n NginxCachePlugin) FetchMetrics() (map[string]interface{}, error) {
	cmd := exec.Command("du", "-sm", n.ProxyCachePath)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	matches := duResultPat.FindStringSubmatch(string(out))

	if len(matches) < 2 {
		return nil, fmt.Errorf("\"%s\" is not matched the supposed result pattern", out)
	}

	usage, err := strconv.ParseUint(matches[1], 0, 64)
	if err != nil {
		return nil, err
	}

	stat := make(map[string]interface{})
	stat["size"] = n.ProxyCacheSize
	stat["usage"] = usage

	return stat, nil
}

func (n NginxCachePlugin) GraphDefinition() map[string](mp.Graphs) {
	return graphdef
}

func main() {
	proxyCachePath := flag.String("path", "", "proxy_cache_path < path >")
	proxyCacheSize := flag.String("size", "", "proxy_cache_path < max_size >")
	tempfile := flag.String("tempfile", "", "temporary file path")
	flag.Parse()

	var (
		nginx NginxCachePlugin
		err   error
	)

	nginx.ProxyCachePath = *proxyCachePath

	if usageUnitPat.MatchString(*proxyCacheSize) {
		proxyCacheSizeStr := *proxyCacheSize
		*proxyCacheSize = proxyCacheSizeStr[:len(proxyCacheSizeStr)-1]
	}
	nginx.ProxyCacheSize, err = strconv.ParseUint(*proxyCacheSize, 0, 64)
	if err != nil {
		os.Exit(1)
	}

	helper := mp.NewMackerelPlugin(nginx)

	if *tempfile != "" {
		helper.Tempfile = *tempfile
	} else {
		helper.Tempfile = buildTempfilePath(*proxyCachePath)
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
		helper.OutputDefinitions()
	} else {
		helper.OutputValues()
	}
}
