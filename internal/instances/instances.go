package instances

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	Endpoint = "https://instances.social/api/1.0/instances/list"
)

func Get(ctx context.Context, cl *http.Client, endpoint string, token string, instances uint) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("User-Agent", "fedinetmap (+https://code.dny.dev/fedinetmap)")

	q := req.URL.Query()
	q.Add("count", fmt.Sprintf("%d", instances))
	// We include instances that instances.social considers down as historically they've had
	// issues correctly detecting this. It's better if we check this ourselves
	q.Add("include_down", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		buf := new(strings.Builder)
		io.Copy(buf, io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("non-200 status code: %d: %s", resp.StatusCode, buf.String())
	}

	var r result
	j := json.NewDecoder(resp.Body)
	err = j.Decode(&r)
	if err != nil {
		return nil, err
	}

	res := make([]string, 0, len(r.Instances))
	for _, ins := range r.Instances {
		u, err := url.Parse("https://" + ins.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skipping instance: %s, could not parse: %v\n", ins.Name, err)
			continue
		}
		res = append(res, u.Hostname())
	}

	return res, nil
}

type result struct {
	Instances []struct {
		Name string `json:"name"`
	} `json:"instances"`
}
