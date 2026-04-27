package util

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
	"go4.org/readerutil"
)

var UnzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "Unzip multi-part zip files",
	Long:  "Unzip multi-part zip files. For example, if you have a zip file named 'archive.zip' that is split into multiple parts (e.g., 'archive.zip.001', 'archive.zip.002', etc.), you can use this command to unzip the files.",
	Run: func(cmd *cobra.Command, args []string) {
		singleReaders := make([]readerutil.SizeReaderAt, 0, len(args))

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

		multiReader := readerutil.NewMultiReaderAt(singleReaders...)
		sequentialReader := io.NewSectionReader(multiReader, 0, multiReader.Size())

		format := archives.Zip{}

		if err := format.Extract(
			cmd.Context(),
			sequentialReader,
			func(ctx context.Context, info archives.FileInfo) error {
				log.Printf("Extracting: %s, size: %s", info.Name(), humanize.Bytes(uint64(info.Size())))

				return nil
			},
		); err != nil {
			log.Fatalf("Error extracting zip file: %v", err)
		}
	},
}
