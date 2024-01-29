package utils

import (
	"fmt"
	"math"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
)

func PrintStatistics(in []byte, compressedSize, uncompressedSize int, list *ll.LinkedList[segments.Segment], t1 time.Time) {
	var segmentCount int
	var minGain, maxGain int = math.MaxInt, 0

	typeCount := map[segments.SegmentType]int{}
	typeGain := map[segments.SegmentType]int{}

	list.ForEach(func(cur *ll.ListEntry[segments.Segment]) {
		segmentCount++
		gain := cur.Value.GetCompressionGains()
		t := cur.Value.GetType()
		typeCount[t] += len(cur.Value.GetPos())
		typeGain[t] += gain
		if gain > maxGain {
			maxGain = gain
		} else if gain < minGain {
			minGain = gain
		}
	})

	ratio := float64(compressedSize) / float64(len(in))
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Statistics", "Value"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Statistics"},
		{Name: "Value"},
	})
	t.AppendRows([]table.Row{
		{"Ratio", ratio},
		{"Compressed Size", compressedSize},
		{"Original Size", len(in)},
		// {"Segments Count", segmentCount},
		// {"Min Gain", minGain},
		// {"Max Gain", maxGain},
		{"Duration", time.Since(t1).String()},
	})
	println(t.Render())

	if list.Len == 0 {
		return
	}

	t = table.NewWriter()
	t.SetTitle("Count")
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Type"},
		{Name: "Count"},
	})
	for segType, count := range typeCount {
		t.AppendRow(table.Row{segments.TypeName[segType], count})
	}
	println(t.Render())

	t = table.NewWriter()
	t.AppendHeader(table.Row{"Type", "Gain ( bytes )", "Percentage"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Type"},
		{Name: "Gain"},
		{Name: "%"},
	})
	for segType, gain := range typeGain {
		t.AppendRow(table.Row{segments.TypeName[segType], gain, fmt.Sprintf("%d%%", gain*100/len(in))})
	}
	println(t.Render())
}
