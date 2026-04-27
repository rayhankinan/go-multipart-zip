package util

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/mholt/archives"
	"github.com/spf13/cobra"
	"github.com/yeka/zip"
	"go4.org/readerutil"
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

		singleReaders := make([]readerutil.SizeReaderAt, 0, len(args))

		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			size, ok := readerutil.Size(f)
			if !ok {
				log.Fatalf("Could not determine size of file: %s", arg)
			}

			singleReaders = append(singleReaders, io.NewSectionReader(f, 0, size))
		}

		// TODO: We need to remove local file header signatures from the multi-reader, otherwise the zip format will be invalid.
		// This is because each part of the multi-part zip file may contain its own local file header signature, which can cause issues when trying to extract the files.
		// We need to ensure that we only have one local file header signature at the beginning of the multi-reader, and that any subsequent signatures are removed or ignored.

		multiReader := readerutil.NewMultiReaderAt(singleReaders...)
		sequentialReader := io.NewSectionReader(multiReader, 0, multiReader.Size())

		format := Zip{
			Password:         password,
			EncryptionMethod: zip.AES256Encryption,
		}

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

func init() {
	UnzipCmd.Flags().StringP("password", "p", "", "Password for encrypted zip files")
}
