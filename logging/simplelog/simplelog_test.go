// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/preview/logging"
	"golang.org/x/net/context"
	"google.golang.org/api/option"

	"github.com/GoogleCloudPlatform/golang-samples/internal/testutil"
)

func TestSimplelog(t *testing.T) {
	tc := testutil.SystemTest(t)
	ctx := context.Background()

	client, err := logging.NewClient(ctx, tc.ProjectID, option.WithScopes(logging.AdminScope))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	defer func() {
		if err := deleteLog(client); err != nil {
			t.Errorf("deleteLog: %v", err)
		}
	}()

	client.OnError = func(err error) {
		t.Errorf("OnError: %v", err)
	}

	writeEntry(client)
	structuredWrite(client)

	testutil.Flaky(t, 10, time.Second, func(r *testutil.R) {
		entries, err := getEntries(client, tc.ProjectID)
		if err != nil {
			r.Failf("getEntries: %v", err)
			return
		}

		if got, want := len(entries), 2; got != want {
			r.Failf("len(entries) = %d; want %d", got, want)
			return
		}

		wantContain := map[string]*logging.Entry{
			"Anything":                            entries[0],
			"The payload can be any type!":        entries[0],
			"infolog is a standard Go log.Logger": entries[1],
		}

		for want, entry := range wantContain {
			msg := fmt.Sprintf("%s", entry.Payload)
			if !strings.Contains(msg, want) {
				r.Failf("want %q to contain %q", msg, want)
			}
		}
	})
}
