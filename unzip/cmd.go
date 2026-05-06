package unzip

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"go-multipart-zip/cloudstorage"

	"cloud.google.com/go/storage"
	"github.com/Match-Made/zip"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go4.org/readerutil"
	"google.golang.org/api/option"
)

var Cmd = &cobra.Command{
	Use:   "unzip",
	Short: "Unzip multi-part zip files",
	Long:  "Unzip multi-part zip files. For example, if you have a zip file named 'archive.zip' that is split into multiple parts (e.g., 'archive.z01', 'archive.z02', ..., and 'archive.zip'), you can use this command to unzip the files.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		credentials := viper.GetString("credentials")
		bucket := viper.GetString("bucket")
		files := viper.GetStringSlice("files")
		password := viper.GetString("password")
		output := viper.GetString("output")

		client, err := storage.NewClient(ctx, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(credentials)))
		if err != nil {
			log.Fatalf("Error creating GCS client: %v", err)
		}

		singleReaders := make([]readerutil.SizeReaderAt, 0, len(files))
		for _, file := range files {
			singleReader, err := cloudstorage.NewGCSReaderAt(ctx, client, bucket, file)
			if err != nil {
				log.Fatalf("Error creating GCS reader: %v", err)
			}
			singleReaders = append(singleReaders, singleReader)
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

			rc, err := f.Open()
			if err != nil {
				log.Fatalf("Error opening file in zip: %v", err)
			}
			defer rc.Close()

			outputPath := filepath.Join(output, f.Name)
			outFile, err := os.Create(outputPath)
			if err != nil {
				log.Fatalf("Error creating output file: %v", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				log.Fatalf("Error writing to output file: %v", err)
			}
		}
	},
}

func init() {
	Cmd.Flags().StringP("credentials", "c", "", "Path to GCS credentials JSON file (if reading from GCS)")
	Cmd.Flags().StringP("bucket", "b", "", "GCS bucket name (if reading from GCS)")
	Cmd.Flags().StringSliceP("files", "f", []string{}, "List of zip file parts (e.g., 'archive.z01', 'archive.z02', ..., 'archive.zip')")
	Cmd.Flags().StringP("password", "p", "", "Password for encrypted zip files")
	Cmd.Flags().StringP("output", "o", ".", "Output directory for extracted files (default: current directory)")

	viper.BindPFlag("credentials", Cmd.Flags().Lookup("credentials"))
	viper.BindPFlag("bucket", Cmd.Flags().Lookup("bucket"))
	viper.BindPFlag("files", Cmd.Flags().Lookup("files"))
	viper.BindPFlag("password", Cmd.Flags().Lookup("password"))
	viper.BindPFlag("output", Cmd.Flags().Lookup("output"))
}
