// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vg_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/gonum/plot"
	"github.com/gonum/plot/internal/cmpimg"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
)

var generateTestData = flag.Bool("regen", false, "Uses the current state to regenerate the test data.")

// TestLineWidth tests output against test images generated by
// running tests with -tag good.
func TestLineWidth(t *testing.T) {
	formats := []string{
		// TODO: Add logic to cope with run to run eps differences.
		"pdf",
		"svg",
		"png",
		"tiff",
		"jpg",
	}

	const (
		width  = 100
		height = 100
	)

	for _, w := range []vg.Length{-1, 0, 1} {
		for _, typ := range formats {
			p, err := lines(w)
			if err != nil {
				log.Fatalf("failed to create plot for %v:%s: %v", w, typ, err)
			}

			c, err := p.WriterTo(width, height, typ)
			if err != nil {
				t.Fatalf("failed to render plot for %v:%s: %v", w, typ, err)
			}

			var buf bytes.Buffer
			if _, err = c.WriteTo(&buf); err != nil {
				t.Fatalf("failed to write plot for %v:%s: %v", w, typ, err)
			}

			name := filepath.Join(".", "testdata", fmt.Sprintf("width_%v.%s", w, typ))

			// Recreate Golden images.
			if *generateTestData {
				err = p.Save(width, height, name)
				if err != nil {
					log.Fatalf("failed to save %q: %v", name, err)
				}
			}

			want, err := ioutil.ReadFile(name)
			if err != nil {
				t.Fatalf("failed to read test image [%s]: %v\n", name, err)
			}

			ok, err := cmpimg.Equal(typ, buf.Bytes(), want)
			if err != nil {
				t.Fatalf("failed to run cmpimg test [%s]: %v\n", name, err)
			}

			if !ok {
				t.Errorf("image mismatch for %v:%s", w, typ)
			}
		}
	}
}

func lines(w vg.Length) (*plot.Plot, error) {
	p, err := plot.New()
	if err != nil {
		return nil, err
	}

	pts := plotter.XYs{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	line, err := plotter.NewLine(pts)
	line.Width = w
	if err != nil {
		return nil, err
	}
	p.Add(line)

	return p, nil
}

func TestParseLength(t *testing.T) {
	for _, table := range []struct {
		str  string
		want vg.Length
		err  error
	}{
		{
			str:  "42.2cm",
			want: 42.2 * vg.Centimeter,
		},
		{
			str:  "42.2mm",
			want: 42.2 * vg.Millimeter,
		},
		{
			str:  "42.2in",
			want: 42.2 * vg.Inch,
		},
		{
			str:  "42.2pt",
			want: 42.2,
		},
		{
			str:  "42.2",
			want: 42.2,
		},
		{
			str: "999bottles",
			err: fmt.Errorf(`strconv.ParseFloat: parsing "999bottles": invalid syntax`),
		},
		{
			str:  "42inch",
			want: 42 * vg.Inch,
			err:  fmt.Errorf(`strconv.ParseFloat: parsing "42inch": invalid syntax`),
		},
	} {
		v, err := vg.ParseLength(table.str)
		if table.err != nil {
			if err == nil {
				t.Errorf("%s: expected an error (%v)\n",
					table.str, table.err,
				)
			}
			if table.err.Error() != err.Error() {
				t.Errorf("%s: got error=%q. want=%q\n",
					table.str, err.Error(), table.err.Error(),
				)
			}
			continue
		}
		if err != nil {
			t.Errorf("error setting flag.Value %q: %v\n",
				table.str,
				err,
			)
		}
		if v != table.want {
			t.Errorf("%s: incorrect value. got %v, want %v\n",
				table.str,
				float64(v), float64(table.want),
			)
		}
	}
}
