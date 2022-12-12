package azurefileshare

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-file-go/azfile"
	"github.com/spf13/cobra"
)

var wg sync.WaitGroup

func accountInfo() (string, string) {
	return os.Getenv("ACCOUNT_NAME"), os.Getenv("ACCOUNT_KEY")
}

var backup = &cobra.Command{
	Use: "backup",
	Run: func(cmd *cobra.Command, args []string) {
		accName, accKey := accountInfo()

		credential, err := azfile.NewSharedKeyCredential(accName, accKey)
		if err != nil {
			log.Fatal(err)
		}

		share, err := cmd.Flags().GetString("share")
		if err != nil {
			log.Fatal(err)
		}

		f, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatal(err)
		}

		d, err := cmd.Flags().GetString("directory")
		if err != nil {
			log.Fatal(err)
		}

		o, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Fatal(err)
		}

		_, err = os.Stat(o)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.MkdirAll(o, 0700)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}

		path := ""
		if len(f) > 0 {
			path = f
		} else if len(d) > 0 {
			path = d
		}

		u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net/%s/%s", accName, share, path))

		if len(f) > 0 {
			downloadFile(o, credential, *u)
		} else if len(d) > 0 {
			downloadDir(o, credential, *u)
		}

		wg.Wait()
	},
}

func downloadFile(output string, c azfile.Credential, u url.URL) {
	fileURL := azfile.NewFileURL(u, azfile.NewPipeline(c, azfile.PipelineOptions{}))

	download, err := fileURL.Download(context.Background(), 0, 0, false)
	if err != nil {
		log.Fatal(err)
	}

	outPath := filepath.Join(output, filepath.Base(u.String()))
	file, err := os.Create(outPath) // Create the file to hold the downloaded file contents.
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	contentLength := download.ContentLength()
	retryReader := download.Body(azfile.RetryReaderOptions{MaxRetryRequests: 3})
	progressReader := pipeline.NewResponseBodyProgress(retryReader,
		func(bytesTransferred int64) {
			fmt.Printf("File %s: downloaded %d of %d bytes.\n", outPath, bytesTransferred, contentLength)
		})

	defer progressReader.Close()

	_, err = io.Copy(file, progressReader) // Write to the file by reading from the file (with intelligent retries).
	if err != nil {
		log.Fatal(err)
	}
}

func downloadDir(output string, c azfile.Credential, u url.URL) {
	dirURL := azfile.NewDirectoryURL(u, azfile.NewPipeline(c, azfile.PipelineOptions{}))

	resp, err := dirURL.ListFilesAndDirectoriesSegment(context.Background(), azfile.Marker{}, azfile.ListFilesAndDirectoriesOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range resp.DirectoryItems {
		newUrl := u
		newUrl.Path = filepath.Join(u.Path, d.Name)
		if err != nil {
			log.Fatal(err)
		}
		downloadDir(output, c, newUrl)
	}

	for _, f := range resp.FileItems {
		newUrl := u
		newUrl.Path = filepath.Join(u.Path, f.Name)

		if err != nil {
			log.Fatal(err)
		}
		wg.Add(1)

		go func() {
			defer wg.Done()
			downloadFile(output, c, newUrl)
		}()
	}

}

func init() {
	backup.Flags().StringP("share", "s", "", "Name of the Azure File Share")
	backup.Flags().StringP("file", "f", "", "File to download")
	backup.Flags().StringP("directory", "d", "", "Directory to download")
	backup.Flags().StringP("output", "o", "output/", "Output path")
	backup.Flags().StringP("key-id", "k", "", "GPG Key ID to use to encrypt the downloaded files")

	rootCmd.AddCommand(backup)
}
