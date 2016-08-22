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
	ProxyCachePath         string
	ProxyCacheSize         uint64
	ProxyCacheKeysZoneName string
	ProxyCacheKeysZoneSize uint64
	Tempfile               string
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

	cmd = exec.Command("sh", "-c", fmt.Sprintf("find %s -type f | wc -l", n.ProxyCachePath))
	out, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	keysZoneUsage, err := strconv.ParseUint(strings.TrimRight(string(out), "\n"), 0, 64)
	if err != nil {
		return nil, err
	}

	stat := make(map[string]interface{})
	stat["size"] = n.ProxyCacheSize
	stat["usage"] = usage
	stat["zone_size"] = n.ProxyCacheKeysZoneSize

	// nginx can store about 8,000 keys per 1MB
	// refs: http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_path
	stat["zone_usage"] = keysZoneUsage / 8000

	return stat, nil
}

func (n NginxCachePlugin) GraphDefinition() map[string](mp.Graphs) {
	dk := fmt.Sprintf("nginx-cache.disk-%s", n.ProxyCacheKeysZoneName)
	kk := fmt.Sprintf("nginx-cache.keys-%s", n.ProxyCacheKeysZoneName)

	var graphdef map[string](mp.Graphs) = map[string](mp.Graphs){
		dk: mp.Graphs{
			Label: fmt.Sprintf("nginx cache usage megabyte: %s", n.ProxyCachePath),
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "usage", Label: "Usage", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "size", Label: "Size", Diff: false, Type: "uint64"},
			},
		},
		kk: mp.Graphs{
			Label: fmt.Sprintf("nginx cache keys zone usage megabyte: %s", n.ProxyCachePath),
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				mp.Metrics{Name: "zone_usage", Label: "Keys Zone Usage", Diff: false, Type: "uint64"},
				mp.Metrics{Name: "zone_size", Label: "Keys Zone Size", Diff: false, Type: "uint64"},
			},
		},
	}

	return graphdef
}

func main() {
	proxyCachePath := flag.String("path", "", "proxy_cache_path $path")
	proxyCacheSize := flag.String("size", "", "proxy_cache_path $max_size")
	proxyCacheKeysZoneName := flag.String("kname", "", "proxy_cache_path $keys_zone_name")
	proxyCacheKeysZoneSize := flag.String("ksize", "", "proxy_cache_path $keys_zone_size")
	tempfile := flag.String("tempfile", "", "temporary file path")
	flag.Parse()

	var (
		nginx NginxCachePlugin
		err   error
	)

	nginx.ProxyCachePath = *proxyCachePath
	nginx.ProxyCacheKeysZoneName = *proxyCacheKeysZoneName

	if usageUnitPat.MatchString(*proxyCacheSize) {
		proxyCacheSizeStr := *proxyCacheSize
		*proxyCacheSize = proxyCacheSizeStr[:len(proxyCacheSizeStr)-1]
	}
	nginx.ProxyCacheSize, err = strconv.ParseUint(*proxyCacheSize, 0, 64)
	if err != nil {
		os.Exit(1)
	}

	if usageUnitPat.MatchString(*proxyCacheKeysZoneSize) {
		proxyCacheKeysZoneSizeStr := *proxyCacheKeysZoneSize
		*proxyCacheKeysZoneSize = proxyCacheKeysZoneSizeStr[:len(proxyCacheKeysZoneSizeStr)-1]
	}
	nginx.ProxyCacheKeysZoneSize, err = strconv.ParseUint(*proxyCacheKeysZoneSize, 0, 64)
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
