package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type FileInfo struct {
	Name string
}

var mu sync.Mutex
var selectedVersion string

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header, and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error Retrieving the File", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header.Get("Content-Type"))

	// Create a file within the "uploads" directory using the original filename
	tempFile, err := os.Create(filepath.Join("uploads", handler.Filename))
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	// Copy the contents of the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File: %s\n", handler.Filename)
}

var lastServedVersion string

func findFileByVersion(requestedVersion string) string {
	files, err := filepath.Glob(filepath.Join("uploads", "esp32_"+requestedVersion+".bin"))
	if err != nil {
		return ""
	}

	if len(files) > 0 {
		return filepath.Base(files[0])
	}
	return ""
}

// func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("update begun.....")
// 	// filename := "esp32.bin"
// 	vars := mux.Vars(r)
// 	// filename := vars["filename"]

// 	version, ok := vars["version"]
// 	if !ok {
// 		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
// 		return
// 	}
// 	if lastServedVersion != "" && version == lastServedVersion {

// 		errorMessage := "Firmware is up to date. No update required."
// 		http.Error(w, errorMessage, http.StatusBadRequest)
// 		fmt.Fprintf(w, "%s", errorMessage)
// 		return

// 	}

// 	// filePath := filepath.Join("uploads", filename)
// 	filename := findFileByVersion(version)
// 	if filename == "" {
// 		http.Error(w, "Requested version not found", http.StatusNotFound)
// 		return
// 	}
// 	filePath := filepath.Join("uploads", filename)
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		http.Error(w, "Error opening file", http.StatusInternalServerError)
// 		return
// 	}
// 	defer file.Close()

// 	fi, err := file.Stat()
// 	if err != nil {
// 		http.Error(w, "Error getting file information", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
// 	w.Header().Set("Content-Type", "application/octet-stream")
// 	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

// 	const bufferSize = 8192
// 	buffer := make([]byte, bufferSize)

// 	mu.Lock()
// 	defer mu.Unlock()

// 	totalBytesRead := int64(0)

// 	for {
// 		n, err := file.Read(buffer)
// 		if err == io.EOF {
// 			break
// 		} else if err != nil {
// 			http.Error(w, "Error reading file", http.StatusInternalServerError)
// 			return
// 		}

// 		_, err = w.Write(buffer[:n])
// 		if err != nil {
// 			return
// 		}

// 		totalBytesRead += int64(n)

// 		// Calculate percentage and send it as a response header
// 		// percentage := (float64(totalBytesRead) / float64(fi.Size())) * 100.0
// 		// w.Header().Set("X-Percentage-Downloaded", fmt.Sprintf("%.2f", percentage))

// 		// Check if the response writer supports flushing
// 		if flusher, ok := w.(http.Flusher); ok {
// 			flusher.Flush()
// 		}
// 	}
// 	lastServedVersion = version
// 	// Optional: Log the completion percentage
// 	// Optional: Log the completion percentage using the log package
// 	// log.Printf("Download of %s completed: %.2f%%\n", filename, 100.0)
// }

func listFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Showing the contents...")
	files, err := getUploadedFiles("uploads")
	if err != nil {
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Encode the list of files to JSON and send it in the response
	json.NewEncoder(w).Encode(files)
}

func selectVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version, ok := vars["version"]
	if !ok {
		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
		return
	}

	// Store the selected version
	mu.Lock()
	defer mu.Unlock()
	selectedVersion = version

	// Print the selected version on the backend
	fmt.Printf("Selected version: %s\n", selectedVersion)

	// Respond with a success message
	fmt.Fprint(w, "Version selected successfully")
}

func getUploadedFiles(uploadDir string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, FileInfo{Name: info.Name()})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
func serveFileWithProgress(w http.ResponseWriter, r *http.Request) {
	fmt.Println("update begun.....")
	vars := mux.Vars(r)

	// Extract version from the URL
	requestedVersion, ok := vars["version"]
	if !ok {
		http.Error(w, "Version not provided in the URL", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if a version has been selected
	if selectedVersion == "" {
		http.Error(w, "No version selected", http.StatusBadRequest)
		return
	}

	// Compare the selected version with the version in the URL
	if selectedVersion != requestedVersion {
		// The selected version and the version in the URL are different
		// Find the file based on the requested version
		filename := findFileByVersion(selectedVersion)
		if filename == "" {
			http.Error(w, "Requested version not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join("uploads", filename)
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			http.Error(w, "Error getting file information", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))

		const bufferSize = 8192
		buffer := make([]byte, bufferSize)

		totalBytesRead := int64(0)

		for {
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			} else if err != nil {
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}

			_, err = w.Write(buffer[:n])
			if err != nil {
				return
			}

			totalBytesRead += int64(n)

			// Check if the response writer supports flushing
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	} else {
		// The selected version and the version in the URL are the same
		errorMessage := "Firmware is up to date. No update required."
		http.Error(w, errorMessage, http.StatusBadRequest)
		fmt.Fprintf(w, "%s", errorMessage)
		return
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/upload", uploadFile).Methods("POST")
	r.HandleFunc("/list", listFiles).Methods("GET")

	r.HandleFunc("/select/{version:[0-9.]+}", selectVersion).Methods("POST")

	// r.HandleFunc("/{filename}", serveFileWithProgress).Methods("GET")
	r.HandleFunc("/{filename}/{version:[0-9.]+}", serveFileWithProgress).Methods("GET")
	// Serve static files from the "static" directory
	r.Use(handlers.CORS(handlers.AllowedOrigins([]string{"http://localhost:3000"})))

	http.Handle("/", http.FileServer(http.Dir("static")))
	err := http.ListenAndServe(":9000", r)
	if err != nil {
		fmt.Println(err)
	}

}
