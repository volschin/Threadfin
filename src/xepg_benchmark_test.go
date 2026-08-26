package src

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"threadfin/src/internal/imgcache"
)

func setupXMLTVBenchmark(b *testing.B, channelCount, programCount int) {
	b.Helper()

	previousSystem := System
	previousSettings := Settings
	previousData := Data
	b.Cleanup(func() {
		System = previousSystem
		Settings = previousSettings
		Data = previousData
	})

	root := b.TempDir()
	dataDir := filepath.Join(root, "data")
	imagesDir := filepath.Join(root, "images")
	for _, dir := range []string{dataDir, imagesDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			b.Fatal(err)
		}
	}

	imageCache, err := imgcache.New(imagesDir+string(os.PathSeparator), "/images/", false)
	if err != nil {
		b.Fatal(err)
	}

	System = SystemStruct{}
	System.Name = "Threadfin"
	System.Version = "benchmark"
	System.Branch = "main"
	System.Folder.Data = dataDir + string(os.PathSeparator)
	System.Folder.ImagesCache = imagesDir + string(os.PathSeparator)
	System.File.XML = filepath.Join(root, "threadfin.xml")
	System.Compressed.GZxml = filepath.Join(root, "threadfin.xml.gz")

	Settings = SettingsStruct{}
	Settings.EnableNonAscii = true
	Settings.Filter = make(map[int64]interface{})

	Data = DataStruct{}
	Data.Cache.Images = imageCache
	Data.XMLTV.Files = []string{"benchmark-guide.xml"}
	Data.XEPG.Channels = make(map[string]interface{}, channelCount)

	for channel := range channelCount {
		sourceID := fmt.Sprintf("source-%03d", channel)
		channelID := fmt.Sprintf("threadfin-%03d", channel)
		Data.XEPG.Channels[channelID] = XEPGChannelStruct{
			Name:       fmt.Sprintf("Channel %03d", channel),
			TvgChno:    strconv.Itoa(channel + 1),
			TvgName:    fmt.Sprintf("Channel %03d", channel),
			XActive:    true,
			XChannelID: channelID,
			XMapping:   sourceID,
			XmltvFile:  "benchmark-guide.xml",
			XName:      fmt.Sprintf("Channel %03d", channel),
		}
	}

	guide := XMLTV{
		Generator: "Threadfin benchmark fixture",
		Source:    "deterministic",
		Program:   make([]*Program, 0, programCount),
	}
	base := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for program := range programCount {
		channel := program % channelCount
		start := base.Add(time.Duration(program) * 30 * time.Minute)
		stop := start.Add(30 * time.Minute)
		guide.Program = append(guide.Program, &Program{
			Channel: fmt.Sprintf("source-%03d", channel),
			Start:   start.Format("20060102150405 -0700"),
			Stop:    stop.Format("20060102150405 -0700"),
			Title: []*Title{{
				Lang:  "en",
				Value: fmt.Sprintf("Program %05d", program),
			}},
			Desc: []*Desc{{
				Lang:  "en",
				Value: fmt.Sprintf("Deterministic benchmark description %05d", program),
			}},
			Category: []*Category{{
				Lang:  "en",
				Value: fmt.Sprintf("Category %02d", program%12),
			}},
		})
	}

	encoded, err := xml.Marshal(guide)
	if err != nil {
		b.Fatal(err)
	}
	encoded = append([]byte(xml.Header), encoded...)
	if err := os.WriteFile(filepath.Join(dataDir, "benchmark-guide.xml"), encoded, 0o600); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkXMLTVGeneration(b *testing.B) {
	cases := []struct {
		name     string
		channels int
		programs int
	}{
		{name: "10Channels_100Programs", channels: 10, programs: 100},
		{name: "100Channels_10000Programs", channels: 100, programs: 10_000},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			setupXMLTVBenchmark(b, tc.channels, tc.programs)
			if err := createXMLTVFile(); err != nil {
				b.Fatalf("fixture validation: %v", err)
			}
			for _, output := range []string{System.File.XML, System.Compressed.GZxml} {
				if _, err := os.Stat(output); err != nil {
					b.Fatalf("output %s: %v", output, err)
				}
			}

			b.ReportAllocs()
			for b.Loop() {
				clear(Data.Cache.XMLTV)
				if err := createXMLTVFile(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
