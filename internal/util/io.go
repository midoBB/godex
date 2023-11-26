package util

import (
	"archive/zip"
	"fmt"
	"godex/internal/mangadex"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// CreateDownloadDir creates a directory for the download path.
// It logs a success message if the directory is created successfully.
// If the directory already exists, it returns nil.
// If there's an error and it's not because the directory already exists, it returns the error.
func CreateDownloadDir(downloadPath string) error {
	err := os.Mkdir(downloadPath, 0755)
	if err == nil {
		log.Printf("Successfully created download path for manga\n")
	}
	if err != nil && os.IsExist(err) {
		return nil
	}
	return err
}

// CreateMangaDir creates a directory for the manga.
// It returns the path to the directory and nil if the directory is created successfully.
// If the directory already exists, it returns the path to the directory and nil.
// If there's an error, it returns nil and the error.
func CreateMangaDir(downloadPath string, manga *mangadex.GodexManga) (string, error) {
	folderPath := filepath.Join(downloadPath, manga.Manga.GetTitle())
	err := os.Mkdir(folderPath, 0755)
	if err != nil {
		if os.IsExist(err) {
			return folderPath, nil
		}
		return "", fmt.Errorf("error when creating manga directory: %v", err)
	}
	return folderPath, nil
}

// CreateCBZ creates a CBZ file from the chapter directory.
// It sorts the files in the directory, creates a zip file, and copies the files into the zip file.
// After the files are copied, it deletes the chapter directory and returns nil.
// If there's an error, it returns the error.
func CreateCBZ(chapterDir string) error {
	files, err := os.ReadDir(chapterDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	zipFile, err := os.Create(chapterDir + ".cbz")
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		fileToZip, err := os.Open(filepath.Join(chapterDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fileToZip.Close()

		info, err := fileToZip.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip file header: %w", err)
		}

		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip writer header: %w", err)
		}

		_, err = io.Copy(writer, fileToZip)
		if err != nil {
			return fmt.Errorf("failed to copy file to zip: %w", err)
		}
	}

	// Delete the chapter directory
	err = os.RemoveAll(chapterDir)
	if err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// CreateChapterDir creates a directory for the chapter.
// It returns the path to the directory and nil if the directory is created successfully.
// If there's an error, it returns an empty string and the error.
func CreateChapterDir(mangaDir string, chapter *mangadex.Chapter) (string, error) {
	folderPath := filepath.Join(mangaDir, chapter.GetChapterNum())
	err := os.Mkdir(folderPath, 0755)
	if err != nil {
		return "", fmt.Errorf("error when creating chapter directory: %v", err)
	}
	return folderPath, nil
}

func CheckChapterAlreadyExists(mangaDir string, chapterNumber string) bool {
	chapterFolderPath := filepath.Join(mangaDir, chapterNumber)
	return CheckFileExists(chapterFolderPath + ".cbz")
}

func CheckFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CreateFileAndDir(filename string) error {
	// Create the parent directory if it doesn't exist
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func MangaCoverExists(mangaDir string) bool {
	return CheckFileExists(filepath.Join(mangaDir, "cover.jpg"))
}
