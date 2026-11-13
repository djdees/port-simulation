// Package client - query subcommands for one-shot API requests.
package client

import (
	"context"
	"encoding/json"
	"io"
)

// QueryArrivals fetches arrival events and pretty-prints them.
func QueryArrivals(ctx context.Context, c *Client, w io.Writer) error {
	return queryCollection(ctx, c, w, "/api/v1/arrivals")
}

// QueryDepartures fetches departure events and pretty-prints them.
func QueryDepartures(ctx context.Context, c *Client, w io.Writer) error {
	return queryCollection(ctx, c, w, "/api/v1/departures")
}

// GetVessel fetches a single vessel by ID and pretty-prints it.
func GetVessel(ctx context.Context, c *Client, w io.Writer, id string) error {
	var v any
	if err := c.Get(ctx, "/api/v1/vessels/"+id, &v); err != nil {
		return err
	}
	return prettyPrint(w, v)
}

// FilterCargo fetches cargo with optional query params and pretty-prints results.
func FilterCargo(ctx context.Context, c *Client, w io.Writer, params string) error {
	path := "/api/v1/cargo"
	if params != "" {
		path += "?" + params
	}
	return queryCollection(ctx, c, w, path)
}

// ListVessels fetches the vessel list and pretty-prints it.
func ListVessels(ctx context.Context, c *Client, w io.Writer) error {
	return queryCollection(ctx, c, w, "/api/v1/vessels")
}

func queryCollection(ctx context.Context, c *Client, w io.Writer, path string) error {
	var resp any
	if err := c.Get(ctx, path, &resp); err != nil {
		return err
	}
	return prettyPrint(w, resp)
}

func prettyPrint(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
