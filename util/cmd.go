package util

import (
	"io"
	"log"
	"os"

	"github.com/KirCute/zip"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var UnzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "Unzip multi-part zip files",
	Long:  "Unzip multi-part zip files. For example, if you have a zip file named 'archive.zip' that is split into multiple parts (e.g., 'archive.zip.001', 'archive.zip.002', etc.), you can use this command to unzip the files.",
	Run: func(cmd *cobra.Command, args []string) {
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			log.Fatalf("Error getting password flag: %v", err)
		}

		singleReaders := make([]zip.SizeReaderAt, 0, len(args))
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			stats, err := f.Stat()
			if err != nil {
				log.Fatalf("Could not stat file: %s, error: %v", arg, err)
			}

			singleReaders = append(singleReaders, io.NewSectionReader(f, 0, stats.Size()))
		}

		multiReader, err := zip.NewMultipartReader(singleReaders)
		if err != nil {
			log.Fatalf("Error creating multi-part reader: %v", err)
		}

		for _, f := range multiReader.File {
			if f.IsEncrypted() {
				f.SetPassword(password)
			}

			log.Printf("Extracting: %s (%s)", f.Name, humanize.Bytes(uint64(f.UncompressedSize64)))
		}
	},
}

func init() {
	UnzipCmd.Flags().StringP("password", "p", "", "Password for encrypted zip files")
}
